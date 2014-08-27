package git

import (
	"encoding/hex"
	"fmt"
)

// A SHA is a SHA-1 hash.
type SHA [20]byte

// String gives the hex representation of s: a9d7100fe1a37926a0a9cb05992e6eb3cdcb4f0d
func (s *SHA) String() string { return hex.EncodeToString((*s)[:]) }

// Short gives the hex representation of the first 4 bytes of s: a9d7100f
func (s *SHA) Short() string { return hex.EncodeToString((*s)[:4]) }

// Bytes gives the s as a byte slice.
func (s *SHA) Bytes() []byte { return []byte((*s)[:]) }

func makeSHA(s string) (*SHA, error) {
	if len(s) != 40 {
		return nil, fmt.Errorf("bad SHA-1 hash: %s", s)
	}
	var sha SHA
	if _, err := hex.Decode(sha[:], []byte(s)); err != nil {
		return nil, err
	}
	return &sha, nil
}
