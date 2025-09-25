package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"silent-code/history"
	"silent-code/ollama"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "silent-code",
	Short: "A coding agent that lives in your terminal",
	Long: `Private Code - A Warp-like AI terminal that runs fully offline on your machine.
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

	fmt.Println("🤖 Private Code - AI Terminal")
	fmt.Printf("📝 Session: %s\n", currentSessionID)
	fmt.Println("Type 'help' for commands, 'exit' to quit")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("private-code> ")
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

func handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "help":
		showHelp()
	case "ask":
		handleAsk(args)
	case "explain":
		handleExplain(args)
	case "generate":
		handleGenerate(args)
	case "refactor":
		handleRefactor(args)
	case "test":
		handleTest(args)
	case "search":
		handleSearch(args)
	case "status":
		handleStatus()
	case "sessions":
		handleSessions()
	case "context":
		handleContext()
	case "prompt":
		handlePrompt(args)
	case "reason":
		handleReason(args)
	case "steps":
		handleSteps()
	default:
		// Treat as a general question
		handleGeneralQuestion(input)
	}
}

func showHelp() {
	fmt.Println("\n📋 Available Commands:")
	fmt.Println("  ask <question>     - Ask questions about your codebase")
	fmt.Println("  explain <file>      - Explain a specific file or function")
	fmt.Println("  generate <what>     - Generate new code")
	fmt.Println("  refactor <file>     - Refactor existing code")
	fmt.Println("  test                - Run tests and analyze results")
	fmt.Println("  search <query>      - Search through codebase semantically")
	fmt.Println("  sessions            - List and manage conversation sessions")
	fmt.Println("  context             - Show current project context")
	fmt.Println("  prompt <file>       - Add specific file to context")
	fmt.Println("  reason <problem>    - Start multi-turn reasoning for a problem")
	fmt.Println("  steps               - Show current reasoning steps")
	fmt.Println("  status              - Show current project status")
	fmt.Println("  help                - Show this help message")
	fmt.Println("  exit/quit           - Exit the terminal")
	fmt.Println("\n💡 You can also just type questions directly!")
	fmt.Println("   Example: 'How does authentication work in this project?'")
}

func handleAsk(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please provide a question. Example: ask 'How does the authentication work?'")
		return
	}
	question := strings.Join(args, " ")
	fmt.Printf("🤔 Question: %s\n", question)
	ollama.TalkToOllama(question, currentSessionID, historyManager)
}

func handleExplain(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify a file or function to explain. Example: explain main.go")
		return
	}
	target := args[0]
	fmt.Printf("📖 Explaining: %s\n", target)
	ollama.TalkToOllama(fmt.Sprintf("Explain this file/function: %s", target), currentSessionID, historyManager)
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

func handleRefactor(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify a file to refactor. Example: refactor main.go")
		return
	}
	file := args[0]
	fmt.Printf("🔧 Refactoring: %s\n", file)
	ollama.TalkToOllama(fmt.Sprintf("Refactor this file: %s", file), currentSessionID, historyManager)
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

func handleStatus() {
	fmt.Println("📊 Project Status:")
	fmt.Println("  • AI Model: Ollama (codellama:13b)")
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
	fmt.Printf("💭 Question: %s\n", input)
	ollama.TalkToOllama(input, currentSessionID, historyManager)
}

func init() {
	// Add command handlers
	rootCmd.AddCommand(&cobra.Command{
		Use:   "ask [question]",
		Short: "Ask a question about your codebase",
		Long:  "Ask questions about your project, code structure, or implementation details",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Println("❌ Please provide a question")
				return
			}
			handleAsk(args)
		},
	})

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

func RootCmd() {
	rootCmd.Execute()
}
