package git

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
)

type ObjectType int

const (
	TypeCommit ObjectType = iota + 1
	TypeTree
	TypeTag
	TypeBlob
)

// An Object is a git object.
type Object struct {
	SHA  *SHA
	Type ObjectType
}

var (
	badObjectFileErr = errors.New("found bad object file")
)

func (r *Repo) objectBySHA(sha *SHA) (*Object, error) {
	path := filepath.Join(r.gitDir, "objects", hex.EncodeToString((*sha)[:1]), hex.EncodeToString((*sha)[1:]))
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	z, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	// Read enough to grab the whole header
	buf := make([]byte, 100)
	n, err := z.Read(buf)
	if err != nil {
		return nil, err
	}
	buf = buf[:n]
	headerLen := bytes.Index(buf, []byte{0})
	if headerLen == -1 {
		return nil, badObjectFileErr
	}
	parts := bytes.SplitN(buf[:headerLen], []byte{' '}, 2)
	var typ ObjectType
	switch string(parts[0]) {
	case "commit":
		typ = TypeCommit
	case "tree":
		typ = TypeTree
	case "tag":
		typ = TypeTag
	case "blob":
		typ = TypeBlob
	default:
		return nil, badObjectFileErr
	}

	return &Object{
		SHA:  sha,
		Type: typ,
	}, nil
}
