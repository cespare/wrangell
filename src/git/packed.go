package git

import (
	"os"
)

// Functions for reading git's pack format.
// References:
// https://www.kernel.org/pub/software/scm/git/docs/technical/pack-format.txt
// http://schacon.github.io/gitbook/7_the_packfile.html

func (r *Repo) packedObjectBySHA(sha *SHA) (*Object, error) {
	// TODO: Handle unindexed packfiles (maybe by getting get to fix them first). (This is an unexpected case
	// though.)
	return nil, os.ErrNotExist
}
