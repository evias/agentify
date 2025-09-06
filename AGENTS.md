# AGENTS

## Overview

### agentify - make your work accessible to AI!

A vibe-coded software that creates AGENTS.md files for your repositories. This software was originally built by Genesìs, a LLM model based on the `codex-mini` model, instructed by re:Software S.L, and later modified by re:Software S.L. to satisfy security and general quality standards.

### Build instructions

The `agentify` tool is executed via the command line.

### Usage instructions

```bash
Usage:
  agentify [flags] <repo>

Flags:
  -h, --help            help for agentify
  -s, --host string     Ollama server base URL (default "http://127.0.0.1:11434")
  -m, --model string    name of the Ollama model (default llama3.2:latest)
  -o, --output string   destination path for AGENTS.md
  -p, --prompt string   path to custom scan prompt (markdown)
  -y, --yes             overwrite existing AGENTS.md file
```

### Examples

`agentify . -o AGENTS.md`
`agentify github.com/evias/agentify -o AGENTS.md`
`agentify github.com/evias/dotsig -o AGENTS.md -m llama3.2:latest`
`agentify . -o AGENTS.md -m gemma3:4b -s http://127.0.0.1:11434`
