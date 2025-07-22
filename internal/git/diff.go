package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	
	"auto-pr/pkg/types"
)

// GetDiff returns the diff for staged and unstaged changes
func (a *Analyzer) GetDiff(staged bool) (string, error) {
	args := []string{"-C", a.repoPath, "diff"}
	if staged {
		args = append(args, "--staged")
	}
	
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	
	return string(output), nil
}

// GetDiffSummary returns a summary of changes in the working directory
func (a *Analyzer) GetDiffSummary() (*types.DiffSummary, error) {
	// Get overall statistics
	cmd := exec.Command("git", "-C", a.repoPath, "diff", "--stat")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff summary: %w", err)
	}
	
	summary, _ := a.parseStatOutput(string(output))
	
	// Get detailed file changes
	fileChanges, err := a.getDetailedFileChanges()
	if err != nil {
		return summary, nil // Return partial summary if detailed parsing fails
	}
	
	summary.FileChanges = fileChanges
	return summary, nil
}

// GetBranchDiff returns diff between current branch and base branch
func (a *Analyzer) GetBranchDiff(baseBranch string) (*types.DiffSummary, error) {
	if baseBranch == "" {
		baseBranch = "main"
	}
	
	// Get diff statistics
	cmd := exec.Command("git", "-C", a.repoPath, 
		"diff", fmt.Sprintf("origin/%s...HEAD", baseBranch), "--stat")
	
	output, err := cmd.Output()
	if err != nil {
		// Fallback to local comparison
		cmd = exec.Command("git", "-C", a.repoPath,
			"diff", fmt.Sprintf("%s...HEAD", baseBranch), "--stat")
		
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get branch diff: %w", err)
		}
	}
	
	summary, _ := a.parseStatOutput(string(output))
	
	// Get detailed file changes for branch comparison
	fileChanges, err := a.getBranchFileChanges(baseBranch)
	if err != nil {
		return summary, nil // Return partial summary
	}
	
	summary.FileChanges = fileChanges
	return summary, nil
}

// getDetailedFileChanges returns detailed file change information
func (a *Analyzer) getDetailedFileChanges() ([]types.FileChange, error) {
	var changes []types.FileChange
	
	// Get staged changes
	stagedChanges, err := a.getFileChangesForStatus("--staged")
	if err == nil {
		changes = append(changes, stagedChanges...)
	}
	
	// Get unstaged changes
	unstagedChanges, err := a.getFileChangesForStatus("")
	if err == nil {
		changes = append(changes, unstagedChanges...)
	}
	
	// Merge changes for the same files
	mergedChanges := a.mergeFileChanges(changes)
	
	return mergedChanges, nil
}

// getBranchFileChanges returns file changes between branches
func (a *Analyzer) getBranchFileChanges(baseBranch string) ([]types.FileChange, error) {
	cmd := exec.Command("git", "-C", a.repoPath,
		"diff", fmt.Sprintf("origin/%s...HEAD", baseBranch), "--name-status")
	
	output, err := cmd.Output()
	if err != nil {
		// Fallback to local comparison
		cmd = exec.Command("git", "-C", a.repoPath,
			"diff", fmt.Sprintf("%s...HEAD", baseBranch), "--name-status")
		
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get branch file changes: %w", err)
		}
	}
	
	return a.parseNameStatus(string(output))
}

// getFileChangesForStatus returns file changes for a specific git diff status
func (a *Analyzer) getFileChangesForStatus(statusFlag string) ([]types.FileChange, error) {
	args := []string{"-C", a.repoPath, "diff", "--name-status"}
	if statusFlag != "" {
		args = append(args, statusFlag)
	}
	
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get file changes: %w", err)
	}
	
	return a.parseNameStatus(string(output))
}

// parseNameStatus parses git diff --name-status output
func (a *Analyzer) parseNameStatus(output string) ([]types.FileChange, error) {
	var changes []types.FileChange
	
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		status := mapGitStatus(parts[0])
		filepath := parts[1]
		
		// Get detailed stats for this file
		additions, deletions, err := a.getFileStats(filepath)
		if err != nil {
			// Continue without detailed stats
			additions, deletions = 0, 0
		}
		
		changes = append(changes, types.FileChange{
			Path:      filepath,
			Status:    status,
			Additions: additions,
			Deletions: deletions,
			IsBinary:  a.isBinaryFile(filepath),
		})
	}
	
	return changes, scanner.Err()
}

// getFileStats returns addition/deletion counts for a specific file
func (a *Analyzer) getFileStats(filepath string) (int, int, error) {
	cmd := exec.Command("git", "-C", a.repoPath, 
		"diff", "--numstat", "--", filepath)
	
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	
	line := strings.TrimSpace(string(output))
	if line == "" {
		return 0, 0, nil
	}
	
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0, 0, nil
	}
	
	additions, err1 := strconv.Atoi(parts[0])
	deletions, err2 := strconv.Atoi(parts[1])
	
	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("failed to parse stats")
	}
	
	return additions, deletions, nil
}

// isBinaryFile checks if a file is binary
func (a *Analyzer) isBinaryFile(filepath string) bool {
	// Simple heuristic based on file extension
	binaryExtensions := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf": true, ".zip": true, ".tar": true, ".gz": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".dat": true, ".db": true,
	}
	
	for ext := range binaryExtensions {
		if strings.HasSuffix(strings.ToLower(filepath), ext) {
			return true
		}
	}
	
	return false
}

// mergeFileChanges merges duplicate file changes
func (a *Analyzer) mergeFileChanges(changes []types.FileChange) []types.FileChange {
	fileMap := make(map[string]types.FileChange)
	
	for _, change := range changes {
		if existing, exists := fileMap[change.Path]; exists {
			// Merge the changes
			merged := types.FileChange{
				Path:      change.Path,
				Status:    change.Status, // Use the latest status
				Additions: existing.Additions + change.Additions,
				Deletions: existing.Deletions + change.Deletions,
				IsBinary:  existing.IsBinary || change.IsBinary,
			}
			fileMap[change.Path] = merged
		} else {
			fileMap[change.Path] = change
		}
	}
	
	// Convert map back to slice
	var merged []types.FileChange
	for _, change := range fileMap {
		merged = append(merged, change)
	}
	
	return merged
}

// mapGitStatus maps git status codes to our internal status types
func mapGitStatus(gitStatus string) types.ChangeStatus {
	switch gitStatus[0] {
	case 'A':
		return types.StatusAdded
	case 'M':
		return types.StatusModified
	case 'D':
		return types.StatusDeleted
	case 'R':
		return types.StatusRenamed
	case 'C':
		return types.StatusCopied
	case '?':
		return types.StatusUntracked
	default:
		return types.StatusModified
	}
}