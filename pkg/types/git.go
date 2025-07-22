package types

import "time"

// GitStatus represents the current status of a git repository
type GitStatus struct {
	IsGitRepo      bool
	CurrentBranch  string
	BaseBranch     string
	RemoteURL      string
	HasChanges     bool
	StagedFiles    []string
	UnstagedFiles  []string
	UntrackedFiles []string
	CommitsAhead   int
	CommitsBehind  int
}

// CommitInfo represents information about a single commit
type CommitInfo struct {
	Hash    string
	Message string
	Author  string
	Email   string
	Date    time.Time
	Files   []string
	Diff    string
}

// FileChange represents a change to a single file
type FileChange struct {
	Path      string
	Status    ChangeStatus
	Additions int
	Deletions int
	IsBinary  bool
}

// ChangeStatus represents the type of change made to a file
type ChangeStatus string

const (
	StatusAdded     ChangeStatus = "added"
	StatusModified  ChangeStatus = "modified"
	StatusDeleted   ChangeStatus = "deleted"
	StatusRenamed   ChangeStatus = "renamed"
	StatusCopied    ChangeStatus = "copied"
	StatusUntracked ChangeStatus = "untracked"
)

// BranchInfo contains information about the current branch and its relationship to other branches
type BranchInfo struct {
	Name          string
	BaseBranch    string
	CommitsAhead  int
	CommitsBehind int
	IsClean       bool
	HasUpstream   bool
}

// DiffSummary contains a summary of all changes in the repository
type DiffSummary struct {
	TotalFiles    int
	TotalLines    int
	Additions     int
	Deletions     int
	FileChanges   []FileChange
	CommitHistory []CommitInfo
}
