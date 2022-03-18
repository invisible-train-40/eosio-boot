package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/token"
)

func init() {
	Register("token.create", &OpCreateToken{})
}

type OpCreateToken struct {
	// TODO: this should have be Issuer
	Account zsw.AccountName `json:"account"`
	// TODO: this should be MaximumSupply
	Amount zsw.Asset `json:"amount"`
}

func (op *OpCreateToken) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(token.NewCreate(op.Account, op.Amount))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}

func (op *OpCreateToken) RequireValidation() bool {
	return true
}
