package cmd

import (
	"fmt"
	"os"

	"auto-pr/internal/ai"
	"auto-pr/internal/git"
	"auto-pr/internal/platforms"
	"auto-pr/pkg/types"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show repository status and Auto PR readiness",
	Long:  `Display comprehensive status information about the repository and Auto PR configuration`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("🔍 Auto PR Status Check")
	fmt.Println("========================")

	// Initialize git analyzer
	gitAnalyzer, err := git.NewAnalyzer(".")
	if err != nil {
		return fmt.Errorf("failed to initialize git analyzer: %w", err)
	}

	// Check git repository
	fmt.Println("\n📁 Repository Information:")
	if !gitAnalyzer.IsGitRepository() {
		fmt.Println("   ❌ Not a git repository")
		return nil
	}
	fmt.Println("   ✅ Git repository detected")

	// Get repository status
	status, err := gitAnalyzer.GetStatus()
	if err != nil {
		fmt.Printf("   ❌ Failed to get repository status: %s\n", err)
		return nil
	}

	fmt.Printf("   📋 Current branch: %s\n", status.CurrentBranch)
	fmt.Printf("   📋 Base branch: %s\n", status.BaseBranch)

	if status.RemoteURL != "" {
		fmt.Printf("   🔗 Remote URL: %s\n", status.RemoteURL)

		// Detect platform
		platform, err := platforms.DetectPlatform(status.RemoteURL)
		if err != nil {
			fmt.Printf("   ❓ Platform: Unknown (%s)\n", err)
		} else {
			fmt.Printf("   🌐 Platform: %s\n", platform)

			// Check platform authentication
			switch platform {
			case types.PlatformGitHub:
				client, err := platforms.NewGitHubClient(status.RemoteURL)
				if err == nil && client.IsAuthenticated() {
					fmt.Println("   ✅ GitHub CLI authenticated")
				} else {
					fmt.Println("   ❌ GitHub CLI not authenticated (run: gh auth login)")
				}
			case types.PlatformGitLab:
				client, err := platforms.NewGitLabClient(status.RemoteURL)
				if err == nil && client.IsAuthenticated() {
					fmt.Println("   ✅ GitLab CLI authenticated")
				} else {
					fmt.Println("   ❌ GitLab CLI not authenticated (run: glab auth login)")
				}
			}
		}
	} else {
		fmt.Println("   ⚠️  No remote repository configured")
	}

	// Show changes
	fmt.Println("\n📝 Working Directory Status:")
	if status.HasChanges {
		if len(status.StagedFiles) > 0 {
			fmt.Printf("   📦 Staged files: %d\n", len(status.StagedFiles))
		}
		if len(status.UnstagedFiles) > 0 {
			fmt.Printf("   📄 Unstaged files: %d\n", len(status.UnstagedFiles))
		}
		if len(status.UntrackedFiles) > 0 {
			fmt.Printf("   ❓ Untracked files: %d\n", len(status.UntrackedFiles))
		}
	} else {
		fmt.Println("   ✅ Working directory clean")
	}

	// Show commit status
	if status.CommitsAhead > 0 || status.CommitsBehind > 0 {
		fmt.Printf("   📊 Branch status: %d ahead, %d behind %s\n",
			status.CommitsAhead, status.CommitsBehind, status.BaseBranch)
	}

	// Check AI providers
	fmt.Println("\n🤖 AI Provider Status:")

	// Check Claude CLI
	if isClaudeAvailable() {
		fmt.Println("   ✅ Claude CLI available")
	} else {
		fmt.Println("   ❌ Claude CLI not found")
	}

	// Check for Gemini API key
	if hasGeminiAPIKey() {
		fmt.Println("   ✅ Gemini API key configured")
	} else {
		fmt.Println("   ❌ Gemini API key not configured")
	}

	// Check configuration
	fmt.Println("\n⚙️  Configuration Status:")

	// Try to load config
	configExists := checkConfigExists()
	if configExists {
		fmt.Println("   ✅ Configuration file found")
	} else {
		fmt.Println("   ❌ Configuration file not found (run: auto-pr config init)")
	}

	// Show commit history if available
	if status.RemoteURL != "" {
		fmt.Println("\n📚 Recent Commits:")
		commits, err := gitAnalyzer.GetCommitHistory(5)
		if err == nil && len(commits) > 0 {
			for _, commit := range commits {
				fmt.Printf("   • %s %s\n", commit.Hash[:8], commit.Message)
			}
		} else {
			fmt.Println("   No commits found")
		}
	}

	// PR readiness check
	fmt.Println("\n🚀 PR Creation Readiness:")
	if !gitAnalyzer.IsGitRepository() {
		fmt.Println("   ❌ Not ready: Not a git repository")
	} else if status.RemoteURL == "" {
		fmt.Println("   ❌ Not ready: No remote repository")
	} else if !status.HasChanges && status.CommitsAhead == 0 {
		fmt.Println("   ⚠️  No changes to create PR from")
	} else if !configExists {
		fmt.Println("   ❌ Not ready: Configuration not initialized")
	} else {
		fmt.Println("   ✅ Ready to create PR/MR!")
	}

	return nil
}

// Helper functions
func isClaudeAvailable() bool {
	return len(ai.GetAvailableProviders()) > 0
}

func hasGeminiAPIKey() bool {
	// Check common environment variable names
	envVars := []string{"GEMINI_API_KEY", "GOOGLE_API_KEY", "GOOGLE_AI_API_KEY"}
	for _, env := range envVars {
		if value := os.Getenv(env); value != "" {
			return true
		}
	}
	return false
}

func checkConfigExists() bool {
	// This is a simple check - in practice we'd use the config manager
	configPath := getConfigPath()
	_, err := os.Stat(configPath)
	return err == nil
}
