# Silent Code

An AI-powered development assistant that lives in your terminal. Built with privacy-first architecture, Silent Code runs fully offline on your machine using local LLMs via Ollama.

## ✨ Features

- **🔍 Smart Code Analysis**: Understands your project structure and provides contextual assistance
- **📝 File Operations**: Create, edit, and analyze files with AI assistance
- **🛠️ Shell Integration**: Execute commands and get AI-powered explanations
- **💬 Natural Conversations**: Ask questions about your codebase in plain English
- **🔒 Privacy-First**: All processing happens locally on your machine
- **⚡ Fast & Responsive**: Optimized for terminal workflows

## Reasons for building
I thought it would be cool to build a AI coding assistant that uses your models locally and privately, no paying just to have access to these coding terminals. Just download ollama and make sure that your computer can run powerful models

## 🚀 Installation

### Prerequisites

1. **Install Ollama** (if not already installed):
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

2. **Pull a coding model** (recommended):
```bash
# For best performance (requires more RAM)
ollama pull codellama:13b

# For lighter usage
ollama pull qwen2.5-coder:7b
```

3. **Start Ollama server**:
```bash
ollama serve
```

### Install Silent Code

```bash
go install github.com/muratbekj/silent-code@latest
```

## 🎯 Usage

### Basic Usage

```bash
silent-code
```

### Available Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/context` | Show current project context |
| `/explain <file>` | Explain a specific file or function |
| `/generate <what>` | Generate new code |
| `/edit <file> <request>` | Edit file with AI assistance |
| `/new <file> <requirements>` | Create new file with AI assistance |
| `/read <file>` | View file contents |
| `/search <query>` | Search through codebase semantically |
| `/config` | Show available Ollama models |
| `/sessions` | Manage conversation sessions |
| `/exit` | Exit the assistant |

### Examples

**Ask questions about your code:**
```bash
silent-code> what is this project?
silent-code> how does this function work?
silent-code> explain the authentication logic
```

**Work with files:**
```bash
silent-code> /edit main.py "add error handling"
silent-code> /new utils.py "create utility functions"
silent-code> /explain config.py
```

**Execute shell commands:**
```bash
silent-code> ls -la
silent-code> git status
silent-code> npm install
```

## 🔧 Configuration

### Model Selection

```bash
silent-code> /config
```

Switch models:
```bash
silent-code> /config models codellama:13b
```

## 🏗️ Architecture

- **Local Processing**: All AI processing happens on your machine
- **Ollama Integration**: Uses any Ollama-compatible model
- **MCP Protocol**: Model Context Protocol for file operations
- **Session Management**: Maintains conversation context
- **Multi-language Detection**: Smart project type detection

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Ollama](https://ollama.com/) for local LLM inference
- Uses [Cobra](https://github.com/spf13/cobra) for CLI framework
- Inspired by modern AI coding assistants




