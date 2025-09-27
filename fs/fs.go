package fs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func WriteFile(path string, data string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, []byte(data), 0644)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DisplayFile(path string) error {
	content, err := ReadFile(path)
	if err != nil {
		return err
	}

	fmt.Printf("\n📄 Contents of %s:\n", path)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%s\n", content)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

func GetEditPrompt(filePath, content, editRequest string) string {
	return fmt.Sprintf(`TASK: Edit the Go file "%s" by making the requested change.

CURRENT FILE CONTENT:
%s

CHANGE REQUESTED: %s

REQUIREMENTS:
- You must return ONLY a unified diff
- Do NOT write any explanations
- Do NOT write any other language
- Return ONLY the diff format shown below

EXAMPLE FORMAT (replace with actual changes):
--- %s
+++ %s
@@ -7,1 +7,2 @@
 func main() {
 	cmd.RootCmd()
+	// HEllo world
 }

RESPOND WITH ONLY THE DIFF - NO OTHER TEXT:`, filePath, content, editRequest, filePath, filePath)
}

func GetGeneratePrompt(filePath, requirements string) string {
	return fmt.Sprintf(`Please generate a new Go file with the following requirements:

File path: %s
Requirements: %s

Please provide the complete file content with proper Go package declaration, imports, and implementation. Format it as a complete, runnable Go file.`, filePath, requirements)
}

func CreateFileWithContent(filePath, content string) error {
	if FileExists(filePath) {
		return fmt.Errorf("file %s already exists", filePath)
	}
	return WriteFile(filePath, content)
}

func BackupFile(filePath string) error {
	if !FileExists(filePath) {
		return fmt.Errorf("file %s does not exist", filePath)
	}

	backupPath := filePath + ".backup"
	content, err := ReadFile(filePath)
	if err != nil {
		return err
	}

	return WriteFile(backupPath, content)
}

func RestoreBackup(filePath string) error {
	backupPath := filePath + ".backup"
	if !FileExists(backupPath) {
		return fmt.Errorf("backup file %s does not exist", backupPath)
	}

	content, err := ReadFile(backupPath)
	if err != nil {
		return err
	}

	return WriteFile(filePath, content)
}

func PromptUser(prompt string) (string, error) {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("failed to read input")
	}
	return strings.TrimSpace(scanner.Text()), nil
}

func ConfirmAction(prompt string) (bool, error) {
	response, err := PromptUser(prompt)
	if err != nil {
		return false, err
	}
	response = strings.ToLower(response)
	return response == "y" || response == "yes", nil
}

// Diff parsing structures
type Diff struct {
	FilePath string
	Hunks    []Hunk
}

type Hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []Line
}

type Line struct {
	Type    LineType
	Content string
	Number  int
}

type LineType int

const (
	Context LineType = iota
	Addition
	Deletion
)

// ParseDiff parses a unified diff string into a Diff struct
func ParseDiff(diffContent string) (*Diff, error) {
	lines := strings.Split(diffContent, "\n")
	diff := &Diff{}

	var currentHunk *Hunk
	var lineNumber int

	for i, line := range lines {
		lineNumber = i + 1

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse file path from --- or +++ lines
		if strings.HasPrefix(line, "--- ") {
			diff.FilePath = strings.TrimSpace(line[4:])
			continue
		}
		if strings.HasPrefix(line, "+++ ") {
			// Use the +++ file path if available, otherwise keep the --- path
			if strings.TrimSpace(line[4:]) != "" {
				diff.FilePath = strings.TrimSpace(line[4:])
			}
			continue
		}

		// Parse hunk header
		if strings.HasPrefix(line, "@@") {
			if currentHunk != nil {
				diff.Hunks = append(diff.Hunks, *currentHunk)
			}

			hunk, err := parseHunkHeader(line)
			if err != nil {
				return nil, fmt.Errorf("error parsing hunk header at line %d: %w", lineNumber, err)
			}
			currentHunk = hunk
			continue
		}

		// Parse diff lines
		if currentHunk == nil {
			continue
		}

		var lineType LineType
		var content string

		if strings.HasPrefix(line, "+") {
			lineType = Addition
			content = line[1:]
		} else if strings.HasPrefix(line, "-") {
			lineType = Deletion
			content = line[1:]
		} else {
			lineType = Context
			content = line[1:]
		}

		currentHunk.Lines = append(currentHunk.Lines, Line{
			Type:    lineType,
			Content: content,
			Number:  lineNumber,
		})
	}

	// Add the last hunk
	if currentHunk != nil {
		diff.Hunks = append(diff.Hunks, *currentHunk)
	}

	return diff, nil
}

