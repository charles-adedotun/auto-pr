package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	
	"auto-pr/pkg/types"
)

// Analyzer provides git repository analysis functionality
type Analyzer struct {
	repoPath string
}

// NewAnalyzer creates a new git analyzer for the specified repository path
func NewAnalyzer(repoPath string) (*Analyzer, error) {
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	
	return &Analyzer{
		repoPath: absPath,
	}, nil
}

// IsGitRepository checks if the current directory is a git repository
func (a *Analyzer) IsGitRepository() bool {
	gitDir := filepath.Join(a.repoPath, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}
	
	// Check if .git is a file (in case of git worktrees)
	if _, err := os.Stat(gitDir); err == nil {
		return true
	}
	
	return false
}

// GetStatus returns the current status of the git repository
func (a *Analyzer) GetStatus() (*types.GitStatus, error) {
	if !a.IsGitRepository() {
		return &types.GitStatus{IsGitRepo: false}, nil
	}
	
	status := &types.GitStatus{IsGitRepo: true}
	
	// Get current branch
	currentBranch, err := a.getCurrentBranch()
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}
	status.CurrentBranch = currentBranch
	
	// Get remote URL
	remoteURL, err := a.getRemoteURL()
	if err == nil {
		status.RemoteURL = remoteURL
	}
	
	// Get base branch (usually main or master)
	baseBranch, err := a.getBaseBranch()
	if err == nil {
		status.BaseBranch = baseBranch
	}
	
	// Get file statuses
	staged, unstaged, untracked, err := a.getFileStatuses()
	if err != nil {
		return nil, fmt.Errorf("failed to get file statuses: %w", err)
	}
	
	status.StagedFiles = staged
	status.UnstagedFiles = unstaged
	status.UntrackedFiles = untracked
	status.HasChanges = len(staged) > 0 || len(unstaged) > 0 || len(untracked) > 0
	
	// Get commit counts
	ahead, behind, err := a.getCommitCounts(baseBranch)
	if err == nil {
		status.CommitsAhead = ahead
		status.CommitsBehind = behind
	}
	
	return status, nil
}

// GetRemoteURL returns the remote URL for the repository
func (a *Analyzer) GetRemoteURL() string {
	remoteURL, _ := a.getRemoteURL()
	return remoteURL
}

// getCurrentBranch returns the current branch name
func (a *Analyzer) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "-C", a.repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// getRemoteURL returns the remote URL for origin
func (a *Analyzer) getRemoteURL() (string, error) {
	cmd := exec.Command("git", "-C", a.repoPath, "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// getBaseBranch attempts to determine the base branch (main/master)
func (a *Analyzer) getBaseBranch() (string, error) {
	// Try to get the default branch from remote
	cmd := exec.Command("git", "-C", a.repoPath, "symbolic-ref", "refs/remotes/origin/HEAD")
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		parts := strings.Split(branch, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}
	
	// Fallback: check common branch names
	commonBranches := []string{"main", "master", "develop"}
	for _, branch := range commonBranches {
		cmd := exec.Command("git", "-C", a.repoPath, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
		if cmd.Run() == nil {
			return branch, nil
		}
	}
	
	return "main", nil // Default fallback
}

// getFileStatuses returns lists of staged, unstaged, and untracked files
func (a *Analyzer) getFileStatuses() (staged, unstaged, untracked []string, err error) {
	cmd := exec.Command("git", "-C", a.repoPath, "status", "--porcelain=v1")
	output, err := cmd.Output()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get git status: %w", err)
	}
	
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 3 {
			continue
		}
		
		indexStatus := line[0]
		workTreeStatus := line[1]
		filename := line[3:]
		
		if indexStatus != ' ' && indexStatus != '?' {
			staged = append(staged, filename)
		}
		if workTreeStatus != ' ' && workTreeStatus != '?' {
			unstaged = append(unstaged, filename)
		}
		if indexStatus == '?' && workTreeStatus == '?' {
			untracked = append(untracked, filename)
		}
	}
	
	return staged, unstaged, untracked, scanner.Err()
}

// getCommitCounts returns the number of commits ahead and behind the base branch
func (a *Analyzer) getCommitCounts(baseBranch string) (ahead, behind int, err error) {
	// Get commits ahead
	cmd := exec.Command("git", "-C", a.repoPath, "rev-list", "--count", fmt.Sprintf("origin/%s..HEAD", baseBranch))
	output, err := cmd.Output()
	if err == nil {
		if count, parseErr := strconv.Atoi(strings.TrimSpace(string(output))); parseErr == nil {
			ahead = count
		}
	}
	
	// Get commits behind
	cmd = exec.Command("git", "-C", a.repoPath, "rev-list", "--count", fmt.Sprintf("HEAD..origin/%s", baseBranch))
	output, err = cmd.Output()
	if err == nil {
		if count, parseErr := strconv.Atoi(strings.TrimSpace(string(output))); parseErr == nil {
			behind = count
		}
	}
	
	return ahead, behind, nil
}