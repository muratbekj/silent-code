package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MCPClient struct {
	BaseURL string
	Client  *http.Client
}

type ToolResult struct {
	Success bool   `json:"success"`
	Content string `json:"content,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Output  string `json:"output,omitempty"`
	Stderr  string `json:"stderr,omitempty"`
	Command string `json:"command,omitempty"`
}

func NewMCPClient(baseURL string) *MCPClient {
	return &MCPClient{
		BaseURL: baseURL,
		Client:  &http.Client{Timeout: 150 * time.Second}, // Increased to 150 seconds
	}
}

func (c *MCPClient) CallTool(toolName string, params map[string]interface{}) (*ToolResult, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": params,
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Post(c.BaseURL+"/mcp", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mcpResp MCPResponse
	if err := json.Unmarshal(body, &mcpResp); err != nil {
		return nil, err
	}

	if mcpResp.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", mcpResp.Error.Message)
	}

	// Parse the result
	result, ok := mcpResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	toolResult := &ToolResult{}
	if success, ok := result["success"].(bool); ok {
		toolResult.Success = success
	}
	if content, ok := result["content"].(string); ok {
		toolResult.Content = content
	}
	if message, ok := result["message"].(string); ok {
		toolResult.Message = message
	}
	if errorMsg, ok := result["error"].(string); ok {
		toolResult.Error = errorMsg
	}
	if output, ok := result["output"].(string); ok {
		toolResult.Output = output
	}
	if stderr, ok := result["stderr"].(string); ok {
		toolResult.Stderr = stderr
	}
	if command, ok := result["command"].(string); ok {
		toolResult.Command = command
	}

	return toolResult, nil
}

// Convenience methods for each tool
func (c *MCPClient) CreateFile(filePath, requirements string) (*ToolResult, error) {
	return c.CallTool("create_file", map[string]interface{}{
		"file_path":    filePath,
		"requirements": requirements,
	})
}

func (c *MCPClient) EditFile(filePath, editRequest string) (*ToolResult, error) {
	return c.CallTool("edit_file", map[string]interface{}{
		"file_path":    filePath,
		"edit_request": editRequest,
	})
}

func (c *MCPClient) ReadFile(filePath string) (*ToolResult, error) {
	return c.CallTool("read_file", map[string]interface{}{
		"file_path": filePath,
	})
}

func (c *MCPClient) AnalyzeCode(filePath, question string) (*ToolResult, error) {
	return c.CallTool("analyze_code", map[string]interface{}{
		"file_path": filePath,
		"question":  question,
	})
}

func (c *MCPClient) ExplainCode(filePath string) (*ToolResult, error) {
	return c.CallTool("explain_code", map[string]interface{}{
		"file_path": filePath,
	})
}

func (c *MCPClient) ExecuteShell(command string) (*ToolResult, error) {
	return c.CallTool("execute_shell", map[string]interface{}{
		"command": command,
	})
}
