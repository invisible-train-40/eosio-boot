package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.buy_ram_bytes", &OpBuyRamBytes{})
}

type OpBuyRamBytes struct {
	Payer    zsw.AccountName
	Receiver zsw.AccountName
	Bytes    uint32
}

func (op *OpBuyRamBytes) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(system.NewBuyRAMBytes(op.Payer, op.Receiver, op.Bytes))
	in <- EndTransaction(opPubkey) // end transaction
	return nil

}

func (op *OpBuyRamBytes) RequireValidation() bool {
	return true
}
