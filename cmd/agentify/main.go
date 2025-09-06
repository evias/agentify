package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
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
		modelName  string
		hostURL    string
		promptFile string
		overwrite  bool
	)

	rootCmd := &cobra.Command{
		Use:   "agentify [flags] <repo>",
		Short: "Generate an AGENTS.md file for a Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArg := args[0]

			// If no model provided, fallback to llama3.2:latest
			if modelName == "" {
				modelName = defaultModelName
			}

			if hostURL == "" {
				hostURL = "http://127.0.0.1:11434"
			}

			var client *ollamago.Client
			client = ollamago.NewClient(
				ollamago.WithBaseURL(hostURL),
				ollamago.WithTimeout(defaultTimeout),
			)

			fmt.Println("Opening repository...")

			// 1) Clone or open
			workdir, cleanup, err := api.PrepareRepository(repoArg)
			if err != nil {
				return err
			}
			defer cleanup()

			// Detect existing AGENTS.md and ask to overwrite
			if _, err := os.Stat(filepath.Join(workdir, "AGENTS.md")); err == nil {
				if !overwrite {
					fmt.Printf("An existing AGENTS.md was found, do you want to overwrite it? [y/N]: ")

					var yes string
					fmt.Scanf("%s", &yes)
					if !slices.Contains([]string{"y", "Y", "yes", "YES"}, yes) {
						return fmt.Errorf("An existing AGENTS.md file was found, please rename it first.")
					}
				}
			}

			fmt.Println("Scanning repository and instructing LLM...")

			// 2) Scan
			summary, err := api.ScanRepository(client, api.ScanOptions{
				Root:       workdir,
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
	rootCmd.Flags().StringVarP(&hostURL, "host", "s", "http://127.0.0.1:11434", "Ollama server base URL")
	rootCmd.Flags().StringVarP(&modelName, "model", "m", "", "name of the Ollama model (default llama3.2:latest)")
	rootCmd.Flags().StringVarP(&promptFile, "prompt", "p", "", "path to custom scan prompt (markdown)")
	rootCmd.Flags().BoolVarP(&overwrite, "yes", "y", false, "overwrite existing AGENTS.md file")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
