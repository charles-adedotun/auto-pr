package templates

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	
	"auto-pr/pkg/types"
)

//go:embed builtin/*.tmpl
var builtinTemplates embed.FS

// Manager handles template operations
type Manager struct {
	customDir string
}

// Template represents a PR/MR template
type Template struct {
	Name        string
	Type        string
	Description string
	Path        string
	IsBuiltIn   bool
}

// NewManager creates a new template manager
func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	customDir := filepath.Join(homeDir, ".auto-pr", "templates")
	
	// Ensure custom templates directory exists
	os.MkdirAll(customDir, 0755)
	
	return &Manager{
		customDir: customDir,
	}
}

// ListBuiltInTemplates returns all built-in templates
func (m *Manager) ListBuiltInTemplates() []Template {
	templates := []Template{
		{
			Name:        "feature",
			Type:        "feature",
			Description: "New feature or enhancement",
			IsBuiltIn:   true,
		},
		{
			Name:        "bugfix",
			Type:        "bugfix",
			Description: "Bug fix or issue resolution",
			IsBuiltIn:   true,
		},
		{
			Name:        "hotfix",
			Type:        "hotfix",
			Description: "Critical production fix",
			IsBuiltIn:   true,
		},
		{
			Name:        "refactor",
			Type:        "refactor",
			Description: "Code refactoring without functionality change",
			IsBuiltIn:   true,
		},
		{
			Name:        "docs",
			Type:        "docs",
			Description: "Documentation updates",
			IsBuiltIn:   true,
		},
		{
			Name:        "test",
			Type:        "test",
			Description: "Test additions or modifications",
			IsBuiltIn:   true,
		},
		{
			Name:        "deps",
			Type:        "deps",
			Description: "Dependency updates",
			IsBuiltIn:   true,
		},
	}
	
	return templates
}

// ListCustomTemplates returns all custom templates
func (m *Manager) ListCustomTemplates() ([]Template, error) {
	var templates []Template
	
	entries, err := os.ReadDir(m.customDir)
	if err != nil {
		if os.IsNotExist(err) {
			return templates, nil
		}
		return nil, err
	}
	
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}
		
		name := strings.TrimSuffix(entry.Name(), ".tmpl")
		templates = append(templates, Template{
			Name:      name,
			Type:      "custom",
			Path:      filepath.Join(m.customDir, entry.Name()),
			IsBuiltIn: false,
		})
	}
	
	return templates, nil
}

// GetTemplate retrieves a template by name
func (m *Manager) GetTemplate(name string) (*Template, error) {
	// Check built-in templates first
	for _, tmpl := range m.ListBuiltInTemplates() {
		if tmpl.Name == name {
			return &tmpl, nil
		}
	}
	
	// Check custom templates
	customPath := filepath.Join(m.customDir, name+".tmpl")
	if _, err := os.Stat(customPath); err == nil {
		return &Template{
			Name:      name,
			Type:      "custom",
			Path:      customPath,
			IsBuiltIn: false,
		}, nil
	}
	
	return nil, fmt.Errorf("template '%s' not found", name)
}

// CreateTemplate creates a new custom template
func (m *Manager) CreateTemplate(name, templateType, fromTemplate string) (*Template, error) {
	// Check if template already exists
	if _, err := m.GetTemplate(name); err == nil {
		return nil, fmt.Errorf("template '%s' already exists", name)
	}
	
	// Create template file
	path := filepath.Join(m.customDir, name+".tmpl")
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create template file: %w", err)
	}
	defer file.Close()
	
	// Write initial content
	var content string
	if fromTemplate != "" {
		// Copy from existing template
		baseTmpl, err := m.GetTemplate(fromTemplate)
		if err != nil {
			return nil, fmt.Errorf("base template '%s' not found", fromTemplate)
		}
		
		baseContent, err := m.LoadTemplateContent(baseTmpl)
		if err != nil {
			return nil, fmt.Errorf("failed to load base template: %w", err)
		}
		content = baseContent
	} else {
		// Use default content based on type
		content = m.getDefaultTemplateContent(templateType)
	}
	
	if _, err := file.WriteString(content); err != nil {
		return nil, fmt.Errorf("failed to write template content: %w", err)
	}
	
	return &Template{
		Name:      name,
		Type:      templateType,
		Path:      path,
		IsBuiltIn: false,
	}, nil
}

// LoadTemplateContent loads the content of a template
func (m *Manager) LoadTemplateContent(tmpl *Template) (string, error) {
	if tmpl.IsBuiltIn {
		// Load from embedded templates
		content, err := builtinTemplates.ReadFile(fmt.Sprintf("builtin/%s.tmpl", tmpl.Name))
		if err != nil {
			return "", fmt.Errorf("failed to load built-in template: %w", err)
		}
		return string(content), nil
	}
	
	// Load from file
	content, err := os.ReadFile(tmpl.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}
	
	return string(content), nil
}

