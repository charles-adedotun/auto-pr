package cmd

import (
	"encoding/json"
	"fmt"
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

	// Smart workflow - only do what's needed
	needsCommit := len(status.UnstagedFiles) > 0 || len(status.UntrackedFiles) > 0 || len(status.StagedFiles) > 0
	needsPush := status.CommitsAhead > 0 // Will be true after we commit
	canCreatePR := needsCommit || status.CommitsAhead > 0 // Can create PR if we have changes or unpushed commits
	
	if !canCreatePR {
		fmt.Println("📭 No changes to ship - working directory is clean and up to date")
		return nil
	}

	// 🧠 SMART: Generate comprehensive AI plan upfront for all workflow data
	fmt.Println("🧠 Analyzing changes and generating comprehensive workflow plan...")
	
	workflowPlan, err := generateComprehensiveWorkflowPlan(gitAnalyzer, status, message, dryRun)
	if err != nil {
		fmt.Printf("⚠️  Failed to generate AI workflow plan: %v\n", err)
		// Continue with fallback behavior
		workflowPlan = &WorkflowPlan{
			BranchName:    fmt.Sprintf("feature/auto-ship-%s", time.Now().Format("2006-01-02-15-04-05")),
			CommitMessage: "feat: auto-generated commit",
			PRTitle:       "Auto-generated PR",
			PRBody:        "Auto-generated changes",
			Labels:        []string{"auto-generated"},
			NeedsBranch:   status.CurrentBranch == "main" || status.CurrentBranch == "master",
			NeedsCommit:   needsCommit,
			NeedsPush:     needsCommit, // Will be true after commit
		}
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
			commitMsg := message
			if commitMsg == "" && workflowPlan.CommitMessage != "" {
				commitMsg = workflowPlan.CommitMessage
			}
			fmt.Printf("   Would commit with message: %s\n", commitMsg)
		} else {
			// Create commit command with proper flags
			commitCmd := &cobra.Command{}
			commitCmd.Flags().Bool("all", true, "")
			// Use AI-generated commit message if no custom message provided
			commitMsg := message
			if commitMsg == "" && workflowPlan.CommitMessage != "" {
				commitMsg = workflowPlan.CommitMessage
			}
			commitCmd.Flags().String("message", commitMsg, "")
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
			if workflowPlan.PRBody != "" {
				fmt.Printf("   PR body preview: %s\n", truncateString(workflowPlan.PRBody, 100))
			}
			if len(workflowPlan.Labels) > 0 {
				fmt.Printf("   Would add labels: %v\n", workflowPlan.Labels)
			}
		} else {
			// Create PR command with proper flags
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

// WorkflowPlan contains all AI-generated data needed for the complete ship workflow
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

// generateComprehensiveWorkflowPlan creates a complete plan with ONE AI call
func generateComprehensiveWorkflowPlan(gitAnalyzer *git.Analyzer, status *types.GitStatus, customMessage string, dryRun bool) (*WorkflowPlan, error) {
	// If custom message provided and not on default branch, minimal AI needed
	if customMessage != "" && status.CurrentBranch != "main" && status.CurrentBranch != "master" {
		return &WorkflowPlan{
			BranchName:    "",
			CommitMessage: customMessage,
			PRTitle:       "Pull Request",
			PRBody:        "Changes made",
			Labels:        []string{},
			NeedsBranch:   false,
			NeedsCommit:   len(status.UnstagedFiles) > 0 || len(status.UntrackedFiles) > 0 || len(status.StagedFiles) > 0,
			NeedsPush:     status.CommitsAhead > 0,
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

	// Get comprehensive context for AI
	diffContent, _ := getGitDiffContent()
	isOnDefault := status.CurrentBranch == "main" || status.CurrentBranch == "master"
	
	// Analyze existing branch patterns for intelligent naming
	branchPattern, _ := analyzeExistingBranchPatterns()

	// Build comprehensive AI context
	context := &ai.AIContext{
		DiffSummary: diffContent,
		FileChanges: buildFileChangesFromStatus(status),
		BranchInfo: types.BranchInfo{
			Name:       status.CurrentBranch,
			BaseBranch: status.BaseBranch,
		},
	}

	// Dynamic prompt based on repository state and context
	prompt := buildComprehensiveWorkflowPrompt(status, diffContent, branchPattern, isOnDefault)

	// Single AI call to get everything
	response, err := client.GenerateContent(context, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Parse the comprehensive JSON response
	var plan WorkflowPlan
	if err := parseAIResponse(response.Title, &plan); err != nil {
		// Fallback with meaningful defaults
		return &WorkflowPlan{
			BranchName:    generateFallbackBranchName(diffContent, branchPattern),
			CommitMessage: generateFallbackCommitMessage(diffContent),
			PRTitle:       "Update repository changes",
			PRBody:        response.Body,
			Labels:        []string{"enhancement"},
			NeedsBranch:   isOnDefault,
			NeedsCommit:   len(status.UnstagedFiles) > 0 || len(status.UntrackedFiles) > 0 || len(status.StagedFiles) > 0,
			NeedsPush:     status.CommitsAhead > 0,
		}, nil
	}

	// Set workflow flags
	plan.NeedsBranch = isOnDefault && (len(status.UnstagedFiles) > 0 || len(status.UntrackedFiles) > 0 || len(status.StagedFiles) > 0)
	plan.NeedsCommit = len(status.UnstagedFiles) > 0 || len(status.UntrackedFiles) > 0 || len(status.StagedFiles) > 0
	plan.NeedsPush = status.CommitsAhead > 0

	return &plan, nil
}

// createFeatureBranch creates and switches to a new feature branch
func createFeatureBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	return cmd.Run()
}

// Helper functions
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
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

func analyzeExistingBranchPatterns() (string, error) {
	cmd := exec.Command("git", "branch", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	branches := strings.Split(string(output), "\n")
	patterns := make(map[string]int)
	
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" || strings.Contains(branch, "HEAD") {
			continue
		}
		
		// Extract pattern (feature/, fix/, docs/, etc.)
		if strings.Contains(branch, "/") {
			parts := strings.Split(branch, "/")
			if len(parts) >= 2 {
				pattern := parts[1] // Skip origin/ part
				if strings.Contains(pattern, "/") {
					prefix := strings.Split(pattern, "/")[0]
					patterns[prefix+"/"]++
				}
			}
		}
	}
	
	// Find most common pattern
	mostCommon := "feature/"
	maxCount := 0
	for pattern, count := range patterns {
		if count > maxCount {
			mostCommon = pattern
			maxCount = count
		}
	}
	
	return mostCommon, nil
}

func buildComprehensiveWorkflowPrompt(status *types.GitStatus, diffContent, branchPattern string, isOnDefault bool) string {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString("Analyze the repository changes and generate a comprehensive workflow plan.\n\n")
	
	// Repository context
	promptBuilder.WriteString("REPOSITORY CONTEXT:\n")
	promptBuilder.WriteString(fmt.Sprintf("- Current branch: %s\n", status.CurrentBranch))
	promptBuilder.WriteString(fmt.Sprintf("- Common branch pattern: %s\n", branchPattern))
	if isOnDefault {
		promptBuilder.WriteString("- STATUS: On default branch - will create feature branch\n")
	} else {
		promptBuilder.WriteString("- STATUS: On feature branch - can commit directly\n")
	}
	
	// Changes context
	promptBuilder.WriteString("\nCHANGES ANALYSIS:\n")
	if diffContent != "" {
		promptBuilder.WriteString(fmt.Sprintf("Diff summary:\n%s\n", diffContent))
	}
	
	files := append(status.UnstagedFiles, status.UntrackedFiles...)
	if len(files) > 0 {
		promptBuilder.WriteString(fmt.Sprintf("Files affected: %s\n", strings.Join(files, ", ")))
	}
	
	// Task specification
	promptBuilder.WriteString("\nTASK: Generate a JSON response with ALL workflow elements:\n")
	
	var jsonFields []string
	
	if isOnDefault {
		promptBuilder.WriteString(fmt.Sprintf("- Create meaningful branch name using pattern '%s' based on actual changes\n", branchPattern))
		jsonFields = append(jsonFields, `"branch_name": "meaningful-name-based-on-changes"`)
	}
	
	jsonFields = append(jsonFields,
		`"commit_message": "conventional commit message based on actual changes"`,
		`"pr_title": "Clear descriptive title"`,
		`"pr_body": "## Summary\n\nDetailed description of changes\n\n## Changes\n- List key changes"`,
		`"labels": ["appropriate", "labels"]`,
		`"priority": "medium"`,
	)
	
	promptBuilder.WriteString("\nRESPOND WITH ONLY THIS JSON:\n{\n  ")
	promptBuilder.WriteString(strings.Join(jsonFields, ",\n  "))
	promptBuilder.WriteString("\n}")
	
	return promptBuilder.String()
}

func parseAIResponse(response string, plan *WorkflowPlan) error {
	// Try to parse as JSON first
	if strings.Contains(response, "{") {
		startIndex := strings.Index(response, "{")
		endIndex := strings.LastIndex(response, "}")
		if startIndex >= 0 && endIndex > startIndex {
			jsonStr := response[startIndex : endIndex+1]
			if err := json.Unmarshal([]byte(jsonStr), plan); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("failed to parse JSON response")
}

func generateFallbackBranchName(diffContent, pattern string) string {
	// Extract meaningful name from diff if possible
	if strings.Contains(diffContent, "README") {
		return pattern + "update-readme"
	}
	if strings.Contains(diffContent, ".go") {
		return pattern + "update-go-code"
	}
	if strings.Contains(diffContent, "test") {
		return pattern + "update-tests"
	}
	if strings.Contains(diffContent, "doc") {
		return pattern + "update-docs"
	}
	
	// Generic fallback
	return pattern + "update-changes"
}

func generateFallbackCommitMessage(diffContent string) string {
	if strings.Contains(diffContent, "README") {
		return "docs: update README"
	}
	if strings.Contains(diffContent, "test") {
		return "test: update tests"
	}
	if strings.Contains(diffContent, "fix") {
		return "fix: resolve issues"
	}
	
	return "feat: update implementation"
}

