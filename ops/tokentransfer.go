package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/token"
)

func init() {
	Register("token.transfer", &OpTransferToken{})
}

type OpTransferToken struct {
	From     zsw.AccountName
	To       zsw.AccountName
	Quantity zsw.Asset
	Memo     string
}

func (op *OpTransferToken) RequireValidation() bool {
	return true
}

func (op *OpTransferToken) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(token.NewTransfer(op.From, op.To, op.Quantity, op.Memo))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
