# MCP Integration Guide

This guide explains how to use Auto PR as an MCP (Model Context Protocol) server with Claude Code and other MCP-compatible AI assistants.

## What is MCP?

The Model Context Protocol (MCP) is an open standard that enables AI assistants to interact with external tools and data sources. Auto PR implements MCP to provide git and pull request functionality directly to AI assistants.

## Adding Auto PR to Claude Code

### Method 1: Using Claude Code CLI

```bash
# Install Auto PR globally
go install github.com/charles-adedotun/auto-pr@latest

# Add Auto PR as an MCP server
claude mcp add auto-pr auto-pr mcp

# Or, if you have a local build
claude mcp add auto-pr /path/to/auto-pr mcp
```

### Method 2: Manual Configuration

Add to your Claude Code configuration:

```json
{
  "mcp": {
    "servers": {
      "auto-pr": {
        "command": "auto-pr",
        "args": ["mcp"],
        "type": "stdio"
      }
    }
  }
}
```

### Method 3: Project-specific Configuration

Create a `.mcp.json` file in your project root:

```json
{
  "auto-pr": {
    "command": "auto-pr",
    "args": ["mcp"],
    "type": "stdio"
  }
}
```

## Available MCP Tools

Once configured, Auto PR provides the following tools to AI assistants:

### 1. `repo_status`
Get current repository status including branch info and uncommitted changes.

**Usage:**
```
Use the repo_status tool to check the current git repository status
```

**Returns:**
- Current branch
- Base branch
- Remote URL
- Uncommitted changes (staged, unstaged, untracked)
- Commits ahead/behind

### 2. `analyze_changes`
Analyze git changes between current branch and base branch.

**Parameters:**
- `base` (optional): Base branch to compare against (defaults to main/master)

**Usage:**
```
Use the analyze_changes tool to see what changes have been made
```

**Returns:**
- Number of commits
- Files changed
- Lines added/deleted
- Detailed commit history
- File-by-file changes

### 3. `create_pr`
Create a pull request with given title and body.

**Parameters:**
- `title` (required): PR/MR title
- `body` (required): PR/MR body/description
- `draft` (optional): Create as draft PR/MR
- `dry_run` (optional): Preview without creating

**Usage:**
```
Use the create_pr tool to create a pull request with title "Fix bug in authentication" and appropriate body
```

## Example Workflows

### 1. Review Changes and Create PR

```
1. First, analyze the changes in this branch
2. Based on the changes, create a pull request with an appropriate title and description
```

### 2. Check Repository Status

```
Check the current repository status and tell me if there are any uncommitted changes
```

### 3. Create Draft PR for Review

```
Create a draft pull request for the current changes with title "WIP: New feature" 
```

## How It Works

1. When you use Auto PR in Claude Code, it runs as an MCP server
2. Claude Code communicates with Auto PR over stdio (standard input/output)
3. Auto PR executes git commands and platform CLI tools locally
4. Results are returned to Claude Code in a structured format

## Requirements

- Git repository with configured remote
- GitHub CLI (`gh`) or GitLab CLI (`glab`) installed and authenticated
- Auto PR binary in PATH or specified location

## Troubleshooting

### "Not in a git repository"
- Ensure you're running Claude Code from within a git repository
- Check that `.git` directory exists

### "Platform not detected"
- Verify your git remote is configured: `git remote -v`
- Ensure the remote URL points to GitHub or GitLab

### "Failed to create PR"
- Check that you're authenticated with GitHub/GitLab CLI
- For GitHub: `gh auth status`
- For GitLab: `glab auth status`

## Advanced Usage

### Environment Variables

You can set environment variables when adding the MCP server:

```bash
claude mcp add auto-pr auto-pr mcp -e GEMINI_API_KEY=your-key
```

### Custom Configuration

Create a custom Auto PR configuration before using MCP:

```bash
auto-pr config init
auto-pr config set ai.provider gemini
```

## Security Considerations

- Auto PR MCP server has access to your git repository
- It can create pull requests on your behalf
- Always review PR content before confirming creation
- Use `dry_run: true` to preview without creating

## Integration with AI Providers

When running as an MCP server, Auto PR relies on the host AI (Claude Code) for intelligent PR generation rather than calling external AI APIs. This provides:

- Seamless integration with your AI assistant
- No additional API keys required
- Context-aware PR generation
- Consistent AI behavior

## Contributing

To contribute to Auto PR's MCP implementation:

1. Check the `internal/mcp` directory for server implementation
2. Tools are defined in `registerTools()` method
3. Follow the MCP specification at https://modelcontextprotocol.io
4. Test with Claude Code before submitting PRs