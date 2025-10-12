package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	DEFAULT_FILE  = "./schema.json"
	PROGRESS_FILE = "./translation_progress.log"
	MAX_WORKERS   = 5
)

// TranslationEntry represents one record from chines.json
type TranslationEntry struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Tag         string `json:"tag"`
	Description string `json:"value"`       // maps to "description"
	Translation string `json:"translation"` // English text
}

// TranslationResult holds the result of a translation attempt
type TranslationResult struct {
	Chinese     string
	Translation string
	Success     bool
	Error       error
}

// translateText calls the external 'trans' command to translate text
func translateText(text string) (string, error) {
	if text == "" {
		return "", fmt.Errorf("empty text")
	}

	// Call trans command: trans zh-CN:en -b "text"
	cmd := exec.Command("trans", "zh-CN:en", "-b", text)

	// Set timeout to prevent hanging
	timer := time.AfterFunc(10*time.Second, func() {
		cmd.Process.Kill()
	})
	defer timer.Stop()

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("translation failed: %v", err)
	}

	// Parse output - take first line only
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", fmt.Errorf("empty translation result")
	}

	return strings.TrimSpace(lines[0]), nil
}

// worker processes translation jobs from the jobs channel
func worker(id int, jobs <-chan string, results chan<- TranslationResult, wg *sync.WaitGroup, counter *atomic.Int64, total int) {
	defer wg.Done()

	for chinese := range jobs {
		translation, err := translateText(chinese)

		result := TranslationResult{
			Chinese: chinese,
			Success: err == nil,
			Error:   err,
		}

		if err == nil {
			result.Translation = translation
		}

		results <- result

		// Update progress counter
		count := counter.Add(1)
		if count%10 == 0 || count == int64(total) {
			fmt.Printf("Progress: %d/%d (%.1f%%)\n", count, total, float64(count)/float64(total)*100)
		}
	}
}

// escapeRegex escapes special regex characters
func escapeRegex(s string) string {
	specialChars := `\.+*?()|[]{}^$`
	var result strings.Builder
	result.Grow(len(s) * 2)

	for _, char := range s {
		if strings.ContainsRune(specialChars, char) {
			result.WriteRune('\\')
		}
		result.WriteRune(char)
	}

	return result.String()
}

// createBackup creates a backup of the original file
func createBackup(filePath string) error {
	backupPath := filePath + ".bak"
	if _, err := os.Stat(backupPath); err == nil {
		return nil // backup already exists
	}

	input, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %v", err)
	}

	if err := os.WriteFile(backupPath, input, 0o644); err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}

	fmt.Printf("✓ Backup created: %s\n", backupPath)
	return nil
}

// extractUniqueTexts finds all Chinese texts that need translation
func extractUniqueTexts(content string) map[string]bool {
	// Pattern matches: "value": "text", "translation": ""
	pattern := regexp.MustCompile(`"value":\s*"([^"]+)",\s*"translation":\s*""`)
	matches := pattern.FindAllStringSubmatch(content, -1)

	uniqueTexts := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			uniqueTexts[match[1]] = true
		}
	}

	return uniqueTexts
}

