package ai

import (
	"fmt"
	
	"auto-pr/pkg/types"
)

// MCPClient implements AIClient for MCP server mode
// When running as an MCP server, we rely on the host AI for intelligence
type MCPClient struct{}

// NewMCPClient creates a new MCP client
func NewMCPClient() *MCPClient {
	return &MCPClient{}
}

// GenerateContent returns a placeholder response in MCP mode
// The actual AI generation happens in the host (e.g., Claude Code)
func (m *MCPClient) GenerateContent(ctx *AIContext, prompt string) (*AIResponse, error) {
	// In MCP mode, the AI host handles the intelligence
	// This is just a placeholder that should not be called
	return nil, fmt.Errorf("AI generation should be handled by MCP host")
}

// IsAvailable always returns true for MCP mode
func (m *MCPClient) IsAvailable() bool {
	return true
}

// GetProvider returns the MCP provider type
func (m *MCPClient) GetProvider() types.AIProvider {
	return types.AIProviderMCP
}

// ValidateConfig always succeeds for MCP
func (m *MCPClient) ValidateConfig() error {
	return nil
}