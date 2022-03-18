package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/token"
)

func init() {
	Register("token.issue", &OpIssueToken{})
}

type OpIssueToken struct {
	Account zsw.AccountName
	Amount  zsw.Asset
	Memo    string
}

func (op *OpIssueToken) RequireValidation() bool {
	return true
}

func (op *OpIssueToken) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(token.NewIssue(op.Account, op.Amount, op.Memo))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
