package ollama

import (
	"fmt"
	"silent-code/agent"
)

// StartReasoning starts a new reasoning session
func StartReasoning(sessionID, problem string) *agent.MultiTurnReasoning {
	if reasoningManager == nil {
		reasoningManager = agent.NewReasoningManager()
	}
	return reasoningManager.StartReasoning(sessionID, problem)
}

// GetReasoningSummary returns the reasoning summary
func GetReasoningSummary(sessionID string) (string, error) {
	if reasoningManager == nil {
		return "", fmt.Errorf("reasoning manager not initialized")
	}
	return reasoningManager.GetReasoningSummary(sessionID)
}

// AddReasoningStep adds a new step to the reasoning process
func AddReasoningStep(sessionID, thought, action string) error {
	if reasoningManager == nil {
		reasoningManager = agent.NewReasoningManager()
	}
	return reasoningManager.AddStep(sessionID, thought, action)
}
