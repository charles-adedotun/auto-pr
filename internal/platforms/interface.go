package platforms

import "auto-pr/pkg/types"

// PlatformClient defines the interface for interacting with different git platforms
type PlatformClient interface {
	// DetectPlatform returns the platform type for the given repository URL
	DetectPlatform(repoURL string) (types.PlatformType, error)
	
	// IsAuthenticated checks if the user is authenticated with the platform
	IsAuthenticated() bool
	
	// CreatePullRequest creates a new pull request or merge request
	CreatePullRequest(req *types.PullRequestRequest) (*types.PullRequest, error)
	
	// GetExistingPR finds an existing PR/MR for the given branch
	GetExistingPR(branch string) (*types.PullRequest, error)
	
	// ValidateRepository checks if the repository is accessible and valid
	ValidateRepository() error
	
	// GetCLIPath returns the path to the platform's CLI tool
	GetCLIPath() string
}