package ai

import "auto-pr/pkg/types"

// AIClient defines the interface for AI service providers
type AIClient interface {
	// GenerateContent generates AI content based on the provided context and prompt
	GenerateContent(ctx *AIContext, prompt string) (*AIResponse, error)

	// IsAvailable checks if the AI service is available and properly configured
	IsAvailable() bool

	// GetProvider returns the provider type
	GetProvider() types.AIProvider

	// ValidateConfig validates the configuration for this provider
	ValidateConfig() error
}

// AIContext contains all the context information for AI generation
type AIContext struct {
	CommitHistory  []types.CommitInfo
	DiffSummary    string
	FileChanges    []types.FileChange
	BranchInfo     types.BranchInfo
	ProjectContext ProjectContext
	PreviousPRs    []types.PullRequest
	Platform       types.PlatformType
	TemplateType   types.TemplateType
}

// ProjectContext contains information about the project
type ProjectContext struct {
	Language    string
	Framework   string
	ProjectType string
	HasTests    bool
	HasCI       bool
	HasDocs     bool
}

// AIResponse represents the response from an AI service
type AIResponse struct {
	Title      string
	Body       string
	Labels     []string
	Reviewers  []string
	Priority   string
	Confidence float32
	TokensUsed int
	Provider   types.AIProvider
}

// PromptTemplate represents a template for AI prompts
type PromptTemplate struct {
	Name        string
	System      string
	User        string
	Examples    []PromptExample
	MaxTokens   int
	Temperature float32
}

// PromptExample represents an example for few-shot learning
type PromptExample struct {
	Input  string
	Output string
}