// parseHunkHeader parses a hunk header like "@@ -5,6 +5,10 @@"
func parseHunkHeader(header string) (*Hunk, error) {
	// Remove @@ markers
	header = strings.TrimSpace(header)
	header = strings.TrimPrefix(header, "@@")
	header = strings.TrimSuffix(header, "@@")
	header = strings.TrimSpace(header)

	// Handle malformed headers with placeholder text
	if strings.Contains(header, "line") || strings.Contains(header, "count") {
		// Try to extract actual line numbers from the content
		return &Hunk{
			OldStart: 1,
			OldCount: 1,
			NewStart: 1,
			NewCount: 1,
		}, nil
	}

	// Parse the ranges
	parts := strings.Split(header, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid hunk header format: %s", header)
	}

	oldRange, err := parseRange(parts[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing old range: %w", err)
	}

	newRange, err := parseRange(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error parsing new range: %w", err)
	}

	return &Hunk{
		OldStart: oldRange.Start,
		OldCount: oldRange.Count,
		NewStart: newRange.Start,
		NewCount: newRange.Count,
	}, nil
}

type Range struct {
	Start int
	Count int
}

func parseRange(rangeStr string) (Range, error) {
	// Remove + or - prefix
	if strings.HasPrefix(rangeStr, "+") || strings.HasPrefix(rangeStr, "-") {
		rangeStr = rangeStr[1:]
	}

	parts := strings.Split(rangeStr, ",")
	if len(parts) == 1 {
		// Single line number
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return Range{}, err
		}
		return Range{Start: start, Count: 1}, nil
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return Range{}, err
	}

	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return Range{}, err
	}

	return Range{Start: start, Count: count}, nil
}

// ApplyDiff applies a parsed diff to a file
func ApplyDiff(filePath string, diff *Diff) error {
	// Read current file content
	content, err := ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(content, "\n")

	// Process hunks in reverse order to maintain line numbers
	for i := len(diff.Hunks) - 1; i >= 0; i-- {
		hunk := diff.Hunks[i]
		lines, err = applyHunk(lines, hunk)
		if err != nil {
			return fmt.Errorf("failed to apply hunk: %w", err)
		}
	}

	// Write the modified content back
	newContent := strings.Join(lines, "\n")
	return WriteFile(filePath, newContent)
}

func applyHunk(lines []string, hunk Hunk) ([]string, error) {
	// Convert to 0-based indexing
	oldStart := hunk.OldStart - 1

	// Validate hunk range
	if oldStart < 0 || oldStart >= len(lines) {
		return nil, fmt.Errorf("hunk start line %d is out of range (file has %d lines)", hunk.OldStart, len(lines))
	}

	// Find the end of the old section
	oldEnd := oldStart + hunk.OldCount
	if oldEnd > len(lines) {
		oldEnd = len(lines)
	}

	// Build new lines
	var newLines []string

	// Add lines before the hunk
	newLines = append(newLines, lines[:oldStart]...)

	// Process hunk lines
	for _, line := range hunk.Lines {
		switch line.Type {
		case Context:
			// Keep the line as-is
			if oldStart < len(lines) {
				newLines = append(newLines, lines[oldStart])
				oldStart++
			}
		case Addition:
			// Add the new line
			newLines = append(newLines, line.Content)
		case Deletion:
			// Skip the old line
			oldStart++
		}
	}

	// Add remaining lines after the hunk
	if oldEnd < len(lines) {
		newLines = append(newLines, lines[oldEnd:]...)
	}

	return newLines, nil
}

