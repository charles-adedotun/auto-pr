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

	// ListLabels returns all label names defined in the repository
	ListLabels() ([]string, error)
}

// FilterExistingLabels returns only those labels from candidates that exist in the repository.
func FilterExistingLabels(client PlatformClient, candidates []string) ([]string, error) {
	if len(candidates) == 0 {
		return []string{}, nil
	}
	existing, err := client.ListLabels()
	if err != nil {
		return []string{}, err
	}
	set := make(map[string]struct{}, len(existing))
	for _, l := range existing {
		set[l] = struct{}{}
	}
	var filtered []string
	for _, c := range candidates {
		if _, ok := set[c]; ok {
			filtered = append(filtered, c)
		}
	}
	if filtered == nil {
		return []string{}, nil
	}
	return filtered, nil
}
