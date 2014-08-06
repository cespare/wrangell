package git

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// A Commmit represents a Git commit.
type Commit struct {
	SHA *SHA
}

func (r *Repo) resolveRefFile(path string) (*Commit, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ref := strings.TrimSpace(string(contents))
	if strings.IndexByte(ref, '\n') != -1 {
		return nil, fmt.Errorf("malformed ref file %s", path)
	}
	if strings.HasPrefix(ref, "ref: ") {
		// Named ref
		refFile := strings.TrimPrefix(ref, "ref: ")
		return r.resolveRefFile(filepath.Join(r.gitDir, refFile))
	}
	// SHA-1
	return r.resolveSHA(ref)
}

var (
	packedRefNotFoundErr = errors.New("cannot locate packed ref")
	badPackedRefsFile    = errors.New("malformed packed-refs file")
)

func (r *Repo) resolvePackedRef(path string) (*Commit, error) {
	pathBytes := []byte(path)
	f, err := os.Open(filepath.Join(r.gitDir, "packed-refs"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		parts := bytes.Fields(line)
		if len(parts) != 2 {
			return nil, badPackedRefsFile
		}
		sha, linePath := parts[0], parts[1]
		if bytes.Equal(linePath, pathBytes) {
			return r.resolveSHA(string(sha))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, packedRefNotFoundErr
}

func (r *Repo) resolveSHA(ref string) (*Commit, error) {
	sha, err := makeSHA(ref)
	if err != nil {
		return nil, err
	}
	obj, err := r.objectBySHA(sha)
	if err != nil {
		return nil, fmt.Errorf("cannot locate ref %s in the object store", ref)
	}
	return r.commitFromObject(obj)
}

func (r *Repo) resolveSHAPrefix(prefix string) (*Commit, error) {
	obj, err := r.objectByPrefix(prefix)
	if err != nil {
		return nil, err
	}
	return r.commitFromObject(obj)
}

func (r *Repo) commitFromObject(obj *Object) (*Commit, error) {
	if obj.Type != TypeCommit {
		return nil, fmt.Errorf("bad object type for commit (%s)", obj.SHA)
	}
	return &Commit{
		SHA: obj.SHA,
	}, nil
}
