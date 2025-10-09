package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileSystemOperations handles file system operations with safety checks
type FileSystemOperations struct {
	dryRun  bool
	verbose bool
}

// NewFileSystemOperations creates a new file system operations handler
func NewFileSystemOperations(dryRun, verbose bool) *FileSystemOperations {
	return &FileSystemOperations{
		dryRun:  dryRun,
		verbose: verbose,
	}
}

// CreateFile creates a file with the given content
func (fso *FileSystemOperations) CreateFile(path, content string) error {
	if fso.verbose {
		if fso.dryRun {
			fmt.Printf("Would create file: %s\n", path)
		} else {
			fmt.Printf("Creating file: %s\n", path)
		}
	}
	
	if fso.dryRun {
		return nil
	}
	
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := fso.EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// Check if file already exists
	if fso.FileExists(path) {
		return fmt.Errorf("file already exists: %s", path)
	}
	
	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer file.Close()
	
	// Write content
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", path, err)
	}
	
	return nil
}

// UpdateFile updates an existing file with new content
func (fso *FileSystemOperations) UpdateFile(path, content string) error {
	if fso.verbose {
		if fso.dryRun {
			fmt.Printf("Would update file: %s\n", path)
		} else {
			fmt.Printf("Updating file: %s\n", path)
		}
	}
	
	if fso.dryRun {
		return nil
	}
	
	// Check if file exists
	if !fso.FileExists(path) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	
	// Create backup if needed
	if err := fso.createBackup(path); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	
	// Write new content
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %w", path, err)
	}
	defer file.Close()
	
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", path, err)
	}
	
	return nil
}

// AppendToFile appends content to an existing file
func (fso *FileSystemOperations) AppendToFile(path, content string) error {
	if fso.verbose {
		if fso.dryRun {
			fmt.Printf("Would append to file: %s\n", path)
		} else {
			fmt.Printf("Appending to file: %s\n", path)
		}
	}
	
	if fso.dryRun {
		return nil
	}
	
	// Check if file exists
	if !fso.FileExists(path) {
		return fmt.Errorf("file does not exist: %s", path)
	}
	
	// Open file for appending
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s for appending: %w", path, err)
	}
	defer file.Close()
	
	// Append content
	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to append content to file %s: %w", path, err)
	}
	
	return nil
}

// EnsureDir creates a directory and all parent directories if they don't exist
func (fso *FileSystemOperations) EnsureDir(path string) error {
	if fso.verbose && !fso.dryRun {
		fmt.Printf("Ensuring directory: %s\n", path)
	}
	
	if fso.dryRun {
		return nil
	}
	
	// Check if directory already exists
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return nil // Directory already exists
		}
		return fmt.Errorf("path exists but is not a directory: %s", path)
	}
	
	// Create directory with all parents
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	
	return nil
}

// FileExists checks if a file exists
func (fso *FileSystemOperations) FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// DirExists checks if a directory exists
func (fso *FileSystemOperations) DirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// ReadFile reads the content of a file
func (fso *FileSystemOperations) ReadFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return string(content), nil
}

// ListFiles lists all files in a directory matching a pattern
func (fso *FileSystemOperations) ListFiles(dir, pattern string) ([]string, error) {
	var files []string
	
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				files = append(files, path)
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %w", dir, err)
	}
	
	return files, nil
}

// ValidateModulePath validates that a module path is safe and follows conventions
func (fso *FileSystemOperations) ValidateModulePath(moduleName string) error {
	// Check for valid characters
	for _, char := range moduleName {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '_') {
			return fmt.Errorf("module name contains invalid character: %c", char)
		}
	}
	
	// Check length
	if len(moduleName) == 0 {
		return fmt.Errorf("module name cannot be empty")
	}
	
	if len(moduleName) > 50 {
		return fmt.Errorf("module name too long (max 50 characters): %s", moduleName)
	}
	
	// Check if it starts with a letter
	firstChar := moduleName[0]
	if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z')) {
		return fmt.Errorf("module name must start with a letter: %s", moduleName)
	}
	
	// Check for reserved names
	reserved := []string{"test", "main", "internal", "cmd", "pkg", "vendor", "build", "bin"}
	for _, reservedName := range reserved {
		if strings.EqualFold(moduleName, reservedName) {
			return fmt.Errorf("module name is reserved: %s", moduleName)
		}
	}
	
	return nil
}

