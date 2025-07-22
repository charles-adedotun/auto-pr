# Auto PR

[![CI](https://github.com/charles-adedotun/auto-pr/actions/workflows/ci.yml/badge.svg)](https://github.com/charles-adedotun/auto-pr/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/charles-adedotun/auto-pr)](https://goreportcard.com/report/github.com/charles-adedotun/auto-pr)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Auto PR is a powerful CLI tool that automatically generates pull requests and merge requests using AI. It analyzes your git changes, commit history, and repository context to create meaningful PR/MR titles, descriptions, and metadata for GitHub and GitLab.

## Features

- 🤖 **AI-Powered**: Supports both Claude CLI and Google Gemini 2.5 Flash for intelligent PR/MR generation
- 🔍 **Smart Analysis**: Analyzes git commits, diffs, and repository context
- 🌐 **Multi-Platform**: Supports both GitHub and GitLab
- 📝 **Template System**: Customizable templates for different change types
- ⚡ **Fast & Efficient**: Single binary, no dependencies
- 🛠️ **Configurable**: Extensive configuration options via YAML/JSON/ENV
- 🔐 **Secure**: Uses existing GitHub CLI and GitLab CLI authentication
- 🔌 **MCP Server**: Use as a tool in Claude Code and other MCP-compatible AI assistants

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
- **AI Provider** (choose one):
  - **Claude CLI** (recommended): `claude` command available in PATH with authentication
  - **Google Gemini**: API key for Gemini 2.5 Flash

### Basic Usage

1. **Initialize configuration:**
   ```bash
   auto-pr config init
   ```

2. **Create a pull request:**
   ```bash
   # Basic usage - analyzes current branch
   auto-pr create
   
   # Preview without creating
   auto-pr create --dry-run
   
   # Interactive mode with confirmation
   auto-pr create --interactive
   
   # Create as draft
   auto-pr create --draft
   ```

## Configuration

Auto PR uses a configuration file at `~/.auto-pr/config.yaml`:

```yaml
ai:
  provider: "auto"  # auto-detect available AI service, or specify "claude" or "gemini"
  max_tokens: 4096
  temperature: 0.7
  
  # Claude CLI configuration (preferred)
  claude:
    cli_path: "claude"  # auto-detected if in PATH
    model: "claude-3-5-sonnet-20241022"
    use_session: true  # use claude CLI session mode
  
  # Gemini API configuration (fallback)
  gemini:
    api_key: "${GEMINI_API_KEY}"
    model: "gemini-2.5-flash"
    project_id: "${GOOGLE_CLOUD_PROJECT_ID}"

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

```bash
# For Gemini API (if using Gemini)
export GEMINI_API_KEY="your-api-key-here"
export GOOGLE_CLOUD_PROJECT_ID="your-project-id"

# For Claude CLI (automatically uses existing authentication)
# No additional environment variables needed if claude CLI is already set up

# Optional: custom config file location
export AUTO_PR_CONFIG="path/to/config.yaml"
```

### AI Provider Setup

#### Claude CLI (Recommended)
1. Install Claude CLI: Follow instructions at [Claude CLI documentation](https://docs.anthropic.com/en/docs/claude-code)
2. Authenticate: The claude CLI should already be authenticated if you're using it
3. Auto PR will automatically detect and use your Claude CLI setup

#### Google Gemini
1. Get an API key from [Google AI Studio](https://aistudio.google.com/)
2. Set your API key in the environment or configuration file
3. Optionally set your Google Cloud Project ID for enhanced features

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

### MCP Server Mode
```bash
auto-pr mcp                       # Run as MCP server for AI assistants

# Add to Claude Code
claude mcp add auto-pr auto-pr mcp
```

### Template Management
```bash
auto-pr template list             # List available templates
auto-pr template create <name>    # Create new template
auto-pr template edit <name>      # Edit existing template
auto-pr template delete <name>    # Delete template
```

## Examples

### Feature Development
```bash
# Work on your feature
git checkout -b feature/user-authentication
# ... make changes ...
git commit -m "Add user authentication system"

# Create PR with feature template
auto-pr create --template feature --reviewer alice,bob
```

### Bug Fix
```bash
# Work on bug fix
git checkout -b fix/login-error
# ... make changes ...
git commit -m "Fix login validation error"

# Create PR as draft for review
auto-pr create --draft --template bugfix
```

### Review Before Creating
```bash
# Preview what would be created
auto-pr create --dry-run --verbose

# Interactive mode for confirmation
auto-pr create --interactive
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

## MCP Integration (Claude Code)

Auto PR can be used as an MCP (Model Context Protocol) server, allowing direct integration with Claude Code:

### Quick Setup
```bash
# Add Auto PR to Claude Code
claude mcp add auto-pr auto-pr mcp
```

### Available Tools
- **repo_status**: Check repository status and changes
- **analyze_changes**: Analyze commits and diffs
- **create_pr**: Create pull requests with AI assistance

### Example Usage in Claude Code
```
"Check the repository status and create a PR for the current changes"
```

See [MCP Integration Guide](docs/MCP_INTEGRATION.md) for detailed setup and usage.

## AI Integration

Auto PR supports multiple AI providers for flexibility and convenience:

### Claude CLI (Recommended)
- **Zero Configuration**: Leverages your existing Claude CLI setup
- **Session Support**: Uses Claude CLI session mode for better context retention
- **Local Integration**: Works seamlessly with local development workflow
- **High Quality**: Latest Claude models for superior PR/MR generation

### Google Gemini 2.5 Flash
- **API-Based**: Direct integration with Google's Gemini API
- **Cost-Effective**: Optimized for high-volume usage
- **Fast Response**: Quick generation for immediate feedback
- **Customizable**: Fine-tuned prompts and temperature settings

### Auto-Detection
- **Smart Selection**: Automatically chooses the best available AI provider
- **Fallback Support**: Falls back to secondary provider if primary fails
- **Configuration-Free**: Works out of the box with minimal setup

### Context Analysis
Auto PR analyzes:
- Commit messages and history
- Code diffs and file changes
- Repository structure and language
- Previous PR/MR patterns
- CODEOWNERS and team structure

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
- For Claude CLI: Ensure `claude --version` works and you're authenticated
- For Gemini: Check your API key is set: `echo $GEMINI_API_KEY`
- Verify API quotas and billing in Google Cloud Console (for Gemini)

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
- [Claude CLI](https://docs.anthropic.com/en/docs/claude-code) and [Google Gemini](https://cloud.google.com/vertex-ai/docs/generative-ai/model-reference/gemini) for AI capabilities