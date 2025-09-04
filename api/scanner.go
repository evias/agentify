package api

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
)

type RepoScan struct {
	Root        string   // local filesystem path
	CmdPackages []string // paths (relative to Root) of cmd/... subdirs
	Readme      string   // contents of README.md (if found)
	ExistingMD  bool     // whether AGENTS.md already exists
}

// PrepareRepository clones the repoArg (URL or local path) into a temp dir.
// Returns the local path, a cleanup func, and any error.
func PrepareRepository(repoArg string) (string, func(), error) {
	// If it's a local directory, just use it in place (no cleanup).
	if fi, err := os.Stat(repoArg); err == nil && fi.IsDir() {
		return repoArg, func() {}, nil
	}

	// Otherwise assume it's a remote and clone into a temp dir.
	tmp, err := ioutil.TempDir("", "agents-gen-*")
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

// ScanRepository inspects the worktree to find cmd packages, README.md, and AGENTS.md.
func ScanRepository(root string) (*RepoScan, error) {
	var rs RepoScan
	rs.Root = root

	// 1) Look for cmd/* directories
	cmdRoot := filepath.Join(root, "cmd")
	entries, err := ioutil.ReadDir(cmdRoot)
	if err == nil {
		for _, fi := range entries {
			if fi.IsDir() {
				rs.CmdPackages = append(rs.CmdPackages, filepath.Join("cmd", fi.Name()))
			}
		}
	}

	// 2) Look for README.md
	readmePath := filepath.Join(root, "README.md")
	if data, err := ioutil.ReadFile(readmePath); err == nil {
		rs.Readme = string(data)
	}

	// 3) Detect existing AGENTS.md
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err == nil {
		rs.ExistingMD = true
	}

	return &rs, nil
}
