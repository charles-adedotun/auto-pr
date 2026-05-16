# Auto PR

[![CI](https://github.com/charles-adedotun/auto-pr/actions/workflows/ci.yml/badge.svg)](https://github.com/charles-adedotun/auto-pr/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/charles-adedotun/auto-pr)](https://goreportcard.com/report/github.com/charles-adedotun/auto-pr)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Auto PR is an experimental Go CLI for generating pull request and merge request text with Claude Code, then creating GitHub PRs or GitLab MRs through the local `gh` or `glab` CLI.

## Current Status

This project is usable as a local workflow tool, but some previously advertised features are still incomplete. Treat the MCP server mode and advanced platform metadata as work in progress.

## What Works Today

- Generate PR/MR title and body text using the local `claude` command.
- Analyze git branch state, commits, and changed files.
- Create GitHub pull requests using `gh pr create`.
- Create GitLab merge requests using `glab mr create`.
- Create commits with AI-generated or user-provided commit messages.
- Run a `ship` workflow that can stage, commit, push, and create a PR.
- Preview create and ship workflows with `--dry-run`.
- Use built-in or custom templates for generated PR/MR bodies.

## Important Limitations

- MCP mode currently lists tools, but tool calls return a work-in-progress response. Use the normal CLI commands for now.
- Labels are intentionally skipped in the main PR creation path to avoid failures on repositories where labels do not exist.
- The `--auto-merge` flag is accepted by the CLI but is not applied by the GitHub or GitLab platform clients.
- Project assignment, CODEOWNERS integration, and automatic PR template discovery are not implemented.
- Homebrew installation is not currently provided by this repository.
- Claude Code must already be installed, authenticated, and available as `claude` in `PATH`, unless configured otherwise.

## Installation

Build from source:

```bash
git clone https://github.com/charles-adedotun/auto-pr.git
cd auto-pr
go build -o auto-pr .
```

The module path is currently `auto-pr`, so `go install github.com/charles-adedotun/auto-pr@latest` is not advertised here until the module path is updated.

## Prerequisites

- A git repository with a GitHub or GitLab remote
- GitHub CLI (`gh`) for GitHub repositories, or GitLab CLI (`glab`) for GitLab repositories
- Claude Code CLI (`claude`) authenticated locally

## Basic Usage

Initialize configuration:

```bash
auto-pr config init
```

Preview a PR or MR:

```bash
auto-pr create --dry-run
```

Create a PR or MR:

```bash
auto-pr create
```

Commit with an AI-generated message:

```bash
auto-pr commit -a
```

Run the one-command workflow:

```bash
auto-pr ship --dry-run
auto-pr ship
```

`ship` may stage files, create a branch, commit, push, and create a PR. Use `--dry-run` first on important branches.

## Example PR/MR Output

Before creating a PR or MR, `auto-pr create --dry-run` gathers branch metadata, recent commits, and file-level diff stats, then asks the local `claude` CLI for structured PR content.

Example generated body:

```markdown
## Summary
Add a dry-run preview path for pull request creation.

## Changes
- Analyze the current branch against the detected base branch.
- Summarize changed files, additions, and deletions for Claude.
- Preview the generated title, body, labels, reviewers, priority, and provider.

## Testing
- go test ./...
- auto-pr create --dry-run
```

With a built-in template such as `--template feature`, the generated body is expanded with file changes, statistics, a checklist, and related issue placeholders.

## Configuration

Auto PR reads configuration from `~/.auto-pr/config.yaml` and environment variables with the `AUTO_PR_` prefix.

Example:

```yaml
ai:
  provider: "claude"
  claude:
    cli_path: "claude"
    model: "claude-3-5-sonnet-20241022"

platforms:
  github:
    default_reviewers: ["teamlead"]
    draft: false

git:
  commit_limit: 10
  diff_context: 3
  max_diff_size: 10000
```

Common environment variables:

```bash
export AUTO_PR_CLAUDE_CLI_PATH="claude"
export AUTO_PR_CLAUDE_MODEL="claude-3-5-sonnet-20241022"
export AUTO_PR_GITHUB_DRAFT="false"
export AUTO_PR_GIT_COMMIT_LIMIT="10"
export AUTO_PR_TEMPLATES_DIR="$HOME/.auto-pr/templates"
```

## Commands

```bash
auto-pr create [--dry-run] [--draft] [--reviewer user]
auto-pr commit -a [-m "message"] [--dry-run]
auto-pr ship [--dry-run] [--no-push] [--no-pr] [--draft]
auto-pr status
auto-pr template list
auto-pr config init
auto-pr config list
```

Aliases:

- `auto-pr pr` and `auto-pr mr` map to `auto-pr create`
- `auto-pr cm` maps to `auto-pr commit`
- `auto-pr go`, `auto-pr send`, and `auto-pr deploy` map to `auto-pr ship`

## MCP Mode

```bash
auto-pr mcp
```

MCP mode is experimental. It advertises `repo_status`, `analyze_changes`, and `create_pr`, but calls currently return a work-in-progress message instead of executing the full CLI behavior.

## Development

```bash
go test ./...
go build -o auto-pr .
```

## License

MIT
