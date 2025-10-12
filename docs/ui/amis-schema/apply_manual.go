package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// ManualTranslation represents one manual translation entry
type ManualTranslation struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Tag         string `json:"tag"`
	Value       string `json:"value"`
	Translation string `json:"translation"`
}

// createBackup makes a .bak copy before modifying
func createBackup(filePath string) error {
	backupPath := filePath + ".bak"
	if _, err := os.Stat(backupPath); err == nil {
		return nil // backup already exists
	}
	input, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read error: %v", err)
	}
	if err := os.WriteFile(backupPath, input, 0o644); err != nil {
		return fmt.Errorf("backup write error: %v", err)
	}
	fmt.Printf("✓ Backup created for %s\n", filePath)
	return nil
}

// readLines reads all lines from file
func readLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// applyManualTranslation applies a single translation to a file
func applyManualTranslation(entry ManualTranslation) error {
	fmt.Printf("Processing %s (line %d)...\n", entry.File, entry.Line)
	
	if _, err := os.Stat(entry.File); err != nil {
		fmt.Printf("⚠️  Skipping missing file: %s\n", entry.File)
		return nil
	}

	_ = createBackup(entry.File)
	lines := readLines(entry.File)

	if entry.Line <= 0 || entry.Line > len(lines) {
		fmt.Printf("⚠️  Invalid line number %d for file %s\n", entry.Line, entry.File)
		return nil
	}

	oldLine := lines[entry.Line-1]
	newLine := oldLine

	// Handle different tag types
	switch entry.Tag {
	case "description":
		// Look for "description": "CHINESE TEXT"
		pattern := regexp.MustCompile(`"description"\s*:\s*"` + regexp.QuoteMeta(entry.Value) + `"`)
		newLine = pattern.ReplaceAllString(oldLine, `"description": "`+strings.ReplaceAll(entry.Translation, `"`, `\"`)+`"`)
	
	case "success", "pending", "fail", "queue", "schedule":
		// Handle status fields like "success": "成功"
		pattern := regexp.MustCompile(`"` + entry.Tag + `"\s*:\s*"` + regexp.QuoteMeta(entry.Value) + `"`)
		newLine = pattern.ReplaceAllString(oldLine, `"`+entry.Tag+`": "`+strings.ReplaceAll(entry.Translation, `"`, `\"`)+`"`)
	
	default:
		// Generic pattern for other tags
		pattern := regexp.MustCompile(`"` + entry.Tag + `"\s*:\s*"` + regexp.QuoteMeta(entry.Value) + `"`)
		newLine = pattern.ReplaceAllString(oldLine, `"`+entry.Tag+`": "`+strings.ReplaceAll(entry.Translation, `"`, `\"`)+`"`)
	}

	if newLine != oldLine {
		lines[entry.Line-1] = newLine
		
		output := strings.Join(lines, "\n")
		if err := os.WriteFile(entry.File, []byte(output), 0o644); err != nil {
			return fmt.Errorf("write error in %s: %v", entry.File, err)
		}
		
		fmt.Printf("  ✓ Line %d updated\n", entry.Line)
		return nil
	} else {
		fmt.Printf("  ⚠️  No changes made to line %d (pattern not found)\n", entry.Line)
		return nil
	}
}

func main() {
	const inputJSON = "./manual_translations.json"

	data, err := os.ReadFile(inputJSON)
	if err != nil {
		fmt.Printf("❌ Failed to read %s: %v\n", inputJSON, err)
		os.Exit(1)
	}

	var entries []ManualTranslation
	if err := json.Unmarshal(data, &entries); err != nil {
		fmt.Printf("❌ JSON parse error: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("No manual translations to apply.")
		return
	}

	fmt.Printf("🚀 Applying %d manual translations...\n\n", len(entries))

	successCount := 0
	for _, entry := range entries {
		if err := applyManualTranslation(entry); err != nil {
			fmt.Printf("❌ Error applying translation: %v\n", err)
		} else {
			successCount++
		}
	}

	fmt.Printf("\n🎉 Applied %d/%d translations successfully!\n", successCount, len(entries))
}