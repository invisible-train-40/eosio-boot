package unregd

import "github.com/zhongshuwen/zswchain-go"

func NewAdd(ethAccount string, balance zsw.Asset) *zsw.Action {
	action := &zsw.Action{
		Account: zsw.AccountName("eosio.unregd"),
		Name:    zsw.ActionName("add"),
		Authorization: []zsw.PermissionLevel{
			{zsw.AccountName("eosio.unregd"), zsw.PermissionName("active")},
		},
		ActionData: zsw.NewActionData(Add{
			EthereumAddress: ethAccount,
			Balance:         balance,
		}),
	}
	return action
}

type Add struct {
	EthereumAddress string    `json:"ethereum_account"`
	Balance         zsw.Asset `json:"balance"`
}
