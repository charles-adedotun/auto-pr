package cmd

import (
	"fmt"
	"os"

	"auto-pr/internal/ai"
	"auto-pr/internal/config"
	"auto-pr/internal/git"
	"auto-pr/internal/platforms"
	"auto-pr/internal/templates"
	"auto-pr/pkg/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"pr", "mr"},
	Short:   "Create a pull request or merge request",
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

	if err := viper.BindPFlags(createCmd.Flags()); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to bind create flags: %v\n", err)
		os.Exit(1)
	}
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

	// Load configuration
	cfg, err := config.LoadConfigWithViper()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create AI client
	aiClient, err := ai.NewClient(cfg.AI)
	if err != nil {
		return fmt.Errorf("failed to create AI client: %w", err)
	}

	if verbose {
		fmt.Printf("Using AI provider: %s\n", aiClient.GetProvider())
	}

	// Get commit history and changes for AI context
	commits, err := gitAnalyzer.GetCommitsSinceBase(status.BaseBranch)
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	// Get diff summary
	diffSummary, err := gitAnalyzer.GetBranchDiff(status.BaseBranch)
	if err != nil {
		return fmt.Errorf("failed to get diff summary: %w", err)
	}

	// Build AI context
	aiContext := &ai.AIContext{
		CommitHistory: commits,
		DiffSummary: fmt.Sprintf("%d files changed, %d additions, %d deletions",
			diffSummary.TotalFiles, diffSummary.Additions, diffSummary.Deletions),
		FileChanges: diffSummary.FileChanges,
		BranchInfo: types.BranchInfo{
			Name:         status.CurrentBranch,
			BaseBranch:   status.BaseBranch,
			CommitsAhead: status.CommitsAhead,
		},
		Platform: platform,
	}

	if verbose {
		fmt.Printf("AI Context: %d commits, %d file changes\n",
			len(commits), len(diffSummary.FileChanges))
	}

	// Generate PR content using AI
	prompt := "Generate a comprehensive pull request title and description based on the provided git changes and commit history."
	aiResponse, err := aiClient.GenerateContent(aiContext, prompt)
	if err != nil {
		return fmt.Errorf("failed to generate AI content: %w", err)
	}

	if verbose {
		fmt.Printf("AI generated content (confidence: %.2f)\n", aiResponse.Confidence)
	}

	// Apply template if specified
	templateName := viper.GetString("template")
	if templateName != "" {
		templateManager := templates.NewManager()
		enhanced, err := templates.EnhanceWithTemplate(templateManager, templateName, aiContext, aiResponse)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: failed to apply template '%s': %v\n", templateName, err)
			}
		} else {
			aiResponse = enhanced
			if verbose {
				fmt.Printf("Applied template: %s\n", templateName)
			}
		}
	} else {
		// Auto-select template based on context
		templateManager := templates.NewManager()
		autoTemplate := templates.SelectTemplateByContext(aiContext)
		if autoTemplate != "" {
			enhanced, err := templates.EnhanceWithTemplate(templateManager, autoTemplate, aiContext, aiResponse)
			if err == nil {
				aiResponse = enhanced
				if verbose {
					fmt.Printf("Auto-selected template: %s\n", autoTemplate)
				}
			}
		}
	}

	if dryRun {
		fmt.Println("🔍 Dry Run - PR/MR Preview")
		fmt.Println("==========================")
		fmt.Printf("📝 Title: %s\n", aiResponse.Title)
		fmt.Printf("📋 Body:\n%s\n", aiResponse.Body)
		if len(aiResponse.Labels) > 0 {
			fmt.Printf("🏷️  Labels: %v\n", aiResponse.Labels)
		}
		if len(aiResponse.Reviewers) > 0 {
			fmt.Printf("👥 Suggested reviewers: %v\n", aiResponse.Reviewers)
		}
		fmt.Printf("⚡ Priority: %s\n", aiResponse.Priority)
		fmt.Printf("🤖 Generated by: %s\n", aiResponse.Provider)
		return nil
	}

	// Create platform client
	var platformClient platforms.PlatformClient
	switch platform {
	case types.PlatformGitHub:
		platformClient, err = platforms.NewGitHubClient(status.RemoteURL)
	case types.PlatformGitLab:
		platformClient, err = platforms.NewGitLabClient(status.RemoteURL)
	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	if err != nil {
		return fmt.Errorf("failed to create platform client: %w", err)
	}

	// Check for existing PR/MR
	existingPR, err := platformClient.GetExistingPR(status.CurrentBranch)
	if err != nil {
		if verbose {
			fmt.Printf("Warning: failed to check for existing PR: %s\n", err)
		}
	}

	if existingPR != nil {
		fmt.Printf("⚠️  A PR/MR already exists for branch '%s': %s\n",
			status.CurrentBranch, existingPR.URL)
		return nil
	}

	// Filter AI-suggested labels to only those that exist in the repository,
	// so we don't attempt to apply a label that hasn't been created yet.
	labels, err := platforms.FilterExistingLabels(platformClient, aiResponse.Labels)
	if err != nil {
		if verbose {
			fmt.Printf("Warning: failed to verify labels, skipping: %v\n", err)
		}
		labels = []string{}
	}

	reviewers := aiResponse.Reviewers
	if len(cfg.Platforms.GitHub.DefaultReviewers) > 0 && platform == types.PlatformGitHub {
		reviewers = append(reviewers, cfg.Platforms.GitHub.DefaultReviewers...)
	}

	// Create PR request
	prRequest := &types.PullRequestRequest{
		Title:      aiResponse.Title,
		Body:       aiResponse.Body,
		HeadBranch: status.CurrentBranch,
		BaseBranch: status.BaseBranch,
		Draft:      viper.GetBool("draft"),
		Labels:     removeDuplicates(labels),
		Reviewers:  removeDuplicates(reviewers),
		AutoMerge:  viper.GetBool("auto-merge"),
	}

	// Create the PR/MR
	fmt.Println("🚀 Creating PR/MR...")
	createdPR, err := platformClient.CreatePullRequest(prRequest)
	if err != nil {
		return fmt.Errorf("failed to create PR/MR: %w", err)
	}

	fmt.Printf("✅ Successfully created %s: %s\n",
		getEntityName(platform), createdPR.URL)
	fmt.Printf("📝 Title: %s\n", createdPR.Title)
	if createdPR.Draft {
		fmt.Println("📋 Status: Draft")
	}

	return nil
}

// Helper functions
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	var result []string
	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

func getEntityName(platform types.PlatformType) string {
	switch platform {
	case types.PlatformGitHub:
		return "Pull Request"
	case types.PlatformGitLab:
		return "Merge Request"
	default:
		return "PR/MR"
	}
}
