package cmd

import (
	"context"
	"fmt"
	
	"auto-pr/internal/git"
	"auto-pr/internal/mcp"
	"auto-pr/internal/platforms"
	
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run Auto PR as an MCP server",
	Long: `Run Auto PR as a Model Context Protocol (MCP) server.
	
This allows Auto PR to be used as a tool by MCP-compatible AI assistants like Claude Code.
The server communicates over stdio and exposes Auto PR functionality as MCP tools.`,
	RunE: runMCPServer,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	// Create git analyzer
	gitAnalyzer, err := git.NewAnalyzer(".")
	if err != nil {
		// Not in a git repo - that's okay for MCP mode
		gitAnalyzer = nil
	}
	
	// Create platform detector
	var platformDetector platforms.PlatformClient
	if gitAnalyzer != nil {
		status, err := gitAnalyzer.GetStatus()
		if err == nil && status.RemoteURL != "" {
			platform, _ := platforms.DetectPlatform(status.RemoteURL)
			switch platform {
			case "github":
				platformDetector, _ = platforms.NewGitHubClient(status.RemoteURL)
			case "gitlab":
				platformDetector, _ = platforms.NewGitLabClient(status.RemoteURL)
			}
		}
	}
	
	// Create and run MCP server
	server, err := mcp.NewSimpleServer(gitAnalyzer, platformDetector)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}
	
	// Run the server (blocks until client disconnects)
	if err := server.Run(context.Background()); err != nil {
		return fmt.Errorf("MCP server error: %w", err)
	}
	
	return nil
}