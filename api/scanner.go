package api

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	git "github.com/go-git/go-git/v5"
)

// CommandInfo holds a single cobra.Command literal’s Use/Short.
type CommandInfo struct {
	Use   string
	Short string
}

// DevSection is a slice of lines for a developer‐oriented markdown subsection.
type DevSection struct {
	Title string   // the header line, e.g. "### Build instructions"
	Lines []string // all raw lines under that header until the next header of same-or-higher level
}

type RepoScan struct {
	Root        string        // local filesystem path
	CmdPackages []string      // paths (relative to Root) of cmd/... subdirs
	Readme      string        // raw README.md contents (if found)
	ExistingMD  bool          // whether AGENTS.md already exists
	CmdInfos    []CommandInfo // all Use/Short pairs found in cmd/ packages
	DevSections []DevSection  // extracted developer‐focused markdown sections
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

	// (1) Look for cmd/* directories
	cmdRoot := filepath.Join(root, "cmd")
	entries, err := ioutil.ReadDir(cmdRoot)
	if err == nil {
		for _, fi := range entries {
			if fi.IsDir() {
				pkg := filepath.Join("cmd", fi.Name())
				rs.CmdPackages = append(rs.CmdPackages, filepath.Join("cmd", fi.Name()))
				scanCobraCommands(filepath.Join(root, pkg), &rs.CmdInfos)
			}
		}
	}

	// (2) Read README.md if present
	readmePath := filepath.Join(root, "README.md")
	if data, err := ioutil.ReadFile(readmePath); err == nil {
		rs.Readme = string(data)
		rs.DevSections = append(rs.DevSections, extractDevSections(rs.Readme)...)
	}

	// (3) Also read any *.md under /docs or /doc or root (except README & AGENTS)
	mdPaths := []string{}
	// root-level files
	if rootEnts, _ := ioutil.ReadDir(root); rootEnts != nil {
		for _, fi := range rootEnts {
			if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".md") &&
				fi.Name() != "README.md" && fi.Name() != "AGENTS.md" {
				mdPaths = append(mdPaths, filepath.Join(root, fi.Name()))
			}
		}
	}
	// docs/doc subfolders
	for _, dn := range []string{"docs", "doc"} {
		if docEnts, _ := ioutil.ReadDir(filepath.Join(root, dn)); docEnts != nil {
			for _, fi := range docEnts {
				if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".md") {
					mdPaths = append(mdPaths, filepath.Join(root, dn, fi.Name()))
				}
			}
		}
	}
	for _, mdp := range mdPaths {
		if data, err := ioutil.ReadFile(mdp); err == nil {
			rs.DevSections = append(rs.DevSections, extractDevSections(string(data))...)
		}
	}

	// 4) Detect existing AGENTS.md
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err == nil {
		rs.ExistingMD = true
	}

	return &rs, nil
}

// scanCobraCommands parses all .go files in pkgDir and appends any cobra.Command{Use:,Short:} to dst.
func scanCobraCommands(pkgDir string, dst *[]CommandInfo) {
	fs, err := ioutil.ReadDir(pkgDir)
	if err != nil {
		return
	}
	// very simple regex-based extraction (could use AST for more robustness)
	reCmd := regexp.MustCompile(`&cobra\.Command\s*\{([^}]*)\}`)
	reUse := regexp.MustCompile(`Use\s*:\s*"([^"]+)"`)
	reShort := regexp.MustCompile(`Short\s*:\s*"([^"]+)"`)

	for _, fi := range fs {
		if fi.IsDir() || !strings.HasSuffix(fi.Name(), ".go") {
			continue
		}
		bytes, err := ioutil.ReadFile(filepath.Join(pkgDir, fi.Name()))
		if err != nil {
			continue
		}
		src := string(bytes)
		for _, block := range reCmd.FindAllStringSubmatch(src, -1) {
			body := block[1]
			useM := reUse.FindStringSubmatch(body)
			shortM := reShort.FindStringSubmatch(body)
			if len(useM) > 1 && len(shortM) > 1 {
				*dst = append(*dst, CommandInfo{Use: useM[1], Short: shortM[1]})
			}
		}
	}
}

// extractDevSections pulls out markdown sections whose header matches developer keywords.
func extractDevSections(md string) []DevSection {
	var out []DevSection
	scanner := bufio.NewScanner(strings.NewReader(md))
	// keywords to match in header (case-insensitive)
	kws := []string{
		"dev notes",
		"developer instructions",
		"dev environment",
		"setup this package locally",
		"how to build",
		"build instructions",
		"how to contribute",
	}
	var current *DevSection
	var currentLevel int

	for scanner.Scan() {
		line := scanner.Text()
		if h := strings.TrimLeft(line, " "); strings.HasPrefix(h, "#") {
			// count heading level
			lvl := strings.IndexFunc(h, func(r rune) bool { return r != '#' })
			title := strings.TrimSpace(h[lvl:])
			// new section?
			found := false
			lowt := strings.ToLower(title)
			for _, kw := range kws {
				if strings.Contains(lowt, kw) {
					found = true
					break
				}
			}
			if found {
				// push prior
				if current != nil {
					out = append(out, *current)
				}
				current = &DevSection{Title: h, Lines: []string{}}
				currentLevel = lvl
				continue
			}
			// any header of same-or-higher level ends section
			if current != nil && lvl <= currentLevel {
				out = append(out, *current)
				current = nil
			}
		}
		if current != nil {
			current.Lines = append(current.Lines, line)
		}
	}
	if current != nil {
		out = append(out, *current)
	}
	return out
}
