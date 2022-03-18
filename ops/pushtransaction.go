package ops

import (
	"encoding/json"
	"fmt"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
)

func init() {
	Register("system.pushtransaction", &OpPushTransaction{})
}

type OpPushTransaction struct {
	Contract   zsw.AccountName
	Action     zsw.ActionName
	Actor      zsw.AccountName
	Permission zsw.PermissionName
	Payload    map[string]interface{}
}

func (op *OpPushTransaction) RequireValidation() bool {
	return true
}

func (op *OpPushTransaction) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	cnt, err := json.Marshal(op.Payload)
	if err != nil {
		return fmt.Errorf("unable to marshal payload: %w", err)
	}

	abi, err := c.AbiCache.GetABI(op.Contract)
	if err != nil {
		return fmt.Errorf("cannot retrieve ABI for account %q encode payload: %w", op.Contract, err)
	}

	actionBinary, err := abi.EncodeAction(op.Action, []byte(cnt))

	action := &zsw.Action{
		Account: op.Contract,
		Name:    op.Action,
		Authorization: []zsw.PermissionLevel{
			{Actor: op.Actor, Permission: op.Permission},
		},
		ActionData: zsw.NewActionDataFromHexData(actionBinary),
	}
	in <- (*TransactionAction)(action)
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}

func encodePayload(payload string) (interface{}, error) {
	var hashData map[string]interface{}
	err := json.Unmarshal([]byte(payload), &hashData)
	if err == nil {
		return hashData, nil
	}

	var data []interface{}
	err = json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return nil, fmt.Errorf("unsupported payload format: %w", err)
	}
	return data, nil
}
