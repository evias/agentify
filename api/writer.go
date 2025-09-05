package api

import (
	"fmt"
	"io/ioutil"
	"regexp"
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

	// MarkdownResult is set when a LLM generates the AGENTS.md content.
	if len(rs.MarkdownResult) > 0 {
		sb.WriteString(rs.MarkdownResult)
		return sb.String()
	}

	sb.WriteString("# AGENTS\n\n")

	// (1) Extracts the README.md content for an Overview
	if rs.Readme != "" {
		sb.WriteString("## Overview\n\n")
		// Prepare regexes to strip inline and reference‑style links/images
		reInline := regexp.MustCompile(`!?\[([^\]]+)\]\([^)]+\)`)
		reRef := regexp.MustCompile(`!?\[([^\]]+)\]\[[^\]]+\]`)

		lines := strings.Split(rs.Readme, "\n")
		for i, l := range lines {
			// stop at first secondary/tertiary heading
			if strings.HasPrefix(l, "## ") || strings.HasPrefix(l, "### ") {
				break
			}

			// flatten/remove any markdown links/images, leaving only the link text
			l = reInline.ReplaceAllString(l, "$1")
			l = reRef.ReplaceAllString(l, "$1")
			l = strings.TrimSpace(l)

			// drop lines that are only links/images
			if strings.HasPrefix(l, "[") || strings.HasPrefix(l, "![") {
				continue
			}

			// promote main title to a third-level heading (### )
			if strings.HasPrefix(l, "# ") {
				sb.WriteString("### " + strings.TrimPrefix(l, "# ") + "\n")
			} else {
				sb.WriteString(l + "\n")
			}
			// if next line is blank and next is "##", we'll still break on next iter.
			_ = i
		}
		sb.WriteString("\n")
	}

	// (2) List cobra commands found under cmd/*
	// TODO: should find bin/, dist/, and/or usage instructions.
	if len(rs.CmdInfos) > 0 {
		sb.WriteString("## Commands (cmd/*)\n\n")
		for _, ci := range rs.CmdInfos {
			sb.WriteString(fmt.Sprintf("- `%s`: %s\n", ci.Use, ci.Short))
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("*(no cmd/* packages found)*\n\n")
	}

	if len(rs.DevSections) > 0 {
		sb.WriteString("## Developer Setup\n\n")
		for _, ds := range rs.DevSections {
			sb.WriteString(ds.Title + "\n\n")
			// decide filtering: build‑related vs deps‑related vs other
			tlow := strings.ToLower(ds.Title)
			keepCode := strings.Contains(tlow, "build")
			keepList := !keepCode
			inCodeBlk := false
			for _, ln := range ds.Lines {
				if strings.HasPrefix(ln, "```") {
					inCodeBlk = !inCodeBlk
					if keepCode {
						sb.WriteString(ln + "\n")
					}
					continue
				}
				if inCodeBlk && keepCode {
					sb.WriteString(ln + "\n")
					continue
				}
				if keepList && strings.HasPrefix(strings.TrimSpace(ln), "- ") {
					sb.WriteString(ln + "\n")
				}
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// filepathBase returns the last element of a slash-separated path.
func filepathBase(p string) string {
	parts := strings.Split(p, "/")
	return parts[len(parts)-1]
}
