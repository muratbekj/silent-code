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
- Detect and work with the appropriate programming language for the project
- Be concise but thorough
- Be natural with your responses
- Ask clarifying questions when needed
- Maintain context across the conversation
- Support multiple programming languages (Python, JavaScript, TypeScript, Go, Java, C++, C#, PHP, Ruby, Rust, Swift, Kotlin, etc.)

When answering questions, be sure to include the following:
- The question being answered
- The answer to the question
- The code that was used to answer the question
- The reasoning behind the answer

You are running locally via Ollama and have access to the project files.`
}

// LoadProjectContext loads relevant project information
func (pb *PromptBuilder) LoadProjectContext(projectPath string) error {
	// Detect project type and load appropriate files
	projectType := detectProjectType(projectPath)

	// Load project-specific configuration files
	configFiles := getConfigFiles(projectType)
	for _, file := range configFiles {
		filePath := filepath.Join(projectPath, file)
		if data, err := os.ReadFile(filePath); err == nil {
			ext := filepath.Ext(file)
			language := getLanguageFromExtension(ext)
			pb.ProjectInfo += fmt.Sprintf("Project Info (%s):\n```%s\n%s\n```\n", file, language, string(data))
		}
	}

	// Load main files for context based on project type
	mainFiles := getMainFiles(projectType)
	var contextParts []string

	for _, file := range mainFiles {
		filePath := filepath.Join(projectPath, file)
		if data, err := os.ReadFile(filePath); err == nil {
			contextParts = append(contextParts, fmt.Sprintf("// %s\n%s", file, string(data)))
		}
	}

	if len(contextParts) > 0 {
		// Use the most common language in the project for code context
		primaryLanguage := getPrimaryLanguage(projectPath)
		pb.CodeContext = fmt.Sprintf("Current Project Files:\n```%s\n%s\n```\n", primaryLanguage, strings.Join(contextParts, "\n\n"))
	}

	return nil
}

// detectProjectType detects the type of project based on configuration files
func detectProjectType(projectPath string) string {
	configFiles := map[string]string{
		"go.mod":           "Go",
		"package.json":     "JavaScript/Node.js",
		"requirements.txt": "Python",
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

// getConfigFiles returns configuration files for a project type
func getConfigFiles(projectType string) []string {
	configMap := map[string][]string{
		"Go":                 {"go.mod", "go.sum"},
		"JavaScript/Node.js": {"package.json", "package-lock.json", "yarn.lock"},
		"Python":             {"requirements.txt", "setup.py", "pyproject.toml"},
		"Java":               {"pom.xml", "build.gradle"},
		"Java/Gradle":        {"build.gradle", "gradle.properties"},
		"Rust":               {"cargo.toml", "Cargo.lock"},
		"PHP":                {"composer.json", "composer.lock"},
		"Ruby":               {"Gemfile", "Gemfile.lock"},
		"Swift/Objective-C":  {"Podfile", "Podfile.lock"},
		"Elixir":             {"mix.exs", "mix.lock"},
		"Dart/Flutter":       {"pubspec.yaml", "pubspec.lock"},
	}

	if files, exists := configMap[projectType]; exists {
		return files
	}

	return []string{}
}

// getMainFiles returns main files for a project type
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

// getLanguageFromExtension returns the language name for a file extension
func getLanguageFromExtension(ext string) string {
	languageMap := map[string]string{
		".go":    "go",
		".js":    "javascript",
		".ts":    "typescript",
		".py":    "python",
		".java":  "java",
		".cpp":   "cpp",
		".c":     "c",
		".cs":    "csharp",
		".php":   "php",
		".rb":    "ruby",
		".rs":    "rust",
		".swift": "swift",
		".kt":    "kotlin",
		".scala": "scala",
		".r":     "r",
		".m":     "objc",
		".mm":    "objcpp",
		".pl":    "perl",
		".sh":    "bash",
		".lua":   "lua",
		".dart":  "dart",
		".vue":   "vue",
		".html":  "html",
		".css":   "css",
		".scss":  "scss",
		".sass":  "sass",
		".less":  "less",
		".xml":   "xml",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".toml":  "toml",
		".ini":   "ini",
		".sql":   "sql",
		".md":    "markdown",
		".txt":   "text",
	}

	if lang, exists := languageMap[strings.ToLower(ext)]; exists {
		return lang
	}

	return "text"
}

// getPrimaryLanguage determines the primary language of the project
func getPrimaryLanguage(projectPath string) string {
	projectType := detectProjectType(projectPath)

	languageMap := map[string]string{
		"Go":                 "go",
		"JavaScript/Node.js": "javascript",
		"Python":             "python",
		"Java":               "java",
		"Java/Gradle":        "java",
		"Rust":               "rust",
		"PHP":                "php",
		"Ruby":               "ruby",
		"Swift/Objective-C":  "swift",
		"Elixir":             "elixir",
		"Dart/Flutter":       "dart",
	}

	if lang, exists := languageMap[projectType]; exists {
		return lang
	}

	return "go" // Default fallback
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
