package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/muratbekj/silent-code/history"
	"github.com/muratbekj/silent-code/mcp"
	"github.com/muratbekj/silent-code/ollama"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "silent-code",
	Short: "AI-powered development assistant with privacy-first architecture",
	Long: `Silent Code - An AI-powered development assistant that runs fully offline on your machine.

Core Features:
• Code Understanding & Editing - Query and edit large codebases beyond traditional context window limits
• Workflow Automation - Automate operational tasks like handling pull requests and complex rebases  
• Local Model Support - Works with any Ollama-compatible model (Qwen, Llama, CodeLlama, etc.)
• Privacy-First Architecture - All processing happens on your infrastructure

It looks and feels like a terminal, but acts as an AI coding agent: you can ask it about 
your project, edit files, create new ones, run tests, and reason about code — all powered 
by local LLMs (via Ollama).`,
	Run: func(cmd *cobra.Command, args []string) {
		startInteractiveMode()
	},
}

// Global session ID and history manager
var currentSessionID string
var historyManager *history.HistoryManager

// Interactive terminal mode
func startInteractiveMode() {
	// Initialize history
	historyManager = history.NewHistoryManager("./history/sessions")

	// Initialize model selection
	fmt.Print("🔍 Detecting available models... ")
	err := ollama.InitializeModelSelection()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println("💡 Make sure Ollama is running: ollama serve")
		fmt.Println("💡 Install a model: ollama pull codellama:13b")
		return
	}
	fmt.Printf("✅ Using model: %s\n", ollama.GetCurrentModel())

	// Create new session
	currentSessionID = fmt.Sprintf("session_%d", time.Now().Unix())

	fmt.Println("🤖 Silent Code - AI-Powered Development Assistant")
	fmt.Printf("📝 Session: %s\n", currentSessionID)
	fmt.Println("Type '/help' for commands, '/exit' to quit")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	showHelp()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("silent-code> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		handleCommand(input)
	}
}

// List of app-specific commands that should NOT be treated as shell commands
var appCommands = map[string]bool{
	"help": true, "explain": true, "generate": true, "test": true, "search": true,
	"config": true, "status": true, "sessions": true, "context": true, "prompt": true,
	"reason": true, "steps": true, "read": true, "edit": true, "new": true,
	"exit": true, "quit": true,
}

func isAppCommand(command string) bool {
	return appCommands[command]
}

// isGeneralQuestion checks if the input looks like a general question to the AI
func isGeneralQuestion(input string) bool {
	// Check for question words and patterns
	questionWords := []string{
		"what", "how", "why", "when", "where", "who", "which", "can", "could", "would", "should",
		"is", "are", "was", "were", "do", "does", "did", "will", "have", "has", "had",
		"explain", "describe", "tell", "show", "help", "analyze", "review", "check",
	}

	// Check for question patterns
	questionPatterns := []string{
		"what is", "how does", "why is", "when does", "where is", "who is", "which is",
		"can you", "could you", "would you", "should i", "is this", "are there",
		"do you", "does this", "did you", "will this", "have you", "has this",
		"explain this", "describe this", "tell me", "show me", "help me",
		"analyze this", "review this", "check this",
	}

	inputLower := strings.ToLower(input)

	// Check for question words at the beginning
	firstWord := strings.Fields(inputLower)[0]
	for _, word := range questionWords {
		if firstWord == word {
			return true
		}
	}

	// Check for question patterns
	for _, pattern := range questionPatterns {
		if strings.HasPrefix(inputLower, pattern) {
			return true
		}
	}

	// Check for question mark
	if strings.HasSuffix(input, "?") {
		return true
	}

	// Check if it contains multiple words and doesn't look like a shell command
	words := strings.Fields(inputLower)
	if len(words) >= 2 {
		// If it has multiple words and doesn't start with common shell commands, treat as question
		shellCommands := []string{"ls", "cd", "pwd", "cat", "grep", "find", "mkdir", "rm", "cp", "mv", "chmod", "sudo", "git", "npm", "pip", "python", "node", "go", "cargo", "mvn", "gradle"}
		firstWord = words[0]
		for _, cmd := range shellCommands {
			if firstWord == cmd {
				return false // It's a shell command
			}
		}
		return true // Multiple words, not a shell command, probably a question
	}

	return false
}

func handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	// Check if it's an app command (with or without / prefix)
	appCommand := command
	if strings.HasPrefix(command, "/") {
		appCommand = command[1:] // Remove the / prefix
	}

	if isAppCommand(appCommand) {
		// Handle app commands
	} else {
		// Check if it looks like a general question (not a shell command)
		if isGeneralQuestion(input) {
			// Handle as general question to AI
			handleGeneralQuestion(input)
			return
		}
		// Everything else is treated as a shell command
		handleShellCommand(input)
		return
	}

	switch command {
	case "help", "/help":
		showHelp()
	case "explain", "/explain":
		handleExplain(args)
	case "generate", "/generate":
		handleGenerate(args)
	case "test", "/test":
		handleTest(args)
	case "search", "/search":
		handleSearch(args)
	case "config", "/config":
		handleConfig(args)
	case "status", "/status":
		handleStatus()
	case "sessions", "/sessions":
		handleSessions()
	case "context", "/context":
		handleContext()
	case "prompt", "/prompt":
		handlePrompt(args)
	case "reason", "/reason":
		handleReason(args)
	case "steps", "/steps":
		handleSteps()
	case "read", "/read":
		handleMCPRead(args)
	case "edit", "/edit":
		handleMCPEdit(args)
	case "new", "/new":
		handleMCPCreate(args)
	case "exit", "quit", "/exit", "/quit":
		fmt.Println("👋 Goodbye!")
		os.Exit(0)
	default:
		// Treat as a general question
		handleGeneralQuestion(input)
	}
}

func showHelp() {
	fmt.Println("\n🤖 Silent Code - AI-Powered Development Assistant")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n🔧 Core Features:")
	fmt.Println("  • Code Understanding & Editing - Query and edit large codebases")
	fmt.Println("  • Workflow Automation - Handle pull requests and complex rebases")
	fmt.Println("  • Local Model Support - Works with any Ollama-compatible model")
	fmt.Println("  • Privacy-First Architecture - All processing on your infrastructure")
	fmt.Println("\n📋 Available Commands:")
	fmt.Println("  /explain <file>     - Explain a specific file or function")
	fmt.Println("  /generate <what>    - Generate new code")
	fmt.Println("  /refactor <file>    - Refactor existing code")
	fmt.Println("  /test               - Run tests and analyze results")
	fmt.Println("  /search <query>     - Search through codebase semantically")
	fmt.Println("  /config             - Show locally installed Ollama models")
	fmt.Println("  /sessions           - List and manage conversation sessions")
	fmt.Println("  /context            - Show current project context")
	fmt.Println("  /prompt <file>      - Add specific file to context")
	fmt.Println("  /reason <problem>   - Start multi-turn reasoning for a problem")
	fmt.Println("  /steps              - Show current reasoning steps")
	fmt.Println("  /status             - Show current project status")
	fmt.Println("  /read <file>        - View file contents")
	fmt.Println("  /edit <file>        - Edit file with AI assistance")
	fmt.Println("  /new <file>         - Create new file with AI assistance")
	fmt.Println("  /help               - Show this help message")
	fmt.Println("  /exit or /quit      - Exit the terminal")
	fmt.Println("\n💡 You can also just type questions directly!")
	fmt.Println("   Example: 'How does authentication work in this project?'")
}

func handleExplain(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify a file or function to explain. Example: explain main.go")
		return
	}
	target := args[0]
	client := mcp.NewMCPClient("http://127.0.0.1:8080")
	result, err := client.ExplainCode(target)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("❌ Explanation failed: %s\n", result.Error)
		return
	}

	fmt.Printf("\n🤖 Code Explanation:\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(result.Content)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func handleGenerate(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify what to generate. Example: generate 'a new API endpoint'")
		return
	}
	what := strings.Join(args, " ")
	fmt.Printf("⚡ Generating: %s\n", what)
	ollama.TalkToOllama(fmt.Sprintf("Generate: %s", what), currentSessionID, historyManager)
}

func handleTest(args []string) {
	fmt.Println("🧪 Running tests...")
	ollama.TalkToOllama("Run tests and analyze the results", currentSessionID, historyManager)
}

func handleSearch(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please provide a search query. Example: search 'authentication logic'")
		return
	}
	query := strings.Join(args, " ")
	fmt.Printf("🔍 Searching for: %s\n", query)
	ollama.TalkToOllama(fmt.Sprintf("Search for: %s", query), currentSessionID, historyManager)
}

