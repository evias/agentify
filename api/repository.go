package api

import (
	"fmt"
	"io/ioutil"
	"os"

	git "github.com/go-git/go-git/v5"
)

// PrepareRepository clones the repoArg (URL or local path) into a temp dir.
// Returns the local path, a cleanup func, and any error.
func PrepareRepository(repoArg string) (string, func(), error) {
	// If it's a local directory, just use it in place (no cleanup).
	if fi, err := os.Stat(repoArg); err == nil && fi.IsDir() {
		return repoArg, func() {}, nil
	}

	// Otherwise assume it's a remote and clone into a temp dir.
	tmp, err := ioutil.TempDir("", "agentify-*")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { os.RemoveAll(tmp) }

	if _, err := git.PlainClone(tmp, false, &git.CloneOptions{
		URL:      repoArg,
		Progress: os.Stdout,
	}); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("could not clone %s: %w", repoArg, err)
	}

	return tmp, cleanup, nil
}
