package ops

import (
	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
)

func init() {
	Register("system.resign_accounts", &OpResignAccounts{})
}

type OpResignAccounts struct {
	Accounts            []zsw.AccountName
	TestnetKeepAccounts bool `json:"TESTNET_KEEP_ACCOUNTS"`
}

func (op *OpResignAccounts) RequireValidation() bool {
	return true
}

func (op *OpResignAccounts) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	if op.TestnetKeepAccounts {
		c.Logger.Debug("keeping system accounts around, for testing purposes.")
		return nil
	}

	systemAccount := AN("eosio")
	prodsAccount := AN("eosio.prods") // this is a special system account that is granted by 2/3 + 1 of the current BP schedule.

	eosioPresent := false
	for _, acct := range op.Accounts {
		if acct == systemAccount {
			eosioPresent = true
			continue
		}

		in <- (*TransactionAction)(system.NewUpdateAuth(acct, PN("active"), PN("owner"), zsw.Authority{
			Threshold: 1,
			Accounts: []zsw.PermissionLevelWeight{
				zsw.PermissionLevelWeight{
					Permission: zsw.PermissionLevel{
						Actor:      AN("eosio"),
						Permission: PN("active"),
					},
					Weight: 1,
				},
			},
		}, PN("active")))
		in <- (*TransactionAction)(system.NewUpdateAuth(acct, PN("owner"), PN(""), zsw.Authority{
			Threshold: 1,
			Accounts: []zsw.PermissionLevelWeight{
				zsw.PermissionLevelWeight{
					Permission: zsw.PermissionLevel{
						Actor:      AN("eosio"),
						Permission: PN("active"),
					},
					Weight: 1,
				},
			},
		}, PN("owner")))

	}

	if eosioPresent {
		in <- (*TransactionAction)(system.NewUpdateAuth(systemAccount, PN("active"), PN("owner"), zsw.Authority{
			Threshold: 1,
			Accounts: []zsw.PermissionLevelWeight{
				zsw.PermissionLevelWeight{
					Permission: zsw.PermissionLevel{
						Actor:      prodsAccount,
						Permission: PN("active"),
					},
					Weight: 1,
				},
			},
		}, PN("active")))
		in <- (*TransactionAction)(system.NewUpdateAuth(systemAccount, PN("owner"), PN(""), zsw.Authority{
			Threshold: 1,
			Accounts: []zsw.PermissionLevelWeight{
				zsw.PermissionLevelWeight{
					Permission: zsw.PermissionLevel{
						Actor:      prodsAccount,
						Permission: PN("active"),
					},
					Weight: 1,
				},
			},
		}, PN("owner")))
	}

	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
