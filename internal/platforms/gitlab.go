package platforms

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	
	"auto-pr/pkg/types"
)

// GitLabClient implements PlatformClient for GitLab
type GitLabClient struct {
	cliPath   string
	projectID string
	baseURL   string
	repoURL   string
}

// NewGitLabClient creates a new GitLab client
func NewGitLabClient(repoURL string) (*GitLabClient, error) {
	// Find glab CLI
	cliPath, err := exec.LookPath("glab")
	if err != nil {
		return nil, fmt.Errorf("GitLab CLI (glab) not found in PATH: %w", err)
	}
	
	// Extract project info
	owner, repo, err := ExtractRepoInfo(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract repo info: %w", err)
	}
	
	projectID := fmt.Sprintf("%s/%s", owner, repo)
	
	// Determine base URL
	baseURL := "https://gitlab.com"
	if strings.Contains(repoURL, "://") {
		parts := strings.Split(strings.Split(repoURL, "://")[1], "/")
		if len(parts) > 0 && !strings.Contains(parts[0], "gitlab.com") {
			baseURL = "https://" + parts[0]
		}
	}
	
	client := &GitLabClient{
		cliPath:   cliPath,
		projectID: projectID,
		baseURL:   baseURL,
		repoURL:   repoURL,
	}
	
	return client, nil
}

// DetectPlatform returns GitLab platform type
func (g *GitLabClient) DetectPlatform(repoURL string) (types.PlatformType, error) {
	return DetectPlatform(repoURL)
}

// IsAuthenticated checks if user is authenticated with GitLab
func (g *GitLabClient) IsAuthenticated() bool {
	cmd := exec.Command(g.cliPath, "auth", "status")
	return cmd.Run() == nil
}

// ValidateRepository checks if the repository is accessible
func (g *GitLabClient) ValidateRepository() error {
	if !g.IsAuthenticated() {
		return fmt.Errorf("not authenticated with GitLab. Run: glab auth login")
	}
	
	// Check repository access
	cmd := exec.Command(g.cliPath, "repo", "view", g.projectID, "--json")
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot access repository %s: %w", g.projectID, err)
	}
	
	return nil
}

// CreatePullRequest creates a new merge request on GitLab
func (g *GitLabClient) CreatePullRequest(req *types.PullRequestRequest) (*types.PullRequest, error) {
	if err := g.ValidateRepository(); err != nil {
		return nil, err
	}
	
	args := []string{
		"mr", "create",
		"--title", req.Title,
		"--description", req.Body,
		"--source-branch", req.HeadBranch,
		"--target-branch", req.BaseBranch,
	}
	
	// Add draft flag
	if req.Draft {
		args = append(args, "--draft")
	}
	
	// Add assignee (GitLab uses assignee instead of reviewers)
	if len(req.Reviewers) > 0 {
		args = append(args, "--assignee", strings.Join(req.Reviewers, ","))
	}
	
	// Add labels
	if len(req.Labels) > 0 {
		args = append(args, "--label", strings.Join(req.Labels, ","))
	}
	
	// Add milestone
	if req.Milestone != "" {
		args = append(args, "--milestone", req.Milestone)
	}
	
	// Execute command
	cmd := exec.Command(g.cliPath, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to create merge request: %w", err)
	}
	
	// Parse the MR URL from output
	mrURL := strings.TrimSpace(string(output))
	
	// Get detailed MR information
	return g.getMRDetails(mrURL)
}

