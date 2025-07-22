package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"auto-pr/internal/ai"
	"auto-pr/internal/config"
	"auto-pr/internal/git"
	"auto-pr/pkg/types"

	"github.com/spf13/cobra"
)

var shipCmd = &cobra.Command{
	Use:     "ship",
	Aliases: []string{"go", "send", "deploy"},
	Short:   "One-command workflow: stage → commit → push → create PR",
	Long: `The ultimate shortcut! This command will:
1. Stage all your changes (git add .)
2. Create a commit with AI-generated message
3. Push to remote
4. Create a pull request with AI-generated content

Perfect for when you just want to ship your changes quickly!`,
	RunE: runShip,
}

func init() {
	rootCmd.AddCommand(shipCmd)
	
	shipCmd.Flags().StringP("message", "m", "", "Custom commit message (skips AI generation)")
	shipCmd.Flags().Bool("draft", false, "Create PR as draft")
	shipCmd.Flags().StringSlice("reviewer", []string{}, "Add reviewers to the PR")
	shipCmd.Flags().Bool("no-push", false, "Don't push to remote (just commit)")
	shipCmd.Flags().Bool("no-pr", false, "Don't create PR (just commit and push)")
}

func runShip(cmd *cobra.Command, args []string) error {
	// Get flags
	message, _ := cmd.Flags().GetString("message")
	draft, _ := cmd.Flags().GetBool("draft")
	reviewers, _ := cmd.Flags().GetStringSlice("reviewer")
	noPush, _ := cmd.Flags().GetBool("no-push")
	noPR, _ := cmd.Flags().GetBool("no-pr")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	fmt.Println("🚀 Starting the ship workflow!")

	// Initialize git analyzer to check what needs to be done
	gitAnalyzer, err := git.NewAnalyzer(".")
	if err != nil {
		return fmt.Errorf("failed to initialize git analyzer: %w", err)
	}

	if !gitAnalyzer.IsGitRepository() {
		return fmt.Errorf("not in a git repository")
	}

	// Get current status
	status, err := gitAnalyzer.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get repository status: %w", err)
	}

	// 🧠 SMART: Analyze everything upfront and get comprehensive AI plan
	fmt.Println("🧠 Analyzing changes and generating comprehensive plan...")
	
	workflowPlan, err := generateWorkflowPlan(gitAnalyzer, status, message, dryRun)
	if err != nil {
		return fmt.Errorf("failed to generate workflow plan: %w", err)
	}

	// Smart workflow - only do what's needed
	needsCommit := len(status.UnstagedFiles) > 0 || len(status.UntrackedFiles) > 0 || len(status.StagedFiles) > 0
	needsPush := status.CommitsAhead > 0 // Will be true after we commit
	
	if !needsCommit && status.CommitsAhead == 0 {
		fmt.Println("📭 No changes to ship - working directory is clean and up to date")
		return nil
	}

	// SUPER SMART: If we're on main/master and have changes, create a feature branch first
	if workflowPlan.NeedsBranch && needsCommit {
		fmt.Println("🌿 On default branch with changes - creating feature branch...")
		
		if dryRun {
			fmt.Printf("   Would create feature branch: %s\n", workflowPlan.BranchName)
		} else {
			if err := createFeatureBranch(workflowPlan.BranchName); err != nil {
				return fmt.Errorf("failed to create feature branch: %w", err)
			}
			fmt.Printf("✅ Created and switched to branch: %s\n", workflowPlan.BranchName)
		}
	}

	stepNum := 1

	// Step 1: Commit (only if needed)
	if needsCommit {
		fmt.Printf("📦 Step %d: Committing changes...\n", stepNum)
		
		if dryRun {
			fmt.Printf("   Would stage %d unstaged, %d untracked, %d staged files\n", 
				len(status.UnstagedFiles), len(status.UntrackedFiles), len(status.StagedFiles))
			if workflowPlan.CommitMessage != "" {
				fmt.Printf("   Would commit with AI message: %s\n", workflowPlan.CommitMessage)
			} else {
				fmt.Println("   Would generate AI commit message based on changes")
			}
		} else {
			// Create commit command with proper flags
			commitCmd := &cobra.Command{}
			commitCmd.Flags().Bool("all", true, "")
			// Use AI-generated message if available, otherwise use provided message
			commitMessage := message
			if workflowPlan.CommitMessage != "" {
				commitMessage = workflowPlan.CommitMessage
			}
			commitCmd.Flags().String("message", commitMessage, "")
			commitCmd.Flags().Bool("dry-run", false, "") // We handle dry-run here
			
			if err := runCommit(commitCmd, []string{}); err != nil {
				return fmt.Errorf("commit failed: %w", err)
			}
		}
		stepNum++
		needsPush = true // We just committed, so we need to push
	}

	// Step 2: Push (only if needed and not disabled)
	if needsPush && !noPush {
		fmt.Printf("🌐 Step %d: Pushing to remote...\n", stepNum)
		
		if dryRun {
			fmt.Println("   Would push commits to remote")
		} else {
			if err := pushChanges(); err != nil {
				return fmt.Errorf("failed to push: %w", err)
			}
			fmt.Println("✅ Pushed to remote")
		}
		stepNum++
	}

	// Step 3: Create PR (only if not disabled)
	if !noPR {
		fmt.Printf("🔀 Step %d: Creating pull request...\n", stepNum)
		
		if dryRun {
			fmt.Printf("   Would create PR with title: %s\n", workflowPlan.PRTitle)
			if len(workflowPlan.Labels) > 0 {
				fmt.Printf("   Would add labels: %v\n", workflowPlan.Labels)
			}
		} else {
			// Create PR using the comprehensive workflow plan data
			// Note: For now we use the existing create command, but in future we could
			// directly use the workflow plan data to create PR with pre-generated content
			createCmd := &cobra.Command{}
			createCmd.Flags().Bool("draft", draft, "")
			createCmd.Flags().StringSlice("reviewer", reviewers, "")
			createCmd.Flags().Bool("dry-run", false, "") // We handle dry-run here
			
			if err := runCreate(createCmd, []string{}); err != nil {
				return fmt.Errorf("PR creation failed: %w", err)
			}
		}
	}

	if dryRun {
		fmt.Println("🔍 Dry run complete - no changes made")
	} else {
		fmt.Println("🎉 Ship complete! Your changes are live!")
		
		if noPR {
			fmt.Println("   💡 Run 'auto-pr pr' to create a pull request")
		}
	}

	return nil
}

