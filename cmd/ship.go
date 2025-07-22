package cmd

import (
	"fmt"
	"os/exec"
	"time"

	"auto-pr/internal/git"

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

	// SUPER SMART: If we're on main/master and have changes, create a feature branch first
	isOnDefaultBranch := status.CurrentBranch == "main" || status.CurrentBranch == "master"
	if isOnDefaultBranch && needsCommit {
		fmt.Println("🌿 On default branch with changes - creating feature branch...")
		
		// Generate feature branch name
		branchName := fmt.Sprintf("feature/auto-ship-%s", time.Now().Format("2006-01-02-15-04-05"))
		
		if dryRun {
			fmt.Printf("   Would create feature branch: %s\n", branchName)
		} else {
			if err := createFeatureBranch(branchName); err != nil {
				return fmt.Errorf("failed to create feature branch: %w", err)
			}
			fmt.Printf("✅ Created and switched to branch: %s\n", branchName)
		}
	}

	stepNum := 1

	// Step 1: Commit (only if needed)
	if needsCommit {
		fmt.Printf("📦 Step %d: Committing changes...\n", stepNum)
		
		if dryRun {
			fmt.Printf("   Would stage %d unstaged, %d untracked, %d staged files\n", 
				len(status.UnstagedFiles), len(status.UntrackedFiles), len(status.StagedFiles))
			if message == "" {
				fmt.Println("   Would generate AI commit message based on changes")
			} else {
				fmt.Printf("   Would commit with message: %s\n", message)
			}
		} else {
			// Create commit command with proper flags
			commitCmd := &cobra.Command{}
			commitCmd.Flags().Bool("all", true, "")
			commitCmd.Flags().String("message", message, "")
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
			fmt.Println("   Would create PR with AI-generated content")
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

// createFeatureBranch creates and switches to a new feature branch
func createFeatureBranch(branchName string) error {
	cmd := exec.Command("git", "checkout", "-b", branchName)
	return cmd.Run()
}

