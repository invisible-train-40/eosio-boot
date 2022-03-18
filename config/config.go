package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/invisible-train-40/eosio-boot/content"
	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
	"go.uber.org/zap"
)

type abiCache struct {
	nodeApi *zsw.API
	abis    map[zsw.AccountName]*zsw.ABI
}

func newAbiCache(nodeApi *zsw.API) *abiCache {
	return &abiCache{
		nodeApi: nodeApi,
		abis:    map[zsw.AccountName]*zsw.ABI{},
	}
}

func (a *abiCache) SetABI(accountName zsw.AccountName, abi *zsw.ABI) {
	a.abis[accountName] = abi
}

func (a *abiCache) GetABI(accountName zsw.AccountName) (*zsw.ABI, error) {
	if abi, found := a.abis[accountName]; found {
		return abi, nil
	}

	resp, err := a.nodeApi.GetABI(context.Background(), accountName)
	if err != nil {
		return nil, fmt.Errorf("ABI not found in cache and could not retrieve from chain: %w", err)
	}

	abi := &resp.ABI
	a.SetABI(accountName, abi)

	return abi, nil
}

type OpConfig struct {
	contentRefs      []*content.ContentRef
	privateKeys      map[string]*ecc.PrivateKey
	contentManager   *content.Manager
	protocolFeatures []zsw.ProtocolFeature
	API              *zsw.API
	AbiCache         *abiCache
	Logger           *zap.Logger
}

func NewOpConfig(contentRefs []*content.ContentRef, contentManager *content.Manager, privateKeys map[string]*ecc.PrivateKey, api *zsw.API, protocolFeatures []zsw.ProtocolFeature, logger *zap.Logger) *OpConfig {
	return &OpConfig{
		contentRefs:      contentRefs,
		privateKeys:      privateKeys,
		contentManager:   contentManager,
		protocolFeatures: protocolFeatures,
		API:              api,
		AbiCache:         newAbiCache(api),
		Logger:           logger,
	}
}

func (c OpConfig) GetProtocolFeature(name string) zsw.Checksum256 {
	name = strings.ToUpper(name)
	for _, protocolFeature := range c.protocolFeatures {
		for _, spec := range protocolFeature.Specification {
			if spec.Value == name {
				return protocolFeature.FeatureDigest
			}
		}
	}
	return nil
}

func (c OpConfig) HackVotingAccounts() bool {
	return false
}

func (c OpConfig) ReadFromCache(ref string) ([]byte, error) {
	return c.contentManager.ReadFromCache(ref)

}

func (c OpConfig) GetContentsCacheRef(filename string) (string, error) {
	for _, fl := range c.contentRefs {
		if fl.Name == filename {
			return fl.URL, nil
		}
	}
	return "", fmt.Errorf("%q not found in target contents", filename)
}

func (c OpConfig) GetPrivateKey(label string) (*ecc.PrivateKey, error) {
	if _, found := c.privateKeys[label]; found {
		return c.privateKeys[label], nil
	}
	return nil, fmt.Errorf("bootseq does not contain key with label %q", label)

}

func (c OpConfig) FileNameFromCache(ref string) string {
	return c.contentManager.FileNameFromCache(ref)
}
