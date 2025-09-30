package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/muratbekj/silent-code/agent"
)

type HistoryManager struct {
	HistoryDir string
	Sessions   map[string]*agent.Conversation
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(historyDir string) *HistoryManager {
	return &HistoryManager{
		HistoryDir: historyDir,
		Sessions:   make(map[string]*agent.Conversation),
	}
}

// SaveSession saves a conversation to disk
func (hm *HistoryManager) SaveSession(sessionID string, conversation *agent.Conversation) error {
	// Ensure history directory exists
	if err := os.MkdirAll(hm.HistoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// Create session file path
	sessionFile := filepath.Join(hm.HistoryDir, fmt.Sprintf("session_%s.json", sessionID))

	// Marshal conversation to JSON
	data, err := json.MarshalIndent(conversation, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	// Write to file
	if err := os.WriteFile(sessionFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	// Update in-memory sessions
	hm.Sessions[sessionID] = conversation

	return nil
}

// LoadSession loads a conversation from disk
func (hm *HistoryManager) LoadSession(sessionID string) (*agent.Conversation, error) {
	// Check if already in memory
	if conv, exists := hm.Sessions[sessionID]; exists {
		return conv, nil
	}

	// Load from disk
	sessionFile := filepath.Join(hm.HistoryDir, fmt.Sprintf("session_%s.json", sessionID))

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var conversation agent.Conversation
	if err := json.Unmarshal(data, &conversation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	// Store in memory
	hm.Sessions[sessionID] = &conversation

	return &conversation, nil
}

// ListSessions returns all available session IDs
func (hm *HistoryManager) ListSessions() ([]string, error) {
	// Ensure history directory exists
	if err := os.MkdirAll(hm.HistoryDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create history directory: %w", err)
	}

	// Read directory
	entries, err := os.ReadDir(hm.HistoryDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read history directory: %w", err)
	}

	var sessions []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			// Extract session ID from filename
			name := entry.Name()
			if len(name) > 8 && name[:8] == "session_" {
				sessionID := name[8 : len(name)-5] // Remove "session_" prefix and ".json" suffix
				sessions = append(sessions, sessionID)
			}
		}
	}

	return sessions, nil
}

// AddMessage adds a message to a session
func (hm *HistoryManager) AddMessage(sessionID string, message agent.Message) error {
	// Load or create session
	conversation, err := hm.LoadSession(sessionID)
	if err != nil {
		// Create new session if it doesn't exist
		conversation = &agent.Conversation{
			SessionID: sessionID,
			CreatedAt: time.Now(),
			Messages:  []agent.Message{},
		}
	}

	// Add message
	conversation.Messages = append(conversation.Messages, message)

	// Save updated session
	return hm.SaveSession(sessionID, conversation)
}

// GetSessionHistory returns all messages for a session
func (hm *HistoryManager) GetSessionHistory(sessionID string) ([]agent.Message, error) {
	conversation, err := hm.LoadSession(sessionID)
	if err != nil {
		return nil, err
	}

	return conversation.Messages, nil
}

// DeleteSession removes a session from disk and memory
func (hm *HistoryManager) DeleteSession(sessionID string) error {
	// Remove from memory
	delete(hm.Sessions, sessionID)

	// Remove from disk
	sessionFile := filepath.Join(hm.HistoryDir, fmt.Sprintf("session_%s.json", sessionID))

	if err := os.Remove(sessionFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete session file: %w", err)
	}

	return nil
}
