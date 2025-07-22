package templates

import (
	"fmt"
	"strings"
	
	"auto-pr/internal/ai"
)

// BuildTemplateContext creates a template context from AI context and response
func BuildTemplateContext(aiCtx *ai.AIContext, aiResp *ai.AIResponse) *TemplateContext {
	ctx := &TemplateContext{
		Title:       aiResp.Title,
		Type:        detectChangeType(aiCtx),
		Summary:     extractSummary(aiResp.Body),
		Description: aiResp.Body,
		Custom:      make(map[string]interface{}),
	}
	
	// Add branch info
	if aiCtx.BranchInfo.Name != "" {
		ctx.Branch = aiCtx.BranchInfo.Name
		ctx.BaseBranch = aiCtx.BranchInfo.BaseBranch
	}
	
	// Add change statistics
	if len(aiCtx.CommitHistory) > 0 {
		ctx.CommitCount = len(aiCtx.CommitHistory)
		ctx.Commits = aiCtx.CommitHistory
	}
	
	if len(aiCtx.FileChanges) > 0 {
		ctx.FilesChanged = len(aiCtx.FileChanges)
		ctx.FileChanges = aiCtx.FileChanges
		
		// Calculate totals
		for _, fc := range aiCtx.FileChanges {
			ctx.Additions += fc.Additions
			ctx.Deletions += fc.Deletions
		}
	}
	
	// Extract changes from AI response
	ctx.Changes = extractBulletPoints(aiResp.Body, "changes", "modifications")
	ctx.TestPlan = extractBulletPoints(aiResp.Body, "test", "testing")
	
	// Add custom fields
	ctx.Custom["labels"] = aiResp.Labels
	ctx.Custom["reviewers"] = aiResp.Reviewers
	ctx.Custom["priority"] = aiResp.Priority
	ctx.Custom["confidence"] = aiResp.Confidence
	
	return ctx
}

// EnhanceWithTemplate enhances AI response using a template
func EnhanceWithTemplate(manager *Manager, templateName string, aiCtx *ai.AIContext, aiResp *ai.AIResponse) (*ai.AIResponse, error) {
	// Build template context
	ctx := BuildTemplateContext(aiCtx, aiResp)
	
	// Render template
	body, err := manager.RenderTemplate(templateName, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}
	
	// Create enhanced response
	enhanced := &ai.AIResponse{
		Title:      aiResp.Title,
		Body:       body,
		Labels:     enhanceLabels(templateName, aiResp.Labels),
		Reviewers:  aiResp.Reviewers,
		Priority:   aiResp.Priority,
		Confidence: aiResp.Confidence,
		Provider:   aiResp.Provider,
		TokensUsed: aiResp.TokensUsed,
	}
	
	return enhanced, nil
}

// detectChangeType attempts to detect the type of change
func detectChangeType(ctx *ai.AIContext) string {
	// Check commit messages
	for _, commit := range ctx.CommitHistory {
		msg := strings.ToLower(commit.Message)
		if strings.Contains(msg, "feat") || strings.Contains(msg, "feature") {
			return "feature"
		}
		if strings.Contains(msg, "fix") || strings.Contains(msg, "bug") {
			return "bugfix"
		}
		if strings.Contains(msg, "hotfix") || strings.Contains(msg, "critical") {
			return "hotfix"
		}
		if strings.Contains(msg, "refactor") {
			return "refactor"
		}
		if strings.Contains(msg, "doc") {
			return "docs"
		}
		if strings.Contains(msg, "test") {
			return "test"
		}
		if strings.Contains(msg, "dep") || strings.Contains(msg, "upgrade") {
			return "deps"
		}
	}
	
	// Check file changes
	hasTests := false
	hasDocs := false
	hasDeps := false
	
	for _, fc := range ctx.FileChanges {
		path := strings.ToLower(fc.Path)
		if strings.Contains(path, "test") || strings.HasSuffix(path, "_test.go") {
			hasTests = true
		}
		if strings.Contains(path, "readme") || strings.Contains(path, "doc") || strings.HasSuffix(path, ".md") {
			hasDocs = true
		}
		if strings.Contains(path, "go.mod") || strings.Contains(path, "package.json") || strings.Contains(path, "requirements.txt") {
			hasDeps = true
		}
	}
	
	if hasTests && !hasDocs && !hasDeps {
		return "test"
	}
	if hasDocs && !hasTests && !hasDeps {
		return "docs"
	}
	if hasDeps {
		return "deps"
	}
	
	return "feature" // default
}

// extractSummary extracts a summary from the body
func extractSummary(body string) string {
	// Look for summary section
	lines := strings.Split(body, "\n")
	inSummary := false
	summaryLines := []string{}
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		
		// Check if we're in summary section
		if strings.Contains(trimmed, "summary") && strings.HasPrefix(line, "#") {
			inSummary = true
			continue
		}
		
		// Check if we're leaving summary section
		if inSummary && strings.HasPrefix(line, "#") {
			break
		}
		
		// Collect summary lines
		if inSummary && strings.TrimSpace(line) != "" {
			summaryLines = append(summaryLines, strings.TrimSpace(line))
		}
	}
	
	if len(summaryLines) > 0 {
		summary := strings.Join(summaryLines, " ")
		// Remove bullet points
		summary = strings.TrimPrefix(summary, "- ")
		if len(summary) > 200 {
			return summary[:197] + "..."
		}
		return summary
	}
	
	// Fallback: use first non-empty, non-header line
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "-") {
			if len(line) > 200 {
				return line[:197] + "..."
			}
			return line
		}
	}
	
	return "No summary provided"
}

// extractBulletPoints extracts bullet points related to keywords
func extractBulletPoints(body string, keywords ...string) []string {
	var points []string
	lines := strings.Split(body, "\n")
	inSection := false
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ToLower(line))
		
		// Check if we're entering a relevant section
		for _, keyword := range keywords {
			if strings.Contains(trimmed, keyword) && strings.HasPrefix(line, "#") {
				inSection = true
				break
			}
		}
		
		// Check if we're leaving the section
		if inSection && strings.HasPrefix(line, "#") && !containsAny(trimmed, keywords...) {
			inSection = false
		}
		
		// Extract bullet points
		if inSection && strings.HasPrefix(strings.TrimSpace(line), "-") {
			point := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
			if point != "" {
				points = append(points, point)
			}
		}
	}
	
	return points
}

// containsAny checks if text contains any of the keywords
func containsAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// enhanceLabels adds template-specific labels
func enhanceLabels(templateName string, existing []string) []string {
	labels := make([]string, len(existing))
	copy(labels, existing)
	
	// Add template-specific label if not present
	templateLabel := templateName
	hasLabel := false
	for _, label := range labels {
		if label == templateLabel {
			hasLabel = true
			break
		}
	}
	
	if !hasLabel {
		labels = append(labels, templateLabel)
	}
	
	return labels
}

// SelectTemplateByContext automatically selects a template based on context
func SelectTemplateByContext(ctx *ai.AIContext) string {
	changeType := detectChangeType(ctx)
	
	// Map change types to template names
	templateMap := map[string]string{
		"feature": "feature",
		"bugfix":  "bugfix",
		"hotfix":  "hotfix",
		"refactor": "refactor",
		"docs":    "docs",
		"test":    "test",
		"deps":    "deps",
	}
	
	if template, ok := templateMap[changeType]; ok {
		return template
	}
	
	return "feature" // default
}