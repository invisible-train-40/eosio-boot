package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.setram", &OpSetRAM{})
}

type OpSetRAM struct {
	MaxRAMSize uint64 `json:"max_ram_size"`
}

func (op *OpSetRAM) RequireValidation() bool {
	return true
}

func (op *OpSetRAM) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(system.NewSetRAM(op.MaxRAMSize))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