// applyTranslations replaces empty translation fields with actual translations
func applyTranslations(content string, translations map[string]string) string {
	for chinese, translation := range translations {
		// Escape special characters
		escapedChinese := escapeRegex(chinese)

		// Escape backslashes and quotes in translation
		escapedTranslation := strings.ReplaceAll(translation, `\`, `\\`)
		escapedTranslation = strings.ReplaceAll(escapedTranslation, `"`, `\"`)

		// Build regex pattern
		pattern := fmt.Sprintf(`"value":\s*"%s",\s*"translation":\s*""`, escapedChinese)

		// Build replacement string
		replacement := fmt.Sprintf(`"value": "%s",
    "translation": "%s"`, chinese, escapedTranslation)

		// Apply replacement
		re := regexp.MustCompile(pattern)
		content = re.ReplaceAllString(content, replacement)
	}

	return content
}

// saveProgressLog writes the progress log to file
func saveProgressLog(logPath string, logs []string) error {
	file, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create progress file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range logs {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write progress log: %v", err)
		}
	}

	return writer.Flush()
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

// runTranslation mode - extract and translate Chinese texts
func runTranslation(filePath string) {
	startTime := time.Now()

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║  Ultra-Fast Batch Translation Tool (Go)               ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Printf("\nFile: %s\n", filePath)
	fmt.Printf("Workers: %d concurrent goroutines\n\n", MAX_WORKERS)

	// Step 1: Create backup
	if err := createBackup(filePath); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Read entire file into memory
	fmt.Println("\n📖 Reading file...")
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("ERROR: Failed to read file: %v\n", err)
		os.Exit(1)
	}
	content := string(contentBytes)
	fmt.Printf("✓ File loaded: %d bytes\n", len(contentBytes))

	// Step 3: Extract unique Chinese texts needing translation
	fmt.Println("\n🔍 Extracting texts to translate...")
	uniqueTexts := extractUniqueTexts(content)
	total := len(uniqueTexts)

	fmt.Printf("✓ Found %d unique texts to translate\n", total)

	if total == 0 {
		fmt.Println("\n✓ No translations needed! All fields are already filled.")
		return
	}

	// Step 4: Set up concurrent translation pipeline
	fmt.Printf("\n🚀 Starting translation with %d workers...\n\n", MAX_WORKERS)

	jobs := make(chan string, total)
	results := make(chan TranslationResult, total)
	var wg sync.WaitGroup
	var counter atomic.Int64

	// Start worker goroutines
	for i := 0; i < MAX_WORKERS; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg, &counter, total)
	}

	// Send all jobs to workers
	for text := range uniqueTexts {
		jobs <- text
	}
	close(jobs)

	// Wait for all workers to finish, then close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Step 5: Collect translation results
	translations := make(map[string]string)
	var progressLog []string
	successCount := 0
	failCount := 0

	for result := range results {
		if result.Success {
			translations[result.Chinese] = result.Translation
			successCount++
			progressLog = append(progressLog,
				fmt.Sprintf("SUCCESS: %s → %s", result.Chinese, result.Translation))
		} else {
			failCount++
			progressLog = append(progressLog,
				fmt.Sprintf("FAILED: %s (Error: %v)", result.Chinese, result.Error))
		}
	}

	fmt.Printf("\n✓ Translation phase complete!\n")
	fmt.Printf("  Success: %d\n", successCount)
	fmt.Printf("  Failed: %d\n", failCount)

	// Step 6: Apply translations to content
	if successCount > 0 {
		fmt.Println("\n📝 Applying translations to file...")
		content = applyTranslations(content, translations)

		// Write updated content back to file
		if err := os.WriteFile(filePath, []byte(content), 0o644); err != nil {
			fmt.Printf("ERROR: Failed to write file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ File updated successfully")
	}

	// Step 7: Save progress log
	fmt.Println("\n💾 Saving progress log...")
	if err := saveProgressLog(PROGRESS_FILE, progressLog); err != nil {
		fmt.Printf("WARNING: Failed to save progress log: %v\n", err)
	} else {
		fmt.Printf("✓ Progress log saved: %s\n", PROGRESS_FILE)
	}

	// Step 8: Final statistics
	remaining := strings.Count(content, `"translation": ""`)
	elapsed := time.Since(startTime)

	fmt.Println("\n" + strings.Repeat("═", 56))
	fmt.Println("TRANSLATION SUMMARY")
	fmt.Println(strings.Repeat("═", 56))
	fmt.Printf("Total unique texts:     %d\n", total)
	fmt.Printf("Successfully translated: %d\n", successCount)
	fmt.Printf("Failed:                 %d\n", failCount)
	fmt.Printf("Remaining empty:        %d\n", remaining)
	fmt.Printf("Time taken:             %v\n", elapsed)
	fmt.Printf("Speed:                  %.1f translations/sec\n", float64(total)/elapsed.Seconds())
	fmt.Println(strings.Repeat("═", 56))

	// Step 9: Post-translation verification (Termux safe)
	fmt.Println("\n🔍 Running verification check for remaining Chinese text...")

	checkCmd := exec.Command("sh", "-c", `find ./output -type f -name "*.json" ! -name "*.bak" -exec grep -q "[一-龯]" {} \; && echo "⚠️  Chinese text still found!" || echo "✅ All translations applied successfully."`)
	output, err := checkCmd.CombinedOutput()

	if err != nil {
		fmt.Printf("⚠️  Verification command failed: %v\n", err)
	} else {
		fmt.Println(strings.TrimSpace(string(output)))
	}

	fmt.Println("\n✨ Done!")
}

