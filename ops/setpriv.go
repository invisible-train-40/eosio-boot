package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.setpriv", &OpSetPriv{})
}

type OpSetPriv struct {
	Account zsw.AccountName
}

func (op *OpSetPriv) RequireValidation() bool {
	return true
}

func (op *OpSetPriv) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(system.NewSetPriv(op.Account))
	in <- EndTransaction(opPubkey) // end transaction
	return nil

}
