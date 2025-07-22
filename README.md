# Auto PR

[![CI](https://github.com/charles-adedotun/auto-pr/actions/workflows/ci.yml/badge.svg)](https://github.com/charles-adedotun/auto-pr/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/charles-adedotun/auto-pr)](https://goreportcard.com/report/github.com/charles-adedotun/auto-pr)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Auto PR is a powerful CLI tool that automatically generates pull requests and merge requests using Claude Code. It analyzes your git changes, commit history, and repository context to create meaningful PR/MR titles, descriptions, and metadata for GitHub and GitLab. Now with super simple one-command workflows!

## Features

- 🤖 **AI-Powered**: Uses Claude Code for intelligent PR/MR generation
- 🔍 **Smart Analysis**: Analyzes git commits, diffs, and repository context
- 🌐 **Multi-Platform**: Supports both GitHub and GitLab
- 📝 **Template System**: Customizable templates for different change types
- ⚡ **Fast & Efficient**: Single binary, no dependencies
- 🛠️ **Configurable**: Extensive configuration options via YAML/JSON/ENV
- 🔐 **Secure**: Uses existing GitHub CLI and GitLab CLI authentication
- 🔌 **MCP Server**: Can be used as an MCP (Model Context Protocol) server for AI assistants

## Quick Start

### Installation

```bash
# Using curl (recommended)
curl -sf https://raw.githubusercontent.com/charles-adedotun/auto-pr/main/scripts/install.sh | sh

# Using Go
go install github.com/charles-adedotun/auto-pr@latest

# Using Homebrew (coming soon)
brew install auto-pr

# Download binary from releases
# Visit: https://github.com/charles-adedotun/auto-pr/releases
```

### Prerequisites

- Git repository with remote origin
- GitHub CLI (`gh`) for GitHub repositories, or GitLab CLI (`glab`) for GitLab
- **Claude Code**: `claude` command available in PATH with authentication

### Basic Usage

1. **Initialize configuration:**
   ```bash
   auto-pr config init
   ```

2. **Super Simple Workflow:**
   ```bash
   # The ultimate shortcut - does everything!
   auto-pr ship
   # This will: stage → commit → push → create PR
   # Perfect for when you just want to ship your changes fast!
   # Zero configuration - AI handles branch naming, commits, and PRs intelligently!
   
   # Other simple aliases:
   auto-pr go          # same as ship
   auto-pr send        # same as ship
   auto-pr deploy      # same as ship
   ```

3. **Individual Commands:**
   ```bash
   # Smart commit with AI message
   auto-pr cm -a       # stage all & commit
   auto-pr commit -a   # same thing
   
   # Create PR/MR
   auto-pr pr          # create pull request
   auto-pr mr          # create merge request
   auto-pr create      # same thing
   
   # Preview before doing anything
   auto-pr ship --dry-run
   auto-pr cm --dry-run -a
   ```

## Configuration

Auto PR uses a configuration file at `~/.auto-pr/config.yaml`:

```yaml
ai:
  provider: "claude"  # Uses Claude Code
  max_tokens: 4096
  temperature: 0.7
  
  # Claude Code configuration
  claude:
    cli_path: "claude"  # auto-detected if in PATH
    model: "claude-3-5-sonnet-20241022"
    use_session: true  # use claude CLI session mode

platforms:
  github:
    default_reviewers: ["teamlead", "senior-dev"]
    labels: ["auto-generated"]
    draft: false
  gitlab:
    default_assignee: "maintainer"
    merge_when_pipeline_succeeds: true

templates:
  feature: "feature-template"
  bugfix: "bugfix-template"
  custom_templates_dir: "~/.auto-pr/templates"

git:
  commit_limit: 10
  diff_context: 3
  ignore_patterns: ["*.log", "node_modules/"]
```

### Environment Variables

Auto PR supports extensive configuration through environment variables. All config file options can be overridden using environment variables with the `AUTO_PR_` prefix:

```bash
# AI Configuration
export AUTO_PR_AI_PROVIDER="claude"              # AI provider: "claude"
export AUTO_PR_AI_MODEL="claude-3-5-sonnet"      # Override default model
export AUTO_PR_AI_MAX_TOKENS="4096"              # Maximum tokens for AI response
export AUTO_PR_AI_TEMPERATURE="0.7"              # AI temperature (0-2)

# Claude Code Configuration
export AUTO_PR_CLAUDE_CLI_PATH="claude"          # Path to Claude CLI
export AUTO_PR_CLAUDE_MODEL="claude-3-5-sonnet-20241022"
export AUTO_PR_CLAUDE_MAX_TOKENS="4096"
export AUTO_PR_CLAUDE_USE_SESSION="true"         # Use Claude session mode

# GitHub Configuration
export AUTO_PR_GITHUB_DRAFT="false"              # Create PRs as draft
export AUTO_PR_GITHUB_AUTO_MERGE="false"         # Enable auto-merge
export AUTO_PR_GITHUB_DELETE_BRANCH="true"       # Delete branch after merge

# GitLab Configuration
export AUTO_PR_GITLAB_AUTO_MERGE="true"          # Merge when pipeline succeeds
export AUTO_PR_GITLAB_REMOVE_SOURCE_BRANCH="true"
export AUTO_PR_GITLAB_DEFAULT_ASSIGNEE="username"

# Git Configuration
export AUTO_PR_GIT_COMMIT_LIMIT="10"             # Number of commits to analyze
export AUTO_PR_GIT_DIFF_CONTEXT="3"              # Lines of context in diffs
export AUTO_PR_GIT_MAX_DIFF_SIZE="10000"         # Maximum diff size

# Template Configuration
export AUTO_PR_TEMPLATES_DIR="~/.auto-pr/templates"  # Custom templates directory

# General Configuration
export AUTO_PR_CONFIG="/path/to/config.yaml"     # Custom config file location
```

### Claude Code Setup

1. Install Claude Code: Follow instructions at [Claude Code documentation](https://docs.anthropic.com/en/docs/claude-code)
2. Authenticate: The claude CLI should already be authenticated if you're using it
3. Auto PR will automatically detect and use your Claude Code setup

## Commands

### Create PR/MR
```bash
auto-pr create [flags]
```

**Flags:**
- `--dry-run`: Preview without creating
- `--interactive`: Interactive mode with confirmation
- `--template <name>`: Use specific template
- `--reviewer <users>`: Override default reviewers
- `--draft`: Create as draft
- `--auto-merge`: Enable auto-merge
- `--force`: Skip validations
- `--commit-range <range>`: Specific commit range
- `--ai-context <file>`: Additional context file

### Configuration Management
```bash
auto-pr config init               # Initialize configuration
auto-pr config set <key> <value>  # Set configuration value
auto-pr config get <key>          # Get configuration value
auto-pr config list               # List all configuration
auto-pr config validate           # Validate current configuration
```

### Repository Status
```bash
auto-pr status                    # Show repository status and readiness
```

### Template Management
```bash
auto-pr template list             # List available templates
auto-pr template create <name>    # Create new template
auto-pr template edit <name>      # Edit existing template
auto-pr template delete <name>    # Delete template
```

## Examples

### Super Quick Workflow
```bash
# Make your changes, then ship everything in one command!
git checkout -b feature/user-auth
# ... edit files ...
auto-pr ship                    # stages, commits, pushes, creates PR

# Want to preview first?
auto-pr ship --dry-run          # see what would happen

# Custom commit message
auto-pr ship -m "Add user authentication system"

# Create as draft with reviewers
auto-pr ship --draft --reviewer alice,bob
```

### Step by Step
```bash
# Just commit changes (with AI message)
auto-pr cm -a                   # stage all and commit

# Just create PR (assumes you already committed)
auto-pr pr                      # create pull request

# Commit with custom message
auto-pr cm -a -m "Fix bug in login system"

# Commit and push (but no PR)
auto-pr cm -a --push
```

### Advanced Usage
```bash
# Ship without creating PR (just commit & push)
auto-pr ship --no-pr

# Ship without pushing (just stage & commit)
auto-pr ship --no-push

# Different aliases for ship
auto-pr go                      # same as ship
auto-pr send                    # same as ship
auto-pr deploy                  # same as ship
```

## Templates

Auto PR supports custom templates for different types of changes:

### Built-in Templates
- **Feature**: New functionality
- **Bugfix**: Bug fixes
- **Hotfix**: Critical fixes
- **Refactor**: Code refactoring
- **Documentation**: Documentation updates
- **Dependency**: Dependency updates

### Custom Templates

Create custom templates in `~/.auto-pr/templates/`:

```yaml
# ~/.auto-pr/templates/my-template.yaml
name: "my-template"
type: "custom"
title_format: "{{.Type}}: {{.Summary}}"
body_format: |
  ## Summary
  {{.Summary}}
  
  ## Changes
  {{.Changes}}
  
  ## Testing
  {{.TestingNotes}}

conditions:
  - field: "files"
    operator: "contains"
    value: "test/"
```

## Platform Integration

### GitHub
- Uses GitHub CLI (`gh`) for authentication and API calls
- Supports all GitHub PR features: reviewers, labels, milestones, projects
- Automatic CODEOWNERS integration
- Template detection from `.github/pull_request_template.md`

### GitLab
- Uses GitLab CLI (`glab`) for authentication and API calls
- Supports GitLab MR features: assignees, labels, milestones
- Works with both GitLab.com and self-hosted instances
- Template detection from `.gitlab/merge_request_templates/`

## AI Integration

Auto PR uses Claude Code for intelligent pull request generation:

### Claude Code Integration
- **Zero Configuration**: Leverages your existing Claude Code setup
- **Session Support**: Uses Claude CLI session mode for better context retention
- **Local Integration**: Works seamlessly with local development workflow
- **High Quality**: Latest Claude models for superior PR/MR generation
- **Automatic Detection**: Automatically detects and uses your Claude Code installation

### Context Analysis
Auto PR analyzes:
- Commit messages and history
- Code diffs and file changes
- Repository structure and language
- Previous PR/MR patterns
- CODEOWNERS and team structure

## MCP Server Mode

Auto PR can run as an MCP (Model Context Protocol) server, allowing AI assistants like Claude Code to use it as a tool for automated PR/MR creation.

### MCP Server Setup

1. **Run as MCP server:**
   ```bash
   auto-pr mcp
   ```

2. **Configure in Claude Code:**
   Add to your Claude Code MCP settings:
   ```json
   {
     "auto-pr": {
       "command": "auto-pr",
       "args": ["mcp"]
     }
   }
   ```

3. **Available MCP Tools:**
   - `repo_status`: Get repository status and branch information
   - `analyze_changes`: Analyze git changes and generate PR content
   - `create_pr`: Create a pull request with AI-generated content

### Using with Claude Code

Once configured, Claude Code can use Auto PR to:
- Analyze repository changes
- Generate PR descriptions automatically
- Create pull requests directly from the chat interface

Example prompts:
- "Check the repository status"
- "Analyze my changes and suggest a PR description"
- "Create a pull request for my current changes"

## Development

### Building from Source
```bash
git clone https://github.com/charles-adedotun/auto-pr.git
cd auto-pr
go build -o auto-pr .
```

### Running Tests
```bash
go test ./...
```

### Contributing
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Troubleshooting

### Common Issues

**"not in a git repository"**
- Ensure you're in a git repository with `git status`
- Check that the repository has a remote origin

**"failed to detect platform"**
- Verify your remote URL with `git remote -v`
- Ensure the URL points to GitHub or GitLab

**"authentication failed"**
- For GitHub: Run `gh auth login`
- For GitLab: Run `glab auth login`

**"AI API error"**
- For Claude Code: Ensure `claude --version` works and you're authenticated
- Verify your Claude Code setup and authentication status

### Debug Mode
```bash
auto-pr create --verbose --dry-run
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [Viper](https://github.com/spf13/viper) for configuration management
- [GitHub CLI](https://github.com/cli/cli) and [GitLab CLI](https://gitlab.com/gitlab-org/cli) for platform integration
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) for AI capabilities