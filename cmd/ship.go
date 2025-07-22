package cmd

import (
	"fmt"

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
	fmt.Println("🚀 Starting the ship workflow!")
	
	// Get flags
	message, _ := cmd.Flags().GetString("message")
	draft, _ := cmd.Flags().GetBool("draft")
	reviewers, _ := cmd.Flags().GetStringSlice("reviewer")
	noPush, _ := cmd.Flags().GetBool("no-push")
	noPR, _ := cmd.Flags().GetBool("no-pr")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	fmt.Println("📦 Step 1: Staging all changes...")
	
	// Stage all changes
	if err := stageAllChanges(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}
	fmt.Println("✅ Changes staged")

	fmt.Println("💾 Step 2: Creating commit...")
	
	// Create commit (reuse the commit command logic)
	commitArgs := []string{}
	if message != "" {
		commitArgs = append(commitArgs, "--message", message)
	}
	if dryRun {
		commitArgs = append(commitArgs, "--dry-run")
	}

	// Call the commit command
	commitCmd.SetArgs(commitArgs)
	if err := commitCmd.Execute(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	if dryRun {
		fmt.Println("🔍 Dry run - would continue with push and PR creation")
		return nil
	}

	if !noPush {
		fmt.Println("🌐 Step 3: Pushing to remote...")
		if err := pushChanges(); err != nil {
			return fmt.Errorf("failed to push: %w", err)
		}
		fmt.Println("✅ Pushed to remote")
	}

	if !noPR {
		fmt.Println("🔀 Step 4: Creating pull request...")
		
		// Build create command args
		createArgs := []string{}
		if draft {
			createArgs = append(createArgs, "--draft")
		}
		for _, reviewer := range reviewers {
			createArgs = append(createArgs, "--reviewer", reviewer)
		}

		// Call the create command
		createCmd.SetArgs(createArgs)
		if err := createCmd.Execute(); err != nil {
			return fmt.Errorf("PR creation failed: %w", err)
		}
	}

	fmt.Println("🎉 Ship complete! Your changes are live!")
	
	if noPR {
		fmt.Println("   Run 'auto-pr pr' to create a pull request")
	}

	return nil
}