// GetExistingPR finds existing MR for the given branch
func (g *GitLabClient) GetExistingPR(branch string) (*types.PullRequest, error) {
	cmd := exec.Command(g.cliPath, "mr", "list", 
		"--source-branch", branch,
		"--json")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list merge requests: %w", err)
	}
	
	var mrs []struct {
		IID         int    `json:"iid"`
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		WebURL      string `json:"web_url"`
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
		Author      struct {
			Username string `json:"username"`
		} `json:"author"`
		Labels    []string `json:"labels"`
		Milestone struct {
			Title string `json:"title"`
		} `json:"milestone"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Draft     bool   `json:"draft"`
	}
	
	if err := json.Unmarshal(output, &mrs); err != nil {
		return nil, fmt.Errorf("failed to parse MR list: %w", err)
	}
	
	if len(mrs) == 0 {
		return nil, nil // No existing MR
	}
	
	mr := mrs[0] // Get the first (most recent) MR
	
	return &types.PullRequest{
		ID:         mr.IID,
		Number:     mr.IID,
		Title:      mr.Title,
		Body:       mr.Description,
		State:      mapGitLabState(mr.State),
		Draft:      mr.Draft,
		URL:        mr.WebURL,
		HeadBranch: mr.SourceBranch,
		BaseBranch: mr.TargetBranch,
		Author:     mr.Author.Username,
		Labels:     mr.Labels,
		Milestone:  mr.Milestone.Title,
		CreatedAt:  mr.CreatedAt,
		UpdatedAt:  mr.UpdatedAt,
	}, nil
}

// GetCLIPath returns the path to GitLab CLI
func (g *GitLabClient) GetCLIPath() string {
	return g.cliPath
}

// getMRDetails gets detailed information about an MR from its URL
func (g *GitLabClient) getMRDetails(mrURL string) (*types.PullRequest, error) {
	// Extract MR IID from URL
	parts := strings.Split(mrURL, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid MR URL: %s", mrURL)
	}
	mrIID := parts[len(parts)-1]
	
	cmd := exec.Command(g.cliPath, "mr", "view", mrIID, "--json")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get MR details: %w", err)
	}
	
	var mr struct {
		IID         int    `json:"iid"`
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		WebURL      string `json:"web_url"`
		SourceBranch string `json:"source_branch"`
		TargetBranch string `json:"target_branch"`
		Author      struct {
			Username string `json:"username"`
		} `json:"author"`
		Labels    []string `json:"labels"`
		Milestone struct {
			Title string `json:"title"`
		} `json:"milestone"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		Draft     bool   `json:"draft"`
	}
	
	if err := json.Unmarshal(output, &mr); err != nil {
		return nil, fmt.Errorf("failed to parse MR details: %w", err)
	}
	
	state := mapGitLabState(mr.State)
	if mr.Draft {
		state = types.PRStateDraft
	}
	
	return &types.PullRequest{
		ID:         mr.IID,
		Number:     mr.IID,
		Title:      mr.Title,
		Body:       mr.Description,
		State:      state,
		Draft:      mr.Draft,
		URL:        mr.WebURL,
		HeadBranch: mr.SourceBranch,
		BaseBranch: mr.TargetBranch,
		Author:     mr.Author.Username,
		Labels:     mr.Labels,
		Milestone:  mr.Milestone.Title,
		CreatedAt:  mr.CreatedAt,
		UpdatedAt:  mr.UpdatedAt,
	}, nil
}

// mapGitLabState maps GitLab MR states to our internal states
func mapGitLabState(glState string) types.PRState {
	switch strings.ToLower(glState) {
	case "opened":
		return types.PRStateOpen
	case "closed":
		return types.PRStateClosed
	case "merged":
		return types.PRStateMerged
	default:
		return types.PRStateOpen
	}
}

// GetPlatform returns the platform type
func (g *GitLabClient) GetPlatform() types.PlatformType {
	return types.PlatformGitLab
}

// GetSupportedFeatures returns supported GitLab features
func (g *GitLabClient) GetSupportedFeatures() []string {
	return []string{
		"merge-requests",
		"labels",
		"reviewers",
		"draft-mr",
		"auto-merge",
		"milestones",
		"assignees",
		"approvals",
		"merge-when-pipeline-succeeds",
		"squash-commits",
		"delete-source-branch",
	}
}