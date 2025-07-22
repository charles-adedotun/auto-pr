package types

// PullRequest represents a pull request or merge request
type PullRequest struct {
	ID          int
	Number      int
	Title       string
	Body        string
	State       PRState
	Draft       bool
	URL         string
	HeadBranch  string
	BaseBranch  string
	Author      string
	Reviewers   []string
	Labels      []string
	Milestone   string
	CreatedAt   string
	UpdatedAt   string
}

// PRState represents the state of a pull request
type PRState string

const (
	PRStateOpen   PRState = "open"
	PRStateClosed PRState = "closed"
	PRStateMerged PRState = "merged"
	PRStateDraft  PRState = "draft"
)

// PullRequestRequest represents a request to create a pull request
type PullRequestRequest struct {
	Title       string
	Body        string
	HeadBranch  string
	BaseBranch  string
	Draft       bool
	Reviewers   []string
	TeamReviewers []string
	Labels      []string
	Milestone   string
	AutoMerge   bool
	DeleteHeadBranch bool
}

// PRTemplate represents a template for generating pull requests
type PRTemplate struct {
	Name        string
	Type        TemplateType
	Conditions  []Condition
	TitleFormat string
	BodyFormat  string
	Variables   map[string]string
}

// TemplateType represents different types of PR templates
type TemplateType string

const (
	TemplateFeature     TemplateType = "feature"
	TemplateBugfix      TemplateType = "bugfix"
	TemplateHotfix      TemplateType = "hotfix"
	TemplateRefactor    TemplateType = "refactor"
	TemplateDocumentation TemplateType = "documentation"
	TemplateDependency  TemplateType = "dependency"
	TemplateCustom      TemplateType = "custom"
)

// Condition represents a condition for template selection
type Condition struct {
	Field    string
	Operator string
	Value    string
}