// ShowDiffPreview displays a formatted preview of the diff
func ShowDiffPreview(filePath, diffContent string) error {
	fmt.Printf("\n📋 Changes to be applied to %s:\n", filePath)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	lines := strings.Split(diffContent, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			continue
		}
		if strings.HasPrefix(line, "@@") {
			fmt.Printf("📍 %s\n", line)
			continue
		}
		if strings.HasPrefix(line, "+") {
			fmt.Printf("➕ %s\n", line[1:])
		} else if strings.HasPrefix(line, "-") {
			fmt.Printf("➖ %s\n", line[1:])
		} else {
			fmt.Printf("   %s\n", line)
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// ApplyDiffToFile is the complete workflow for applying diffs
func ApplyDiffToFile(filePath, diffContent string) error {
	// Check if the response contains unwanted content
	if containsUnwantedContent(diffContent) {
		fmt.Printf("⚠️  Warning: AI returned unexpected content, attempting to extract changes manually...\n")
		return applyChangesManually(filePath, diffContent)
	}

	// Parse the diff
	diff, err := ParseDiff(diffContent)
	if err != nil {
		// If parsing fails, try to extract changes manually
		fmt.Printf("⚠️  Warning: Could not parse diff format, attempting to extract changes manually...\n")
		return applyChangesManually(filePath, diffContent)
	}

	// Show preview
	if err := ShowDiffPreview(filePath, diffContent); err != nil {
		return fmt.Errorf("failed to show preview: %w", err)
	}

	// Get user confirmation
	confirm, err := ConfirmAction("\n❓ Do you want to apply these changes? (y/N): ")
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirm {
		fmt.Println("❌ Changes not applied")
		return nil
	}

	// Create backup
	if err := BackupFile(filePath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Apply the diff
	if err := ApplyDiff(filePath, diff); err != nil {
		// Try to restore backup on failure
		if restoreErr := RestoreBackup(filePath); restoreErr != nil {
			return fmt.Errorf("failed to apply diff and restore backup: %w, restore error: %v", err, restoreErr)
		}
		return fmt.Errorf("failed to apply diff: %w", err)
	}

	fmt.Printf("✅ Changes applied successfully to %s\n", filePath)
	return nil
}

// applyChangesManually tries to extract and apply changes from malformed diff content
func applyChangesManually(filePath, diffContent string) error {
	// Read current file content
	content, err := ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(content, "\n")

	// Try to extract a complete file from the AI response
	extractedContent, err := extractCompleteFileFromResponse(diffContent)
	if err == nil && extractedContent != "" {
		// Show preview of the complete file
		fmt.Printf("\n📋 Complete file content from AI:\n")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		extractedLines := strings.Split(extractedContent, "\n")
		for i, line := range extractedLines {
			fmt.Printf("%3d│ %s\n", i+1, line)
		}
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		// Get user confirmation
		confirm, err := ConfirmAction("\n❓ Do you want to replace the entire file with this content? (y/N): ")
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !confirm {
			fmt.Println("❌ Changes not applied")
			return nil
		}

		// Create backup
		if err := BackupFile(filePath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		// Write the new content
		if err := WriteFile(filePath, extractedContent); err != nil {
			// Try to restore backup on failure
			if restoreErr := RestoreBackup(filePath); restoreErr != nil {
				return fmt.Errorf("failed to apply changes and restore backup: %w, restore error: %v", err, restoreErr)
			}
			return fmt.Errorf("failed to apply changes: %w", err)
		}

		fmt.Printf("✅ File updated successfully: %s\n", filePath)
		return nil
	}

	// Fallback to line-by-line changes
	diffLines := strings.Split(diffContent, "\n")
	var changes []struct {
		oldLine string
		newLine string
	}

	for _, line := range diffLines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			// This is a deletion line
			oldLine := line[1:]
			// Look for the corresponding + line
			for _, nextLine := range diffLines {
				nextLine = strings.TrimSpace(nextLine)
				if strings.HasPrefix(nextLine, "+") && !strings.HasPrefix(nextLine, "+++") {
					newLine := nextLine[1:]
					changes = append(changes, struct {
						oldLine string
						newLine string
					}{oldLine, newLine})
					break
				}
			}
		}
	}

	// Apply changes
	for _, change := range changes {
		for i, line := range lines {
			if strings.TrimSpace(line) == strings.TrimSpace(change.oldLine) {
				lines[i] = change.newLine
				break
			}
		}
	}

	// Show preview of changes
	fmt.Printf("\n📋 Manual changes to be applied to %s:\n", filePath)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for _, change := range changes {
		fmt.Printf("➖ %s\n", change.oldLine)
		fmt.Printf("➕ %s\n", change.newLine)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Get user confirmation
	confirm, err := ConfirmAction("\n❓ Do you want to apply these changes? (y/N): ")
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirm {
		fmt.Println("❌ Changes not applied")
		return nil
	}

	// Create backup
	if err := BackupFile(filePath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Write the modified content
	newContent := strings.Join(lines, "\n")
	if err := WriteFile(filePath, newContent); err != nil {
		// Try to restore backup on failure
		if restoreErr := RestoreBackup(filePath); restoreErr != nil {
			return fmt.Errorf("failed to apply changes and restore backup: %w, restore error: %v", err, restoreErr)
		}
		return fmt.Errorf("failed to apply changes: %w", err)
	}

	fmt.Printf("✅ Changes applied successfully to %s\n", filePath)
	return nil
}

// extractCompleteFileFromResponse tries to extract a complete Go file from AI response
func extractCompleteFileFromResponse(content string) (string, error) {
	// Look for code blocks first
	codeBlocks, err := ExtractCodeBlocks(content)
	if err == nil && len(codeBlocks) > 0 {
		// Use the first code block
		cleanCode := CleanGoCode(codeBlocks[0])
		if strings.HasPrefix(cleanCode, "package ") {
			return cleanCode, nil
		}
	}

	// Try to extract from the content directly
	lines := strings.Split(content, "\n")
	var extractedLines []string
	inCodeBlock := false
	foundPackage := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for package declaration to start extraction
		if strings.HasPrefix(line, "package ") {
			foundPackage = true
			inCodeBlock = true
		}

		// Stop if we hit explanatory text
		if inCodeBlock && (strings.HasPrefix(line, "In this example") ||
			strings.HasPrefix(line, "I hope this helps") ||
			strings.HasPrefix(line, "This example") ||
			strings.HasPrefix(line, "The code above") ||
			strings.HasPrefix(line, "This code") ||
			strings.HasPrefix(line, "The function") ||
			strings.HasPrefix(line, "We've defined") ||
			strings.HasPrefix(line, "Finally") ||
			strings.HasPrefix(line, "In this case")) {
			break
		}

		if inCodeBlock {
			extractedLines = append(extractedLines, line)
		}
	}

	if foundPackage && len(extractedLines) > 0 {
		return strings.Join(extractedLines, "\n"), nil
	}

	return "", fmt.Errorf("no complete Go file found in response")
}

// containsUnwantedContent checks if the AI response contains unwanted content
func containsUnwantedContent(content string) bool {
	content = strings.ToLower(content)

	// Check for Python code
	if strings.Contains(content, "def ") || strings.Contains(content, "import ") ||
		strings.Contains(content, "python") || strings.Contains(content, "with open(") {
		return true
	}

	// Check for explanatory text instead of diffs
	if strings.Contains(content, "here's how") || strings.Contains(content, "you can use") ||
		strings.Contains(content, "here are a few") || strings.Contains(content, "you could generate") {
		return true
	}

	// Check if it's missing diff markers
	if !strings.Contains(content, "---") && !strings.Contains(content, "+++") &&
		!strings.Contains(content, "@@") {
		return true
	}

	return false
}

// Content parsing functions for AI-generated content

// ExtractCodeBlocks extracts code blocks from markdown or mixed content
func ExtractCodeBlocks(content string) ([]string, error) {
	var codeBlocks []string

	// Look for markdown code blocks
	codeBlockRegex := regexp.MustCompile("```(?:go|golang)?\\s*\\n([\\s\\S]*?)```")
	matches := codeBlockRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			codeBlocks = append(codeBlocks, strings.TrimSpace(match[1]))
		}
	}

	// If no code blocks found, try to extract Go code directly
	if len(codeBlocks) == 0 {
		// Look for package declaration as a sign of Go code
		if strings.Contains(content, "package ") {
			codeBlocks = append(codeBlocks, content)
		}
	}

	return codeBlocks, nil
}

// CleanGoCode removes markdown formatting and cleans up Go code
func CleanGoCode(code string) string {
	// Remove markdown code block markers if present
	code = regexp.MustCompile("```(?:go|golang)?\\s*\\n?").ReplaceAllString(code, "")
	code = regexp.MustCompile("```\\s*$").ReplaceAllString(code, "")

	// Remove common AI response prefixes
	prefixes := []string{
		"Here's the Go code:",
		"Here is the Go code:",
		"Here's the complete Go file:",
		"Here is the complete Go file:",
		"The Go code is:",
		"Here's your Go file:",
		"Here is your Go file:",
		"Sure! Here is the complete, runnable Go file you requested:",
		"Here is the complete, runnable Go file you requested:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(code, prefix) {
			code = strings.TrimSpace(code[len(prefix):])
			break
		}
	}

	// Extract only the Go code part (from package to the end of the last function)
	lines := strings.Split(code, "\n")
	var cleanLines []string
	foundPackage := false
	inGoCode := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Start collecting when we find package declaration
		if strings.HasPrefix(line, "package ") {
			foundPackage = true
			inGoCode = true
		}

		// Stop collecting when we hit explanatory text (usually starts with "In this example" or similar)
		if inGoCode && (strings.HasPrefix(line, "In this example") ||
			strings.HasPrefix(line, "I hope this helps") ||
			strings.HasPrefix(line, "This example") ||
			strings.HasPrefix(line, "The code above") ||
			strings.HasPrefix(line, "This code") ||
			strings.HasPrefix(line, "The function") ||
			strings.HasPrefix(line, "We've defined") ||
			strings.HasPrefix(line, "Finally") ||
			strings.HasPrefix(line, "In this case")) {
			break
		}

		// Collect lines that are part of the Go code
		if inGoCode {
			cleanLines = append(cleanLines, line)
		}
	}

	// If we didn't find package declaration, try to extract from the beginning
	if !foundPackage {
		cleanLines = nil
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Look for Go keywords to identify the start
			if strings.HasPrefix(line, "package ") ||
				strings.HasPrefix(line, "import ") ||
				strings.HasPrefix(line, "func ") ||
				strings.HasPrefix(line, "type ") ||
				strings.HasPrefix(line, "var ") ||
				strings.HasPrefix(line, "const ") {
				cleanLines = append(cleanLines, line)
			} else if len(cleanLines) > 0 {
				// Continue collecting if we're already in Go code
				cleanLines = append(cleanLines, line)
			}
		}
	}

	code = strings.Join(cleanLines, "\n")
	code = strings.TrimSpace(code)

	return code
}

