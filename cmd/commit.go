package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"auto-pr/internal/ai"
	"auto-pr/internal/config"
	"auto-pr/internal/git"
	"auto-pr/pkg/types"

	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:     "commit",
	Aliases: []string{"cm", "c"},
	Short:   "Smart commit with AI-generated message",
	Long:    `Automatically stage changes and create a commit with AI-generated message based on your changes`,
	RunE:    runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)
	
	commitCmd.Flags().BoolP("all", "a", false, "Stage all changes before committing")
	commitCmd.Flags().StringP("message", "m", "", "Custom commit message (skips AI generation)")
	commitCmd.Flags().Bool("amend", false, "Amend the last commit")
	commitCmd.Flags().Bool("push", false, "Push after committing")
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Initialize git analyzer
	gitAnalyzer, err := git.NewAnalyzer(".")
	if err != nil {
		return fmt.Errorf("failed to initialize git analyzer: %w", err)
	}

	if !gitAnalyzer.IsGitRepository() {
		return fmt.Errorf("not in a git repository")
	}

	// Get flags
	stageAll, _ := cmd.Flags().GetBool("all")
	customMessage, _ := cmd.Flags().GetString("message")
	amend, _ := cmd.Flags().GetBool("amend")
	pushAfter, _ := cmd.Flags().GetBool("push")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Get repository status first
	status, err := gitAnalyzer.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get repository status: %w", err)
	}

	// Stage files if requested
	if stageAll {
		if dryRun {
			fmt.Printf("🔄 Would stage %d unstaged and %d untracked files\n", 
				len(status.UnstagedFiles), len(status.UntrackedFiles))
		} else {
			fmt.Println("🔄 Staging all changes...")
			if err := stageAllChanges(); err != nil {
				return fmt.Errorf("failed to stage changes: %w", err)
			}
			// Refresh status after staging
			status, err = gitAnalyzer.GetStatus()
			if err != nil {
				return fmt.Errorf("failed to get updated repository status: %w", err)
			}
			fmt.Println("✅ Changes staged")
		}
	}

	if len(status.StagedFiles) == 0 && !amend && !stageAll {
		return fmt.Errorf("no changes staged for commit. Use --all to stage all changes")
	}

	var commitMessage string
	
	if customMessage != "" {
		commitMessage = customMessage
	} else {
		fmt.Println("🤖 Generating commit message with AI...")
		
		// Generate AI commit message
		commitMessage, err = generateCommitMessage(gitAnalyzer, status)
		if err != nil {
			return fmt.Errorf("failed to generate commit message: %w", err)
		}
	}

	fmt.Printf("📝 Commit message:\n%s\n\n", commitMessage)

	if dryRun {
		fmt.Println("🔍 Dry run - would commit with above message")
		return nil
	}

	// Create the commit
	fmt.Println("💾 Creating commit...")
	if err := createCommit(commitMessage, amend); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	fmt.Println("✅ Commit created successfully!")

	// Push if requested
	if pushAfter {
		fmt.Println("🚀 Pushing to remote...")
		if err := pushChanges(); err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
		fmt.Println("✅ Changes pushed!")
	}

	return nil
}

func stageAllChanges() error {
	cmd := exec.Command("git", "add", ".")
	return cmd.Run()
}

func createCommit(message string, amend bool) error {
	args := []string{"commit", "-m", message}
	if amend {
		args = []string{"commit", "--amend", "-m", message}
	}
	
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func pushChanges() error {
	cmd := exec.Command("git", "push")
	return cmd.Run()
}

func generateCommitMessage(gitAnalyzer *git.Analyzer, status *types.GitStatus) (string, error) {
	// Load configuration
	cfg, err := config.LoadConfigWithViper()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	// Create AI client
	client, err := ai.NewClient(cfg.AI)
	if err != nil {
		return "", fmt.Errorf("failed to create AI client: %w", err)
	}

	// Get diff for staged files
	diffSummary, err := getStagedDiff()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	// Build AI context
	context := &ai.AIContext{
		DiffSummary: diffSummary,
		FileChanges: buildFileChanges(status),
		BranchInfo: types.BranchInfo{
			Name:       status.CurrentBranch,
			BaseBranch: status.BaseBranch,
		},
	}

	// Generate commit message
	prompt := `Generate a concise, clear commit message for these changes.

Rules:
- Use conventional commit format (feat:, fix:, docs:, refactor:, etc.)
- First line should be 50 characters or less
- Be specific about what changed
- Don't include explanations, just the action

Example formats:
- feat: add user authentication
- fix: resolve memory leak in parser
- docs: update API documentation
- refactor: simplify error handling

Focus on WHAT changed, not HOW or WHY.`

	response, err := client.GenerateContent(context, prompt)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	// Extract just the commit message (first line of the response)
	lines := strings.Split(strings.TrimSpace(response.Title), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return response.Title, nil
}

func getStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--stat")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func buildFileChanges(status *types.GitStatus) []types.FileChange {
	var changes []types.FileChange
	
	for _, file := range status.StagedFiles {
		changes = append(changes, types.FileChange{
			Path:   file,
			Status: types.StatusModified, // Simplified for now
		})
	}
	
	return changes
}