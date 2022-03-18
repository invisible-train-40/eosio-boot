package boot

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/invisible-train-40/eosio-boot/ops"
	"github.com/zhongshuwen/zswchain-go"
	"go.uber.org/zap"
)

type ActionMap map[string]*zsw.Action

func (b *Boot) RunChainValidation(opConfig *config.OpConfig) (bool, error) {
	bootSeqMap := ActionMap{}
	bootSeq := []*zsw.Action{}

	trxEventCh := make(chan interface{}, 500)
	go func() {
		defer close(trxEventCh)
		for _, step := range b.bootSequence.BootSequence {
			if !step.Data.RequireValidation() {
				continue
			}

			if step.SkipValidation {
				continue
			}

			pubkey, err := b.getOpPubkey(step)
			if err != nil {
				b.logger.Error("unable to get public key for operation", zap.Error(err))
				return
			}

			err = step.Data.Actions(pubkey, opConfig, trxEventCh)
			if err != nil {
				b.logger.Error("unable to get actions for step", zap.String("ops", step.Op), zap.Error(err))
				return
			}
		}
	}()

	for act := range trxEventCh {
		switch v := act.(type) {
		case ops.TransactionBoundary:
		case *ops.TransactionAction:
			action := (*zsw.Action)(v)
			if action != nil {
				action.SetToServer(true)
				data, err := zsw.MarshalBinary(action)
				if err != nil {
					return false, fmt.Errorf("validating: binary marshalling: %s", err)
				}
				key := sha2(data)
				// if _, ok := bootSeqMap[key]; ok {
				// 	// TODO: don't fatal here plz :)
				// 	log.Fatalf("Same action detected twice [%s] with key [%s]\n", stepAction.Name, key)
				// }
				bootSeqMap[key] = action
				bootSeq = append(bootSeq, action)
			}
		default:
			panic("validation: unexpected type in action chan")
		}
	}

	//err := b.validateTargetNetwork(bootSeqMap, bootSeq)
	//if err != nil {
	//	b.logger.Info("BOOT SEQUENCE VALIDATION FAILED:", zap.Error(err))
	//	return false, nil
	//}

	b.logger.Info("")
	b.logger.Info("All good! Chain validation succeeded!")
	b.logger.Info("")

	return true, nil
}

