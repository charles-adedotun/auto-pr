package config

import (
	"fmt"
	"os"
	"path/filepath"

	"auto-pr/pkg/types"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*types.Config, error) {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, ".auto-pr", "config.yaml")
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config types.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge with defaults
	mergeWithDefaults(&config)

	return &config, nil
}

// WriteConfig writes configuration to file
func WriteConfig(configPath string, config *types.Config) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ValidateConfig validates a configuration
func ValidateConfig(config *types.Config) error {
	// Validate AI configuration
	if err := validateAIConfig(&config.AI); err != nil {
		return fmt.Errorf("AI configuration error: %w", err)
	}

	// Validate Git configuration
	if err := validateGitConfig(&config.Git); err != nil {
		return fmt.Errorf("git configuration error: %w", err)
	}

	return nil
}

// validateAIConfig validates AI configuration
func validateAIConfig(ai *types.AIConfig) error {
	// Check provider
	switch ai.Provider {
	case types.AIProviderClaude:
		// Valid provider
	case "", "auto":
		ai.Provider = types.AIProviderClaude // Default to Claude or convert auto to Claude
	case "gemini":
		return fmt.Errorf("gemini provider is no longer supported. Please use Claude Code instead")
	default:
		return fmt.Errorf("invalid AI provider: %s", ai.Provider)
	}

	// Validate max tokens (0 means use default)
	if ai.MaxTokens != 0 && (ai.MaxTokens < 100 || ai.MaxTokens > 100000) {
		return fmt.Errorf("max_tokens must be 0 (default) or between 100 and 100000, got %d", ai.MaxTokens)
	}

	// Validate temperature
	if ai.Temperature < 0 || ai.Temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2, got %f", ai.Temperature)
	}

	// Validate Claude configuration
	if ai.Provider == types.AIProviderClaude {
		if ai.Claude.MaxTokens < 0 {
			return fmt.Errorf("claude max_tokens must be non-negative")
		}
	}

	return nil
}

// validateGitConfig validates Git configuration
func validateGitConfig(git *types.GitConfig) error {
	// Validate commit limit (0 means use default)
	if git.CommitLimit != 0 && (git.CommitLimit < 1 || git.CommitLimit > 100) {
		return fmt.Errorf("commit_limit must be 0 (default) or between 1 and 100, got %d", git.CommitLimit)
	}

	if git.DiffContext < 0 || git.DiffContext > 20 {
		return fmt.Errorf("diff_context must be between 0 and 20, got %d", git.DiffContext)
	}

	if git.MaxDiffSize < 0 {
		return fmt.Errorf("max_diff_size must be non-negative, got %d", git.MaxDiffSize)
	}

	return nil
}

// getDefaultConfig returns default configuration
func getDefaultConfig() *types.Config {
	return &types.Config{
		AI: types.AIConfig{
			Provider:    types.AIProviderClaude,
			MaxTokens:   4096,
			Temperature: 0.7,
			Claude: types.ClaudeConfig{
				CLIPath:    "claude",
				Model:      "claude-3-5-sonnet-20241022",
				MaxTokens:  4096,
				UseSession: true,
			},
		},
		Platforms: types.PlatformConfig{
			GitHub: types.GitHubConfig{
				DefaultReviewers: []string{},
				Labels:           []string{"auto-generated"},
				Draft:            false,
				AutoMerge:        false,
				DeleteBranch:     true,
			},
			GitLab: types.GitLabConfig{
				DefaultAssignee:           "",
				MergeWhenPipelineSucceeds: false,
				RemoveSourceBranch:        true,
			},
		},
		Templates: types.TemplateConfig{
			Feature:           "feature-template",
			Bugfix:            "bugfix-template",
			CustomTemplateDir: "~/.auto-pr/templates",
		},
		Git: types.GitConfig{
			CommitLimit:    10,
			DiffContext:    3,
			IgnorePatterns: []string{"*.log", "node_modules/", "*.tmp"},
			MaxDiffSize:    10000,
		},
	}
}

// LoadConfigWithViper loads configuration using viper (with env var support)
func LoadConfigWithViper() (*types.Config, error) {
	config := getDefaultConfig()

	// Load from viper (includes env vars and config file)
	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply environment variable overrides that viper might have loaded
	applyEnvOverrides(config)

	// Validate the final configuration
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// applyEnvOverrides applies any environment variable overrides
func applyEnvOverrides(config *types.Config) {
	// AI provider override
	if provider := viper.GetString("ai.provider"); provider != "" {
		config.AI.Provider = types.AIProvider(provider)
	}

	// No API keys needed for Claude CLI

	// Apply numeric overrides
	if maxTokens := viper.GetInt("ai.max_tokens"); maxTokens > 0 {
		config.AI.MaxTokens = maxTokens
	}
	if temp := viper.GetFloat64("ai.temperature"); temp > 0 {
		config.AI.Temperature = float32(temp)
	}

	// Git config overrides
	if commitLimit := viper.GetInt("git.commit_limit"); commitLimit > 0 {
		config.Git.CommitLimit = commitLimit
	}
	if diffContext := viper.GetInt("git.diff_context"); diffContext > 0 {
		config.Git.DiffContext = diffContext
	}
	if maxDiffSize := viper.GetInt("git.max_diff_size"); maxDiffSize > 0 {
		config.Git.MaxDiffSize = maxDiffSize
	}
}

// mergeWithDefaults merges configuration with defaults
func mergeWithDefaults(config *types.Config) {
	defaults := getDefaultConfig()

	// Merge AI config
	if config.AI.MaxTokens == 0 {
		config.AI.MaxTokens = defaults.AI.MaxTokens
	}
	if config.AI.Temperature == 0 {
		config.AI.Temperature = defaults.AI.Temperature
	}
	if config.AI.Provider == "" {
		config.AI.Provider = defaults.AI.Provider
	}

	// Merge Claude config
	if config.AI.Claude.CLIPath == "" {
		config.AI.Claude.CLIPath = defaults.AI.Claude.CLIPath
	}
	if config.AI.Claude.Model == "" {
		config.AI.Claude.Model = defaults.AI.Claude.Model
	}
	if config.AI.Claude.MaxTokens == 0 {
		config.AI.Claude.MaxTokens = defaults.AI.Claude.MaxTokens
	}


	// Merge Git config
	if config.Git.CommitLimit == 0 {
		config.Git.CommitLimit = defaults.Git.CommitLimit
	}
	if config.Git.DiffContext == 0 {
		config.Git.DiffContext = defaults.Git.DiffContext
	}
	if config.Git.MaxDiffSize == 0 {
		config.Git.MaxDiffSize = defaults.Git.MaxDiffSize
	}
	if len(config.Git.IgnorePatterns) == 0 {
		config.Git.IgnorePatterns = defaults.Git.IgnorePatterns
	}

	// Merge platform config defaults
	if len(config.Platforms.GitHub.Labels) == 0 {
		config.Platforms.GitHub.Labels = defaults.Platforms.GitHub.Labels
	}
}
