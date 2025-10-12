package generator

import (
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

// DocsGenerator handles documentation generation for modules
type DocsGenerator struct {
	fs   *FileSystemOperations
	tmpl *template.Template
}

// NewDocsGenerator creates a new documentation generator
func NewDocsGenerator() (*DocsGenerator, error) {
	tmpl, err := template.New("docs").Funcs(TemplateFunctions()).ParseFS(templateFS, "templates/module/*")
	if err != nil {
		return nil, fmt.Errorf("failed to parse documentation templates: %w", err)
	}
	
	return &DocsGenerator{
		tmpl: tmpl,
	}, nil
}

// Generate generates documentation for a module
func (g *DocsGenerator) Generate(config DocsConfig) error {
	// Initialize file system operations
	g.fs = NewFileSystemOperations(config.DryRun, config.Verbose)
	
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Validate project structure
	if err := g.fs.ValidateProjectStructure(); err != nil {
		return fmt.Errorf("invalid project structure: %w", err)
	}
	
	templateData := config.ToTemplateData()
	
	// Define the documentation files to generate
	docFiles := []FileTemplate{
		{
			TemplatePath: "README.md.tmpl",
			OutputPath:   filepath.Join("docs", "reference", "modules", templateData.ModuleNameSnake, "README.md"),
			Data:         templateData,
		},
		{
			TemplatePath: "api-reference.md.tmpl", 
			OutputPath:   filepath.Join("docs", "reference", "modules", templateData.ModuleNameSnake, "api-reference.md"),
			Data:         templateData,
		},
		{
			TemplatePath: "testing.md.tmpl",
			OutputPath:   filepath.Join("docs", "reference", "modules", templateData.ModuleNameSnake, "testing.md"),
			Data:         templateData,
		},
		{
			TemplatePath: "architecture-guide.md.tmpl",
			OutputPath:   filepath.Join("docs", "reference", "modules", templateData.ModuleNameSnake, "architecture-guide.md"),
			Data:         templateData,
		},
	}
	
	if config.Verbose {
		fmt.Printf("Generating documentation for module: %s\n", templateData.ModuleNamePascal)
	}
	
	// Generate each documentation file
	for _, docFile := range docFiles {
		if err := g.generateFile(docFile); err != nil {
			return fmt.Errorf("failed to generate file %s: %w", docFile.OutputPath, err)
		}
	}
	
	if config.DryRun {
		fmt.Printf("🔍 Dry run completed - no files were created\n")
	} else {
		fmt.Printf("📚 Documentation generation completed successfully\n")
	}
	
	return nil
}

func (g *DocsGenerator) generateFile(file FileTemplate) error {
	// Check if file exists and handle accordingly
	if g.fs.FileExists(file.OutputPath) {
		if g.fs.IsVerbose() {
			fmt.Printf("Overwriting existing file: %s\n", file.OutputPath)
		}
	}

	// Execute template
	var content strings.Builder
	if err := g.tmpl.ExecuteTemplate(&content, file.TemplatePath, file.Data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", file.TemplatePath, err)
	}

	// Create the file
	if err := g.fs.CreateFile(file.OutputPath, content.String()); err != nil {
		return fmt.Errorf("failed to create file %s: %w", file.OutputPath, err)
	}

	if g.fs.IsVerbose() {
		fmt.Printf("✅ Generated: %s\n", file.OutputPath)
	}

	return nil
}