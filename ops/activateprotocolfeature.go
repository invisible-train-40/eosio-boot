package ops

import (
	"encoding/hex"
	"fmt"
	"regexp"

	"github.com/zhongshuwen/zswchain-go/system"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
)

func init() {
	Register("system.activate_protocol_features", &ActivateProtocolFeatures{})
}

type ActivateProtocolFeatures struct {
	Features []string
}

func (op *ActivateProtocolFeatures) RequireValidation() bool {
	return false
}

func (op *ActivateProtocolFeatures) Actions(opPubkey ecc.PublicKey, c *config.OpConfig, in chan interface{}) (err error) {
	actions := []*zsw.Action{}

	protocolRegExp := regexp.MustCompile(`^[a-zA-Z][a-zA-Z_]+[a-zA-Z]$`)

	for _, feature := range op.Features {
		var featureDigest zsw.Checksum256

		if protocolRegExp.Match([]byte(feature)) {
			featureDigest = c.GetProtocolFeature(feature)
			if featureDigest == nil {
				return fmt.Errorf("cannot determined '%q' feature digest", feature)
			}
			actions = append(actions, system.NewActivateFeature(featureDigest))
			continue
		} else {
			featureDigest, err = hex.DecodeString(feature)
			if err != nil {
				return fmt.Errorf("unable to unmarshal feature into feature digest (checksum256) %q: %w", feature, err)

			}
		}
		actions = append(actions, system.NewActivateFeature(featureDigest))
	}

	for _, act := range actions {
		in <- (*TransactionAction)(act)
	}

	in <- EndTransaction(opPubkey) // end transaction
	return nil
}
