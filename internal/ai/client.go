package ai

import (
	"fmt"
	"os/exec"

	"auto-pr/pkg/types"
)

// NewClient creates a new AI client based on the configuration
func NewClient(config types.AIConfig) (AIClient, error) {
	switch config.Provider {
	case types.AIProviderClaude:
		return NewClaudeClient(config.Claude)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", config.Provider)
	}
}


// isClaudeAvailable checks if Claude CLI is available in the system
func isClaudeAvailable() bool {
	_, err := exec.LookPath("claude")
	return err == nil
}

// GetAvailableProviders returns a list of available AI providers
func GetAvailableProviders() []types.AIProvider {
	var providers []types.AIProvider

	if isClaudeAvailable() {
		providers = append(providers, types.AIProviderClaude)
	}

	return providers
}

// DetectBestProvider returns the recommended AI provider based on availability
func DetectBestProvider() types.AIProvider {
	return types.AIProviderClaude
}
