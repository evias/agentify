# AGENTS

## Overview
### agentify - make your work accessible to AI!

A vibe-coded software that creates `AGENTS.md` files for your repositories. This software was initially developed by Genesìs, a Large Language Model (LLM) based on the `codex-mini` model, instructed by re:Software S.L., and subsequently modified by re:Software S.L. to meet security and quality standards. The project aims to enhance the accessibility of code for AI models by providing a structured documentation format.

## Commands
- `agentify [flags] <repo>`: Generates an `AGENTS.md` file for a Git repository.

## Library
- `api.PrepareRepository`: Handles cloning a Git repository or opening a folder.
- `api.ScanRepository`: Scans a Git repository for commands and libraries.
- `api.WriteAgentsFile`: Generates the final `AGENTS.md` file.

## Examples
- `agentify . -o AGENTS.md`
- `agentify github.com/evias/agentify -o AGENTS.md`
- `agentify github.com/evias/dotsiig -o ~/AGENTS.md`
