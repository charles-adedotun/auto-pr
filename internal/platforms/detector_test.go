package platforms

import (
	"testing"

	"auto-pr/pkg/types"
)

func TestDetectPlatform(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected types.PlatformType
		wantErr  bool
	}{
		{
			name:     "GitHub HTTPS URL",
			url:      "https://github.com/user/repo.git",
			expected: types.PlatformGitHub,
			wantErr:  false,
		},
		{
			name:     "GitHub SSH URL",
			url:      "git@github.com:user/repo.git",
			expected: types.PlatformGitHub,
			wantErr:  false,
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/user/repo.git",
			expected: types.PlatformGitLab,
			wantErr:  false,
		},
		{
			name:     "GitLab SSH URL",
			url:      "git@gitlab.com:user/repo.git",
			expected: types.PlatformGitLab,
			wantErr:  false,
		},
		{
			name:     "Self-hosted GitLab",
			url:      "https://gitlab.company.com/user/repo.git",
			expected: types.PlatformGitLab,
			wantErr:  false,
		},
		{
			name:     "Unknown platform",
			url:      "https://bitbucket.org/user/repo.git",
			expected: types.PlatformUnknown,
			wantErr:  false,
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: types.PlatformUnknown,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectPlatform(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectPlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("DetectPlatform() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestExtractRepoInfo(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "GitHub HTTPS URL",
			url:       "https://github.com/user/repo.git",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "GitHub SSH URL",
			url:       "git@github.com:user/repo.git",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
		{
			name:      "GitLab with namespace",
			url:       "https://gitlab.com/namespace/user/repo.git",
			wantOwner: "namespace",
			wantRepo:  "user",
			wantErr:   false,
		},
		{
			name:      "URL without .git suffix",
			url:       "https://github.com/user/repo",
			wantOwner: "user",
			wantRepo:  "repo",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOwner, gotRepo, err := ExtractRepoInfo(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractRepoInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOwner != tt.wantOwner {
				t.Errorf("ExtractRepoInfo() owner = %v, want %v", gotOwner, tt.wantOwner)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("ExtractRepoInfo() repo = %v, want %v", gotRepo, tt.wantRepo)
			}
		})
	}
}

func TestCleanRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "SSH to HTTPS conversion",
			url:      "git@github.com:user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTPS with .git suffix",
			url:      "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTPS without .git suffix",
			url:      "https://github.com/user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "SSH GitLab URL",
			url:      "git@gitlab.com:user/repo.git",
			expected: "https://gitlab.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanRemoteURL(tt.url)
			if got != tt.expected {
				t.Errorf("cleanRemoteURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}
