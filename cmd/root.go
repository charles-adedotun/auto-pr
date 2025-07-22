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
	viper.BindEnv("ai.provider", "AUTO_PR_AI_PROVIDER")
	viper.BindEnv("ai.model", "AUTO_PR_AI_MODEL")
	viper.BindEnv("ai.api_key", "AUTO_PR_AI_API_KEY")
	viper.BindEnv("ai.max_tokens", "AUTO_PR_AI_MAX_TOKENS")
	viper.BindEnv("ai.temperature", "AUTO_PR_AI_TEMPERATURE")
	viper.BindEnv("ai.project_id", "AUTO_PR_AI_PROJECT_ID")

	// Claude specific
	viper.BindEnv("ai.claude.cli_path", "AUTO_PR_CLAUDE_CLI_PATH")
	viper.BindEnv("ai.claude.model", "AUTO_PR_CLAUDE_MODEL")
	viper.BindEnv("ai.claude.max_tokens", "AUTO_PR_CLAUDE_MAX_TOKENS")
	viper.BindEnv("ai.claude.use_session", "AUTO_PR_CLAUDE_USE_SESSION")

	// Gemini specific
	viper.BindEnv("ai.gemini.api_key", "AUTO_PR_GEMINI_API_KEY", "GEMINI_API_KEY")
	viper.BindEnv("ai.gemini.project_id", "AUTO_PR_GEMINI_PROJECT_ID")
	viper.BindEnv("ai.gemini.model", "AUTO_PR_GEMINI_MODEL")
	viper.BindEnv("ai.gemini.max_tokens", "AUTO_PR_GEMINI_MAX_TOKENS")
	viper.BindEnv("ai.gemini.temperature", "AUTO_PR_GEMINI_TEMPERATURE")

	// GitHub configuration
	viper.BindEnv("platforms.github.draft", "AUTO_PR_GITHUB_DRAFT")
	viper.BindEnv("platforms.github.auto_merge", "AUTO_PR_GITHUB_AUTO_MERGE")
	viper.BindEnv("platforms.github.delete_branch", "AUTO_PR_GITHUB_DELETE_BRANCH")

	// GitLab configuration
	viper.BindEnv("platforms.gitlab.merge_when_pipeline_succeeds", "AUTO_PR_GITLAB_AUTO_MERGE")
	viper.BindEnv("platforms.gitlab.remove_source_branch", "AUTO_PR_GITLAB_REMOVE_SOURCE_BRANCH")
	viper.BindEnv("platforms.gitlab.default_assignee", "AUTO_PR_GITLAB_DEFAULT_ASSIGNEE")

	// Git configuration
	viper.BindEnv("git.commit_limit", "AUTO_PR_GIT_COMMIT_LIMIT")
	viper.BindEnv("git.diff_context", "AUTO_PR_GIT_DIFF_CONTEXT")
	viper.BindEnv("git.max_diff_size", "AUTO_PR_GIT_MAX_DIFF_SIZE")

	// Template configuration
	viper.BindEnv("templates.custom_templates_dir", "AUTO_PR_TEMPLATES_DIR")

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