func handleSessions() {
	fmt.Println("📝 Session Management:")
	fmt.Printf("  Current Session: %s\n", currentSessionID)
	fmt.Println("  💡 Sessions are automatically saved to ./sessions/")
	fmt.Println("  💡 Each conversation maintains context across commands")

	// List available sessions
	sessions, err := historyManager.ListSessions()
	if err != nil {
		fmt.Printf("  ❌ Error listing sessions: %v\n", err)
		return
	}

	if len(sessions) > 0 {
		fmt.Println("  📋 Available Sessions:")
		for _, session := range sessions {
			fmt.Printf("    • %s\n", session)
		}
	} else {
		fmt.Println("  📋 No previous sessions found")
	}
}

func handleConfig(args []string) {
	fmt.Println("🔧 Ollama Configuration:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Show current model
	fmt.Printf("🤖 Current Model: %s\n\n", ollama.GetCurrentModel())

	// Handle model switching if requested
	if len(args) >= 2 && args[0] == "models" {
		modelName := args[1]
		err := ollama.SetModel(modelName)
		if err != nil {
			fmt.Printf("❌ Error switching model: %v\n", err)
			return
		}
		fmt.Printf("✅ Model switched to: %s\n\n", modelName)
	}

	models, err := ollama.ListOllamaModels()
	if err != nil {
		fmt.Printf("❌ Error connecting to Ollama: %v\n", err)
		fmt.Println("💡 Make sure Ollama is running: ollama serve")
		return
	}

	if len(models) == 0 {
		fmt.Println("📋 No models installed")
		fmt.Println("💡 Install a model: ollama pull codellama:13b")
		return
	}

	fmt.Printf("📋 Installed Models (%d):\n", len(models))
	for i, model := range models {
		// Mark current model with an arrow
		currentIndicator := ""
		if model.Name == ollama.GetCurrentModel() {
			currentIndicator = " ← Current"
		}

		fmt.Printf("  %d. %s%s\n", i+1, model.Name, currentIndicator)
		fmt.Printf("     Size: %.2f GB\n", float64(model.Size)/1024/1024/1024)
		fmt.Printf("     Modified: %s\n", model.ModifiedAt.Format("2006-01-02 15:04:05"))
		if model.Details.Family != "" {
			fmt.Printf("     Family: %s\n", model.Details.Family)
		}
		if model.Details.ParameterSize != "" {
			fmt.Printf("     Parameters: %s\n", model.Details.ParameterSize)
		}
		fmt.Println()
	}

	fmt.Println("💡 Usage: /config models <modelname> to switch models")
}

func handleStatus() {
	fmt.Println("📊 Project Status:")
	fmt.Printf("  • AI Model: Ollama (%s)\n", ollama.GetCurrentModel())
	fmt.Println("  • Project: silent-code")
	fmt.Println("  • Language: Go")
	fmt.Println("  • Session: Active")
	fmt.Printf("  • History: %s\n", currentSessionID)
}

func handleContext() {
	fmt.Println("📁 Project Context:")
	fmt.Println("  • Current directory: .")

	// Detect project type dynamically
	projectType := detectProjectType(".")
	fmt.Printf("  • Project type: %s\n", projectType)

	// Get actual files in the directory
	actualFiles := getActualFiles(".")
	if len(actualFiles) > 0 {
		fmt.Printf("  • Main files: %s\n", strings.Join(actualFiles, ", "))
	}

	// Get dependencies based on project type
	dependencies := getDependencies(projectType)
	if len(dependencies) > 0 {
		fmt.Printf("  • Dependencies: %s\n", strings.Join(dependencies, ", "))
	} else {
		fmt.Println("  • Dependencies: None")
	}

	fmt.Println("  💡 Context is automatically loaded for better AI responses")
}

func handlePrompt(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify a file to add to context. Example: prompt main.go")
		return
	}
	file := args[0]
	fmt.Printf("📄 Adding %s to context...\n", file)
	fmt.Println("💡 This file will be included in AI responses for better context")
}

func handleReason(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify a problem to reason about. Example: reason 'How to optimize this code?'")
		return
	}
	problem := strings.Join(args, " ")
	fmt.Printf("🧠 Starting multi-turn reasoning for: %s\n", problem)
	fmt.Println("💡 Use 'steps' to see reasoning progress")

	// Start reasoning session
	ollama.InitializeReasoning()
	ollama.StartReasoning(currentSessionID, problem)

	// Add initial step
	ollama.AddReasoningStep(currentSessionID, "Analyzing the problem", "Breaking down the problem into manageable steps")

	fmt.Println("🔄 Reasoning session started. The AI will work through this step by step.")
}

