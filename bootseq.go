package boot

import (
	"crypto/sha256"
	"fmt"
	"github.com/invisible-train-40/eosio-boot/content"
	"github.com/invisible-train-40/eosio-boot/ops"
	"io/ioutil"
)

type BootSeq struct {
	Keys         map[string]string    `json:"keys"`
	Contents     []*content.ContentRef        `json:"contents"`
	BootSequence []*ops.OperationType `json:"boot_sequence"`
	Checksum     string
}

func readBootSeq(filename string) (out *BootSeq, err error) {
	rawBootSeq, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading boot seq: %w", err)
	}

	if err := yamlUnmarshal(rawBootSeq, &out); err != nil {
		return nil, fmt.Errorf("parsing boot seq yaml: %w", err)
	}
	out.Checksum = fmt.Sprintf("%x", sha256.Sum256(rawBootSeq))
	return
}

