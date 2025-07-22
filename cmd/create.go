package cmd

import (
	"fmt"
	"auto-pr/internal/git"
	"auto-pr/internal/platforms"
	
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a pull request or merge request",
	Long: `Analyze your git changes and create a pull request (GitHub) or merge request (GitLab)
with AI-generated title, description, and appropriate metadata.`,
	RunE: runCreate,
}

func init() {
	rootCmd.AddCommand(createCmd)
	
	createCmd.Flags().Bool("interactive", false, "Interactive mode with confirmation")
	createCmd.Flags().String("template", "", "Use specific template")
	createCmd.Flags().StringSlice("reviewer", []string{}, "Override default reviewers")
	createCmd.Flags().Bool("draft", false, "Create as draft")
	createCmd.Flags().Bool("auto-merge", false, "Enable auto-merge")
	createCmd.Flags().Bool("force", false, "Skip validations")
	createCmd.Flags().String("commit-range", "", "Specific commit range")
	createCmd.Flags().String("ai-context", "", "Additional context file")
	
	viper.BindPFlags(createCmd.Flags())
}

func runCreate(cmd *cobra.Command, args []string) error {
	verbose := viper.GetBool("verbose")
	dryRun := viper.GetBool("dry-run")
	
	if verbose {
		fmt.Println("Starting Auto PR creation...")
	}
	
	// Initialize git analyzer
	gitAnalyzer, err := git.NewAnalyzer(".")
	if err != nil {
		return fmt.Errorf("failed to initialize git analyzer: %w", err)
	}
	
	// Check if we're in a git repository
	if !gitAnalyzer.IsGitRepository() {
		return fmt.Errorf("not in a git repository")
	}
	
	// Detect platform (GitHub/GitLab)
	platform, err := platforms.DetectPlatform(gitAnalyzer.GetRemoteURL())
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}
	
	if verbose {
		fmt.Printf("Detected platform: %s\n", platform)
	}
	
	// Get repository status
	status, err := gitAnalyzer.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get repository status: %w", err)
	}
	
	if verbose {
		fmt.Printf("Repository status: %+v\n", status)
	}
	
	if dryRun {
		fmt.Println("Dry run mode - would create PR/MR here")
		return nil
	}
	
	fmt.Println("PR/MR creation not yet implemented - coming soon!")
	return nil
}