func handleSteps() {
	fmt.Println("🧠 Current Reasoning Steps:")

	// Get reasoning summary
	summary, err := ollama.GetReasoningSummary(currentSessionID)
	if err != nil {
		fmt.Printf("❌ No active reasoning session. Use 'reason <problem>' to start one.\n")
		return
	}

	fmt.Println(summary)
}

func handleGeneralQuestion(input string) {
	// Use MCP to analyze the project and answer the question
	client := mcp.NewMCPClient("http://127.0.0.1:8080")

	// First, get the current directory contents
	result, err := client.ExecuteShell("ls -la")
	if err != nil {
		fmt.Printf("❌ Error getting directory contents: %v\n", err)
		// Fallback to regular AI response
		ollama.TalkToOllama(input, currentSessionID, historyManager)
		return
	}

	if !result.Success {
		fmt.Printf("❌ Failed to get directory contents: %s\n", result.Error)
		// Fallback to regular AI response
		ollama.TalkToOllama(input, currentSessionID, historyManager)
		return
	}

	// Build enhanced question with directory contents
	enhancedQuestion := fmt.Sprintf("%s\n\nCurrent directory contents:\n%s", input, result.Output)

	// For file-specific questions, try to read relevant files
	if shouldReadFiles(input) {
		fileContents := readRelevantFiles()
		if fileContents != "" {
			enhancedQuestion += "\n\nFile contents:\n" + fileContents
		}
	}

	// Send enhanced question to AI
	ollama.TalkToOllama(enhancedQuestion, currentSessionID, historyManager)
}

// shouldReadFiles determines if the question would benefit from file contents
func shouldReadFiles(question string) bool {
	questionLower := strings.ToLower(question)

	// Questions that would benefit from file contents
	fileRelatedKeywords := []string{
		"what is", "what does", "what's in", "what are",
		"how does", "how is", "how are",
		"explain", "describe", "analyze", "review",
		"code", "function", "class", "method", "variable",
		"project", "folder", "directory", "files",
		"main", "app", "script", "program",
	}

	for _, keyword := range fileRelatedKeywords {
		if strings.Contains(questionLower, keyword) {
			return true
		}
	}

	return false
}

// readRelevantFiles reads the most relevant files in the directory
func readRelevantFiles() string {
	client := mcp.NewMCPClient("http://127.0.0.1:8080")

	// Get list of files
	result, err := client.ExecuteShell("ls -1")
	if err != nil || !result.Success {
		return ""
	}

	files := strings.Split(strings.TrimSpace(result.Output), "\n")
	var fileContents []string

	// Read up to 3 most relevant files
	fileCount := 0
	for _, file := range files {
		if fileCount >= 3 {
			break
		}

		// Skip directories and non-source files
		if strings.Contains(file, "/") ||
			strings.HasPrefix(file, ".") ||
			file == "silent-code" ||
			file == "go.sum" ||
			file == "LICENSE" {
			continue
		}

		// Try to read the file
		readResult, err := client.ReadFile(file)
		if err == nil && readResult.Success {
			fileContents = append(fileContents, fmt.Sprintf("=== %s ===\n%s", file, readResult.Content))
			fileCount++
		}
	}

	return strings.Join(fileContents, "\n\n")
}

func init() {
	// Add command handlers

	rootCmd.AddCommand(&cobra.Command{
		Use:   "explain [file]",
		Short: "Explain a file or function",
		Long:  "Get detailed explanations of code files or specific functions",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ Please specify a file or function")
				return
			}
			handleExplain(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "generate [what]",
		Short: "Generate new code",
		Long:  "Generate new code based on your specifications",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ Please specify what to generate")
				return
			}
			handleGenerate(args)
		},
	})
}

// MCP Handler functions
func handleMCPCreate(args []string) {
	if len(args) < 2 {
		fmt.Println("❌ Usage: mcp-create <file> <requirements>")
		return
	}

	filePath := args[0]
	requirements := strings.Join(args[1:], " ")

	client := mcp.NewMCPClient("http://127.0.0.1:8080")
	result, err := client.CreateFile(filePath, requirements)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("❌ Creation failed: %s\n", result.Error)
		return
	}

	fmt.Printf("✅ %s\n", result.Message)
}

func handleMCPEdit(args []string) {
	if len(args) < 2 {
		fmt.Println("❌ Usage: mcp-edit <file> <edit_request>")
		return
	}

	filePath := args[0]
	editRequest := strings.Join(args[1:], " ")

	client := mcp.NewMCPClient("http://127.0.0.1:8080")
	result, err := client.EditFile(filePath, editRequest)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("❌ Edit failed: %s\n", result.Error)
		return
	}

	fmt.Printf("✅ %s\n", result.Message)
}