// DeleteTemplate deletes a custom template
func (m *Manager) DeleteTemplate(name string) error {
	tmpl, err := m.GetTemplate(name)
	if err != nil {
		return err
	}
	
	if tmpl.IsBuiltIn {
		return fmt.Errorf("cannot delete built-in template")
	}
	
	return os.Remove(tmpl.Path)
}

// IsBuiltInTemplate checks if a template is built-in
func (m *Manager) IsBuiltInTemplate(name string) bool {
	for _, tmpl := range m.ListBuiltInTemplates() {
		if tmpl.Name == name {
			return true
		}
	}
	return false
}

// RenderTemplate renders a template with the given context
func (m *Manager) RenderTemplate(templateName string, ctx *TemplateContext) (string, error) {
	tmpl, err := m.GetTemplate(templateName)
	if err != nil {
		return "", err
	}
	
	content, err := m.LoadTemplateContent(tmpl)
	if err != nil {
		return "", err
	}
	
	// Parse and execute template
	t, err := template.New(templateName).Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	
	var buf strings.Builder
	if err := t.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	
	return buf.String(), nil
}

// getDefaultTemplateContent returns default content for a template type
func (m *Manager) getDefaultTemplateContent(templateType string) string {
	switch templateType {
	case "feature":
		return defaultFeatureTemplate
	case "bugfix":
		return defaultBugfixTemplate
	case "hotfix":
		return defaultHotfixTemplate
	case "refactor":
		return defaultRefactorTemplate
	case "docs":
		return defaultDocsTemplate
	default:
		return defaultCustomTemplate
	}
}

// TemplateContext contains data for rendering templates
type TemplateContext struct {
	// PR/MR metadata
	Title       string
	Type        string
	Branch      string
	BaseBranch  string
	Author      string
	Date        time.Time
	
	// Change information
	CommitCount   int
	FilesChanged  int
	Additions     int
	Deletions     int
	Commits       []types.CommitInfo
	FileChanges   []types.FileChange
	
	// Content sections
	Summary       string
	Description   string
	Changes       []string
	TestPlan      []string
	Checklist     []string
	
	// Custom fields
	Custom        map[string]interface{}
}

// Default template contents
const (
	defaultFeatureTemplate = `{{.Title}}

## Summary
{{.Summary}}

## Changes
{{range .Changes}}- {{.}}
{{end}}

## Testing
{{range .TestPlan}}- [ ] {{.}}
{{end}}

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Tests added/updated
- [ ] Documentation updated
`

	defaultBugfixTemplate = `{{.Title}}

## Bug Description
{{.Description}}

## Root Cause
[Describe the root cause of the bug]

## Solution
{{.Summary}}

## Changes
{{range .Changes}}- {{.}}
{{end}}

## Testing
- [ ] Bug reproduced before fix
- [ ] Bug resolved after fix
- [ ] Regression tests added
{{range .TestPlan}}- [ ] {{.}}
{{end}}
`

	defaultHotfixTemplate = `{{.Title}}

## Critical Issue
{{.Description}}

## Impact
- [ ] Production affected
- [ ] Users affected: [number/percentage]
- [ ] Revenue impact: [if applicable]

## Fix
{{.Summary}}

## Changes
{{range .Changes}}- {{.}}
{{end}}

## Deployment Plan
- [ ] Tested in staging
- [ ] Rollback plan prepared
- [ ] Monitoring alerts configured

## Post-Deployment
- [ ] Verify fix in production
- [ ] Monitor for 24 hours
- [ ] Create follow-up ticket for permanent fix
`

	defaultRefactorTemplate = `{{.Title}}

## Motivation
{{.Description}}

## Changes
{{range .Changes}}- {{.}}
{{end}}

## Benefits
- Improved code readability
- Better performance
- Reduced technical debt

## Testing
- [ ] All existing tests pass
- [ ] No functionality changes
- [ ] Performance benchmarks (if applicable)
`

	defaultDocsTemplate = `{{.Title}}

## Documentation Updates
{{.Summary}}

## Changes
{{range .Changes}}- {{.}}
{{end}}

## Review Checklist
- [ ] Grammar and spelling checked
- [ ] Technical accuracy verified
- [ ] Examples tested
- [ ] Links validated
`

	defaultCustomTemplate = `{{.Title}}

## Summary
{{.Summary}}

## Changes
{{range .Changes}}- {{.}}
{{end}}

## Notes
{{.Description}}
`
)