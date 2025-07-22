package ai

import (
	"strings"
	"testing"

	"auto-pr/pkg/types"
)

func TestClaudeParseResponse(t *testing.T) {
	client := &ClaudeClient{}

	tests := []struct {
		name        string
		output      string
		wantTitle   string
		wantHasBody bool
		wantErr     bool
	}{
		{
			name: "Valid JSON response",
			output: `{
				"title": "Add new feature",
				"body": "This PR adds a new feature",
				"labels": ["feature"],
				"reviewers": ["alice"],
				"priority": "high",
				"confidence": 0.9
			}`,
			wantTitle:   "Add new feature",
			wantHasBody: true,
			wantErr:     false,
		},
		{
			name: "JSON embedded in text",
			output: `Here's the PR information:
			{
				"title": "Fix bug",
				"body": "This fixes a critical bug",
				"labels": ["bugfix"],
				"reviewers": [],
				"priority": "high",
				"confidence": 0.95
			}
			That's all!`,
			wantTitle:   "Fix bug",
			wantHasBody: true,
			wantErr:     false,
		},
		{
			name: "Malformed JSON with title extraction",
			output: `Title: Update documentation
			
			This PR updates the documentation for the new API.
			
			Changes:
			- Updated README
			- Added examples`,
			wantTitle:   "Update documentation",
			wantHasBody: true,
			wantErr:     false,
		},
		{
			name:        "Plain text without structure",
			output:      "This is just some plain text response without any structure",
			wantTitle:   "Auto-generated PR",
			wantHasBody: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.parseResponse(tt.output)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp.Title != tt.wantTitle {
				t.Errorf("parseResponse() Title = %v, want %v", resp.Title, tt.wantTitle)
			}

			if tt.wantHasBody && resp.Body == "" {
				t.Error("parseResponse() Body is empty, want non-empty")
			}
		})
	}
}

func TestClaudeBuildPrompt(t *testing.T) {
	client := &ClaudeClient{}

	ctx := &AIContext{
		CommitHistory: []types.CommitInfo{
			{Hash: "abc123", Message: "Initial commit"},
			{Hash: "def456", Message: "Add feature"},
		},
		DiffSummary: "2 files changed, 50 additions, 10 deletions",
		FileChanges: []types.FileChange{
			{Path: "main.go", Status: types.StatusModified, Additions: 40, Deletions: 5},
			{Path: "README.md", Status: types.StatusModified, Additions: 10, Deletions: 5},
		},
		ProjectContext: ProjectContext{
			Language:  "Go",
			Framework: "Cobra",
		},
	}

	prompt := client.buildPrompt(ctx, "Generate a PR")

	// Check that prompt contains expected sections
	if !strings.Contains(prompt, "Recent Commits:") {
		t.Error("Prompt missing Recent Commits section")
	}
	if !strings.Contains(prompt, "Changes Summary:") {
		t.Error("Prompt missing Changes Summary section")
	}
	if !strings.Contains(prompt, "Files Changed:") {
		t.Error("Prompt missing Files Changed section")
	}
	if !strings.Contains(prompt, "Project Info:") {
		t.Error("Prompt missing Project Info section")
	}
	if !strings.Contains(prompt, "IMPORTANT: Respond with ONLY a valid JSON object") {
		t.Error("Prompt missing JSON format instruction")
	}
}
