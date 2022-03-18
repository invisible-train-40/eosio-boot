package snapshot

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/zhongshuwen/zswchain-go"
	"github.com/zhongshuwen/zswchain-go/ecc"
)

type Snapshot []SnapshotLine

type SnapshotLine struct {
	EthereumAddress string
	EOSPublicKey    ecc.PublicKey
	Balance         zsw.Asset
	AccountName     string
}

func New(content []byte) (out Snapshot, err error) {
	reader := csv.NewReader(bytes.NewBuffer(content))
	allRecords, err := reader.ReadAll()
	if err != nil {
		return
	}

	for _, el := range allRecords {
		if len(el) != 4 {
			return nil, fmt.Errorf("should have 4 elements per line")
		}

		newAsset, err := zsw.NewEOSAssetFromString(el[3])
		if err != nil {
			return out, err
		}

		pubKey, err := ecc.NewPublicKey(el[2])
		if err != nil {
			return out, err
		}

		out = append(out, SnapshotLine{el[0], pubKey, newAsset, el[1]})
	}

	return
}

type UnregdSnapshot []UnregdSnapshotLine

type UnregdSnapshotLine struct {
	EthereumAddress string
	AccountName     string
	Balance         zsw.Asset
}

func NewUnregd(content []byte) (out UnregdSnapshot, err error) {
	reader := csv.NewReader(bytes.NewBuffer(content))
	allRecords, err := reader.ReadAll()
	if err != nil {
		return
	}

	for _, el := range allRecords {
		if len(el) != 3 {
			return nil, fmt.Errorf("should have 2 elements per line")
		}

		newAsset, err := zsw.NewEOSAssetFromString(el[2])
		if err != nil {
			return out, err
		}

		out = append(out, UnregdSnapshotLine{el[0], el[1], newAsset})
	}

	return
}
