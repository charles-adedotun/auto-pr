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
	case types.AIProviderGemini:
		return NewGeminiClient(config.Gemini)
	case types.AIProviderAuto:
		return NewAutoClient(config)
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", config.Provider)
	}
}

// NewAutoClient creates a client by auto-detecting available AI services
func NewAutoClient(config types.AIConfig) (AIClient, error) {
	// Try Claude CLI first (prefer local setup)
	if isClaudeAvailable() {
		claudeConfig := config.Claude
		if claudeConfig.CLIPath == "" {
			if path, err := exec.LookPath("claude"); err == nil {
				claudeConfig.CLIPath = path
			}
		}

		if client, err := NewClaudeClient(claudeConfig); err == nil && client.IsAvailable() {
			return client, nil
		}
	}

	// Fall back to Gemini if API key is available
	if config.Gemini.APIKey != "" || config.APIKey != "" {
		geminiConfig := config.Gemini
		if geminiConfig.APIKey == "" && config.APIKey != "" {
			geminiConfig.APIKey = config.APIKey
		}
		if geminiConfig.ProjectID == "" && config.ProjectID != "" {
			geminiConfig.ProjectID = config.ProjectID
		}
		if geminiConfig.Model == "" && config.Model != "" {
			geminiConfig.Model = config.Model
		}
		if geminiConfig.MaxTokens == 0 && config.MaxTokens > 0 {
			geminiConfig.MaxTokens = config.MaxTokens
		}
		if geminiConfig.Temperature == 0 && config.Temperature > 0 {
			geminiConfig.Temperature = config.Temperature
		}

		if client, err := NewGeminiClient(geminiConfig); err == nil && client.IsAvailable() {
			return client, nil
		}
	}

	return nil, fmt.Errorf("no available AI providers found. Please configure Claude CLI or Gemini API")
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

	// Gemini is always potentially available if API key is provided
	providers = append(providers, types.AIProviderGemini)

	return providers
}

// DetectBestProvider returns the recommended AI provider based on availability
func DetectBestProvider() types.AIProvider {
	if isClaudeAvailable() {
		return types.AIProviderClaude
	}
	return types.AIProviderGemini
}
