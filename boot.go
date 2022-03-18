package boot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/invisible-train-40/eosio-boot/content"
	"github.com/invisible-train-40/eosio-boot/ops"
	"github.com/invisible-train-40/eosio-boot/snapshot"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
	"go.uber.org/zap"
)

type Option interface {
	apply(b *Boot)
}

type funcOption func(b *Boot)

func (f funcOption) apply(b *Boot) {
	f(b)
}

var traceEnable = false

func init() {
	traceEnable = os.Getenv("TRACE") == "true"
}

func WithMaxActionCountPerTrx(max int) Option {
	return funcOption(func(b *Boot) {
		b.maxActionCountPerTrx = max
	})
}

func WithKeyBag(keyBag *zsw.KeyBag) Option {
	return funcOption(func(b *Boot) {
		b.keyBag = keyBag
	})
}

func WithLogger(logger *zap.Logger) Option {
	return funcOption(func(b *Boot) {
		b.logger = logger
		b.contentManager.SetLogger(logger)
	})
}

func WithHackVotingActions() Option {
	return funcOption(func(b *Boot) {
		b.hackVotingAccounts = true
	})
}

type Boot struct {
	Snapshot snapshot.Snapshot

	bootSequencePath     string
	targetNetAPI         *zsw.API
	bootstrappingEnabled bool
	genesisPath          string
	bootSequence         *BootSeq
	contentManager       *content.Manager
	keyBag               *zsw.KeyBag
	bootseqKeys          map[string]*ecc.PrivateKey
	maxActionCountPerTrx int
	hackVotingAccounts   bool

	logger *zap.Logger
}

func New(bootSequencePath string, targetAPI *zsw.API, cachePath string, opts ...Option) (b *Boot, err error) {
	b = &Boot{
		targetNetAPI:         targetAPI,
		bootSequencePath:     bootSequencePath,
		contentManager:       content.NewManager(cachePath),
		maxActionCountPerTrx: 500,
		bootseqKeys:          map[string]*ecc.PrivateKey{},
		logger:               zap.NewNop(),
	}
	for _, opt := range opts {
		opt.apply(b)
	}

	b.bootSequence, err = readBootSeq(b.bootSequencePath)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Boot) Revision() string {
	return b.bootSequence.Checksum
}

func (b *Boot) getBootseqKey(label string) (*ecc.PrivateKey, error) {
	if _, found := b.bootseqKeys[label]; found {
		return b.bootseqKeys[label], nil
	}
	return nil, fmt.Errorf("bootseq does not contain key with label %q", label)
}

