package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.buy_ram", &OpBuyRam{})
}

type OpBuyRam struct {
	Payer       zsw.AccountName
	Receiver    zsw.AccountName
	EOSQuantity uint64 `json:"eos_quantity"`
}

func (op *OpBuyRam) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(system.NewBuyRAM(op.Payer, op.Receiver, op.EOSQuantity))
	in <- EndTransaction(opPubkey) // end transaction
	return nil

}

func (op *OpBuyRam) RequireValidation() bool {
	return true
}
