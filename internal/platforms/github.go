package platforms

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	
	"auto-pr/pkg/types"
)

// GitHubClient implements PlatformClient for GitHub
type GitHubClient struct {
	cliPath   string
	repoOwner string
	repoName  string
	repoURL   string
}

// NewGitHubClient creates a new GitHub client
func NewGitHubClient(repoURL string) (*GitHubClient, error) {
	// Find gh CLI
	cliPath, err := exec.LookPath("gh")
	if err != nil {
		return nil, fmt.Errorf("GitHub CLI (gh) not found in PATH: %w", err)
	}
	
	// Extract repo info
	owner, repo, err := ExtractRepoInfo(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract repo info: %w", err)
	}
	
	client := &GitHubClient{
		cliPath:   cliPath,
		repoOwner: owner,
		repoName:  repo,
		repoURL:   repoURL,
	}
	
	return client, nil
}

// DetectPlatform returns GitHub platform type
func (g *GitHubClient) DetectPlatform(repoURL string) (types.PlatformType, error) {
	return DetectPlatform(repoURL)
}

// IsAuthenticated checks if user is authenticated with GitHub
func (g *GitHubClient) IsAuthenticated() bool {
	cmd := exec.Command(g.cliPath, "auth", "status")
	return cmd.Run() == nil
}

// ValidateRepository checks if the repository is accessible
func (g *GitHubClient) ValidateRepository() error {
	if !g.IsAuthenticated() {
		return fmt.Errorf("not authenticated with GitHub. Run: gh auth login")
	}
	
	// Check repository access
	cmd := exec.Command(g.cliPath, "repo", "view", fmt.Sprintf("%s/%s", g.repoOwner, g.repoName), "--json", "name")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot access repository %s/%s: %w", g.repoOwner, g.repoName, err)
	}
	
	var repoInfo struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(output, &repoInfo); err != nil {
		return fmt.Errorf("failed to parse repository info: %w", err)
	}
	
	return nil
}

// CreatePullRequest creates a new pull request on GitHub
func (g *GitHubClient) CreatePullRequest(req *types.PullRequestRequest) (*types.PullRequest, error) {
	if err := g.ValidateRepository(); err != nil {
		return nil, err
	}
	
	args := []string{
		"pr", "create",
		"--title", req.Title,
		"--body", req.Body,
		"--head", req.HeadBranch,
		"--base", req.BaseBranch,
	}
	
	// Add draft flag
	if req.Draft {
		args = append(args, "--draft")
	}
	
	// Add reviewers
	if len(req.Reviewers) > 0 {
		args = append(args, "--reviewer", strings.Join(req.Reviewers, ","))
	}
	
	// Add team reviewers
	if len(req.TeamReviewers) > 0 {
		args = append(args, "--reviewer", strings.Join(req.TeamReviewers, ","))
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
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}
	
	// Parse the PR URL from output
	prURL := strings.TrimSpace(string(output))
	
	// Get detailed PR information
	return g.getPRDetails(prURL)
}

// GetExistingPR finds existing PR for the given branch
func (g *GitHubClient) GetExistingPR(branch string) (*types.PullRequest, error) {
	cmd := exec.Command(g.cliPath, "pr", "list", 
		"--head", branch,
		"--json", "number,title,body,state,url,headRefName,baseRefName,author,labels,milestone,createdAt,updatedAt")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}
	
	var prs []struct {
		Number      int    `json:"number"`
		Title       string `json:"title"`
		Body        string `json:"body"`
		State       string `json:"state"`
		URL         string `json:"url"`
		HeadRefName string `json:"headRefName"`
		BaseRefName string `json:"baseRefName"`
		Author      struct {
			Login string `json:"login"`
		} `json:"author"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Milestone struct {
			Title string `json:"title"`
		} `json:"milestone"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	}
	
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR list: %w", err)
	}
	
	if len(prs) == 0 {
		return nil, nil // No existing PR
	}
	
	pr := prs[0] // Get the first (most recent) PR
	
	// Extract labels
	labels := make([]string, len(pr.Labels))
	for i, label := range pr.Labels {
		labels[i] = label.Name
	}
	
	return &types.PullRequest{
		ID:         pr.Number,
		Number:     pr.Number,
		Title:      pr.Title,
		Body:       pr.Body,
		State:      mapGitHubState(pr.State),
		URL:        pr.URL,
		HeadBranch: pr.HeadRefName,
		BaseBranch: pr.BaseRefName,
		Author:     pr.Author.Login,
		Labels:     labels,
		Milestone:  pr.Milestone.Title,
		CreatedAt:  pr.CreatedAt,
		UpdatedAt:  pr.UpdatedAt,
	}, nil
}

// GetCLIPath returns the path to GitHub CLI
func (g *GitHubClient) GetCLIPath() string {
	return g.cliPath
}

// getPRDetails gets detailed information about a PR from its URL
func (g *GitHubClient) getPRDetails(prURL string) (*types.PullRequest, error) {
	// Extract PR number from URL
	parts := strings.Split(prURL, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid PR URL: %s", prURL)
	}
	prNumber := parts[len(parts)-1]
	
	cmd := exec.Command(g.cliPath, "pr", "view", prNumber,
		"--json", "number,title,body,state,url,headRefName,baseRefName,author,labels,milestone,createdAt,updatedAt,isDraft")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get PR details: %w", err)
	}
	
	var pr struct {
		Number      int    `json:"number"`
		Title       string `json:"title"`
		Body        string `json:"body"`
		State       string `json:"state"`
		URL         string `json:"url"`
		HeadRefName string `json:"headRefName"`
		BaseRefName string `json:"baseRefName"`
		Author      struct {
			Login string `json:"login"`
		} `json:"author"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Milestone struct {
			Title string `json:"title"`
		} `json:"milestone"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		IsDraft   bool   `json:"isDraft"`
	}
	
	if err := json.Unmarshal(output, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse PR details: %w", err)
	}
	
	// Extract labels
	labels := make([]string, len(pr.Labels))
	for i, label := range pr.Labels {
		labels[i] = label.Name
	}
	
	state := mapGitHubState(pr.State)
	if pr.IsDraft {
		state = types.PRStateDraft
	}
	
	return &types.PullRequest{
		ID:         pr.Number,
		Number:     pr.Number,
		Title:      pr.Title,
		Body:       pr.Body,
		State:      state,
		Draft:      pr.IsDraft,
		URL:        pr.URL,
		HeadBranch: pr.HeadRefName,
		BaseBranch: pr.BaseRefName,
		Author:     pr.Author.Login,
		Labels:     labels,
		Milestone:  pr.Milestone.Title,
		CreatedAt:  pr.CreatedAt,
		UpdatedAt:  pr.UpdatedAt,
	}, nil
}

// mapGitHubState maps GitHub PR states to our internal states
func mapGitHubState(ghState string) types.PRState {
	switch strings.ToLower(ghState) {
	case "open":
		return types.PRStateOpen
	case "closed":
		return types.PRStateClosed
	case "merged":
		return types.PRStateMerged
	default:
		return types.PRStateOpen
	}
}