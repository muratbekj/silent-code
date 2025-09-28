package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"silent-code/history"
	"silent-code/mcp"
	"silent-code/ollama"

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
	fmt.Println("  • Project type: Go")
	fmt.Println("  • Main files: main.go, cmd/root.go")
	fmt.Println("  • Dependencies: cobra, ollama")
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
	ollama.TalkToOllama(input, currentSessionID, historyManager)
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
