package config

import (
	"os"
	"path/filepath"
	"testing"

	"auto-pr/pkg/types"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.Config
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: &types.Config{
				AI: types.AIConfig{
					Provider:    types.AIProviderAuto,
					MaxTokens:   4096,
					Temperature: 0.7,
				},
				Git: types.GitConfig{
					CommitLimit: 10,
					DiffContext: 3,
					MaxDiffSize: 10000,
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid max tokens",
			config: &types.Config{
				AI: types.AIConfig{
					Provider:    types.AIProviderAuto,
					MaxTokens:   50, // Too low
					Temperature: 0.7,
				},
				Git: types.GitConfig{
					CommitLimit: 10,
					DiffContext: 3,
					MaxDiffSize: 10000,
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid temperature",
			config: &types.Config{
				AI: types.AIConfig{
					Provider:    types.AIProviderAuto,
					MaxTokens:   4096,
					Temperature: 3.0, // Too high
				},
				Git: types.GitConfig{
					CommitLimit: 10,
					DiffContext: 3,
					MaxDiffSize: 10000,
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid commit limit",
			config: &types.Config{
				AI: types.AIConfig{
					Provider:    types.AIProviderAuto,
					MaxTokens:   4096,
					Temperature: 0.7,
				},
				Git: types.GitConfig{
					CommitLimit: 0, // Too low
					DiffContext: 3,
					MaxDiffSize: 10000,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `ai:
  provider: auto
  max_tokens: 4096
  temperature: 0.7
git:
  commit_limit: 10
  diff_context: 3
  max_diff_size: 10000
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("LoadConfig() error = %v", err)
		return
	}

	// Verify loaded values
	if config.AI.Provider != types.AIProviderAuto {
		t.Errorf("LoadConfig() AI.Provider = %v, want %v", config.AI.Provider, types.AIProviderAuto)
	}
	if config.AI.MaxTokens != 4096 {
		t.Errorf("LoadConfig() AI.MaxTokens = %v, want %v", config.AI.MaxTokens, 4096)
	}
	if config.Git.CommitLimit != 10 {
		t.Errorf("LoadConfig() Git.CommitLimit = %v, want %v", config.Git.CommitLimit, 10)
	}
}

func TestWriteConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	config := &types.Config{
		AI: types.AIConfig{
			Provider:    types.AIProviderClaude,
			MaxTokens:   4096,
			Temperature: 0.7,
		},
		Git: types.GitConfig{
			CommitLimit: 10,
			DiffContext: 3,
			MaxDiffSize: 10000,
		},
	}

	// Write config
	err := WriteConfig(configPath, config)
	if err != nil {
		t.Errorf("WriteConfig() error = %v", err)
		return
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("WriteConfig() did not create config file")
		return
	}

	// Load it back to verify
	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Failed to load written config: %v", err)
		return
	}

	if loaded.AI.Provider != config.AI.Provider {
		t.Errorf("Written config has different provider: got %v, want %v",
			loaded.AI.Provider, config.AI.Provider)
	}
}
