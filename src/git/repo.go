package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// A Repo represents a Git repository.
type Repo struct {
	gitDir    string // GIT_DIR
	workTree  string // GIT_WORK_TREE
	indexFile string // GIT_INDEX_FILE
}

// Open initializes a Repo for a git repository at dir.
func Open(dir string) (*Repo, error) {
	gitDir := filepath.Join(dir, ".git")
	stat, err := os.Stat(gitDir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		// TODO: support a gitdir file (and GIT_DIR)
		return nil, fmt.Errorf("cannot open git dir at %s; it is not a dir (separate gitdirs not supported)",
			gitDir)
	}
	return &Repo{
		gitDir:   gitDir,
		workTree: dir,
		// TODO: support GIT_INDEX_FILE
		indexFile: filepath.Join(gitDir, "index"),
	}, nil
}

// Dir returns the path to the root of the working directory tree of r.
func (r *Repo) Dir() string { return r.workTree }

// Returns the commit that HEAD ultimately references.
func (r *Repo) Head() (*Commit, error) {
	return r.resolveRefFile(filepath.Join(r.gitDir, "HEAD"))
}

// Ref resolves some ref name and returns the corresponding Commit, if it exists. It tries, in preference
// order:
//
//   - Full SHA-1 hashes (if name is 40 bytes of hex)
//   - Unique SHA-1 prefixes (if name is 4 <= N < 40 bytes of hex)
//   - Fully qualified ref names ("heads/master")
//   - Tag names ("v1.2.3")
//   - Unqualified head name ("master")
//   - Unqualified remote name ("origin/master")
func (r *Repo) Ref(name string) (*Commit, error) {
	if maybeHex(name) {
		if len(name) == 40 {
			return r.resolveSHA(name)
		}
		if len(name) < 40 && len(name) >= 4 {
			//if c, err := r.resolveSHAPrefix(name); err == nil {
			//return c, nil
			//}
		}
	}

	for _, path := range []string{
		filepath.Join(r.gitDir, "refs", name),
		filepath.Join(r.gitDir, "refs", "tags", name),
		filepath.Join(r.gitDir, "refs", "heads", name),
		filepath.Join(r.gitDir, "refs", "remotes", name),
	} {
		c, err := r.resolveRefFile(path)
		if err == nil {
			return c, nil
		}
		if !os.IsNotExist(err) {
			return c, err
		}
		// TODO: check for path in packed-refs
	}
	return nil, fmt.Errorf("bad ref name: %s", name)
}

func maybeHex(s string) bool {
	for i := range s {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}

func (r *Repo) resolveSHAPrefix(prefix string) (*Commit, error) {
	panic("unimplemented")
}

func (r *Repo) resolveSHA(ref string) (*Commit, error) {
	// TODO: locate in object store
	sha, err := makeSHA(ref)
	if err != nil {
		return nil, err
	}
	return &Commit{
		sha: sha,
	}, nil
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
