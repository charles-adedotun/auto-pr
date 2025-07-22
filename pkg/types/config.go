package types

// Config represents the application configuration
type Config struct {
	AI        AIConfig       `yaml:"ai"`
	Platforms PlatformConfig `yaml:"platforms"`
	Templates TemplateConfig `yaml:"templates"`
	Git       GitConfig      `yaml:"git"`
}

// AIConfig contains AI service configuration
type AIConfig struct {
	Provider    AIProvider   `yaml:"provider"`
	Model       string       `yaml:"model"`
	MaxTokens   int          `yaml:"max_tokens"`
	Temperature float32      `yaml:"temperature"`
	Claude      ClaudeConfig `yaml:"claude,omitempty"`
}

// AIProvider represents different AI service providers
type AIProvider string

const (
	AIProviderClaude AIProvider = "claude"
)

// ClaudeConfig contains Claude-specific configuration
type ClaudeConfig struct {
	CLIPath    string `yaml:"cli_path,omitempty"`
	Model      string `yaml:"model,omitempty"`
	MaxTokens  int    `yaml:"max_tokens,omitempty"`
	UseSession bool   `yaml:"use_session,omitempty"`
}


// PlatformConfig contains platform-specific settings
type PlatformConfig struct {
	GitHub GitHubConfig `yaml:"github"`
	GitLab GitLabConfig `yaml:"gitlab"`
}

// GitHubConfig contains GitHub-specific settings
type GitHubConfig struct {
	DefaultReviewers []string `yaml:"default_reviewers"`
	Labels           []string `yaml:"labels"`
	Draft            bool     `yaml:"draft"`
	AutoMerge        bool     `yaml:"auto_merge"`
	DeleteBranch     bool     `yaml:"delete_branch"`
}

// GitLabConfig contains GitLab-specific settings
type GitLabConfig struct {
	DefaultAssignee           string `yaml:"default_assignee"`
	MergeWhenPipelineSucceeds bool   `yaml:"merge_when_pipeline_succeeds"`
	RemoveSourceBranch        bool   `yaml:"remove_source_branch"`
}

// TemplateConfig contains template-related settings
type TemplateConfig struct {
	Feature           string `yaml:"feature"`
	Bugfix            string `yaml:"bugfix"`
	CustomTemplateDir string `yaml:"custom_templates_dir"`
}

// GitConfig contains git-related settings
type GitConfig struct {
	CommitLimit    int      `yaml:"commit_limit"`
	DiffContext    int      `yaml:"diff_context"`
	IgnorePatterns []string `yaml:"ignore_patterns"`
	MaxDiffSize    int      `yaml:"max_diff_size"`
}

// PlatformType represents different git platforms
type PlatformType string

const (
	PlatformGitHub  PlatformType = "github"
	PlatformGitLab  PlatformType = "gitlab"
	PlatformUnknown PlatformType = "unknown"
)
