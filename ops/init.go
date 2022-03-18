package ops

import (
	"fmt"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.init", &OpSystemInit{})
}

type OpSystemInit struct {
	Version zsw.Varuint32 `json:"version"`
	Core    string        `json:"core"`
}

func (op *OpSystemInit) RequireValidation() bool {
	return true
}

func (op *OpSystemInit) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	core, err := zsw.StringToSymbol(op.Core)
	if err != nil {
		return fmt.Errorf("unable to convert system.init core %q to symbol: %w", op.Core, err)
	}
	in <- (*TransactionAction)(system.NewInitSystem(op.Version, core))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
