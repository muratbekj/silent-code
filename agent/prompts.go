package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PromptBuilder struct {
	SystemPrompt string
	CodeContext  string
	ProjectInfo  string
}

// NewPromptBuilder creates a new prompt builder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		SystemPrompt: getSystemPrompt(),
		CodeContext:  "",
		ProjectInfo:  "",
	}
}

// getSystemPrompt returns the base system prompt for coding assistance
func getSystemPrompt() string {
	return `You are Private Code, an AI coding assistant that lives in the terminal. You help developers with:

- Code analysis and explanation
- Code generation and refactoring  
- Debugging and optimization
- Project structure understanding
- Best practices and patterns

Guidelines:
- Always format code blocks with proper syntax highlighting
- Provide clear, actionable explanations
- Focus on Go development when relevant
- Be concise but thorough
- Be natural with your responses
- Ask clarifying questions when needed
- Maintain context across the conversation

When answering questions, be sure to include the following:
- The question being answered
- The answer to the question
- The code that was used to answer the question
- The reasoning behind the answer

You are running locally via Ollama and have access to the project files.`
}

// LoadProjectContext loads relevant project information
func (pb *PromptBuilder) LoadProjectContext(projectPath string) error {
	// Load go.mod for project info
	goModPath := filepath.Join(projectPath, "go.mod")
	if data, err := os.ReadFile(goModPath); err == nil {
		pb.ProjectInfo = fmt.Sprintf("Project Info:\n```go\n%s\n```\n", string(data))
	}

	// Load main files for context
	mainFiles := []string{"main.go", "cmd/root.go", "README.md"}
	var contextParts []string

	for _, file := range mainFiles {
		filePath := filepath.Join(projectPath, file)
		if data, err := os.ReadFile(filePath); err == nil {
			contextParts = append(contextParts, fmt.Sprintf("// %s\n%s", file, string(data)))
		}
	}

	if len(contextParts) > 0 {
		pb.CodeContext = fmt.Sprintf("Current Project Files:\n```go\n%s\n```\n", strings.Join(contextParts, "\n\n"))
	}

	return nil
}

// BuildPrompt constructs the full prompt with context
func (pb *PromptBuilder) BuildPrompt(userInput string, conversationHistory []string) string {
	var parts []string

	// Add system prompt
	parts = append(parts, fmt.Sprintf("System: %s", pb.SystemPrompt))

	// Add project context if available
	if pb.ProjectInfo != "" {
		parts = append(parts, pb.ProjectInfo)
	}

	if pb.CodeContext != "" {
		parts = append(parts, pb.CodeContext)
	}

	// Add conversation history for context
	if len(conversationHistory) > 0 {
		parts = append(parts, fmt.Sprintf("Previous conversation:\n%s", strings.Join(conversationHistory, "\n")))
	}

	// Add current user input
	parts = append(parts, fmt.Sprintf("User: %s", userInput))

	return strings.Join(parts, "\n\n")
}

// GetCodeContext returns the current code context
func (pb *PromptBuilder) GetCodeContext() string {
	return pb.CodeContext
}

// GetProjectInfo returns the current project information
func (pb *PromptBuilder) GetProjectInfo() string {
	return pb.ProjectInfo
}

// UpdateSystemPrompt allows customizing the system prompt
func (pb *PromptBuilder) UpdateSystemPrompt(newPrompt string) {
	pb.SystemPrompt = newPrompt
}

// AddFileContext adds a specific file to the context
func (pb *PromptBuilder) AddFileContext(filePath string) error {
	if data, err := os.ReadFile(filePath); err == nil {
		fileName := filepath.Base(filePath)
		fileContext := fmt.Sprintf("// %s\n%s", fileName, string(data))

		if pb.CodeContext == "" {
			pb.CodeContext = fmt.Sprintf("Current Project Files:\n```go\n%s\n```\n", fileContext)
		} else {
			// Append to existing context
			pb.CodeContext = strings.TrimSuffix(pb.CodeContext, "```\n") + "\n\n" + fileContext + "\n```\n"
		}
		return nil
	}
	return fmt.Errorf("failed to read file: %s", filePath)
}