// runApplication mode - apply existing translations to files
func runApplication(inputJSON string) {
	data, err := os.ReadFile(inputJSON)
	if err != nil {
		fmt.Printf("❌ Failed to read %s: %v\n", inputJSON, err)
		os.Exit(1)
	}

	var entries []TranslationEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		fmt.Printf("❌ JSON parse error: %v\n", err)
		os.Exit(1)
	}

	// Group entries by file
	grouped := make(map[string][]TranslationEntry)
	for _, e := range entries {
		if strings.TrimSpace(e.Translation) == "" {
			continue // skip untranslated lines
		}
		grouped[e.File] = append(grouped[e.File], e)
	}

	if len(grouped) == 0 {
		fmt.Println("No translations to apply.")
		return
	}

	for filePath, list := range grouped {
		fmt.Printf("\n📄 Processing %s ...\n", filePath)

		if _, err := os.Stat(filePath); err != nil {
			fmt.Printf("⚠️  Skipping missing file: %s\n", filePath)
			continue
		}

		_ = createBackup(filePath)
		lines := readLines(filePath)

		sort.Slice(list, func(i, j int) bool { return list[i].Line > list[j].Line })

		for _, entry := range list {
			if entry.Line <= 0 || entry.Line > len(lines) {
				continue
			}
			oldLine := lines[entry.Line-1]

			// Look for "description": "CHINESE TEXT"
			pattern := regexp.MustCompile(`"description"\s*:\s*"` + regexp.QuoteMeta(entry.Description) + `"`)
			newLine := pattern.ReplaceAllString(oldLine, `"description": "`+strings.ReplaceAll(entry.Translation, `"`, `\"`)+`"`)

			if newLine != oldLine {
				lines[entry.Line-1] = newLine
				fmt.Printf("  ✓ Line %d updated\n", entry.Line)
			}
		}

		output := strings.Join(lines, "\n")
		if err := os.WriteFile(filePath, []byte(output), 0o644); err != nil {
			fmt.Printf("❌ Write error in %s: %v\n", filePath, err)
			continue
		}
		fmt.Printf("✅ Done: %s\n", filePath)
	}

	fmt.Println("\n🎉 All translations applied successfully!")
}

func main() {
	var (
		translate = flag.Bool("translate", false, "Run translation mode")
		apply     = flag.Bool("apply", false, "Run application mode")
		file      = flag.String("file", DEFAULT_FILE, "Input file path")
	)
	flag.Parse()

	if *translate && *apply {
		fmt.Println("❌ Cannot use both -translate and -apply flags at the same time")
		os.Exit(1)
	}

	if !*translate && !*apply {
		fmt.Println("Usage: go run main.go [OPTIONS]")
		fmt.Println("\nOptions:")
		fmt.Println("  -translate    Extract and translate Chinese texts from JSON file")
		fmt.Println("  -apply        Apply existing translations to schema files")
		fmt.Println("  -file string  Input file path (default: ./chines.json)")
		fmt.Println("\nExamples:")
		fmt.Println("  go run main.go -translate -file ./chines.json")
		fmt.Println("  go run main.go -apply -file ./chines.json")
		os.Exit(1)
	}

	if *translate {
		runTranslation(*file)
	} else if *apply {
		runApplication(*file)
	}
}
