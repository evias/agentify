package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/prathyushnallamothu/ollamago"
	"github.com/spf13/cobra"

	"github.com/evias/agentify/api"
)

const (
	defaultModelName = "llama3.2:latest"
	defaultTimeout   = 5 * time.Minute
)

func main() {
	var (
		outputPath string
		useOllama  bool
		modelName  string
		hostURL    string
		promptFile string
	)

	rootCmd := &cobra.Command{
		Use:   "agentify [flags] <repo>",
		Short: "Generate an AGENTS.md file for a Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArg := args[0]

			// If --ollama but no model provided, fallback:
			if useOllama && modelName == "" {
				modelName = defaultModelName
			}

			if useOllama && hostURL == "" {
				hostURL = "http://127.0.0.1:11434"
			}

			var client *ollamago.Client
			if useOllama {
				client = ollamago.NewClient(
					ollamago.WithBaseURL(hostURL),
					ollamago.WithTimeout(defaultTimeout),
				)
			}

			// 1) Clone or open
			workdir, cleanup, err := api.PrepareRepository(repoArg)
			if err != nil {
				return err
			}
			defer cleanup()

			// 2) Scan
			summary, err := api.ScanRepository(api.ScanOptions{
				Root:       workdir,
				UseOllama:  useOllama,
				Ollama:     client,
				Model:      modelName,
				PromptFile: promptFile,
			})
			if err != nil {
				return err
			}

			// 3) Write AGENTS.md
			dest := filepath.Join(workdir, "AGENTS.md")
			if outputPath != "" {
				dest = outputPath
			}
			if err := api.WriteAgentsFile(dest, summary); err != nil {
				return err
			}

			fmt.Printf("Wrote AGENTS.md → %s\n", dest)
			return nil
		},
	}

	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "destination path for AGENTS.md")
	rootCmd.Flags().BoolVar(&useOllama, "ollama", false, "enable Ollama LLM scanning")
	rootCmd.Flags().StringVarP(&modelName, "model", "m", "", "name of the Ollama model (default llama3.2:latest)")
	rootCmd.Flags().StringVarP(&hostURL, "host", "s", "http://127.0.0.1:11434", "Ollama server base URL")
	rootCmd.Flags().StringVarP(&promptFile, "prompt", "p", "", "path to custom scan prompt (markdown)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