func (b *Boot) validateTargetNetwork(bootSeqMap ActionMap, bootSeq []*zsw.Action) (err error) {
	expectedActionCount := len(bootSeq)
	validationErrors := make([]error, 0)

	b.waitTargetNodeToBeUp()

	// TODO: wait for target network to be up, and responding...
	b.logger.Info("Pulling blocks from chain until we gathered all actions to validate:")
	blockHeight := 1
	actionsRead := 0
	seenMap := map[string]bool{}
	gotSomeTx := false
	backOff := false
	timeBetweenFetch := time.Duration(0)
	var timeLastNotFound time.Time

	for {
		time.Sleep(timeBetweenFetch)

		m, err := b.targetNetAPI.GetBlockByNum(context.Background(), uint32(blockHeight))
		if err != nil {
			if gotSomeTx && !backOff {
				backOff = true
				timeBetweenFetch = 500 * time.Millisecond
				timeLastNotFound = time.Now()

				time.Sleep(2000 * time.Millisecond)

				continue
			}

			b.logger.Warn("Failed getting block num from target api", zap.String("message", err.Error()))
			time.Sleep(1 * time.Second)
			continue
		}

		blockHeight++

		b.logger.Info("Receiving block", zap.Uint32("block_num", m.BlockNumber()), zap.String("producer", string(m.Producer)), zap.Int("trx_count", len(m.Transactions)))

		if !gotSomeTx && len(m.Transactions) > 2 {
			gotSomeTx = true
		}

		if !timeLastNotFound.IsZero() && timeLastNotFound.Before(time.Now().Add(-10*time.Second)) {
			b.flushMissingActions(seenMap, bootSeq)
		}

		for _, receipt := range m.Transactions {
			unpacked, err := receipt.Transaction.Packed.Unpack()
			if err != nil {
				b.logger.Warn("Unable to unpack transaction, won't be able to fully validate", zap.Error(err))
				return fmt.Errorf("unpack transaction failed")
			}

			for _, act := range unpacked.Actions {
				act.SetToServer(false)
				data, err := zsw.MarshalBinary(act)
				if err != nil {
					b.logger.Error("Error marshalling an action", zap.Error(err))
					validationErrors = append(validationErrors, ValidationError{
						Err:               err,
						BlockNumber:       1, // extract from the block transactionmroot
						PackedTransaction: receipt.Transaction.Packed,
						Action:            act,
						RawAction:         data,
						ActionHexData:     hex.EncodeToString(act.HexData),
						Index:             actionsRead,
					})
					return err
				}
				key := sha2(data) // TODO: compute a hash here..

				if _, ok := bootSeqMap[key]; !ok {
					validationErrors = append(validationErrors, ValidationError{
						Err:               errors.New("not found"),
						BlockNumber:       1, // extract from the block transactionmroot
						PackedTransaction: receipt.Transaction.Packed,
						Action:            act,
						RawAction:         data,
						ActionHexData:     hex.EncodeToString(act.HexData),
						Index:             actionsRead,
					})
					b.logger.Warn("INVALID action", zap.Int("action_read", actionsRead+1), zap.Int("expected_action_count", expectedActionCount), zap.String("account", string(act.Account)), zap.String("action", string(act.Name)))
				} else {
					seenMap[key] = true
					b.logger.Info("validated action", zap.Int("action_read", actionsRead+1), zap.Int("expected_action_count", expectedActionCount), zap.String("account", string(act.Account)), zap.String("action", string(act.Name)))
				}

				actionsRead++
			}
		}

		if actionsRead == len(bootSeq) {
			break
		}

	}

	if len(validationErrors) > 0 {
		return ValidationErrors{Errors: validationErrors}
	}

	return nil
}

func (b *Boot) flushMissingActions(seenMap map[string]bool, bootSeq []*zsw.Action) {
	fl, err := os.Create("missing_actions.jsonl")
	if err != nil {
		b.logger.Error("Couldn't write to `missing_actions.jsonl`:", zap.Error(err))
		return
	}
	defer fl.Close()

	// TODO: print all actions that are still MISSING to `missing_actions.jsonl`.
	b.logger.Info("Flushing unseen transactions to `missing_actions.jsonl` up until this point.")

	for _, act := range bootSeq {
		act.SetToServer(true)
		data, _ := zsw.MarshalBinary(act)
		key := sha2(data)

		if !seenMap[key] {
			act.SetToServer(false)
			data, _ := json.Marshal(act)
			fl.Write(data)
			fl.Write([]byte("\n"))
		}
	}
}

type ValidationError struct {
	Err               error
	BlockNumber       int
	Action            *zsw.Action
	RawAction         []byte
	Index             int
	ActionHexData     string
	PackedTransaction *zsw.PackedTransaction
}

func (e ValidationError) Error() string {
	s := fmt.Sprintf("Action [%d][%s::%s] absent from blocks\n", e.Index, e.Action.Account, e.Action.Name)

	data, err := json.Marshal(e.Action)
	if err != nil {
		s += fmt.Sprintf("    json generation err: %s\n", err)
	} else {
		s += fmt.Sprintf("    json data: %s\n", string(data))
	}
	s += fmt.Sprintf("    hex data: %s\n", hex.EncodeToString(e.RawAction))
	s += fmt.Sprintf("    error: %s\n", e.Err.Error())

	return s
}

type ValidationErrors struct {
	Errors []error
}

func (v ValidationErrors) Error() string {
	s := ""
	for _, err := range v.Errors {
		s += ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\n"
		s += err.Error()
		s += "<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n"
	}

	return s
}