// generateFeatureBranchName creates a smart branch name using environment variables and defaults
func generateFeatureBranchName(gitAnalyzer *git.Analyzer, status *types.GitStatus) (string, error) {
	// Use environment variables for customization
	branchPrefix := getEnvOrDefault("AUTO_PR_BRANCH_PREFIX", "feature")
	branchFormat := getEnvOrDefault("AUTO_PR_BRANCH_FORMAT", "auto-ship")
	
	// Smart timestamp-based naming (configurable via env vars)
	timestamp := time.Now().Format(getEnvOrDefault("AUTO_PR_BRANCH_TIME_FORMAT", "2006-01-02-15-04"))
	
	// Create branch name
	branchName := fmt.Sprintf("%s/%s-%s", branchPrefix, branchFormat, timestamp)
	
	return branchName, nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// WorkflowPlan contains all the data needed for the entire ship workflow
type WorkflowPlan struct {
	BranchName    string   `json:"branch_name"`
	CommitMessage string   `json:"commit_message"`
	PRTitle       string   `json:"pr_title"`
	PRBody        string   `json:"pr_body"`
	Labels        []string `json:"labels"`
	Priority      string   `json:"priority"`
	NeedsBranch   bool     `json:"needs_branch"`
	NeedsCommit   bool     `json:"needs_commit"`
	NeedsPush     bool     `json:"needs_push"`
}

// generateWorkflowPlan analyzes everything and creates a comprehensive plan with one AI call
func generateWorkflowPlan(gitAnalyzer *git.Analyzer, status *types.GitStatus, customMessage string, dryRun bool) (*WorkflowPlan, error) {
	// If custom message provided, skip AI for commit message
	if customMessage != "" && !isOnDefaultBranch(status.CurrentBranch) {
		return &WorkflowPlan{
			BranchName:    "", // Not needed if not on default branch
			CommitMessage: customMessage,
			PRTitle:       "Pull Request", // Will be generated later
			PRBody:        "",
			Labels:        []string{},
			NeedsBranch:   false,
			NeedsCommit:   hasChangesToCommit(status),
			NeedsPush:     hasChangesToPush(status),
		}, nil
	}

	// Load AI configuration
	cfg, err := config.LoadConfigWithViper()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create AI client
	client, err := ai.NewClient(cfg.AI)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	// Get comprehensive context
	diffContent, _ := getGitDiffContent()
	fileList := append(status.UnstagedFiles, status.UntrackedFiles...)
	isOnDefault := isOnDefaultBranch(status.CurrentBranch)

	// Build comprehensive AI context
	context := &ai.AIContext{
		DiffSummary: diffContent,
		FileChanges: buildFileChangesFromStatus(status),
		BranchInfo: types.BranchInfo{
			Name:       status.CurrentBranch,
			BaseBranch: status.BaseBranch,
		},
	}

	// Dynamic prompt based on current state
	prompt := buildDynamicPrompt(status, diffContent, fileList, isOnDefault)

	// Get AI response
	response, err := client.GenerateContent(context, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse JSON response
	var plan WorkflowPlan
	if err := json.Unmarshal([]byte(response.Title), &plan); err != nil {
		// Fallback to basic plan if JSON parsing fails
		branchName := ""
		if isOnDefault {
			timestamp := time.Now().Format("2006-01-02-15-04")
			branchName = fmt.Sprintf("feature/auto-ship-%s", timestamp)
		}
		
		return &WorkflowPlan{
			BranchName:    branchName,
			CommitMessage: "feat: auto-generated commit",
			PRTitle:       "Auto-generated PR",
			PRBody:        response.Body,
			Labels:        []string{"auto-generated"},
			NeedsBranch:   isOnDefault,
			NeedsCommit:   hasChangesToCommit(status),
			NeedsPush:     hasChangesToPush(status),
		}, nil
	}

	return &plan, nil
}

// Helper functions
func isOnDefaultBranch(branch string) bool {
	return branch == "main" || branch == "master"
}

func hasChangesToCommit(status *types.GitStatus) bool {
	return len(status.UnstagedFiles) > 0 || len(status.UntrackedFiles) > 0 || len(status.StagedFiles) > 0
}

func hasChangesToPush(status *types.GitStatus) bool {
	return status.CommitsAhead > 0
}

func getGitDiffContent() (string, error) {
	cmd := exec.Command("git", "diff", "--stat")
	output, err := cmd.Output()
	return string(output), err
}

func buildFileChangesFromStatus(status *types.GitStatus) []types.FileChange {
	var changes []types.FileChange
	
	for _, file := range status.UnstagedFiles {
		changes = append(changes, types.FileChange{
			Path:   file,
			Status: types.StatusModified,
		})
	}
	
	for _, file := range status.UntrackedFiles {
		changes = append(changes, types.FileChange{
			Path:   file,
			Status: types.StatusUntracked,
		})
	}
	
	for _, file := range status.StagedFiles {
		changes = append(changes, types.FileChange{
			Path:   file,
			Status: types.StatusModified,
		})
	}
	
	return changes
}

// buildDynamicPrompt creates a dynamic prompt based on current repository state
func buildDynamicPrompt(status *types.GitStatus, diffContent string, fileList []string, isOnDefault bool) string {
	var promptBuilder strings.Builder
	
	// Base analysis
	promptBuilder.WriteString("Analyze the current git repository state and generate a comprehensive workflow plan.\n\n")
	
	// Dynamic state analysis
	promptBuilder.WriteString("CURRENT SITUATION:\n")
	promptBuilder.WriteString(fmt.Sprintf("- Current branch: %s\n", status.CurrentBranch))
	
	if isOnDefault {
		promptBuilder.WriteString("- STATUS: On default branch - will need to create feature branch\n")
	} else {
		promptBuilder.WriteString("- STATUS: On feature branch - can commit directly\n")
	}
	
	// Analyze what needs to be done
	hasUnstaged := len(status.UnstagedFiles) > 0
	hasUntracked := len(status.UntrackedFiles) > 0
	hasStaged := len(status.StagedFiles) > 0
	hasCommitsAhead := status.CommitsAhead > 0
	
	promptBuilder.WriteString("\nWORKFLOW REQUIREMENTS:\n")
	
	if hasUnstaged || hasUntracked {
		promptBuilder.WriteString(fmt.Sprintf("- COMMIT NEEDED: %d unstaged, %d untracked files\n", 
			len(status.UnstagedFiles), len(status.UntrackedFiles)))
	}
	if hasStaged {
		promptBuilder.WriteString(fmt.Sprintf("- STAGED READY: %d files already staged\n", len(status.StagedFiles)))
	}
	if hasCommitsAhead {
		promptBuilder.WriteString(fmt.Sprintf("- PUSH NEEDED: %d commits ahead of remote\n", status.CommitsAhead))
	}
	
	// Files being changed
	if len(fileList) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("\nFILES AFFECTED: %s\n", strings.Join(fileList, ", ")))
	}
	
	// Diff content if available
	if diffContent != "" {
		promptBuilder.WriteString(fmt.Sprintf("\nCHANGES SUMMARY:\n%s\n", diffContent))
	}
	
	// Dynamic task based on state
	promptBuilder.WriteString("\nTASK: Provide a JSON response with the required workflow elements:\n")
	
	var jsonFields []string
	
	if isOnDefault {
		jsonFields = append(jsonFields, `"branch_name": "meaningful-branch-name-based-on-changes"`)
	}
	
	if hasUnstaged || hasUntracked || hasStaged {
		jsonFields = append(jsonFields, `"commit_message": "conventional commit message based on actual changes"`)
	}
	
	// Always need PR elements
	jsonFields = append(jsonFields, 
		`"pr_title": "Clear descriptive title"`,
		`"pr_body": "## Summary\n\nDetailed description..."`,
		`"labels": ["appropriate", "labels"]`,
	)
	
	// Add boolean flags for what's needed
	jsonFields = append(jsonFields,
		fmt.Sprintf(`"needs_branch": %v`, isOnDefault),
		fmt.Sprintf(`"needs_commit": %v`, hasUnstaged || hasUntracked || hasStaged),
		fmt.Sprintf(`"needs_push": %v`, hasCommitsAhead || hasUnstaged || hasUntracked || hasStaged),
	)
	
	promptBuilder.WriteString("\nRESPOND WITH ONLY THIS JSON:\n{\n  ")
	promptBuilder.WriteString(strings.Join(jsonFields, ",\n  "))
	promptBuilder.WriteString("\n}")
	
	return promptBuilder.String()
}

// createFeatureBranch creates and switches to a new feature branch
func createFeatureBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	return cmd.Run()
}

