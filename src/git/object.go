package git

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ObjectType int

const (
	// These constants match the type value encoded in 3 bits in the pack format.
	TypeCommit ObjectType = 1
	TypeTree   ObjectType = 2
	TypeBlob   ObjectType = 3
	TypeTag    ObjectType = 4
	// The following two are only used for delta encoded objects in packfiles.
	TypeOfsDelta ObjectType = 6
	TypeRefDelta ObjectType = 7
)

func (t ObjectType) String() string {
	switch t {
	case TypeCommit:
		return "commit"
	case TypeTree:
		return "tree"
	case TypeBlob:
		return "blob"
	case TypeTag:
		return "tag"
	case TypeOfsDelta:
		return "ofs delta"
	case TypeRefDelta:
		return "ref delta"
	}
	panic("unexpected object type")
}

// An Object is a git object.
type Object struct {
	SHA  *SHA
	Type ObjectType
}

var (
	badObjectFileErr  = errors.New("found bad object file")
	objectNotExistErr = errors.New("object does not exist")
)

func (r *Repo) objectBySHA(sha *SHA) (*Object, error) {
	path := filepath.Join(r.gitDir, "objects", hex.EncodeToString((*sha)[:1]), hex.EncodeToString((*sha)[1:]))
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Is it packed?
			obj, err := r.packedObjectBySHA(sha)
			if os.IsNotExist(err) {
				return nil, objectNotExistErr
			}
			return obj, err
		}
		return nil, err
	}
	defer f.Close()
	return r.readObject(f, sha)
}

var (
	prefixNotFoundErr  = errors.New("no object was found with the given prefix")
	prefixNotUniqueErr = errors.New("more than one object matches the given prefix")
)

func (r *Repo) objectByPrefix(prefix string) (*Object, error) {
	if len(prefix) < 2 {
		return nil, fmt.Errorf("prefix %q is too short", prefix)
	}
	dirName, prefix := prefix[:2], prefix[2:]
	dirPath := filepath.Join(r.gitDir, "objects", dirName)
	dir, err := os.Open(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, prefixNotFoundErr
		}
		return nil, err
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var match string
	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			if match != "" {
				return nil, prefixNotUniqueErr
			}
			match = name
		}
	}
	if match == "" {
		return nil, prefixNotFoundErr
	}
	sha, err := makeSHA(dirName + match)
	if err != nil {
		return nil, err
	}
	return r.objectBySHA(sha)
}

func (r *Repo) readObject(rdr io.Reader, sha *SHA) (*Object, error) {
	z, err := zlib.NewReader(rdr)
	if err != nil {
		return nil, err
	}
	defer z.Close()

	// Read enough to grab the whole header
	buf := make([]byte, 100)
	n, err := io.ReadFull(z, buf)
	switch err {
	case io.EOF, io.ErrUnexpectedEOF, nil:
	default:
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
	case "blob":
		typ = TypeBlob
	case "tag":
		typ = TypeTag
	default:
		return nil, badObjectFileErr
	}

	return &Object{
		SHA:  sha,
		Type: typ,
	}, nil
}
