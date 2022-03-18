package ops

import (
	"bytes"
	"fmt"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
	"github.com/zhongshuwen/zswchain-go/token"
)

func init() {
	Register("system.create_voters", &OpCreateVoters{})
}

type OpCreateVoters struct {
	Creator zsw.AccountName
	Pubkey  string
	Count   int
}

func (op *OpCreateVoters) RequireValidation() bool {
	return true
}

func (op *OpCreateVoters) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	pubKey, err := decodeOpPublicKey(c, op.Pubkey)
	if err != nil {
		return err
	}

	for i := 0; i < op.Count; i++ {
		voterName := zsw.AccountName(voterName(i))
		fmt.Println("Creating voter: ", voterName)

		in <- (*TransactionAction)(system.NewNewAccount(op.Creator, voterName, pubKey))
		in <- (*TransactionAction)(token.NewTransfer(op.Creator, voterName, zsw.NewZSWAsset(1000000000), ""))
		in <- (*TransactionAction)(system.NewBuyRAMBytes(AN("eosio"), voterName, 8192)) // 8kb gift !
		in <- (*TransactionAction)(system.NewDelegateBW(AN("eosio"), voterName, zsw.NewZSWAsset(10000), zsw.NewZSWAsset(10000), true))
	}
	in <- EndTransaction(opPubkey) // end transaction
	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyz"

func voterName(index int) string {
	padding := string(bytes.Repeat([]byte{charset[index]}, 7))
	return "voter" + padding
}
