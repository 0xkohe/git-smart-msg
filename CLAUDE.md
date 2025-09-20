# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`git-smartmsg` is a Go CLI tool that uses OpenAI's API to generate improved commit messages for Git repositories. It operates in two phases:

1. **Plan phase**: Analyzes commit history and generates AI-suggested commit messages, saving them to `plan.json`
2. **Apply phase**: Creates a new branch with rewritten commit history using the improved messages

## Core Architecture

- **Single-file Go application** (`main.go`) with clear separation of concerns
- **OpenAI Integration**: Uses OpenAI SDK v2 with configurable API base and model
- **Git Operations**: Direct `git` command execution via `exec.Command`
- **Plan Storage**: JSON-based plan file format for staging commit message improvements

Key components:
- `AIClient` interface with `OpenAIClient` implementation
- `Plan` and `PlanItem` structs for managing commit rewriting plans
- Git helper functions for repository operations
- Commit metadata extraction and diff generation

## Environment Variables

Required:
- `OPENAI_API_KEY`: OpenAI API authentication key

Optional:
- `OPENAI_API_BASE`: Custom API endpoint (for OpenAI-compatible services)
- `OPENAI_MODEL`: Model to use (defaults to "gpt-5-nano")

## Common Commands

### Development
```bash
# Build the application
go build -o git-smartmsg main.go

# Run directly
go run main.go plan --limit 10
go run main.go apply --branch rewrite/improved-messages
```

### Usage Examples
```bash
# Generate AI commit messages for last 20 commits
./git-smartmsg plan --limit 20 --model gpt-4o

# Generate commit messages with emoji prefixes
./git-smartmsg plan --emoji --limit 10

# Apply with custom range
./git-smartmsg plan --range HEAD~10..HEAD

# Apply the plan to a new branch
./git-smartmsg apply --branch improved-history --in plan.json

# Include merge commits (not recommended)
./git-smartmsg plan --allow-merges
./git-smartmsg apply --allow-merges --branch with-merges
```

## Important Implementation Details

### Commit Message Processing
- Uses Conventional Commits style when appropriate (normal mode)
- Emoji mode: Uses present tense, imperative mood with contextual emoji prefixes
- Sanitizes messages to remove markdown formatting artifacts
- Limits first line to ~72 characters
- Preserves original author information and timestamps

### Emoji Mode
When `--emoji` flag is used, commit messages start with contextual emojis:
- 🎨 `:art:` when improving code structure/format
- 🐛 `:bug:` when fixing bugs
- 📝 `:memo:` when writing docs
- ✅ `:white_check_mark:` when adding tests
- 🔒 `:lock:` when dealing with security
- ⬆️ `:arrow_up:` when upgrading dependencies
- And more based on the change context

### Safety Features
- Requires clean worktree before applying changes
- Creates new branches for rewritten history (never modifies current branch)
- Skips merge commits by default (linear history preference)
- Uses cherry-pick with `--no-verify` to avoid hooks during rewriting

### Git Operations
- All git commands use `exec.Command` with proper error handling
- Preserves author name, email, and date during commit rewriting
- Uses `--force-with-lease` recommendation for safe force-pushing

## File Structure
- `main.go`: Complete application logic
- `plan.json`: Generated plan file (git-ignored, working file)
- `go.mod/go.sum`: Go module dependencies (OpenAI SDK v2)