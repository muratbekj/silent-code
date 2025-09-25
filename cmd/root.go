package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

// Interactive terminal mode
func startInteractiveMode() {
	fmt.Println("🤖 Private Code - AI Terminal")
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
	default:
		// Treat as a general question
		handleGeneralQuestion(input)
	}
}

func showHelp() {
	fmt.Println("\n📋 Available Commands:")
	fmt.Println("  ask <question>     - Ask questions about your codebase")
	fmt.Println("  explain <file>     - Explain a specific file or function")
	fmt.Println("  generate <what>    - Generate new code")
	fmt.Println("  refactor <file>    - Refactor existing code")
	fmt.Println("  test               - Run tests and analyze results")
	fmt.Println("  search <query>     - Search through codebase semantically")
	fmt.Println("  status             - Show current project status")
	fmt.Println("  help               - Show this help message")
	fmt.Println("  exit/quit          - Exit the terminal")
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
	fmt.Println("🔄 [AI Response would go here - Ollama integration needed]")
}

func handleExplain(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify a file or function to explain. Example: explain main.go")
		return
	}
	target := args[0]
	fmt.Printf("📖 Explaining: %s\n", target)
	fmt.Println("🔄 [AI Explanation would go here - Ollama integration needed]")
}

func handleGenerate(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify what to generate. Example: generate 'a new API endpoint'")
		return
	}
	what := strings.Join(args, " ")
	fmt.Printf("⚡ Generating: %s\n", what)
	fmt.Println("🔄 [AI Code Generation would go here - Ollama integration needed]")
}

func handleRefactor(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please specify a file to refactor. Example: refactor main.go")
		return
	}
	file := args[0]
	fmt.Printf("🔧 Refactoring: %s\n", file)
	fmt.Println("🔄 [AI Refactoring would go here - Ollama integration needed]")
}

func handleTest(args []string) {
	fmt.Println("🧪 Running tests...")
	fmt.Println("🔄 [Test execution and AI analysis would go here - Ollama integration needed]")
}

func handleSearch(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please provide a search query. Example: search 'authentication logic'")
		return
	}
	query := strings.Join(args, " ")
	fmt.Printf("🔍 Searching for: %s\n", query)
	fmt.Println("🔄 [Semantic search would go here - Ollama integration needed]")
}

func handleStatus() {
	fmt.Println("📊 Project Status:")
	fmt.Println("  • AI Model: Not connected (Ollama integration pending)")
	fmt.Println("  • Project: silent-code")
	fmt.Println("  • Language: Go")
	fmt.Println("  • Status: Ready for AI integration")
}

func handleGeneralQuestion(input string) {
	fmt.Printf("💭 Question: %s\n", input)
	fmt.Println("🔄 [AI Response would go here - Ollama integration needed]")
}

func init() {
	// Remove the old greet command and add new commands
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
