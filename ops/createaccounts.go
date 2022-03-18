package ops

import (
	"fmt"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/invisible-train-40/eosio-boot/snapshot"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"github.com/zhongshuwen/zswchain-go/system"
	"github.com/zhongshuwen/zswchain-go/token"
	"go.uber.org/zap"
)

func init() {
	Register("snapshot.create_accounts", &OpSnapshotCreateAccounts{})
}

type OpSnapshotCreateAccounts struct {
	BuyRAMBytes             uint64 `json:"buy_ram_bytes"`
	TestnetTruncateSnapshot int    `json:"TESTNET_TRUNCATE_SNAPSHOT"`
}

func (op *OpSnapshotCreateAccounts) RequireValidation() bool {
	return true
}

func (op *OpSnapshotCreateAccounts) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	snapshotFile, err := c.GetContentsCacheRef("snapshot.csv")
	if err != nil {
		return err
	}

	rawSnapshot, err := c.ReadFromCache(snapshotFile)
	if err != nil {
		return fmt.Errorf("reading snapshot file: %s", err)
	}

	snapshotData, err := snapshot.New(rawSnapshot)
	if err != nil {
		return fmt.Errorf("loading snapshot csv: %s", err)
	}

	if len(snapshotData) == 0 {
		return fmt.Errorf("snapshot is empty or not loaded")
	}

	wellKnownPubkey, _ := ecc.NewPublicKey("EOS6MRyAjQq8ud7hVNYcfnVPJqcVpscN5So8BhtHuGYqET5GDW5CV")

	for idx, hodler := range snapshotData {
		if trunc := op.TestnetTruncateSnapshot; trunc != 0 {
			if idx == trunc {
				c.Logger.Debug("truncated snapshot", zap.Int("at_row", trunc))
				break
			}
		}

		destAccount := AN(hodler.AccountName)
		destPubKey := hodler.EOSPublicKey
		if c.HackVotingAccounts() {
			destPubKey = wellKnownPubkey
		}

		cpuStake, netStake, rest := splitSnapshotStakes(hodler.Balance)

		in <- system.NewNewAccount(AN("eosio"), destAccount, destPubKey)
		// special case `transfer` for `b1` ?
		in <- (*TransactionAction)(system.NewDelegateBW(AN("eosio"), destAccount, cpuStake, netStake, true))
		in <- (*TransactionAction)(system.NewBuyRAMBytes(AN("eosio"), destAccount, uint32(op.BuyRAMBytes)))
		in <- EndTransaction(opPubkey) // end transaction

		memo := "Welcome " + hodler.EthereumAddress[len(hodler.EthereumAddress)-6:]
		in <- (*TransactionAction)(token.NewTransfer(AN("eosio"), destAccount, rest, memo))
		in <- EndTransaction(opPubkey) // end transaction
	}
	return nil
}

func splitSnapshotStakes(balance zsw.Asset) (cpu, net, xfer zsw.Asset) {
	if balance.Amount < 5000 {
		return
	}

	// everyone has minimum 0.25 EOS staked
	// some 10 EOS unstaked
	// the rest split between the two

	cpu = zsw.NewEOSAsset(2500)
	net = zsw.NewEOSAsset(2500)

	remainder := zsw.NewEOSAsset(int64(balance.Amount - cpu.Amount - net.Amount))

	if remainder.Amount <= 100000 /* 10.0 EOS */ {
		return cpu, net, remainder
	}

	remainder.Amount -= 100000 // keep them floating, unstaked

	firstHalf := remainder.Amount / 2
	cpu.Amount += firstHalf
	net.Amount += remainder.Amount - firstHalf

	return cpu, net, zsw.NewEOSAsset(100000)
}
