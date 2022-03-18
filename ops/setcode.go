package ops

import (
	"fmt"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.setcode", &OpSetCode{})
}

type OpSetCode struct {
	Account         zsw.AccountName
	ContractNameRef string `json:"contract_name_ref"`
	PermissionLevel string `json:"permission_level"`
}

func (op *OpSetCode) RequireValidation() bool {
	return true
}

func (op *OpSetCode) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	wasmFileRef, err := c.GetContentsCacheRef(fmt.Sprintf("%s.wasm", op.ContractNameRef))
	if err != nil {
		return err
	}

	abiFileRef, err := c.GetContentsCacheRef(fmt.Sprintf("%s.abi", op.ContractNameRef))
	if err != nil {
		return err
	}

	codeAction, err := system.NewSetCode(op.Account, c.FileNameFromCache(wasmFileRef))
	if err != nil {
		return fmt.Errorf("NewSetCode %s: %s", op.ContractNameRef, err)
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

	if op.PermissionLevel != "" {
		permissionLevel, err := zsw.NewPermissionLevel(op.PermissionLevel)
		if err != nil {
			return fmt.Errorf("unable to read permission level: %w", err)
		}

		codeAction.Authorization = []zsw.PermissionLevel{permissionLevel}
		abiAction.Authorization = []zsw.PermissionLevel{permissionLevel}
	}

	in <- (*TransactionAction)(codeAction)
	in <- (*TransactionAction)(abiAction)

	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
