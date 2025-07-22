package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	
	"auto-pr/internal/git"
	"auto-pr/internal/platforms"
	"auto-pr/pkg/types"
	
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SimpleServer represents a simplified MCP server for Auto PR
type SimpleServer struct {
	mcpServer        *server.MCPServer
	gitAnalyzer      *git.Analyzer
	platformDetector platforms.PlatformClient
}

// NewSimpleServer creates a new simplified MCP server
func NewSimpleServer(gitAnalyzer *git.Analyzer, platformDetector platforms.PlatformClient) (*SimpleServer, error) {
	// Create MCP server with tool capabilities only
	mcpServer := server.NewMCPServer(
		"Auto PR MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	
	s := &SimpleServer{
		mcpServer:        mcpServer,
		gitAnalyzer:      gitAnalyzer,
		platformDetector: platformDetector,
	}
	
	// Register tools
	s.registerTools()
	
	return s, nil
}

// Run starts the MCP server
func (s *SimpleServer) Run(ctx context.Context) error {
	return server.ServeStdio(s.mcpServer)
}

// registerTools registers all available MCP tools
func (s *SimpleServer) registerTools() {
	// Repository status tool
	statusTool := mcp.NewTool("repo_status",
		mcp.WithDescription("Get current repository status including branch info and uncommitted changes"),
	)
	s.mcpServer.AddTool(statusTool, s.handleRepoStatus)
	
	// Analyze changes tool
	analyzeChangesTool := mcp.NewTool("analyze_changes",
		mcp.WithDescription("Analyze git changes between current branch and base branch"),
		mcp.WithString("base",
			mcp.Description("Base branch to compare against (optional - defaults to main/master)"),
		),
	)
	s.mcpServer.AddTool(analyzeChangesTool, s.handleAnalyzeChanges)
	
	// Create PR tool (simplified)
	createPRTool := mcp.NewTool("create_pr",
		mcp.WithDescription("Create a pull request with given title and body"),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("PR/MR title"),
		),
		mcp.WithString("body",
			mcp.Required(),
			mcp.Description("PR/MR body/description"),
		),
		mcp.WithBoolean("draft",
			mcp.Description("Create as draft PR/MR"),
		),
		mcp.WithBoolean("dry_run",
			mcp.Description("Preview without creating"),
		),
	)
	s.mcpServer.AddTool(createPRTool, s.handleCreatePR)
}

// Tool Handlers

func (s *SimpleServer) handleRepoStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.gitAnalyzer == nil {
		return mcp.NewToolResultError("Not in a git repository"), nil
	}
	
	status, err := s.gitAnalyzer.GetStatus()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get repository status: %v", err)), nil
	}
	
	// Format status as JSON
	statusJSON, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format status: %v", err)), nil
	}
	
	return mcp.NewToolResultText(string(statusJSON)), nil
}

func (s *SimpleServer) handleAnalyzeChanges(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.gitAnalyzer == nil {
		return mcp.NewToolResultError("Not in a git repository"), nil
	}
	
	args := request.GetArguments()
	base, _ := args["base"].(string)
	
	if base == "" {
		status, _ := s.gitAnalyzer.GetStatus()
		base = status.BaseBranch
	}
	
	// Get commits
	commits, err := s.gitAnalyzer.GetCommitsSinceBase(base)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get commits: %v", err)), nil
	}
	
	// Get diff summary
	diffSummary, err := s.gitAnalyzer.GetBranchDiff(base)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get diff: %v", err)), nil
	}
	
	result := map[string]interface{}{
		"base_branch":   base,
		"commits_count": len(commits),
		"files_changed": diffSummary.TotalFiles,
		"lines_added":   diffSummary.Additions,
		"lines_deleted": diffSummary.Deletions,
		"commits":       commits,
		"file_changes":  diffSummary.FileChanges,
	}
	
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(string(resultJSON)), nil
}

func (s *SimpleServer) handleCreatePR(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if s.gitAnalyzer == nil {
		return mcp.NewToolResultError("Not in a git repository"), nil
	}
	
	args := request.GetArguments()
	
	// Extract required parameters
	title, err := request.RequireString("title")
	if err != nil {
		return mcp.NewToolResultError("Title is required"), nil
	}
	
	body, err := request.RequireString("body")
	if err != nil {
		return mcp.NewToolResultError("Body is required"), nil
	}
	
	draft, _ := args["draft"].(bool)
	dryRun, _ := args["dry_run"].(bool)
	
	// Get repository status
	status, err := s.gitAnalyzer.GetStatus()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get repository status: %v", err)), nil
	}
	
	// Preview in dry run mode
	if dryRun {
		preview := fmt.Sprintf("PR Preview:\n"+
			"Title: %s\n"+
			"Branch: %s -> %s\n"+
			"Draft: %v\n\n"+
			"Body:\n%s",
			title, status.CurrentBranch, status.BaseBranch, draft, body)
		return mcp.NewToolResultText(preview), nil
	}
	
	// Create actual PR
	if s.platformDetector == nil {
		return mcp.NewToolResultError("Platform not detected. Ensure you have a valid remote repository."), nil
	}
	
	prRequest := &types.PullRequestRequest{
		Title:      title,
		Body:       body,
		HeadBranch: status.CurrentBranch,
		BaseBranch: status.BaseBranch,
		Draft:      draft,
	}
	
	pr, err := s.platformDetector.CreatePullRequest(prRequest)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create PR: %v", err)), nil
	}
	
	result := fmt.Sprintf("✅ Successfully created PR #%d: %s\nURL: %s", pr.Number, pr.Title, pr.URL)
	return mcp.NewToolResultText(result), nil
}