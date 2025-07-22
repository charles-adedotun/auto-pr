package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"auto-pr/internal/config"
	"auto-pr/pkg/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `Manage Auto PR configuration settings`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  `Create a default configuration file with common settings`,
	RunE:  runConfigInit,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long:  `Set a configuration value. Use dot notation for nested keys (e.g., ai.provider)`,
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get configuration value",
	Long:  `Get a configuration value. Use dot notation for nested keys (e.g., ai.provider)`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration",
	Long:  `List all configuration values`,
	RunE:  runConfigList,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Validate the current configuration for errors`,
	RunE:  runConfigValidate,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd, configSetCmd, configGetCmd, configListCmd, configValidateCmd)

	configInitCmd.Flags().Bool("force", false, "Overwrite existing configuration")
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	force := cmd.Flags().Changed("force")

	configPath := getConfigPath()

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("configuration file already exists at %s. Use --force to overwrite", configPath)
	}

	// Create config directory
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create default configuration
	defaultConfig := getDefaultConfig()

	// Write configuration file
	if err := config.WriteConfig(configPath, defaultConfig); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	fmt.Printf("Configuration initialized at: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Ensure Claude Code is installed and authenticated")
	fmt.Println("2. Test the configuration:")
	fmt.Println("   auto-pr config validate")

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Load existing config
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set the value
	viper.Set(key, value)

	// Write the configuration back
	configPath := getConfigPath()
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Load configuration
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	value := viper.Get(key)
	if value == nil {
		fmt.Printf("Configuration key '%s' not found\n", key)
		return nil
	}

	fmt.Printf("%s = %v\n", key, value)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	// Load configuration
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	settings := viper.AllSettings()
	if len(settings) == 0 {
		fmt.Println("No configuration found. Run 'auto-pr config init' to create default configuration.")
		return nil
	}

	fmt.Println("Current configuration:")
	printNestedMap(settings, "")

	return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	// Load configuration
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Parse configuration into struct
	var cfg types.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Validate configuration
	if err := config.ValidateConfig(&cfg); err != nil {
		fmt.Printf("Configuration validation failed: %s\n", err)
		return err
	}

	fmt.Println("Configuration is valid ✓")
	return nil
}

// getConfigPath returns the configuration file path
func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".auto-pr/config.yaml"
	}

	return filepath.Join(home, ".auto-pr", "config.yaml")
}

// loadConfig loads the configuration file
func loadConfig() error {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file not found. Run 'auto-pr config init' to create it")
	}

	viper.SetConfigFile(configPath)
	return viper.ReadInConfig()
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

// printNestedMap recursively prints nested configuration
func printNestedMap(data map[string]interface{}, prefix string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			fmt.Printf("%s:\n", fullKey)
			printNestedMap(v, fullKey)
		case []interface{}:
			fmt.Printf("%s = [%s]\n", fullKey, formatSlice(v))
		default:
			fmt.Printf("%s = %v\n", fullKey, v)
		}
	}
}

// formatSlice formats slice values for display
func formatSlice(slice []interface{}) string {
	var items []string
	for _, item := range slice {
		items = append(items, fmt.Sprintf("%v", item))
	}
	return strings.Join(items, ", ")
}
