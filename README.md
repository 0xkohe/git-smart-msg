# git-smartmsg

🤖 AI-powered Git commit message improver using OpenAI API

**git-smartmsg** is a command-line tool that analyzes your Git commit history and generates improved commit messages using AI. It operates in two phases: planning (analyzing commits and generating suggestions) and applying (rewriting history with improved messages on a new branch).

[日本語版 README はこちら](./README.ja.md) | [Japanese README here](./README.ja.md)

## Features

- 🎯 **AI-Powered**: Uses OpenAI API to generate meaningful commit messages
- 🔒 **Safe by Design**: Creates new branches, never modifies your current branch
- 📋 **Two-Phase Process**: Plan first, review, then apply
- 🎨 **Emoji Mode**: Optional emoji prefixes for visual commit categorization
- ⚡ **Conventional Commits**: Supports conventional commit format
- 🛡️ **History Preservation**: Maintains author information and timestamps
- 🔄 **Flexible Range**: Process specific commit ranges or recent commits

## Installation

### Prerequisites

- Go 1.25+
- Git
- OpenAI API Key

### Build from Source

```bash
git clone https://github.com/yourusername/git-smartmsg
cd git-smartmsg
go build -o git-smartmsg main.go
```

### Add to PATH (Optional)

```bash
# Move to a directory in your PATH
sudo mv git-smartmsg /usr/local/bin/
# Or create a symlink
ln -s $(pwd)/git-smartmsg /usr/local/bin/git-smartmsg
```

## Environment Variables

### Required

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

### Optional

```bash
# Custom API endpoint (for OpenAI-compatible services)
export OPENAI_API_BASE="https://api.openai.com/v1"

# Default model (defaults to gpt-5-nano)
export OPENAI_MODEL="gpt-4o"
```

## Quick Start

1. **Navigate to your Git repository**
   ```bash
   cd your-git-repo
   ```

2. **Generate improved commit messages**
   ```bash
   ./git-smartmsg plan --limit 5
   ```

3. **Review the generated plan**
   ```bash
   cat plan.json
   ```

4. **Apply the improved messages to a new branch**
   ```bash
   ./git-smartmsg apply --branch improved-messages
   ```

## Usage

### Command Overview

```bash
git-smartmsg <subcommand> [options]
```

### Subcommands

#### `plan` - Generate AI commit messages

```bash
git-smartmsg plan [options]
```

**Options:**
- `--limit <n>`: Number of commits from HEAD to include (default: 20)
- `--range <range>`: Explicit git range (e.g., `HEAD~10..HEAD`)
- `--model <model>`: LLM model to use (default: from env or `gpt-5-nano`)
- `--emoji`: Use emoji-style commit messages
- `--allow-merges`: Include merge commits (not recommended)
- `--out <file>`: Output plan file (default: `plan.json`)
- `--timeout <duration>`: Per-commit AI timeout (default: 25s)

#### `apply` - Apply plan to new branch

```bash
git-smartmsg apply [options]
```

**Options:**
- `--branch <name>`: New branch name (required)
- `--in <file>`: Plan file path (default: `plan.json`)
- `--allow-merges`: Attempt to preserve merge commits (experimental)

## Examples

### Basic Usage

```bash
# Improve last 10 commits
./git-smartmsg plan --limit 10

# Use specific model
./git-smartmsg plan --limit 5 --model gpt-4o

# Apply to new branch
./git-smartmsg apply --branch feature/improved-commits
```

### Advanced Usage

```bash
# Process specific range
./git-smartmsg plan --range v1.0.0..HEAD

# Use emoji mode
./git-smartmsg plan --emoji --limit 15

# Include merge commits (experimental)
./git-smartmsg plan --allow-merges --limit 20
./git-smartmsg apply --allow-merges --branch with-merges
```

### Workflow Example

```bash
# 1. Check what commits you want to improve
git log --oneline -10

# 2. Generate plan with emoji mode
./git-smartmsg plan --emoji --limit 10

# 3. Review the suggestions
cat plan.json | jq '.items[] | {old: .old_message, new: .new_message}'

# 4. Apply to a new branch
./git-smartmsg apply --branch feature/ai-improved-messages

# 5. Review the new branch
git log --oneline -10

# 6. Push when satisfied (optional)
git push --force-with-lease origin feature/ai-improved-messages
```

## Emoji Mode

When using `--emoji` flag, commit messages are prefixed with contextual emojis:

| Emoji | Code | Usage |
|-------|------|-------|
| 🎨 | `:art:` | Improving code structure/format |
| 🐛 | `:bug:` | Fixing bugs |
| 🔥 | `:fire:` | Removing code or files |
| 📝 | `:memo:` | Writing docs |
| ⚡ | `:zap:` | Improving performance |
| ✅ | `:white_check_mark:` | Adding tests |
| 🔒 | `:lock:` | Security fixes |
| ⬆️ | `:arrow_up:` | Upgrading dependencies |

**Example Output:**
```
🎨 Refactor user authentication module
🐛 Fix null pointer exception in data parser
📝 Update API documentation for v2 endpoints
✅ Add unit tests for payment processing
```

## Safety & Best Practices

### Safety Features

- **Clean Worktree Required**: Ensures no uncommitted changes (ignores `plan.json`)
- **New Branch Creation**: Never modifies your current branch
- **Author Preservation**: Maintains original author info and timestamps
- **Backup Recommendations**: Original commits remain accessible

### Best Practices

1. **Review Before Applying**: Always check `plan.json` before running `apply`
2. **Use Small Batches**: Process 10-20 commits at a time for better results
3. **Test Branch**: Review the generated branch before merging
4. **Team Coordination**: Coordinate with team before force-pushing rewritten history
5. **Backup**: Consider creating a backup branch before major rewrites

### Force Push Safely

```bash
# Use --force-with-lease for safer force pushing
git push --force-with-lease origin your-branch-name
```

## File Structure

```
.
├── main.go           # Complete application
├── go.mod            # Go dependencies
├── go.sum            # Dependency checksums
├── plan.json         # Generated plan (ignored by git)
├── CLAUDE.md         # Claude Code guidance
└── README.md         # This file
```

## Troubleshooting

### Common Issues

**"worktree is not clean"**
- Commit or stash your changes first
- `plan.json` is automatically ignored

**"AI failed for commit"**
- Check your OpenAI API key
- Verify API quota/limits
- Try a smaller batch size

**"cherry-pick failed"**
- Complex conflicts may require manual resolution
- Consider excluding merge commits with default settings

### Getting Help

```bash
# Show available commands
./git-smartmsg

# Show command-specific help
./git-smartmsg plan --help
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [OpenAI API](https://openai.com/api/) for AI-powered message generation
- [Conventional Commits](https://www.conventionalcommits.org/) for commit message standards
- [gitmoji](https://gitmoji.dev/) for emoji inspiration