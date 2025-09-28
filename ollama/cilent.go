package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"silent-code/agent"
	"silent-code/history"
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

// InitializeReasoning sets up the reasoning manager
func InitializeReasoning() {
	reasoningManager = agent.NewReasoningManager()
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
		Model:    "codellama:13b",
		Stream:   true, // Enable streaming
		Messages: messages,
	}

	// Show typing indicator
	fmt.Print("🤖 AI: ")
	showTypingIndicator()

	// Store AI response
	var aiResponse string

	err := talkToOllamaStream(defaultOllamaURL, req, func(content string) {
		aiResponse += content
	})

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
		Model:    "codellama:13b",
		Stream:   true, // Enable streaming
		Messages: messages,
	}

	// Show typing indicator
	fmt.Print("🤖 AI: ")
	showTypingIndicator()

	// Store AI response
	var aiResponse string

	err := talkToOllamaStream(defaultOllamaURL, req, func(content string) {
		aiResponse += content
	})

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
func showTypingIndicator() {
	// Start thinking indicator in background
	go func() {
		time.Sleep(200 * time.Millisecond) // Small delay before showing thinking

		// Show "AI is thinking" with animated dots
		thinkingPhrases := []string{"I am thinking", "I am thinking.", "I am thinking..", "I am thinking..."}

		for i := 0; i < 500; i++ { // Run for about 5 seconds max
			phrase := thinkingPhrases[i%len(thinkingPhrases)]
			fmt.Print("\r🤖: " + phrase + "   ")
			time.Sleep(100 * time.Millisecond)
		}
	}()
}

// talkToOllamaStream handles streaming responses with enhanced typing effect
func talkToOllamaStream(url string, ollamaReq Request, onContent func(string)) error {
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
		Model:    "codellama:13b",
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
