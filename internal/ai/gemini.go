package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	
	"auto-pr/pkg/types"
)

// GeminiClient implements AIClient for Google Gemini API
type GeminiClient struct {
	apiKey      string
	projectID   string
	model       string
	maxTokens   int
	temperature float32
	httpClient  *http.Client
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(config types.GeminiConfig) (*GeminiClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("gemini API key is required")
	}
	
	model := config.Model
	if model == "" {
		model = "gemini-2.5-flash" // Default model
	}
	
	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 2048 // Default max tokens
	}
	
	temperature := config.Temperature
	if temperature == 0 {
		temperature = 0.7 // Default temperature
	}
	
	return &GeminiClient{
		apiKey:      config.APIKey,
		projectID:   config.ProjectID,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GenerateContent generates AI content using Gemini API
func (g *GeminiClient) GenerateContent(ctx *AIContext, prompt string) (*AIResponse, error) {
	// Build the full prompt with context
	fullPrompt := g.buildPrompt(ctx, prompt)
	
	// Prepare API request
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": fullPrompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":   g.temperature,
			"maxOutputTokens": g.maxTokens,
			"responseMimeType": "application/json",
		},
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Build API URL
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.model, g.apiKey)
	
	// Make API request
	resp, err := g.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	
	// Parse response
	var apiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			TotalTokenCount int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	if len(apiResponse.Candidates) == 0 || len(apiResponse.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content generated")
	}
	
	content := apiResponse.Candidates[0].Content.Parts[0].Text
	
	// Parse the AI response
	response, err := g.parseResponse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}
	
	response.Provider = types.AIProviderGemini
	response.TokensUsed = apiResponse.UsageMetadata.TotalTokenCount
	
	return response, nil
}

// IsAvailable checks if Gemini API is available
func (g *GeminiClient) IsAvailable() bool {
	return g.apiKey != ""
}

// GetProvider returns the provider type
func (g *GeminiClient) GetProvider() types.AIProvider {
	return types.AIProviderGemini
}

// ValidateConfig validates the Gemini configuration
func (g *GeminiClient) ValidateConfig() error {
	if g.apiKey == "" {
		return fmt.Errorf("gemini API key is required")
	}
	
	if g.maxTokens <= 0 {
		return fmt.Errorf("max tokens must be positive")
	}
	
	if g.temperature < 0 || g.temperature > 2 {
		return fmt.Errorf("temperature must be between 0 and 2")
	}
	
	return nil
}

// buildPrompt builds the full prompt with context for Gemini
func (g *GeminiClient) buildPrompt(ctx *AIContext, basePrompt string) string {
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
	prompt.WriteString("IMPORTANT: Respond with ONLY a valid JSON object, no other text. The JSON must contain:\n")
	prompt.WriteString(`{
  "title": "Brief, descriptive PR title",
  "body": "Detailed PR description in markdown",
  "labels": ["suggested", "labels"],
  "reviewers": ["suggested", "reviewers"],
  "priority": "low|medium|high",
  "confidence": 0.85
}`)
	prompt.WriteString("\n\nDo not include any text before or after the JSON object.")
	
	return prompt.String()
}

// parseResponse parses the Gemini API response
func (g *GeminiClient) parseResponse(output string) (*AIResponse, error) {
	// Clean the output to extract JSON
	output = strings.TrimSpace(output)
	
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
		}, nil
	}
	
	return &AIResponse{
		Title:      parsed.Title,
		Body:       parsed.Body,
		Labels:     parsed.Labels,
		Reviewers:  parsed.Reviewers,
		Priority:   parsed.Priority,
		Confidence: parsed.Confidence,
	}, nil
}