// CheckModuleExists checks if a module already exists
func (fso *FileSystemOperations) CheckModuleExists(moduleName string) bool {
	modulePath := filepath.Join("internal", "core", ToSnakeCase(moduleName))
	return fso.DirExists(modulePath)
}

// GetModulePath returns the path for a module
func (fso *FileSystemOperations) GetModulePath(moduleName string) string {
	return filepath.Join("internal", "core", ToSnakeCase(moduleName))
}

// GetAPIPath returns the API path for a module
func (fso *FileSystemOperations) GetAPIPath(moduleName string) string {
	return filepath.Join("internal", "api", "design", "services", ToSnakeCase(moduleName))
}

// GetDBPath returns the database queries path for a module
func (fso *FileSystemOperations) GetDBPath(moduleName string) string {
	return filepath.Join("db", "queries", ToSnakeCase(moduleName)+".sql")
}

// GetTestPath returns the test path for a module
func (fso *FileSystemOperations) GetTestPath(moduleName string) string {
	return filepath.Join("internal", "core", ToSnakeCase(moduleName), "test")
}

// createBackup creates a backup of an existing file
func (fso *FileSystemOperations) createBackup(path string) error {
	if !fso.FileExists(path) {
		return nil // No backup needed if file doesn't exist
	}
	
	backupPath := path + ".backup"
	content, err := fso.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read original file for backup: %w", err)
	}
	
	backupFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer backupFile.Close()
	
	if _, err := backupFile.WriteString(content); err != nil {
		return fmt.Errorf("failed to write backup content: %w", err)
	}
	
	if fso.verbose {
		fmt.Printf("Created backup: %s\n", backupPath)
	}
	
	return nil
}

// GetProjectRoot finds the project root by looking for go.mod
func (fso *FileSystemOperations) GetProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}
	
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if fso.FileExists(goModPath) {
			return dir, nil
		}
		
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}
	
	return "", fmt.Errorf("could not find go.mod file - make sure you're in a Go project")
}

// IsInProjectRoot checks if we're currently in the project root
func (fso *FileSystemOperations) IsInProjectRoot() bool {
	return fso.FileExists("go.mod") && fso.DirExists("internal")
}

// GetRelativePath returns a path relative to the project root
func (fso *FileSystemOperations) GetRelativePath(absolutePath string) (string, error) {
	projectRoot, err := fso.GetProjectRoot()
	if err != nil {
		return "", err
	}
	
	relativePath, err := filepath.Rel(projectRoot, absolutePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}
	
	return relativePath, nil
}

// ValidateProjectStructure validates that we're in a valid ERP project
func (fso *FileSystemOperations) ValidateProjectStructure() error {
	if !fso.FileExists("go.mod") {
		return fmt.Errorf("go.mod not found - make sure you're in a Go project root")
	}
	
	requiredDirs := []string{
		"internal",
		"internal/core",
		"internal/api",
		"db",
		"cmd",
	}
	
	for _, dir := range requiredDirs {
		if !fso.DirExists(dir) {
			return fmt.Errorf("required directory not found: %s", dir)
		}
	}
	
	// Check for ERP-specific files
	erpFiles := []string{
		"internal/api/design",
		"db/queries",
		"db/migration",
	}
	
	for _, file := range erpFiles {
		if !fso.DirExists(file) {
			return fmt.Errorf("ERP-specific directory not found: %s", file)
		}
	}
	
	return nil
}