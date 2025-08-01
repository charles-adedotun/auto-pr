package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "auto-pr",
	Short: "Automatically generate pull requests and merge requests using AI",
	Long: `Auto PR is a CLI tool that analyzes your git changes and uses AI to 
generate comprehensive pull requests and merge requests for GitHub and GitLab.

It analyzes your commits, code changes, and repository context to create
meaningful PR/MR titles, descriptions, and metadata automatically.`,
	Version: "0.1.0",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.auto-pr/config.yaml)")
	rootCmd.PersistentFlags().Bool("verbose", false, "verbose output")
	rootCmd.PersistentFlags().Bool("dry-run", false, "preview changes without executing")

	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		panic(fmt.Errorf("failed to bind verbose flag: %w", err))
	}
	if err := viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run")); err != nil {
		panic(fmt.Errorf("failed to bind dry-run flag: %w", err))
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home + "/.auto-pr")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Set environment variable prefix
	viper.SetEnvPrefix("AUTO_PR")
	viper.AutomaticEnv()

	// Bind environment variables for AI configuration
	_ = viper.BindEnv("ai.provider", "AUTO_PR_AI_PROVIDER")
	_ = viper.BindEnv("ai.model", "AUTO_PR_AI_MODEL")
	_ = viper.BindEnv("ai.max_tokens", "AUTO_PR_AI_MAX_TOKENS")
	_ = viper.BindEnv("ai.temperature", "AUTO_PR_AI_TEMPERATURE")

	// Claude specific
	_ = viper.BindEnv("ai.claude.cli_path", "AUTO_PR_CLAUDE_CLI_PATH")
	_ = viper.BindEnv("ai.claude.model", "AUTO_PR_CLAUDE_MODEL")
	_ = viper.BindEnv("ai.claude.max_tokens", "AUTO_PR_CLAUDE_MAX_TOKENS")
	_ = viper.BindEnv("ai.claude.use_session", "AUTO_PR_CLAUDE_USE_SESSION")

	// GitHub configuration
	_ = viper.BindEnv("platforms.github.draft", "AUTO_PR_GITHUB_DRAFT")
	_ = viper.BindEnv("platforms.github.auto_merge", "AUTO_PR_GITHUB_AUTO_MERGE")
	_ = viper.BindEnv("platforms.github.delete_branch", "AUTO_PR_GITHUB_DELETE_BRANCH")

	// GitLab configuration
	_ = viper.BindEnv("platforms.gitlab.merge_when_pipeline_succeeds", "AUTO_PR_GITLAB_AUTO_MERGE")
	_ = viper.BindEnv("platforms.gitlab.remove_source_branch", "AUTO_PR_GITLAB_REMOVE_SOURCE_BRANCH")
	_ = viper.BindEnv("platforms.gitlab.default_assignee", "AUTO_PR_GITLAB_DEFAULT_ASSIGNEE")

	// Git configuration
	_ = viper.BindEnv("git.commit_limit", "AUTO_PR_GIT_COMMIT_LIMIT")
	_ = viper.BindEnv("git.diff_context", "AUTO_PR_GIT_DIFF_CONTEXT")
	_ = viper.BindEnv("git.max_diff_size", "AUTO_PR_GIT_MAX_DIFF_SIZE")

	// Template configuration
	_ = viper.BindEnv("templates.custom_templates_dir", "AUTO_PR_TEMPLATES_DIR")

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
