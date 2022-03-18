package ops

import (
	"fmt"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.setabi", &OpSetABI{})
}

type OpSetABI struct {
	Account         zsw.AccountName
	ContractNameRef string `json:"contract_name_ref"`
}

func (op *OpSetABI) RequireValidation() bool {
	return true
}

func (op *OpSetABI) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	abiFileRef, err := c.GetContentsCacheRef(fmt.Sprintf("%s.abi", op.ContractNameRef))
	if err != nil {
		return err
	}

	abi, err := retrieveABIfromRef(c.FileNameFromCache(abiFileRef))
	if err != nil {
		return fmt.Errorf("unable to read ABI %s: %s", abiFileRef, err)
	}

	abiAction, err := system.NewSetAbiFromAbi(op.Account, *abi)
	if err != nil {
		return fmt.Errorf("NewSetAbiFromAbi %s: %s", op.ContractNameRef, err)
	}

	c.AbiCache.SetABI(op.Account, abi)
	for _, act := range []*zsw.Action{abiAction} {
		in <- (*TransactionAction)(act)
	}

	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
