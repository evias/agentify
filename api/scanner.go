package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/prathyushnallamothu/ollamago"
)

// ScanOptions configures ScanRepository behavior.
type ScanOptions struct {
	Root       string // repository root directory
	Model      string // model name (llama3.2:latest default)
	PromptFile string // path to prompt markdown (default prompts/scan.default.md)
}

// ScanRepository inspects the worktree to find cmd packages, README.md, and AGENTS.md.
func ScanRepository(ollama *ollamago.Client, opts ScanOptions) (*RepoScan, error) {
	defaultPromptFile := filepath.Join("prompts", "scan.default.md")
	defaultAgentsFile := filepath.Join("prompts", "agent.example.md")

	var rs RepoScan
	rs.Root = opts.Root

	// load prompt template
	pf := opts.PromptFile
	if pf == "" {
		pf = defaultPromptFile
	}
	tmplBytes, err := ioutil.ReadFile(pf)
	if err != nil {
		return nil, fmt.Errorf("reading prompt template: %w", err)
	}

	// load AGENTS.md template
	af := defaultAgentsFile
	agt, err := ioutil.ReadFile(af)
	if err != nil {
		return nil, fmt.Errorf("reading AGENTS.md template: %w", err)
	}

	// read the repo README
	readmePath := filepath.Join(rs.Root, "README.md")
	rd, err := ioutil.ReadFile(readmePath)
	if err != nil {
		return nil, fmt.Errorf("reading README.md for LLM scan: %w", err)
	}

	// substitute the actual README contents into the prompt
	rs.Readme = string(rd)
	prompt := strings.Replace(string(tmplBytes), "<ATTACH_README_HERE>", string(rd), 1)
	prompt = strings.Replace(prompt, "<ATTACH_EXAMPLE_AGENTS>", string(agt), 1)

	// call ollama
	resp, err := ollama.Generate(context.Background(), ollamago.GenerateRequest{
		Model:  opts.Model,
		Prompt: prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	rs.Result = resp.Response
	return &rs, nil
}
