package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	
	"auto-pr/internal/templates"
	
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage PR/MR templates",
	Long:  `Manage templates for different types of pull requests and merge requests.`,
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE:  runTemplateList,
}

var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new template",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateCreate,
}

var templateEditCmd = &cobra.Command{
	Use:   "edit <name>",
	Short: "Edit an existing template",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateEdit,
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a template",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateDelete,
}

var templateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show template content",
	Args:  cobra.ExactArgs(1),
	RunE:  runTemplateShow,
}

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateEditCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	templateCmd.AddCommand(templateShowCmd)
	
	// Add flags
	templateCreateCmd.Flags().String("type", "custom", "Template type (feature, bugfix, hotfix, refactor, docs, custom)")
	templateCreateCmd.Flags().String("from", "", "Base template on existing template")
	templateCreateCmd.Flags().Bool("edit", true, "Open editor after creating")
}

func runTemplateList(cmd *cobra.Command, args []string) error {
	manager := templates.NewManager()
	
	// Get built-in templates
	builtIn := manager.ListBuiltInTemplates()
	if len(builtIn) > 0 {
		fmt.Println("Built-in Templates:")
		fmt.Println("==================")
		for _, tmpl := range builtIn {
			fmt.Printf("- %-15s %s\n", tmpl.Name, tmpl.Description)
		}
		fmt.Println()
	}
	
	// Get custom templates
	custom, err := manager.ListCustomTemplates()
	if err != nil {
		return fmt.Errorf("failed to list custom templates: %w", err)
	}
	
	if len(custom) > 0 {
		fmt.Println("Custom Templates:")
		fmt.Println("=================")
		for _, tmpl := range custom {
			fmt.Printf("- %-15s %s\n", tmpl.Name, tmpl.Description)
		}
		fmt.Println()
	}
	
	if len(builtIn) == 0 && len(custom) == 0 {
		fmt.Println("No templates found.")
		fmt.Println("\nCreate a new template with: auto-pr template create <name>")
	}
	
	return nil
}

func runTemplateCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	templateType, _ := cmd.Flags().GetString("type")
	fromTemplate, _ := cmd.Flags().GetString("from")
	shouldEdit, _ := cmd.Flags().GetBool("edit")
	
	manager := templates.NewManager()
	
	// Create template
	tmpl, err := manager.CreateTemplate(name, templateType, fromTemplate)
	if err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}
	
	fmt.Printf("✅ Created template '%s' at %s\n", name, tmpl.Path)
	
	// Open editor if requested
	if shouldEdit {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // Default to vi
		}
		
		editorCmd := exec.Command(editor, tmpl.Path)
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		
		if err := editorCmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to open editor: %v\n", err)
			fmt.Printf("You can edit the template manually at: %s\n", tmpl.Path)
		}
	}
	
	return nil
}

func runTemplateEdit(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	manager := templates.NewManager()
	tmpl, err := manager.GetTemplate(name)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}
	
	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	
	editorCmd := exec.Command(editor, tmpl.Path)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr
	
	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}
	
	fmt.Printf("✅ Template '%s' updated\n", name)
	return nil
}

func runTemplateDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	manager := templates.NewManager()
	
	// Check if it's a built-in template
	if manager.IsBuiltInTemplate(name) {
		return fmt.Errorf("cannot delete built-in template '%s'", name)
	}
	
	// Confirm deletion
	if !viper.GetBool("force") {
		fmt.Printf("Are you sure you want to delete template '%s'? [y/N] ", name)
		var response string
		fmt.Scanln(&response)
		if !strings.HasPrefix(strings.ToLower(response), "y") {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}
	
	if err := manager.DeleteTemplate(name); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	
	fmt.Printf("✅ Deleted template '%s'\n", name)
	return nil
}

func runTemplateShow(cmd *cobra.Command, args []string) error {
	name := args[0]
	
	manager := templates.NewManager()
	tmpl, err := manager.GetTemplate(name)
	if err != nil {
		return fmt.Errorf("failed to get template: %w", err)
	}
	
	// Load and display template content
	content, err := manager.LoadTemplateContent(tmpl)
	if err != nil {
		return fmt.Errorf("failed to load template content: %w", err)
	}
	
	fmt.Printf("Template: %s\n", tmpl.Name)
	fmt.Printf("Type: %s\n", tmpl.Type)
	if tmpl.Description != "" {
		fmt.Printf("Description: %s\n", tmpl.Description)
	}
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println(content)
	
	return nil
}