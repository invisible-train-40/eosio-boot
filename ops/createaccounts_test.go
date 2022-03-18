package ops

import (
	"fmt"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSnapshotDelegationAmounts(t *testing.T) {
	tests := []struct {
		balance  zsw.Asset
		cpuStake zsw.Asset
		netStake zsw.Asset
		xfer     zsw.Asset
	}{
		{
			zsw.NewZSWAsset(10000), // 1.0 EOS
			zsw.NewZSWAsset(2500),
			zsw.NewZSWAsset(2500),
			zsw.NewZSWAsset(5000), // 0.5 EOS
		},
		{
			zsw.NewZSWAsset(100000), // 10.0 EOS
			zsw.NewZSWAsset(2500),   // 0.25 EOS
			zsw.NewZSWAsset(2500),   // 0.25 EOS
			zsw.NewZSWAsset(95000),  // 9.5 EOS
		},
		{
			zsw.NewZSWAsset(105000), // 10.5 EOS
			zsw.NewZSWAsset(2500),   // 0.25 EOS
			zsw.NewZSWAsset(2500),   // 0.25 EOS
			zsw.NewZSWAsset(100000), // 10.0 EOS
		},
		{
			zsw.NewZSWAsset(107000), // 10.7 EOS
			zsw.NewZSWAsset(3500),   // 0.35 EOS
			zsw.NewZSWAsset(3500),   // 0.35 EOS
			zsw.NewZSWAsset(100000), // 10.0 EOS
		},
		{
			zsw.NewZSWAsset(120000), // 12.0 EOS
			zsw.NewZSWAsset(10000),  // 0.25 + 0.75 EOS
			zsw.NewZSWAsset(10000),  // 0.25 + 0.75 EOS
			zsw.NewZSWAsset(100000), // 10.0 EOS
		},
		{
			zsw.NewZSWAsset(99990000), // 9999.0 EOS
			zsw.NewZSWAsset(49945000), // 4994.5 EOS
			zsw.NewZSWAsset(49945000), // 4994.5 EOS, 10.0 EOS remaining :) yessir!
			zsw.NewZSWAsset(100000),   // 10.0 EOS
		},
	}

	for idx, test := range tests {
		cpuStake, netStake, xfer := splitSnapshotStakes(test.balance)
		assert.Equal(t, test.cpuStake, cpuStake, fmt.Sprintf("idx=%d", idx))
		assert.Equal(t, test.netStake, netStake, fmt.Sprintf("idx=%d", idx))
		assert.Equal(t, test.xfer, xfer, fmt.Sprintf("idx=%d", idx))
	}
}
