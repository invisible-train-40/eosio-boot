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
			zsw.NewEOSAsset(10000), // 1.0 EOS
			zsw.NewEOSAsset(2500),
			zsw.NewEOSAsset(2500),
			zsw.NewEOSAsset(5000), // 0.5 EOS
		},
		{
			zsw.NewEOSAsset(100000), // 10.0 EOS
			zsw.NewEOSAsset(2500),   // 0.25 EOS
			zsw.NewEOSAsset(2500),   // 0.25 EOS
			zsw.NewEOSAsset(95000),  // 9.5 EOS
		},
		{
			zsw.NewEOSAsset(105000), // 10.5 EOS
			zsw.NewEOSAsset(2500),   // 0.25 EOS
			zsw.NewEOSAsset(2500),   // 0.25 EOS
			zsw.NewEOSAsset(100000), // 10.0 EOS
		},
		{
			zsw.NewEOSAsset(107000), // 10.7 EOS
			zsw.NewEOSAsset(3500),   // 0.35 EOS
			zsw.NewEOSAsset(3500),   // 0.35 EOS
			zsw.NewEOSAsset(100000), // 10.0 EOS
		},
		{
			zsw.NewEOSAsset(120000), // 12.0 EOS
			zsw.NewEOSAsset(10000),  // 0.25 + 0.75 EOS
			zsw.NewEOSAsset(10000),  // 0.25 + 0.75 EOS
			zsw.NewEOSAsset(100000), // 10.0 EOS
		},
		{
			zsw.NewEOSAsset(99990000), // 9999.0 EOS
			zsw.NewEOSAsset(49945000), // 4994.5 EOS
			zsw.NewEOSAsset(49945000), // 4994.5 EOS, 10.0 EOS remaining :) yessir!
			zsw.NewEOSAsset(100000),   // 10.0 EOS
		},
	}

	for idx, test := range tests {
		cpuStake, netStake, xfer := splitSnapshotStakes(test.balance)
		assert.Equal(t, test.cpuStake, cpuStake, fmt.Sprintf("idx=%d", idx))
		assert.Equal(t, test.netStake, netStake, fmt.Sprintf("idx=%d", idx))
		assert.Equal(t, test.xfer, xfer, fmt.Sprintf("idx=%d", idx))
	}
}
