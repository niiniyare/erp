package migration

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFiles writes all files to outDir, skipping any file that already exists.
// Returns the list of files written.
func WriteFiles(files []File, outDir string) ([]string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("migration.WriteFiles: mkdir %s: %w", outDir, err)
	}

	var written []string
	for _, f := range files {
		path := filepath.Join(outDir, f.Name)
		if _, err := os.Stat(path); err == nil {
			// Skip existing — never clobber hand-edited migrations.
			continue
		}
		if err := os.WriteFile(path, []byte(f.Content), 0o644); err != nil {
			return written, fmt.Errorf("migration.WriteFiles: write %s: %w", path, err)
		}
		written = append(written, path)
	}
	return written, nil
}

// Preview returns the content of all files without writing them.
// Useful for dry-run / review before committing.
func Preview(files []File) string {
	var b []byte
	for _, f := range files {
		b = append(b, fmt.Sprintf("\n-- === %s ===\n", f.Name)...)
		b = append(b, f.Content...)
	}
	return string(b)
}
