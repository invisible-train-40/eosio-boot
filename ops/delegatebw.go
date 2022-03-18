package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.delegate_bw", &OpDelegateBW{})
}

type OpDelegateBW struct {
	From     zsw.AccountName
	To       zsw.AccountName
	StakeCPU int64 `json:"stake_cpu"`
	StakeNet int64 `json:"stake_net"`
	Transfer bool
}

func (op *OpDelegateBW) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	in <- (*TransactionAction)(system.NewDelegateBW(op.From, op.To, zsw.NewEOSAsset(op.StakeCPU), zsw.NewEOSAsset(op.StakeNet), op.Transfer))
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}

func (op *OpDelegateBW) RequireValidation() bool {
	return true
}
