package api

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/prathyushnallamothu/ollamago"
)

// ScanOptions configures ScanRepository behavior.
type ScanOptions struct {
	Root       string           // repository root directory
	UseOllama  bool             // enable LLM scanning
	Ollama     *ollamago.Client // ~nil if not using LLM
	Model      string           // model name (llama3.2:latest default)
	PromptFile string           // path to prompt markdown (default prompts/scan.default.md)
}

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
	Root       string // local filesystem path
	Readme     string // raw README.md contents (if found)
	ExistingMD bool   // whether AGENTS.md already exists

	CmdPackages []string      // paths (relative to Root) of cmd/... subdirs
	CmdInfos    []CommandInfo // all Use/Short pairs found in cmd/ packages
	DevSections []DevSection  // extracted developer‐focused markdown sections

	MarkdownResult string
}

// ScanRepository inspects the worktree to find cmd packages, README.md, and AGENTS.md.
func ScanRepository(opts ScanOptions) (*RepoScan, error) {
	defaultPromptFile := filepath.Join("prompts", "scan.default.md")
	defaultAgentsFile := filepath.Join("prompts", "agent.example.md")

	var rs RepoScan
	rs.Root = opts.Root

	// 4) Detect existing AGENTS.md
	if _, err := os.Stat(filepath.Join(rs.Root, "AGENTS.md")); err == nil {
		rs.ExistingMD = true
	}

	// if LLM scanning requested, call ollama instead of standard scan
	if opts.UseOllama {
		// load prompt template
		pf := opts.PromptFile
		if pf == "" {
			pf = defaultPromptFile
		}
		tmplBytes, err := ioutil.ReadFile(pf)
		if err != nil {
			return nil, fmt.Errorf("reading prompt template: %w", err)
		}

		// read the repo README
		readmePath := filepath.Join(rs.Root, "README.md")
		rd, err := ioutil.ReadFile(readmePath)
		if err != nil {
			return nil, fmt.Errorf("reading README.md for LLM scan: %w", err)
		}

		// load AGENTS.md template
		af := defaultAgentsFile
		agt, err := ioutil.ReadFile(af)
		if err != nil {
			return nil, fmt.Errorf("reading AGENTS.md template: %w", err)
		}

		// substitute the actual README contents into the prompt
		rs.Readme = string(rd)
		prompt := strings.Replace(string(tmplBytes), "<ATTACH_README_HERE>", string(rd), 1)
		prompt = strings.Replace(prompt, "<ATTACH_EXAMPLE_AGENTS>", string(agt), 1)

		// call ollama
		resp, err := opts.Ollama.Generate(context.Background(), ollamago.GenerateRequest{
			Model:  opts.Model,
			Prompt: prompt,
		})
		if err != nil {
			return nil, fmt.Errorf("LLM generation failed: %w", err)
		}

		rs.MarkdownResult = resp.Response
		return &rs, nil
	}

	// --- fallback to the previous scan behavior ---
	manualScanGolang(&rs)

	return &rs, nil
}

// manualScanGolang scans a repository *manually* (not using LLM).
func manualScanGolang(rs *RepoScan) {
	// (1) Look for cmd/* directories
	cmdRoot := filepath.Join(rs.Root, "cmd")
	entries, err := ioutil.ReadDir(cmdRoot)
	if err == nil {
		for _, fi := range entries {
			if fi.IsDir() {
				pkg := filepath.Join("cmd", fi.Name())
				rs.CmdPackages = append(rs.CmdPackages, filepath.Join("cmd", fi.Name()))
				scanCobraCommands(filepath.Join(rs.Root, pkg), &rs.CmdInfos)
			}
		}
	}

	// (2) Read README.md if present
	readmePath := filepath.Join(rs.Root, "README.md")
	if data, err := ioutil.ReadFile(readmePath); err == nil {
		rs.Readme = string(data)
		rs.DevSections = append(rs.DevSections, extractDevSections(rs.Readme)...)
	}

	// (3) Also read any *.md under /docs or /doc or root (except README & AGENTS)
	mdPaths := []string{}
	// root-level files
	if rootEnts, _ := ioutil.ReadDir(rs.Root); rootEnts != nil {
		for _, fi := range rootEnts {
			if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".md") &&
				fi.Name() != "README.md" && fi.Name() != "AGENTS.md" {
				mdPaths = append(mdPaths, filepath.Join(rs.Root, fi.Name()))
			}
		}
	}
	// docs/doc subfolders
	for _, dn := range []string{"docs", "doc"} {
		if docEnts, _ := ioutil.ReadDir(filepath.Join(rs.Root, dn)); docEnts != nil {
			for _, fi := range docEnts {
				if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".md") {
					mdPaths = append(mdPaths, filepath.Join(rs.Root, dn, fi.Name()))
				}
			}
		}
	}
	for _, mdp := range mdPaths {
		if data, err := ioutil.ReadFile(mdp); err == nil {
			rs.DevSections = append(rs.DevSections, extractDevSections(string(data))...)
		}
	}
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
			lvl := 0
			for lvl < len(h) && h[lvl] == '#' {
				lvl++
			}

			title := "Instructions"
			if lvl < len(h) {
				title = strings.TrimSpace(h[lvl:])
			}

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