func (b *Boot) Run() (checksums string, err error) {
	ctx := context.Background()

	b.logger.Debug("parsing boot sequence keys")
	if err := b.parseBootseqKeys(); err != nil {
		return "", fmt.Errorf("unable to parse boot sequence: %w", err)
	}

	b.logger.Debug("downloading references")
	if err := b.contentManager.Download(b.bootSequence.Contents); err != nil {
		return "", fmt.Errorf("unable to download content references: %w", err)
	}

	b.logger.Debug("setting boot keys")
	if err := b.setKeys(); err != nil {
		return "", fmt.Errorf("unable to set boot keys: %w", err)
	}

	if err := b.attachKeysOnTargetNode(ctx); err != nil {
		return "", fmt.Errorf("unable to attach keys on target node: %w", err)
	}

	// We need to wait for target node to be up prior calling the get producer protocol features below
	b.waitTargetNodeToBeUp()

	features, err := b.targetNetAPI.GetProducerProtocolFeatures(ctx)
	if err != nil {
		return "", fmt.Errorf("unable to get producer protocol features: %w", err)
	}

	opConfig := config.NewOpConfig(
		b.bootSequence.Contents,
		b.contentManager,
		b.bootseqKeys,
		b.targetNetAPI,
		features,
		b.logger,
	)

	trxEventCh := make(chan interface{}, 500)
	go func() {
		defer close(trxEventCh)
		for _, step := range b.bootSequence.BootSequence {
			b.logger.Info("executing bootseq op",
				zap.String("label", step.Label),
				zap.String("op", step.Op),
				zap.String("signer", step.Signer),
				zap.Bool("skip_validation", step.SkipValidation),
			)
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

	index := 0
	for {
		index++
		trxBundle := b.chunkifyActionChan(trxEventCh)
		if trxBundle == nil {
			// chunkify exited without given any chunks, channel must be closed
			break
		}

		if len(trxBundle.actions) == 0 {
			// nothing to execute skip
			continue
		}

		str := []string{}
		for _, t := range trxBundle.actions {
			str = append(str, fmt.Sprintf("%s:%s", t.Account, t.Name))
		}
		b.logger.Debug("pushing transaction",
			zap.Int("index", index),
			zap.Int("action_count", len(trxBundle.actions)),
			zap.String("actions", strings.Join(str, ", ")),
		)
		b.targetNetAPI.SetCustomGetRequiredKeys(func(ctx context.Context, tx *zsw.Transaction) (out []ecc.PublicKey, err error) {
			out = append(out, trxBundle.signer)
			return out, nil
		})

		err := Retry(25, time.Second, func() error {
			_, err := b.targetNetAPI.SignPushActions(ctx, trxBundle.actions...)
			if err != nil {
				b.logger.Error("error pushing transaction bundle",
					zap.Error(err),
					zap.Int("index", index),
				)
				if traceEnable {
					trxBundle.debugPrint(b.logger)
				}
				return fmt.Errorf("push actions of transaciton bundle: %w", err)
			}

			return nil
		})
		if err != nil {
			b.logger.Error("failed to push transaction bundle", zap.Error(err))
			return "", err
		}
	}

	b.logger.Info("waiting 2 seconds for transactions to flush to blocks")
	time.Sleep(2 * time.Second)

	// FIXME: don't do chain validation here..
	isValid, err := b.RunChainValidation(opConfig)
	if err != nil {
		return "", fmt.Errorf("chain validation: %s", err)
	}
	if !isValid {
		b.logger.Info("WARNING: chain invalid, destroying network if possible")
		os.Exit(0)
	}

	return b.bootSequence.Checksum, nil
}

type transactionBundle struct {
	actions []*zsw.Action
	signer  ecc.PublicKey
}

// helpful for debug puropses
func (t *transactionBundle) debugPrint(logger *zap.Logger) {
	acts := []string{}
	logger.Debug("transaction bundle dump start", zap.Int("action_count", len(t.actions)))
	for _, action := range t.actions {
		actionKey := fmt.Sprintf("%s:%s", action.Account, action.Name)
		var str string
		switch actionKey {
		case "eosio:newaccount":
			logger.Debug("action: new account",
				zap.Reflect("account", (action.ActionData.Data).(system.NewAccount)),
			)
		case "eosio:setabi":
			setABIAction := (action.ActionData.Data).(system.SetABI)
			h := sha256.New()
			h.Write(setABIAction.ABI)
			logger.Debug("action: set abi",
				zap.String("account", string(setABIAction.Account)),
				zap.ByteString("abi_sha256", h.Sum(nil)),
			)
		case "eosio:setcode":
			setCodeAction := (action.ActionData.Data).(system.SetCode)
			h := sha256.New()
			h.Write(setCodeAction.Code)
			logger.Debug("action: set code",
				zap.String("account", string(setCodeAction.Account)),
				zap.ByteString("code_sha256", h.Sum(nil)),
			)
		case "eosio:linkauth":
			logger.Debug("action: link auth code",
				zap.Reflect("link_auth", (action.ActionData.Data).(system.LinkAuth)),
			)
		case "eosio:updateauth":
			logger.Debug("action: update auth", zap.Reflect("update_auth", (action.ActionData.Data).(system.UpdateAuth)))
		case "eosio:init":
			initAction := (action.ActionData.Data).(system.Init)
			logger.Debug("action: eosio init",
				zap.Uint32("version", uint32(initAction.Version)),
				zap.String("core", initAction.Core.String()),
			)

		}
		acts = append(acts, str)
	}
	logger.Debug("transaction bundle dump end")

}

func (b *Boot) chunkifyActionChan(trxEventCh chan interface{}) *transactionBundle {
	out := &transactionBundle{
		actions: []*zsw.Action{},
	}
	for {
		if len(out.actions) > b.maxActionCountPerTrx {
			return out
		}
		act, ok := <-trxEventCh
		if !ok {
			// channel is closed, there is not transaction to process
			return nil
		}
		switch v := act.(type) {
		case ops.TransactionBoundary:
			out.signer = v.Signer
			return out
		case *ops.TransactionAction:
			out.actions = append(out.actions, (*zsw.Action)(v))
		default:
			panic(fmt.Sprintf("chunkify: unexpected type in action chan"))
		}
	}
}

func (b *Boot) getOpPubkey(op *ops.OperationType) (ecc.PublicKey, error) {
	if op.Signer != "" {
		if privKey, found := b.bootseqKeys[op.Signer]; found {
			return privKey.PublicKey(), nil
		}
		return ecc.PublicKey{}, fmt.Errorf("cannot find private key in boot sequence with label %q", op.Signer)
	}

	pubKey, err := b.getBootKey()
	if err != nil {
		return ecc.PublicKey{}, err
	}
	return pubKey, nil
}
