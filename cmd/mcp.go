package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"auto-pr/internal/git"

	"github.com/spf13/cobra"
)

// MCP Protocol types
type MCPRequest struct {
	JsonRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id"`
}

type MCPResponse struct {
	JsonRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MCPToolList struct {
	Tools []MCPTool `json:"tools"`
}

type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

var mcpCmd = &cobra.Command{
	Use:    "mcp",
	Short:  "Run as MCP (Model Context Protocol) server",
	Long:   `Run Auto PR as an MCP server for integration with AI assistants like Claude Code.`,
	RunE:   runMCPServer,
	Hidden: false,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	// Initialize git analyzer
	gitAnalyzer, err := git.NewAnalyzer(".")
	if err != nil {
		gitAnalyzer = nil // Allow MCP to work in non-git directories
	}

	return runMCPLoop(gitAnalyzer)
}

func runMCPLoop(gitAnalyzer *git.Analyzer) error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var request MCPRequest
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode request: %w", err)
		}

		response := handleMCPRequest(request, gitAnalyzer)
		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}
	}

	return nil
}

func handleMCPRequest(request MCPRequest, gitAnalyzer *git.Analyzer) MCPResponse {
	switch request.Method {
	case "tools/list":
		return MCPResponse{
			JsonRPC: "2.0",
			ID:      request.ID,
			Result: MCPToolList{
				Tools: []MCPTool{
					{
						Name:        "repo_status",
						Description: "Get repository status and branch information",
						InputSchema: map[string]interface{}{
							"type":       "object",
							"properties": map[string]interface{}{},
						},
					},
					{
						Name:        "analyze_changes",
						Description: "Analyze git changes and generate PR content",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"commit_range": map[string]interface{}{
									"type":        "string",
									"description": "Specific commit range to analyze",
								},
								"context_file": map[string]interface{}{
									"type":        "string",
									"description": "Additional context file path",
								},
							},
						},
					},
					{
						Name:        "create_pr",
						Description: "Create a pull request with AI-generated content",
						InputSchema: map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"title": map[string]interface{}{
									"type":        "string",
									"description": "Pull request title",
								},
								"body": map[string]interface{}{
									"type":        "string",
									"description": "Pull request body/description",
								},
								"draft": map[string]interface{}{
									"type":        "boolean",
									"description": "Create as draft",
								},
								"reviewers": map[string]interface{}{
									"type":        "array",
									"description": "List of reviewers",
									"items":       map[string]interface{}{"type": "string"},
								},
								"labels": map[string]interface{}{
									"type":        "array",
									"description": "List of labels",
									"items":       map[string]interface{}{"type": "string"},
								},
							},
							"required": []string{"title", "body"},
						},
					},
				},
			},
		}

	case "tools/call":
		return handleToolCall(request, gitAnalyzer)

	default:
		return MCPResponse{
			JsonRPC: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", request.Method),
			},
		}
	}
}

func handleToolCall(request MCPRequest, gitAnalyzer *git.Analyzer) MCPResponse {
	// This is a simplified implementation
	// In a full implementation, you would parse the params and call the appropriate tool
	return MCPResponse{
		JsonRPC: "2.0",
		ID:      request.ID,
		Result: map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "MCP tool implementation is a work in progress. Use the regular CLI commands for now.",
				},
			},
		},
	}
}