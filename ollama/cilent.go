package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/muratbekj/silent-code/agent"
	"github.com/muratbekj/silent-code/history"
)

type Request struct {
	Model    string          `json:"model"`
	Messages []agent.Message `json:"messages"`
	Stream   bool            `json:"stream"`
}

type Response struct {
	Model              string        `json:"model"`
	CreatedAt          time.Time     `json:"created_at"`
	Message            agent.Message `json:"message"`
	Done               bool          `json:"done"`
	TotalDuration      int64         `json:"total_duration"`
	LoadDuration       int           `json:"load_duration"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	PromptEvalDuration int           `json:"prompt_eval_duration"`
	EvalCount          int           `json:"eval_count"`
	EvalDuration       int64         `json:"eval_duration"`
}
type agentStreamResponse struct {
	Message            agent.Message `json:"message"`
	Done               bool          `json:"done"`
	TotalDuration      int64         `json:"total_duration"`
	LoadDuration       int           `json:"load_duration"`
	PromptEvalCount    int           `json:"prompt_eval_count"`
	PromptEvalDuration int           `json:"prompt_eval_duration"`
	EvalCount          int           `json:"eval_count"`
	EvalDuration       int64         `json:"eval_duration"`
}

const defaultOllamaURL = "http://localhost:11434/api/chat"
const ollamaListURL = "http://localhost:11434/api/tags"

// Global reasoning manager
var reasoningManager *agent.ReasoningManager

// Global model configuration
var currentModel = ""

// InitializeReasoning sets up the reasoning manager
func InitializeReasoning() {
	reasoningManager = agent.NewReasoningManager()
}

// InitializeModelSelection automatically selects the best available model
func InitializeModelSelection() error {
	models, err := ListOllamaModels()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	if len(models) == 0 {
		return fmt.Errorf("no models available. Please install a model first: ollama pull codellama:13b")
	}

	// Select the best model based on priority
	selectedModel := selectBestModel(models)
	currentModel = selectedModel.Name

	return nil
}

// selectBestModel chooses the best model based on coding capabilities and performance
func selectBestModel(models []OllamaModel) OllamaModel {
	// Define model priorities for coding tasks
	// Higher priority models are better for coding
	modelPriorities := map[string]int{
		"qwen2.5-coder:7b":    100,
		"codellama:13b":       95,
		"codellama:34b":       90,
		"qwen2.5-coder:32b":   85,
		"qwen2.5-coder:14b":   80,
		"deepseek-coder:33b":  75,
		"magicoder:15b":       70,
		"deepseek-coder:6.7b": 65,
		"starcoder2:15b":      60,
		"magicoder:7b":        55,
		"starcoder2:7b":       50,
		"starcoder2:3b":       45,
		"qwen2.5:32b":         40,
		"qwen2.5:14b":         35,
		"qwen2.5:7b":          30,
		"gemma2:27b":          25,
		"gemma2:9b":           20,
	}

	var bestModel OllamaModel
	bestScore := -1

	for _, model := range models {
		score := 0

		// Check for exact match in priorities
		if priority, exists := modelPriorities[model.Name]; exists {
			score = priority
		} else {
			// Fallback scoring based on model characteristics
			score = calculateFallbackScore(model)
		}

		// Prefer larger models if scores are close (within 5 points)
		if score > bestScore || (score == bestScore && model.Size > bestModel.Size) {
			bestScore = score
			bestModel = model
		}
	}

	return bestModel
}

// calculateFallbackScore provides a score for models not in the priority list
func calculateFallbackScore(model OllamaModel) int {
	score := 30 // Base score for unknown models

	// Boost score for coding-related keywords in name
	name := strings.ToLower(model.Name)
	if strings.Contains(name, "code") {
		score += 30
	}
	if strings.Contains(name, "coder") {
		score += 25
	}
	if strings.Contains(name, "star") {
		score += 20
	}
	if strings.Contains(name, "wizard") {
		score += 15
	}
	if strings.Contains(name, "magic") {
		score += 15
	}

	// Boost score for larger models (more parameters)
	if model.Details.ParameterSize != "" {
		if strings.Contains(model.Details.ParameterSize, "13B") || strings.Contains(model.Details.ParameterSize, "13b") {
			score += 20
		} else if strings.Contains(model.Details.ParameterSize, "7B") || strings.Contains(model.Details.ParameterSize, "7b") {
			score += 15
		} else if strings.Contains(model.Details.ParameterSize, "34B") || strings.Contains(model.Details.ParameterSize, "34b") {
			score += 25
		} else if strings.Contains(model.Details.ParameterSize, "70B") || strings.Contains(model.Details.ParameterSize, "70b") {
			score += 30
		}
	}

	// Boost score for recent models (based on modification date)
	// This is a simple heuristic - newer models are often better
	daysSinceModified := time.Since(model.ModifiedAt).Hours() / 24
	if daysSinceModified < 30 { // Modified within last 30 days
		score += 10
	} else if daysSinceModified < 90 { // Modified within last 90 days
		score += 5
	}

	return score
}

// SetModel sets the current model for all Ollama requests
func SetModel(modelName string) error {
	// Validate that the model exists
	models, err := ListOllamaModels()
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	// Check if the model exists
	for _, model := range models {
		if model.Name == modelName {
			currentModel = modelName
			return nil
		}
	}

	return fmt.Errorf("model '%s' not found. Use '/config' to see available models", modelName)
}

// GetCurrentModel returns the currently configured model
func GetCurrentModel() string {
	return currentModel
}

func TalkToOllama(userInput string, sessionID string, historyManager *history.HistoryManager) {
	start := time.Now()

	// Initialize prompt builder
	promptBuilder := agent.NewPromptBuilder()

	// Load project context
	promptBuilder.LoadProjectContext(".")

	// Add user message to history
	userMessage := agent.Message{
		Role:    "user",
		Content: userInput,
	}

	if historyManager != nil {
		historyManager.AddMessage(sessionID, userMessage)
	}

	// Get conversation history for context
	var conversationHistory []string
	if historyManager != nil {
		history, err := historyManager.GetSessionHistory(sessionID)
		if err == nil {
			// Convert history to conversation format
			for _, msg := range history {
				conversationHistory = append(conversationHistory, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
			}
		}
	}

	// Build enhanced prompt with context
	enhancedPrompt := promptBuilder.BuildPrompt(userInput, conversationHistory)

	// Create messages with system prompt
	messages := []agent.Message{
		{
			Role:    "system",
			Content: promptBuilder.SystemPrompt,
		},
		{
			Role:    "user",
			Content: enhancedPrompt,
		},
	}

	req := Request{
		Model:    currentModel,
		Stream:   true, // Enable streaming
		Messages: messages,
	}

	// Show typing indicator
	fmt.Print("🤖 AI: ")
	stopTyping := showTypingIndicator()

	// Store AI response
	var aiResponse string

	err := talkToOllamaStream(defaultOllamaURL, req, func(content string) {
		aiResponse += content
	}, stopTyping)

	if err != nil {
		fmt.Printf("❌ Error talking to Ollama: %v\n", err)
		return
	}

	// Add AI response to history
	if historyManager != nil && aiResponse != "" {
		aiMessage := agent.Message{
			Role:    "assistant",
			Content: aiResponse,
		}
		historyManager.AddMessage(sessionID, aiMessage)
	}

	fmt.Printf("\n⏱️  Completed in %v\n", time.Since(start))
}

// TalkToOllamaWithResponse returns the AI response as a string
func TalkToOllamaWithResponse(userInput string, sessionID string, historyManager *history.HistoryManager) (string, error) {
	start := time.Now()

	// Initialize prompt builder
	promptBuilder := agent.NewPromptBuilder()

	// Load project context
	promptBuilder.LoadProjectContext(".")

	// Add user message to history
	userMessage := agent.Message{
		Role:    "user",
		Content: userInput,
	}

	if historyManager != nil {
		historyManager.AddMessage(sessionID, userMessage)
	}

	// Get conversation history for context
	var conversationHistory []string
	if historyManager != nil {
		history, err := historyManager.GetSessionHistory(sessionID)
		if err == nil {
			// Convert history to conversation format
			for _, msg := range history {
				conversationHistory = append(conversationHistory, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
			}
		}
	}

	// Build enhanced prompt with context
	enhancedPrompt := promptBuilder.BuildPrompt(userInput, conversationHistory)

	// Create messages with system prompt
	messages := []agent.Message{
		{
			Role:    "system",
			Content: promptBuilder.SystemPrompt,
		},
		{
			Role:    "user",
			Content: enhancedPrompt,
		},
	}

	req := Request{
		Model:    currentModel,
		Stream:   true, // Enable streaming
		Messages: messages,
	}

	// Show typing indicator
	fmt.Print("🤖 AI: ")
	stopTyping := showTypingIndicator()

	// Store AI response
	var aiResponse string

	err := talkToOllamaStream(defaultOllamaURL, req, func(content string) {
		aiResponse += content
	}, stopTyping)

	if err != nil {
		return "", fmt.Errorf("error talking to Ollama: %w", err)
	}

	// Add AI response to history
	if historyManager != nil && aiResponse != "" {
		aiMessage := agent.Message{
			Role:    "assistant",
			Content: aiResponse,
		}
		historyManager.AddMessage(sessionID, aiMessage)
	}

	fmt.Printf("\n⏱️  Completed in %v\n", time.Since(start))
	return aiResponse, nil
}

// showTypingIndicator displays an "AI is thinking" animation
func showTypingIndicator() chan bool {
	stopChan := make(chan bool, 1)

	// Start thinking indicator in background
	go func() {
		time.Sleep(200 * time.Millisecond) // Small delay before showing thinking

		// Show "AI is thinking" with animated dots
		thinkingPhrases := []string{"I am thinking", "I am thinking.", "I am thinking..", "I am thinking..."}

		for i := 0; i < 50; i++ { // Run for about 5 seconds max
			select {
			case <-stopChan:
				return // Stop if signaled
			default:
				phrase := thinkingPhrases[i%len(thinkingPhrases)]
				fmt.Print("\r🤖: " + phrase + "   ")
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	return stopChan
}

// talkToOllamaStream handles streaming responses with enhanced typing effect
func talkToOllamaStream(url string, ollamaReq Request, onContent func(string), stopTyping chan bool) error {
	js, err := json.Marshal(&ollamaReq)
	if err != nil {
		return err
	}

	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(js))
	if err != nil {
		return err
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	// Read streaming response line by line
	scanner := bufio.NewScanner(httpResp.Body)
	firstToken := true

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse each JSON line from the stream
		var streamResp agentStreamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			continue // Skip malformed JSON lines
		}

		// Print the content as it streams
		if streamResp.Message.Content != "" {
			// Clear thinking indicator on first token
			if firstToken {
				// Stop the thinking indicator
				select {
				case stopTyping <- true:
				default:
				}
				fmt.Print("\r🤖 AI: ") // Clear thinking indicator and reset to AI prompt
				firstToken = false
			}

			// Add small delay to simulate typing speed
			time.Sleep(10 * time.Millisecond)
			fmt.Print(streamResp.Message.Content)

			// Call the callback to store content
			if onContent != nil {
				onContent(streamResp.Message.Content)
			}
		}

		// Check if streaming is done
		if streamResp.Done {
			break
		}
	}

	return scanner.Err()
}

// OllamaModel represents a model from Ollama
type OllamaModel struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    struct {
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}

// OllamaModelsResponse represents the response from Ollama list API
type OllamaModelsResponse struct {
	Models []OllamaModel `json:"models"`
}

// ListOllamaModels fetches and returns the list of installed Ollama models
func ListOllamaModels() ([]OllamaModel, error) {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ollamaListURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var modelsResponse OllamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return modelsResponse.Models, nil
}

// TalkToOllamaWithTyping provides enhanced typing simulation
func TalkToOllamaWithTyping(userInput string) {
	start := time.Now()

	msg := agent.Message{
		Role:    "user",
		Content: userInput,
	}

	req := Request{
		Model:    currentModel,
		Stream:   true,
		Messages: []agent.Message{msg},
	}

	fmt.Print("🤖 AI: ")

	// Enhanced typing indicator
	go func() {
		time.Sleep(200 * time.Millisecond)
		for i := 0; i < 3; i++ {
			fmt.Print(".")
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Print("\b\b\b   \b\b\b") // Clear dots
	}()

	err := talkToOllamaStreamEnhanced(defaultOllamaURL, req)
	if err != nil {
		fmt.Printf("❌ Error talking to Ollama: %v\n", err)
		return
	}

	fmt.Printf("\n⏱️  Completed in %v\n", time.Since(start))
}

func talkToOllamaStreamEnhanced(url string, ollamaReq Request) error {
	js, err := json.Marshal(&ollamaReq)
	if err != nil {
		return err
	}

	client := http.Client{}
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(js))
	if err != nil {
		return err
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	scanner := bufio.NewScanner(httpResp.Body)
	firstToken := true

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var streamResp agentStreamResponse
		if err := json.Unmarshal([]byte(line), &streamResp); err != nil {
			continue
		}

		if streamResp.Message.Content != "" {
			// Clear typing indicator on first token
			if firstToken {
				fmt.Print("\b\b\b   \b\b\b") // Clear typing dots
				firstToken = false
			}

			// Simulate realistic typing speed
			time.Sleep(15 * time.Millisecond)
			fmt.Print(streamResp.Message.Content)
		}

		if streamResp.Done {
			break
		}
	}

	return scanner.Err()
}
