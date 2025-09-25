package agent

import (
	"fmt"
	"strings"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Conversation struct {
	Messages  []Message
	SessionID string
	CreatedAt time.Time
}

type SessionManager struct {
	sessions       map[string]Conversation
	currentSession string
}

type Prompt struct {
	SystemPrompt string
	CodeContext  string
	ProjectInfo  string
}

// ReasoningStep represents a single step in multi-turn reasoning
type ReasoningStep struct {
	Step    int    `json:"step"`
	Thought string `json:"thought"`
	Action  string `json:"action"`
	Result  string `json:"result"`
	Status  string `json:"status"` // "pending", "in_progress", "completed", "failed"
}

// MultiTurnReasoning manages step-by-step problem solving
type MultiTurnReasoning struct {
	Steps       []ReasoningStep `json:"steps"`
	CurrentStep int             `json:"current_step"`
	IsComplete  bool            `json:"is_complete"`
	Problem     string          `json:"problem"`
	Solution    string          `json:"solution"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ReasoningManager handles multi-turn reasoning sessions
type ReasoningManager struct {
	ActiveReasoning map[string]*MultiTurnReasoning
	MaxSteps        int
}

// NewReasoningManager creates a new reasoning manager
func NewReasoningManager() *ReasoningManager {
	return &ReasoningManager{
		ActiveReasoning: make(map[string]*MultiTurnReasoning),
		MaxSteps:        10, // Maximum steps per reasoning session
	}
}

// StartReasoning begins a new multi-turn reasoning session
func (rm *ReasoningManager) StartReasoning(sessionID, problem string) *MultiTurnReasoning {
	reasoning := &MultiTurnReasoning{
		Steps:       []ReasoningStep{},
		CurrentStep: 0,
		IsComplete:  false,
		Problem:     problem,
		Solution:    "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	rm.ActiveReasoning[sessionID] = reasoning
	return reasoning
}

// AddStep adds a new reasoning step
func (rm *ReasoningManager) AddStep(sessionID string, thought, action string) error {
	reasoning, exists := rm.ActiveReasoning[sessionID]
	if !exists {
		return fmt.Errorf("no active reasoning session for session %s", sessionID)
	}

	if len(reasoning.Steps) >= rm.MaxSteps {
		return fmt.Errorf("maximum steps reached (%d)", rm.MaxSteps)
	}

	step := ReasoningStep{
		Step:    len(reasoning.Steps) + 1,
		Thought: thought,
		Action:  action,
		Result:  "",
		Status:  "pending",
	}

	reasoning.Steps = append(reasoning.Steps, step)
	reasoning.CurrentStep = len(reasoning.Steps)
	reasoning.UpdatedAt = time.Now()

	return nil
}

// UpdateStepResult updates the result of the current step
func (rm *ReasoningManager) UpdateStepResult(sessionID string, result string, status string) error {
	reasoning, exists := rm.ActiveReasoning[sessionID]
	if !exists {
		return fmt.Errorf("no active reasoning session for session %s", sessionID)
	}

	if reasoning.CurrentStep < 1 || reasoning.CurrentStep > len(reasoning.Steps) {
		return fmt.Errorf("invalid current step: %d", reasoning.CurrentStep)
	}

	stepIndex := reasoning.CurrentStep - 1
	reasoning.Steps[stepIndex].Result = result
	reasoning.Steps[stepIndex].Status = status
	reasoning.UpdatedAt = time.Now()

	return nil
}

// CompleteReasoning marks the reasoning session as complete
func (rm *ReasoningManager) CompleteReasoning(sessionID, solution string) error {
	reasoning, exists := rm.ActiveReasoning[sessionID]
	if !exists {
		return fmt.Errorf("no active reasoning session for session %s", sessionID)
	}

	reasoning.Solution = solution
	reasoning.IsComplete = true
	reasoning.UpdatedAt = time.Now()

	return nil
}

// GetReasoning returns the current reasoning session
func (rm *ReasoningManager) GetReasoning(sessionID string) (*MultiTurnReasoning, error) {
	reasoning, exists := rm.ActiveReasoning[sessionID]
	if !exists {
		return nil, fmt.Errorf("no active reasoning session for session %s", sessionID)
	}

	return reasoning, nil
}

// GetReasoningSummary returns a formatted summary of the reasoning process
func (rm *ReasoningManager) GetReasoningSummary(sessionID string) (string, error) {
	reasoning, err := rm.GetReasoning(sessionID)
	if err != nil {
		return "", err
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("🧠 Reasoning Process for: %s\n", reasoning.Problem))
	summary.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for i, step := range reasoning.Steps {
		statusIcon := "⏳"
		switch step.Status {
		case "completed":
			statusIcon = "✅"
		case "failed":
			statusIcon = "❌"
		case "in_progress":
			statusIcon = "🔄"
		}

		summary.WriteString(fmt.Sprintf("%s Step %d: %s\n", statusIcon, i+1, step.Thought))
		summary.WriteString(fmt.Sprintf("   Action: %s\n", step.Action))
		if step.Result != "" {
			summary.WriteString(fmt.Sprintf("   Result: %s\n", step.Result))
		}
		summary.WriteString("\n")
	}

	if reasoning.IsComplete && reasoning.Solution != "" {
		summary.WriteString("🎯 Final Solution:\n")
		summary.WriteString(reasoning.Solution)
	}

	return summary.String(), nil
}
