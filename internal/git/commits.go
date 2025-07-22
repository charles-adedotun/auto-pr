package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"auto-pr/pkg/types"
)

// GetCommitHistory returns the commit history for the current branch
func (a *Analyzer) GetCommitHistory(limit int) ([]types.CommitInfo, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	cmd := exec.Command("git", "-C", a.repoPath,
		"log",
		fmt.Sprintf("-%d", limit),
		"--pretty=format:%H|%s|%an|%ae|%at",
		"--name-only")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}

	return a.parseCommitHistory(string(output))
}

// GetCommitsSinceBase returns commits since the base branch
func (a *Analyzer) GetCommitsSinceBase(baseBranch string) ([]types.CommitInfo, error) {
	if baseBranch == "" {
		baseBranch = "main"
	}

	// Check if base branch exists on remote
	cmd := exec.Command("git", "-C", a.repoPath,
		"rev-parse", "--verify", fmt.Sprintf("origin/%s", baseBranch))
	if err := cmd.Run(); err != nil {
		// Fallback to local base branch
		cmd = exec.Command("git", "-C", a.repoPath,
			"rev-parse", "--verify", baseBranch)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("base branch %s not found", baseBranch)
		}
	}

	// Get commits between base and HEAD
	cmd = exec.Command("git", "-C", a.repoPath,
		"log",
		fmt.Sprintf("origin/%s..HEAD", baseBranch),
		"--pretty=format:%H|%s|%an|%ae|%at",
		"--name-only")

	output, err := cmd.Output()
	if err != nil {
		// Fallback to local base branch comparison
		cmd = exec.Command("git", "-C", a.repoPath,
			"log",
			fmt.Sprintf("%s..HEAD", baseBranch),
			"--pretty=format:%H|%s|%an|%ae|%at",
			"--name-only")

		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get commits since base: %w", err)
		}
	}

	if strings.TrimSpace(string(output)) == "" {
		return []types.CommitInfo{}, nil // No commits ahead
	}

	return a.parseCommitHistory(string(output))
}

// parseCommitHistory parses git log output into CommitInfo structs
func (a *Analyzer) parseCommitHistory(output string) ([]types.CommitInfo, error) {
	var commits []types.CommitInfo
	var currentCommit *types.CommitInfo

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check if this is a commit header line (contains |)
		if strings.Contains(line, "|") {
			// Save previous commit if exists
			if currentCommit != nil {
				commits = append(commits, *currentCommit)
			}

			// Parse new commit header
			parts := strings.Split(line, "|")
			if len(parts) >= 5 {
				timestamp, err := strconv.ParseInt(parts[4], 10, 64)
				if err != nil {
					timestamp = time.Now().Unix()
				}

				currentCommit = &types.CommitInfo{
					Hash:    parts[0],
					Message: parts[1],
					Author:  parts[2],
					Email:   parts[3],
					Date:    time.Unix(timestamp, 0),
					Files:   []string{},
				}
			}
		} else if currentCommit != nil {
			// This is a file name from the commit
			currentCommit.Files = append(currentCommit.Files, line)
		}
	}

	// Add the last commit
	if currentCommit != nil {
		commits = append(commits, *currentCommit)
	}

	return commits, scanner.Err()
}

// GetCommitDiff returns the diff for a specific commit
func (a *Analyzer) GetCommitDiff(commitHash string) (string, error) {
	cmd := exec.Command("git", "-C", a.repoPath,
		"show", commitHash, "--pretty=format:", "--name-only")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit diff: %w", err)
	}

	return string(output), nil
}

// GetCommitStats returns statistics for a commit
func (a *Analyzer) GetCommitStats(commitHash string) (*types.DiffSummary, error) {
	cmd := exec.Command("git", "-C", a.repoPath,
		"show", commitHash, "--stat", "--format=")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit stats: %w", err)
	}

	return a.parseStatOutput(string(output))
}

// parseStatOutput parses git stat output
func (a *Analyzer) parseStatOutput(output string) (*types.DiffSummary, error) {
	summary := &types.DiffSummary{
		FileChanges: []types.FileChange{},
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse file change line: "filename | 10 +++++++---"
		if strings.Contains(line, "|") {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				filename := strings.TrimSpace(parts[0])
				statsStr := strings.TrimSpace(parts[1])

				// Count additions and deletions from the visual representation
				additions := strings.Count(statsStr, "+")
				deletions := strings.Count(statsStr, "-")

				// Determine file status
				status := types.StatusModified
				if additions > 0 && deletions == 0 {
					status = types.StatusAdded
				} else if additions == 0 && deletions > 0 {
					status = types.StatusDeleted
				}

				summary.FileChanges = append(summary.FileChanges, types.FileChange{
					Path:      filename,
					Status:    status,
					Additions: additions,
					Deletions: deletions,
				})

				summary.TotalFiles++
				summary.Additions += additions
				summary.Deletions += deletions
				summary.TotalLines += additions + deletions
			}
		}

		// Parse summary line: "5 files changed, 100 insertions(+), 20 deletions(-)"
		if strings.Contains(line, "file") && strings.Contains(line, "changed") {
			// This is handled by individual file parsing above
			continue
		}
	}

	return summary, nil
}