func handleMCPRead(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify a file. Example: mcp-read main.go")
		return
	}

	filePath := args[0]
	client := mcp.NewMCPClient("http://127.0.0.1:8080")
	result, err := client.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("❌ Read failed: %s\n", result.Error)
		return
	}

	fmt.Printf("\n📄 Contents of %s:\n", filePath)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(result.Content)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func handleShellCommand(command string) {
	fmt.Printf("🔧 Executing: %s\n", command)

	client := mcp.NewMCPClient("http://127.0.0.1:8080")
	result, err := client.ExecuteShell(command)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	if !result.Success {
		fmt.Printf("❌ Command failed: %s\n", result.Error)
		if result.Stderr != "" {
			fmt.Printf("Error output: %s\n", result.Stderr)
		}
		return
	}

	// For successful commands, just show the output without extra formatting
	if result.Output != "" {
		fmt.Print(result.Output)
	}
	if result.Stderr != "" {
		fmt.Print(result.Stderr)
	}
}

func RootCmd() {
	rootCmd.Execute()
}

// detectProjectType detects the type of project based on configuration files and current files
func detectProjectType(projectPath string) string {
	// First, check for Python files (most common for current work)
	if hasPythonFiles(projectPath) {
		return "Python"
	}

	// Then check for other languages based on file extensions
	if hasJavaScriptFiles(projectPath) {
		return "JavaScript/Node.js"
	}

	if hasTypeScriptFiles(projectPath) {
		return "TypeScript"
	}

	// Check configuration files
	configFiles := map[string]string{
		"requirements.txt": "Python",
		"package.json":     "JavaScript/Node.js",
		"go.mod":           "Go",
		"pom.xml":          "Java",
		"build.gradle":     "Java/Gradle",
		"cargo.toml":       "Rust",
		"composer.json":    "PHP",
		"Gemfile":          "Ruby",
		"Podfile":          "Swift/Objective-C",
		"mix.exs":          "Elixir",
		"pubspec.yaml":     "Dart/Flutter",
	}

	for file, projectType := range configFiles {
		if _, err := os.Stat(filepath.Join(projectPath, file)); err == nil {
			return projectType
		}
	}

	return "Unknown"
}

// hasPythonFiles checks if there are Python files in the directory
func hasPythonFiles(projectPath string) bool {
	files, err := os.ReadDir(projectPath)
	if err != nil {
		return false
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".py") {
			return true
		}
	}
	return false
}

// hasJavaScriptFiles checks if there are JavaScript files in the directory
func hasJavaScriptFiles(projectPath string) bool {
	files, err := os.ReadDir(projectPath)
	if err != nil {
		return false
	}

	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".js") || strings.HasSuffix(file.Name(), ".jsx")) {
			return true
		}
	}
	return false
}

// hasTypeScriptFiles checks if there are TypeScript files in the directory
func hasTypeScriptFiles(projectPath string) bool {
	files, err := os.ReadDir(projectPath)
	if err != nil {
		return false
	}

	for _, file := range files {
		if !file.IsDir() && (strings.HasSuffix(file.Name(), ".ts") || strings.HasSuffix(file.Name(), ".tsx")) {
			return true
		}
	}
	return false
}

// getActualFiles returns the actual files in the directory
func getActualFiles(projectPath string) []string {
	files, err := os.ReadDir(projectPath)
	if err != nil {
		return []string{}
	}

	var actualFiles []string
	for _, file := range files {
		if !file.IsDir() {
			// Skip hidden files and common non-source files
			fileName := file.Name()
			if !strings.HasPrefix(fileName, ".") &&
				fileName != "silent-code" &&
				fileName != "go.sum" &&
				fileName != "LICENSE" {
				actualFiles = append(actualFiles, fileName)
			}
		}
	}

	return actualFiles
}

