package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Utilities for running git directly

var gitBin string

func init() {
	var err error
	gitBin, err = exec.LookPath("git")
	if err != nil {
		fatalf("Cannot locate git binary.\n")
	}
	// Sanity checks
	stdout, stderr, err := gitExec("/", "version")
	if err != nil {
		fatalf("Cannot invoke git: %s\n%s", err, stderr)
	}
	version := strings.TrimSpace(strings.TrimPrefix(string(stdout), "git version "))
	if !strings.HasPrefix(version, "2.") {
		fatalf("Invalid git version %s; need at least 2.0.0\n", version)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

// exec runs some git command in the given dir.
func gitExec(dir string, args ...string) (stdout, stderr []byte, err error) {
	cmd := exec.Command(gitBin, args...)
	var outBuf, errBuf bytes.Buffer
	cmd.Dir = dir
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}

// exec runs some git command in the repo root.
func (r *Repo) exec(args ...string) (stdout, stderr []byte, err error) {
	return gitExec(r.gitDir, args...)
}
