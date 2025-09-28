package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type OllamaClient struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		Client:  &http.Client{Timeout: 300 * time.Second}, // Increased to 5 minutes
	}
}

func (o *OllamaClient) Generate(prompt string) (string, error) {
	// Add timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // Increased to 5 minutes
	defer cancel()

	reqBody := OllamaRequest{
		Model:  o.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "POST", o.BaseURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return ollamaResp.Response, nil
}

func StartServer() {
	// Initialize Ollama client
	ollamaClient := NewOllamaClient("http://localhost:11434", "codellama:13b")

	// HTTP server for MCP-like functionality
	http.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req MCPRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		response := processMCPRequest(req, ollamaClient)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Add test endpoint
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "test successful"})
	})

	fmt.Println("🚀 Starting Silent Code MCP Server on port 8080...")
	fmt.Println("💡 Make sure Ollama is running on localhost:11434")
	fmt.Println("🔧 Available tools: create_file, edit_file, read_file, analyze_code, execute_shell")
	fmt.Println("📡 Server will start on http://localhost:8080")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}

func processMCPRequest(req MCPRequest, ollamaClient *OllamaClient) MCPResponse {
	switch req.Method {
	case "tools/call":
		return handleToolCall(req, ollamaClient)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func handleToolCall(req MCPRequest, ollamaClient *OllamaClient) MCPResponse {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	toolName, ok := params["name"].(string)
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Missing tool name",
			},
		}
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Missing arguments",
			},
		}
	}

	var result interface{}
	var err error

	switch toolName {
	case "create_file":
		result, err = handleCreateFile(arguments, ollamaClient)
	case "edit_file":
		result, err = handleEditFile(arguments, ollamaClient)
	case "read_file":
		result, err = handleReadFile(arguments)
	case "analyze_code":
		result, err = handleAnalyzeCode(arguments, ollamaClient)
	case "explain_code":
		result, err = handleExplainCode(arguments, ollamaClient)
	case "execute_shell":
		result, err = handleExecuteShell(arguments)
	default:
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Tool not found",
			},
		}
	}

	if err != nil {
		return MCPResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			},
		}
	}

	return MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func handleCreateFile(params map[string]interface{}, ollamaClient *OllamaClient) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	requirements, ok := params["requirements"].(string)
	if !ok {
		return nil, fmt.Errorf("requirements parameter is required")
	}

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return map[string]interface{}{
			"success": false,
			"error":   "File already exists",
		}, nil
	}

	// Generate file content using Ollama
	prompt := fmt.Sprintf(`Create a new Go file with the following requirements:

FILE PATH: %s
REQUIREMENTS: %s

Return ONLY the complete Go file content with proper package declaration, imports, and implementation. Do not include explanations or markdown formatting.`, filePath, requirements)

	response, err := ollamaClient.Generate(prompt)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("AI generation failed: %v", err),
		}, nil
	}

	// Clean the response
	cleanContent := cleanAIResponse(response)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to create directory: %v", err),
		}, nil
	}

	// Write the file
	if err := os.WriteFile(filePath, []byte(cleanContent), 0644); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to write file: %v", err),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"content": cleanContent,
		"message": fmt.Sprintf("File created successfully: %s", filePath),
	}, nil
}

func handleEditFile(params map[string]interface{}, ollamaClient *OllamaClient) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	editRequest, ok := params["edit_request"].(string)
	if !ok {
		return nil, fmt.Errorf("edit_request parameter is required")
	}

	// Read current file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}

	// Generate edit using Ollama
	prompt := fmt.Sprintf(`Edit this Go file by making the requested change.

FILE: %s
CURRENT CONTENT:
%s

REQUESTED CHANGE: %s

Return ONLY the complete modified file content. Do not include explanations or markdown formatting.`, filePath, string(content), editRequest)

	response, err := ollamaClient.Generate(prompt)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("AI edit failed: %v", err),
		}, nil
	}

	// Clean the response
	cleanContent := cleanAIResponse(response)

	// Write the modified file directly (no backup)
	if err := os.WriteFile(filePath, []byte(cleanContent), 0644); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to write file: %v", err),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"content": cleanContent,
		"message": fmt.Sprintf("File edited successfully: %s", filePath),
	}, nil
}

func handleReadFile(params map[string]interface{}) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"content": string(content),
		"message": fmt.Sprintf("File read successfully: %s", filePath),
	}, nil
}

func handleAnalyzeCode(params map[string]interface{}, ollamaClient *OllamaClient) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	question, ok := params["question"].(string)
	if !ok {
		return nil, fmt.Errorf("question parameter is required")
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}

	// Generate analysis using Ollama
	prompt := fmt.Sprintf(`Analyze this Go code and answer the question.

FILE: %s
CODE:
%s

QUESTION: %s

Provide a detailed analysis and answer.`, filePath, string(content), question)

	response, err := ollamaClient.Generate(prompt)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("AI analysis failed: %v", err),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"content": response,
		"message": "Analysis completed successfully",
	}, nil
}

func handleExplainCode(params map[string]interface{}, ollamaClient *OllamaClient) (interface{}, error) {
	filePath, ok := params["file_path"].(string)
	if !ok {
		return nil, fmt.Errorf("file_path parameter is required")
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to read file: %v", err),
		}, nil
	}

	// Generate detailed explanation using Ollama
	prompt := fmt.Sprintf(`Explain this Go code in detail. Provide a comprehensive explanation covering:

1. What this code does overall
2. Key functions and their purposes
3. Important variables and data structures
4. Control flow and logic
5. Any notable patterns or design decisions
6. How different parts work together

FILE: %s
CODE:
%s

Provide a clear, detailed explanation that would help someone understand this code.`, filePath, string(content))

	response, err := ollamaClient.Generate(prompt)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("AI explanation failed: %v", err),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"content": response,
		"message": "Code explanation completed successfully",
	}, nil
}

func cleanAIResponse(response string) string {
	// Remove markdown code blocks
	response = strings.TrimPrefix(response, "```go")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")

	// Remove common AI prefixes
	prefixes := []string{
		"Here's the Go code:",
		"Here is the Go code:",
		"Here's the complete file:",
		"Here is the complete file:",
		"The Go code is:",
		"Here's your file:",
		"Here is your file:",
		"Here's the modified file:",
		"Here is the modified file:",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(response, prefix) {
			response = strings.TrimSpace(response[len(prefix):])
			break
		}
	}

	return strings.TrimSpace(response)
}

func handleExecuteShell(params map[string]interface{}) (interface{}, error) {
	command, ok := params["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command parameter is required")
	}

	// Parse command and arguments
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return map[string]interface{}{
			"success": false,
			"error":   "Empty command provided",
		}, nil
	}

	// Create command with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)

	// Set working directory to current directory
	cmd.Dir = "."

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Get output
	output := stdout.String()
	errorOutput := stderr.String()

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		return map[string]interface{}{
			"success": false,
			"error":   "Command timed out after 30 seconds",
			"output":  output,
			"stderr":  errorOutput,
		}, nil
	}

	// Determine success based on exit code
	success := err == nil
	message := "Command executed successfully"
	if !success {
		message = fmt.Sprintf("Command failed with error: %v", err)
	}

	return map[string]interface{}{
		"success": success,
		"output":  output,
		"stderr":  errorOutput,
		"message": message,
		"command": command,
	}, nil
}
