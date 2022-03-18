package ops

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
)

type Operation interface {
	Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error
	RequireValidation() bool
}

var operationsRegistry = map[string]Operation{}

type TransactionAction zsw.Action

type TransactionBoundary struct {
	Signer ecc.PublicKey
}

func EndTransaction(signer ecc.PublicKey) TransactionBoundary {
	return TransactionBoundary{
		Signer: signer,
	}
}

func Register(key string, operation Operation) {
	if key == "" {
		panic(fmt.Sprintf("canont register an ops with  a blank key"))
	} else if _, ok := operationsRegistry[key]; ok {
		panic(fmt.Sprintf("ops already registered: %q", key))
	}
	operationsRegistry[key] = operation

}

type OperationType struct {
	Op             string
	Signer         string
	Label          string
	SkipValidation bool
	Data           Operation
}

func (o *OperationType) UnmarshalJSON(data []byte) error {
	opData := struct {
		Op             string
		Signer         string
		Label          string
		SkipValidation bool
		Data           json.RawMessage
	}{}
	if err := json.Unmarshal(data, &opData); err != nil {
		return err
	}

	opType, found := operationsRegistry[opData.Op]
	if !found {
		return fmt.Errorf("operation type %q invalid, use one of: %q", opData.Op, operationsRegistry)
	}

	objType := reflect.TypeOf(opType).Elem()
	obj := reflect.New(objType).Interface()

	if len(opData.Data) != 0 {
		err := json.Unmarshal(opData.Data, &obj)
		if err != nil {
			return fmt.Errorf("operation type %q invalid, error decoding: %s", opData.Op, err)
		}
	}

	opIface, ok := obj.(Operation)
	if !ok {
		return fmt.Errorf("operation type %q isn't an op", opData.Op)
	}

	*o = OperationType{
		Op:             opData.Op,
		Label:          opData.Label,
		Signer:         opData.Signer,
		SkipValidation: opData.SkipValidation,
		Data:           opIface,
	}

	return nil
}
