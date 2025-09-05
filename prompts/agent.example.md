# AGENTS

## Overview

### agentify - make your work accessible to AI!

A vibe-coded software that creates AGENTS.md files for your repositories.

This software was originally built by Genesìs, a LLM model based on
the `codex-mini` model, instructed by re:Software S.L, and later modified
by re:Software S.L. to satisfy security and general quality standards.

## Commands

- `agentify [flags] <repo>`: Generate an AGENTS.md file for a Git repository

## Library

- `api.PrepareRepository`: Clones a git repository or opens a folder.
- `api.ScanRepository`: Scans a git repository for commands and libraries.
- `api.WriteAgentsFile`: Produces the resulting `AGENTS.md` file.

## Examples

`agentify . -o AGENTS.md`
`agentify github.com/evias/agentify -o AGENTS.md`
`agentify github.com/evias/dotsig -o ~/AGENTS.md`
