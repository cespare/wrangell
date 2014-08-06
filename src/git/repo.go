package git

import (
	"fmt"
	"os"
	"path/filepath"
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

	paths := []string{
		"refs/" + name,
		"refs/tags/" + name,
		"refs/heads/" + name,
		"refs/remotes/" + name,
	}
	for _, path := range paths {
		c, err := r.resolveRefFile(filepath.Join(r.gitDir, path))
		if err == nil {
			return c, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	for _, path := range paths {
		c, err := r.resolvePackedRef(path)
		if err == nil {
			return c, nil
		}
		if !(os.IsNotExist(err) || err == packedRefNotFoundErr) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("bad ref name: %s", name)
}

func maybeHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
