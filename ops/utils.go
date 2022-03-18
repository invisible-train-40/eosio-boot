package ops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/invisible-train-40/eosio-boot/config"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
)

var AN = zsw.AN
var ActN = zsw.ActN
var PN = zsw.PN

func decodeOpPublicKey(c *config.OpConfig, opPubKey string) (ecc.PublicKey, error) {
	privateKey, err := c.GetPrivateKey(opPubKey)
	if err == nil {
		return privateKey.PublicKey(), nil
	}

	pubKey, err := ecc.NewPublicKey(opPubKey)
	if err != nil {
		return ecc.PublicKey{}, fmt.Errorf("reading pubkey: %s", err)
	}
	return pubKey, nil
}

// this is use to support ephemeral key
func getBootKey(c *config.OpConfig) (ecc.PublicKey, error) {
	privateKey, err := c.GetPrivateKey("boot")
	if err == nil {
		return privateKey.PublicKey(), nil
	}

	privateKey, err = c.GetPrivateKey("ephemeral")
	if err == nil {
		return privateKey.PublicKey(), nil
	}

	return ecc.PublicKey{}, fmt.Errorf("cannot find boot/ephemeral key")
}

func retrieveABIfromRef(abiPath string) (*zsw.ABI, error) {
	abiContent, err := ioutil.ReadFile(abiPath)
	if err != nil {
		return nil, err
	}
	if len(abiContent) == 0 {
		return nil, fmt.Errorf("unable to unmarshal abi with 0 bytes")
	}

	var abiDef zsw.ABI
	if err := json.Unmarshal(abiContent, &abiDef); err != nil {
		return nil, fmt.Errorf("unmarshal ABI file: %s", err)
	}

	return &abiDef, nil
}