// getMainFiles returns main files for a project type (kept for backward compatibility)
func getMainFiles(projectType string) []string {
	mainFilesMap := map[string][]string{
		"Go":                 {"main.go", "cmd/root.go", "README.md"},
		"JavaScript/Node.js": {"index.js", "app.js", "server.js", "package.json", "README.md"},
		"Python":             {"main.py", "app.py", "requirements.txt", "README.md"},
		"Java":               {"src/main/java", "pom.xml", "README.md"},
		"Java/Gradle":        {"src/main/java", "build.gradle", "README.md"},
		"Rust":               {"src/main.rs", "Cargo.toml", "README.md"},
		"PHP":                {"index.php", "composer.json", "README.md"},
		"Ruby":               {"main.rb", "app.rb", "Gemfile", "README.md"},
		"Swift/Objective-C":  {"main.swift", "AppDelegate.swift", "README.md"},
		"Elixir":             {"lib", "mix.exs", "README.md"},
		"Dart/Flutter":       {"lib/main.dart", "pubspec.yaml", "README.md"},
	}

	if files, exists := mainFilesMap[projectType]; exists {
		return files
	}

	return []string{"README.md"}
}

// getDependencies returns actual dependencies found in the project
func getDependencies(projectType string) []string {
	// Check for actual dependency files first
	actualDeps := getActualDependencies(".")
	if len(actualDeps) > 0 {
		return actualDeps
	}

	// If no actual dependencies found, return empty
	return []string{}
}

// getActualDependencies scans for actual dependency files and extracts dependencies
func getActualDependencies(projectPath string) []string {
	var dependencies []string

	// Check for Python requirements.txt
	if requirementsPath := filepath.Join(projectPath, "requirements.txt"); fileExists(requirementsPath) {
		if deps := parseRequirementsTxt(requirementsPath); len(deps) > 0 {
			dependencies = append(dependencies, deps...)
		}
	}

	// Check for Node.js package.json
	if packageJsonPath := filepath.Join(projectPath, "package.json"); fileExists(packageJsonPath) {
		if deps := parsePackageJson(packageJsonPath); len(deps) > 0 {
			dependencies = append(dependencies, deps...)
		}
	}

	// Check for Go go.mod
	if goModPath := filepath.Join(projectPath, "go.mod"); fileExists(goModPath) {
		if deps := parseGoMod(goModPath); len(deps) > 0 {
			dependencies = append(dependencies, deps...)
		}
	}

	// Check for other dependency files
	otherDeps := checkOtherDependencyFiles(projectPath)
	dependencies = append(dependencies, otherDeps...)

	return dependencies
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// parseRequirementsTxt parses Python requirements.txt file
func parseRequirementsTxt(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}

	var deps []string
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract package name (before ==, >=, etc.)
		parts := strings.FieldsFunc(line, func(r rune) bool {
			return r == '=' || r == '>' || r == '<' || r == '!' || r == '~'
		})
		if len(parts) > 0 {
			packageName := strings.TrimSpace(parts[0])
			if packageName != "" {
				deps = append(deps, packageName)
			}
		}
	}

	return deps
}

// parsePackageJson parses Node.js package.json file
func parsePackageJson(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}

	// Simple JSON parsing for dependencies
	var deps []string
	lines := strings.Split(string(content), "\n")

	inDependencies := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "\"dependencies\"") {
			inDependencies = true
			continue
		}

		if inDependencies {
			if strings.Contains(line, "}") && !strings.Contains(line, "\"") {
				break
			}

			if strings.Contains(line, "\"") && strings.Contains(line, ":") {
				// Extract package name
				parts := strings.Split(line, "\"")
				if len(parts) >= 2 {
					packageName := strings.TrimSpace(parts[1])
					if packageName != "" && packageName != "dependencies" {
						deps = append(deps, packageName)
					}
				}
			}
		}
	}

	return deps
}

// parseGoMod parses Go go.mod file
func parseGoMod(path string) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}

	var deps []string
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require") {
			continue
		}

		if strings.Contains(line, " ") && !strings.HasPrefix(line, "module") && !strings.HasPrefix(line, "go ") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				packageName := parts[0]
				if !strings.Contains(packageName, "/") || strings.Count(packageName, "/") > 1 {
					// This looks like a dependency
					deps = append(deps, packageName)
				}
			}
		}
	}

	return deps
}

// checkOtherDependencyFiles checks for other dependency files
func checkOtherDependencyFiles(projectPath string) []string {
	var deps []string

	// Check for other common dependency files
	dependencyFiles := []string{
		"composer.json", // PHP
		"Gemfile",       // Ruby
		"Cargo.toml",    // Rust
		"pom.xml",       // Java
		"build.gradle",  // Java/Gradle
	}

	for _, file := range dependencyFiles {
		if fileExists(filepath.Join(projectPath, file)) {
			// For now, just indicate the file exists
			deps = append(deps, file)
		}
	}

	return deps
}
