package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/evias/agentify/api"
)

func main() {
	var outputPath string

	rootCmd := &cobra.Command{
		Use:   "agentify [flags] <repo>",
		Short: "Generate an AGENTS.md file for a Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoArg := args[0]

			// 1) Clone or open
			workdir, cleanup, err := api.PrepareRepository(repoArg)
			if err != nil {
				return err
			}
			defer cleanup()

			// 2) Scan
			summary, err := api.ScanRepository(workdir)
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
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
