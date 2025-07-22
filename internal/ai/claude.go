package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	
	"auto-pr/pkg/types"
)

// ClaudeClient implements AIClient for Claude CLI integration
type ClaudeClient struct {
	cliPath    string
	model      string
	maxTokens  int
	useSession bool
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(config types.ClaudeConfig) (*ClaudeClient, error) {
	cliPath := config.CLIPath
	if cliPath == "" {
		// Try to find claude CLI in PATH
		path, err := exec.LookPath("claude")
		if err != nil {
			return nil, fmt.Errorf("claude CLI not found in PATH and no explicit path provided: %w", err)
		}
		cliPath = path
	}
	
	// Validate that claude CLI is available
	if err := exec.Command(cliPath, "--version").Run(); err != nil {
		return nil, fmt.Errorf("claude CLI not available at %s: %w", cliPath, err)
	}
	
	model := config.Model
	if model == "" {
		model = "claude-3-5-sonnet-20241022" // Default to latest Sonnet
	}
	
	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096 // Default max tokens
	}
	
	return &ClaudeClient{
		cliPath:    cliPath,
		model:      model,
		maxTokens:  maxTokens,
		useSession: config.UseSession,
	}, nil
}

// GenerateContent generates AI content using Claude CLI
func (c *ClaudeClient) GenerateContent(ctx *AIContext, prompt string) (*AIResponse, error) {
	// Build the full prompt with context
	fullPrompt := c.buildPrompt(ctx, prompt)
	
	// Prepare claude CLI command
	args := []string{
		"--print", // Non-interactive mode
		"--output-format", "text", // Text output format
		"--model", c.model, // Specify model
	}
	
	// Execute claude CLI with prompt via stdin
	cmd := exec.Command(c.cliPath, args...)
	cmd.Stdin = strings.NewReader(fullPrompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("claude CLI execution failed: %w\nOutput: %s", err, string(output))
	}
	
	// Parse the response
	response, err := c.parseResponse(string(output))
	if err != nil {
		return nil, fmt.Errorf("failed to parse claude response: %w", err)
	}
	
	response.Provider = types.AIProviderClaude
	return response, nil
}

// IsAvailable checks if Claude CLI is available
func (c *ClaudeClient) IsAvailable() bool {
	return exec.Command(c.cliPath, "--version").Run() == nil
}

// GetProvider returns the provider type
func (c *ClaudeClient) GetProvider() types.AIProvider {
	return types.AIProviderClaude
}

// ValidateConfig validates the Claude configuration
func (c *ClaudeClient) ValidateConfig() error {
	if c.cliPath == "" {
		return fmt.Errorf("claude CLI path not specified")
	}
	
	if _, err := os.Stat(c.cliPath); os.IsNotExist(err) {
		return fmt.Errorf("claude CLI not found at path: %s", c.cliPath)
	}
	
	// Check if Claude CLI is authenticated
	cmd := exec.Command(c.cliPath, "--help")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude CLI not properly configured: %w", err)
	}
	
	return nil
}

// buildPrompt builds the full prompt with context for Claude
func (c *ClaudeClient) buildPrompt(ctx *AIContext, basePrompt string) string {
	var prompt strings.Builder
	
	prompt.WriteString("You are an expert software engineer helping to create a pull request. ")
	prompt.WriteString("Analyze the provided git changes and generate an appropriate PR title and description.\n\n")
	
	// Add context information
	if len(ctx.CommitHistory) > 0 {
		prompt.WriteString("## Recent Commits:\n")
		for _, commit := range ctx.CommitHistory {
			prompt.WriteString(fmt.Sprintf("- %s: %s\n", commit.Hash[:8], commit.Message))
		}
		prompt.WriteString("\n")
	}
	
	if ctx.DiffSummary != "" {
		prompt.WriteString("## Changes Summary:\n")
		prompt.WriteString(ctx.DiffSummary)
		prompt.WriteString("\n\n")
	}
	
	if len(ctx.FileChanges) > 0 {
		prompt.WriteString("## Files Changed:\n")
		for _, file := range ctx.FileChanges {
			prompt.WriteString(fmt.Sprintf("- %s (%s): +%d -%d\n", 
				file.Path, file.Status, file.Additions, file.Deletions))
		}
		prompt.WriteString("\n")
	}
	
	// Add project context
	if ctx.ProjectContext.Language != "" {
		prompt.WriteString(fmt.Sprintf("## Project Info:\n- Language: %s\n", ctx.ProjectContext.Language))
		if ctx.ProjectContext.Framework != "" {
			prompt.WriteString(fmt.Sprintf("- Framework: %s\n", ctx.ProjectContext.Framework))
		}
		prompt.WriteString("\n")
	}
	
	// Add the base prompt
	prompt.WriteString("## Task:\n")
	prompt.WriteString(basePrompt)
	prompt.WriteString("\n\n")
	
	// Add output format requirements
	prompt.WriteString("Please respond with a JSON object containing:\n")
	prompt.WriteString(`{
  "title": "Brief, descriptive PR title",
  "body": "Detailed PR description in markdown",
  "labels": ["suggested", "labels"],
  "reviewers": ["suggested", "reviewers"],
  "priority": "low|medium|high",
  "confidence": 0.85
}`)
	
	return prompt.String()
}

// parseResponse parses the Claude CLI response
func (c *ClaudeClient) parseResponse(output string) (*AIResponse, error) {
	// Try to find JSON in the output
	start := strings.Index(output, "{")
	end := strings.LastIndex(output, "}")
	
	if start == -1 || end == -1 || start >= end {
		// Fallback: create a basic response from the raw output
		return &AIResponse{
			Title:      "Auto-generated PR",
			Body:       output,
			Labels:     []string{"auto-generated"},
			Reviewers:  []string{},
			Priority:   "medium",
			Confidence: 0.5,
			TokensUsed: len(output) / 4, // Rough estimation
		}, nil
	}
	
	jsonStr := output[start : end+1]
	
	var parsed struct {
		Title      string   `json:"title"`
		Body       string   `json:"body"`
		Labels     []string `json:"labels"`
		Reviewers  []string `json:"reviewers"`
		Priority   string   `json:"priority"`
		Confidence float32  `json:"confidence"`
	}
	
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		// Fallback to raw response if JSON parsing fails
		return &AIResponse{
			Title:      "Auto-generated PR",
			Body:       output,
			Labels:     []string{"auto-generated"},
			Reviewers:  []string{},
			Priority:   "medium",
			Confidence: 0.5,
			TokensUsed: len(output) / 4,
		}, nil
	}
	
	return &AIResponse{
		Title:      parsed.Title,
		Body:       parsed.Body,
		Labels:     parsed.Labels,
		Reviewers:  parsed.Reviewers,
		Priority:   parsed.Priority,
		Confidence: parsed.Confidence,
		TokensUsed: len(output) / 4, // Rough estimation
	}, nil
}