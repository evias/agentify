package api

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// WriteAgentsFile renders and writes the AGENTS.md content.
func WriteAgentsFile(dest string, rs *RepoScan) error {
	if rs.ExistingMD {
		return fmt.Errorf("AGENTS.md already exists in %s; remove it or rename it first", rs.Root)
	}
	content := BuildAgentsContent(rs)
	return ioutil.WriteFile(dest, []byte(content), 0644)
}

// BuildAgentsContent constructs the markdown body.
func BuildAgentsContent(rs *RepoScan) string {
	sb := &strings.Builder{}
	sb.WriteString("# AGENTS\n\n")

	if rs.Readme != "" {
		sb.WriteString("## Overview\n\n")
		// grab first paragraph of README
		paragraphs := strings.SplitN(rs.Readme, "\n\n", 2)
		sb.WriteString(paragraphs[0])
		sb.WriteString("\n\n")
	}

	if len(rs.CmdPackages) > 0 {
		sb.WriteString("## Commands (cmd/*)\n\n")
		for _, pkg := range rs.CmdPackages {
			sb.WriteString(fmt.Sprintf("- **%s**: CLI tool under `%s`\n", filepathBase(pkg), pkg))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("*(no cmd/* packages found)*\n\n")
	}

	return sb.String()
}

// filepathBase returns the last element of a slash-separated path.
func filepathBase(p string) string {
	parts := strings.Split(p, "/")
	return parts[len(parts)-1]
}
