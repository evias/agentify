 Title: agentify

## Overview

### Agentify - Facilitate AI access to your repositories

Agentify is a tool, coded by vibe, designed to generate `AGENTS.md` files for Git repositories. Originally developed by Genesìs, a LLM model based on the `codex-mini` model, Agentify was later refined by re:Software S.L to ensure security and quality standards.

#### Build instructions

1. Install Go (https://golang.org/doc/install)
2. Clone this repository: `go get -u github.com/evias/agentify`
3. Run the command: `go run agentify.go [flags] <repo>`

#### Usage instructions

1. Navigate to the cloned repository directory
2. Execute Agentify with the desired flags and repository path: `./agentify [flags] <repo>` or `go run agentify.go [flags] <repo>`
3. The generated `AGENTS.md` file will be placed in the current working directory

## Library

- `api.PrepareRepository`: Clones a Git repository, or opens an existing local one.
- `api.ScanRepository`: Scans a Git repository for commands and libraries.
- `api.WriteAgentsFile`: Writes the resulting `AGENTS.md` file.

## Examples

- `agentify . -o AGENTS.md` generates an `AGENTS.md` file in the current working directory
- `agentify github.com/evias/agentify -o AGENTS.md` generates an `AGENTS.md` file for the agentify GitHub repository, placing it in the current working directory
- `agentify github.com/evias/dotsig -o ~/AGENTS.md` generates an `AGENTS.md` file for the dotsig GitHub repository and saves it in your home directory (~/)