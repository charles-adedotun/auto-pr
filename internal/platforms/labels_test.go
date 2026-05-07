package platforms

import (
	"testing"

	"auto-pr/pkg/types"
)

// stubClient implements PlatformClient with a fixed label set for testing.
type stubClient struct {
	labels []string
	err    error
}

func (s *stubClient) DetectPlatform(repoURL string) (types.PlatformType, error) { return types.PlatformGitHub, nil }
func (s *stubClient) IsAuthenticated() bool                                      { return true }
func (s *stubClient) CreatePullRequest(req *types.PullRequestRequest) (*types.PullRequest, error) {
	return nil, nil
}
func (s *stubClient) GetExistingPR(branch string) (*types.PullRequest, error) { return nil, nil }
func (s *stubClient) ValidateRepository() error                                { return nil }
func (s *stubClient) GetCLIPath() string                                       { return "" }
func (s *stubClient) ListLabels() ([]string, error)                            { return s.labels, s.err }

func TestFilterExistingLabels(t *testing.T) {
	tests := []struct {
		name           string
		repoLabels     []string
		candidates     []string
		want           []string
		wantErr        bool
	}{
		{
			name:       "keeps only labels that exist",
			repoLabels: []string{"bug", "feature", "docs"},
			candidates: []string{"bug", "feature", "nonexistent"},
			want:       []string{"bug", "feature"},
		},
		{
			name:       "all candidates nonexistent returns empty slice",
			repoLabels: []string{"bug", "feature"},
			candidates: []string{"nonexistent", "also-missing"},
			want:       []string{},
		},
		{
			name:       "empty candidates returns empty slice",
			repoLabels: []string{"bug"},
			candidates: []string{},
			want:       []string{},
		},
		{
			name:       "all candidates exist",
			repoLabels: []string{"bug", "feature"},
			candidates: []string{"bug", "feature"},
			want:       []string{"bug", "feature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &stubClient{labels: tt.repoLabels}
			got, err := FilterExistingLabels(client, tt.candidates)
			if (err != nil) != tt.wantErr {
				t.Fatalf("FilterExistingLabels() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("FilterExistingLabels() = %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("FilterExistingLabels()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
