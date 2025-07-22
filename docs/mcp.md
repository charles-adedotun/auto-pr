# MCP (Model Context Protocol) Integration

Auto PR can function as an MCP server, enabling AI assistants like Claude Code to use it as a tool for automated pull request creation and repository analysis.

## Overview

The MCP server mode provides three main tools:
1. **repo_status** - Get repository information and current branch status
2. **analyze_changes** - Analyze git changes and generate PR content suggestions
3. **create_pr** - Create a pull request with AI-generated content

## Installation & Setup

### 1. Install Auto PR

First, ensure Auto PR is installed and available in your PATH:

```bash
# Using the install script
curl -sf https://raw.githubusercontent.com/charles-adedotun/auto-pr/main/scripts/install.sh | sh

# Or download from releases
# https://github.com/charles-adedotun/auto-pr/releases
```

### 2. Configure Claude Code

Add Auto PR to your Claude Code MCP configuration. The configuration file location varies by platform:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

Add the following to your configuration:

```json
{
  "mcpServers": {
    "auto-pr": {
      "command": "auto-pr",
      "args": ["mcp"],
      "env": {
        // Optional: Add any environment variables needed
        "AUTO_PR_AI_PROVIDER": "claude",
        "AUTO_PR_GITHUB_DRAFT": "false"
      }
    }
  }
}
```

### 3. Restart Claude Code

After updating the configuration, restart Claude Code to load the MCP server.

## Available Tools

### repo_status

Get current repository status including branch information, uncommitted changes, and platform details.

**Request Example:**
```json
{
  "tool": "repo_status",
  "arguments": {}
}
```

**Response Example:**
```json
{
  "is_git_repo": true,
  "current_branch": "feature/new-api",
  "base_branch": "main",
  "has_uncommitted_changes": true,
  "platform": "github",
  "remote_url": "github.com/user/repo"
}
```

### analyze_changes

Analyze git changes and generate suggested PR title and body based on commits and diffs.

**Request Example:**
```json
{
  "tool": "analyze_changes",
  "arguments": {
    "commit_range": "main..HEAD",  // optional
    "context_file": "CONTEXT.md"   // optional
  }
}
```

**Response Example:**
```json
{
  "title": "Add user authentication API endpoints",
  "body": "## Summary\n\nImplemented JWT-based authentication...",
  "commit_count": 5,
  "files_changed": 12,
  "additions": 450,
  "deletions": 23
}
```

### create_pr

Create a pull request with the analyzed changes.

**Request Example:**
```json
{
  "tool": "create_pr",
  "arguments": {
    "title": "Add user authentication API",           // required
    "body": "## Summary\n\nImplemented JWT auth...",  // required
    "draft": false,                                    // optional
    "reviewers": ["alice", "bob"],                     // optional
    "labels": ["feature", "api"],                      // optional
    "auto_merge": false                                // optional
  }
}
```

**Response Example:**
```json
{
  "pr_url": "https://github.com/user/repo/pull/123",
  "pr_number": 123,
  "status": "created"
}
```

## Usage Examples

### Basic Workflow in Claude Code

1. **Check repository status:**
   ```
   User: "What's the current status of my repository?"
   Claude: *uses repo_status tool*
   ```

2. **Analyze changes:**
   ```
   User: "Analyze my changes and suggest a PR description"
   Claude: *uses analyze_changes tool*
   ```

3. **Create PR:**
   ```
   User: "Create a pull request for these changes"
   Claude: *uses create_pr tool*
   ```

### Advanced Usage

**With custom context:**
```
User: "Create a PR for my authentication feature. Use the ARCHITECTURE.md file for additional context."
Claude: *uses analyze_changes with context_file, then create_pr*
```

**Draft PR with specific reviewers:**
```
User: "Create a draft PR and assign alice and bob as reviewers"
Claude: *uses create_pr with draft=true and reviewers*
```

## Configuration

The MCP server respects all Auto PR configuration options. You can:

1. Use environment variables in the MCP configuration
2. Have a config file at `~/.auto-pr/config.yaml`
3. Pass configuration through tool arguments

### Environment Variables in MCP Config

```json
{
  "mcpServers": {
    "auto-pr": {
      "command": "auto-pr",
      "args": ["mcp"],
      "env": {
        "AUTO_PR_AI_PROVIDER": "claude",
        "AUTO_PR_GITHUB_DRAFT": "true",
        "AUTO_PR_GIT_COMMIT_LIMIT": "20"
      }
    }
  }
}
```

## Troubleshooting

### MCP Server Not Available

If Claude Code doesn't show Auto PR tools:

1. Verify Auto PR is in PATH: `which auto-pr`
2. Check Claude Code logs for errors
3. Validate JSON configuration syntax
4. Restart Claude Code

### Authentication Issues

The MCP server uses the same authentication as the CLI:

- **GitHub**: Requires `gh` CLI authenticated
- **GitLab**: Requires `glab` CLI authenticated

### Tool Errors

Common errors and solutions:

- **"not in a git repository"**: Ensure Claude Code is opened in a git repository
- **"no changes to analyze"**: Make some commits or changes first
- **"platform authentication failed"**: Run `gh auth login` or `glab auth login`

## Security Considerations

1. The MCP server runs with the same permissions as your user account
2. It can only access repositories you have access to
3. API keys in environment variables are only accessible to the MCP server process
4. The server does not store any data between requests

## Limitations

- The MCP server is stateless - each request is independent
- File paths must be absolute or relative to the current working directory
- Large repositories may take longer to analyze
- Binary files are excluded from analysis

## Development

To run the MCP server in development mode:

```bash
# Build and run
go build -o auto-pr .
./auto-pr mcp

# With verbose logging
AUTO_PR_VERBOSE=true ./auto-pr mcp
```

The server communicates over stdio using the MCP protocol. You can test it manually:

```bash
# Send a test request
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./auto-pr mcp
```