// ParseGeneratedContent extracts and cleans Go code from AI response
func ParseGeneratedContent(content string) (string, error) {
	// First try to extract code blocks
	codeBlocks, err := ExtractCodeBlocks(content)
	if err != nil {
		return "", fmt.Errorf("failed to extract code blocks: %w", err)
	}

	if len(codeBlocks) == 0 {
		return "", fmt.Errorf("no code blocks found in content")
	}

	// Use the first (and usually only) code block
	cleanCode := CleanGoCode(codeBlocks[0])

	// Validate that it's valid Go code
	if !strings.HasPrefix(cleanCode, "package ") {
		return "", fmt.Errorf("extracted content doesn't appear to be valid Go code (missing package declaration). Got: %q", cleanCode[:min(50, len(cleanCode))])
	}

	return cleanCode, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ShowFilePreview displays a formatted preview of the new file content
func ShowFilePreview(filePath, content string) error {
	fmt.Printf("\n📄 New file: %s\n", filePath)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		fmt.Printf("%3d│ %s\n", i+1, line)
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

// CreateFileFromContent is the complete workflow for creating files from AI-generated content
func CreateFileFromContent(filePath, content string) error {
	// Parse and clean the content
	cleanContent, err := ParseGeneratedContent(content)
	if err != nil {
		return fmt.Errorf("failed to parse generated content: %w", err)
	}

	// Show preview
	if err := ShowFilePreview(filePath, cleanContent); err != nil {
		return fmt.Errorf("failed to show preview: %w", err)
	}

	// Get user confirmation
	confirm, err := ConfirmAction("\n❓ Do you want to create this file? (y/N): ")
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if !confirm {
		fmt.Println("❌ File not created")
		return nil
	}

	// Create the file
	if err := CreateFileWithContent(filePath, cleanContent); err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	fmt.Printf("✅ File created successfully: %s\n", filePath)
	return nil
}
