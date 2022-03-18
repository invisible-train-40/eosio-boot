package ops

import (
	"context"
	"fmt"
	"time"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"go.uber.org/zap"
)

func init() {
	Register("system.enable_protocol_features", &OpEnableProtocolFeature{})
}

type OpEnableProtocolFeature struct {
}

func (op *OpEnableProtocolFeature) RequireValidation() bool {
	return false
}

func (op *OpEnableProtocolFeature) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) error {
	ctx := context.Background()
	preactivateFeature := "PREACTIVATE_FEATURE"
	featureDigest := c.GetProtocolFeature(preactivateFeature)
	if featureDigest == nil {
		return fmt.Errorf("cannot enable protocol features: cannot determined %q feature digest", preactivateFeature)
	}

	c.Logger.Info("activating protocol features!")
	err := c.API.ScheduleProducerProtocolFeatureActivations(ctx, []zsw.Checksum256{featureDigest})
	if err != nil {
		c.Logger.Error("cannot enable protocol feature %q: %w",
			zap.Error(err),
			zap.String("protocol", preactivateFeature),
		)
		//return fmt.Errorf("cannot enable protocol feature %q: %w", preactivateFeature, err)
	}
	c.Logger.Info("successfully enabled protcol features")

	// TODO: we need to sleep here to make sure that the chain processed a block
	time.Sleep(2 * time.Second)
	return nil
}
