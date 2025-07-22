package platforms

import (
	"net/url"
	"strings"

	"auto-pr/pkg/types"
)

// DetectPlatform detects the git platform from a remote URL
func DetectPlatform(remoteURL string) (types.PlatformType, error) {
	if remoteURL == "" {
		return types.PlatformUnknown, nil
	}

	// Clean up the URL for parsing
	cleanURL := cleanRemoteURL(remoteURL)

	parsedURL, err := url.Parse(cleanURL)
	if err != nil {
		return types.PlatformUnknown, err
	}

	hostname := parsedURL.Hostname()

	// Check for GitHub
	if hostname == "github.com" || strings.HasSuffix(hostname, ".github.com") {
		return types.PlatformGitHub, nil
	}

	// Check for GitLab
	if hostname == "gitlab.com" || strings.Contains(hostname, "gitlab") {
		return types.PlatformGitLab, nil
	}

	return types.PlatformUnknown, nil
}

// cleanRemoteURL converts git SSH URLs to HTTPS format for easier parsing
func cleanRemoteURL(remoteURL string) string {
	// Handle SSH format: git@github.com:user/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		parts := strings.SplitN(remoteURL, ":", 2)
		if len(parts) == 2 {
			host := strings.TrimPrefix(parts[0], "git@")
			path := parts[1]

			// Remove .git suffix if present
			path = strings.TrimSuffix(path, ".git")

			return "https://" + host + "/" + path
		}
	}

	// Handle HTTPS format - just remove .git suffix if present
	if strings.HasPrefix(remoteURL, "https://") {
		return strings.TrimSuffix(remoteURL, ".git")
	}

	return remoteURL
}

// ExtractRepoInfo extracts owner and repository name from a remote URL
func ExtractRepoInfo(remoteURL string) (owner, repo string, err error) {
	cleanURL := cleanRemoteURL(remoteURL)

	parsedURL, err := url.Parse(cleanURL)
	if err != nil {
		return "", "", err
	}

	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) >= 2 {
		owner = pathParts[0]
		repo = pathParts[1]
	}

	return owner, repo, nil
}
