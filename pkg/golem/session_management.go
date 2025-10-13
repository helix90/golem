package golem

import (
	"fmt"
	"strings"
	"time"
)

// SessionManagement provides session and history management utilities
type SessionManagement struct {
	golem *Golem
}

// NewSessionManagement creates a new session management instance
func NewSessionManagement(golem *Golem) *SessionManagement {
	return &SessionManagement{golem: golem}
}

// Note: ChatSession, ContextConfig, ContextAnalytics, and ContextItem are defined in aiml_native.go

// NewChatSession creates a new chat session
func (sm *SessionManagement) NewChatSession(id string) *ChatSession {
	now := time.Now().Format(time.RFC3339)
	return &ChatSession{
		ID:                id,
		CreatedAt:         now,
		LastActivity:      now,
		Variables:         make(map[string]string),
		RequestHistory:    make([]string, 0),
		ResponseHistory:   make([]string, 0),
		ThatHistory:       make([]string, 0),
		ContextConfig:     sm.getDefaultContextConfig(),
		ContextWeights:    make(map[string]float64),
		ContextUsage:      make(map[string]int),
		ContextTags:       make(map[string][]string),
		ContextMetadata:   make(map[string]interface{}),
		LearnedCategories: make([]Category, 0),
		LearningStats:     &SessionLearningStats{},
	}
}

// getDefaultContextConfig returns default context configuration
func (sm *SessionManagement) getDefaultContextConfig() *ContextConfig {
	return &ContextConfig{
		MaxThatDepth:         20,
		MaxRequestDepth:      20,
		MaxResponseDepth:     20,
		MaxTotalContext:      100,
		CompressionThreshold: 50,
		WeightDecay:          0.9,
		EnableCompression:    true,
		EnableAnalytics:      true,
		EnablePruning:        true,
	}
}

// SetSessionTopic sets the topic for a session
func (sm *SessionManagement) SetSessionTopic(session *ChatSession, topic string) {
	session.Topic = topic
	session.LastActivity = time.Now().Format(time.RFC3339)
}

// GetSessionTopic gets the topic for a session
func (sm *SessionManagement) GetSessionTopic(session *ChatSession) string {
	return session.Topic
}

// AddToThatHistory adds a response to the that history
func (sm *SessionManagement) AddToThatHistory(session *ChatSession, response string) {
	if session == nil {
		return
	}

	// Add to history
	session.ThatHistory = append(session.ThatHistory, response)

	// Maintain max size
	if len(session.ThatHistory) > session.ContextConfig.MaxThatDepth {
		session.ThatHistory = session.ThatHistory[1:]
	}

	session.LastActivity = time.Now().Format(time.RFC3339)
}

// GetLastThat gets the last response from that history
func (sm *SessionManagement) GetLastThat(session *ChatSession) string {
	if session == nil || len(session.ThatHistory) == 0 {
		return ""
	}
	return session.ThatHistory[len(session.ThatHistory)-1]
}

// GetThatHistory gets the that history
func (sm *SessionManagement) GetThatHistory(session *ChatSession) []string {
	if session == nil {
		return []string{}
	}
	return session.ThatHistory
}

// AddToRequestHistory adds a request to the request history
func (sm *SessionManagement) AddToRequestHistory(session *ChatSession, request string) {
	if session == nil {
		return
	}

	session.RequestHistory = append(session.RequestHistory, request)

	// Maintain max size
	if len(session.RequestHistory) > session.ContextConfig.MaxRequestDepth {
		session.RequestHistory = session.RequestHistory[1:]
	}

	session.LastActivity = time.Now().Format(time.RFC3339)
}

// GetRequestHistory gets the request history
func (sm *SessionManagement) GetRequestHistory(session *ChatSession) []string {
	if session == nil {
		return []string{}
	}
	return session.RequestHistory
}

// GetRequestByIndex gets a request by index
func (sm *SessionManagement) GetRequestByIndex(session *ChatSession, index int) string {
	if session == nil || index < 0 || index >= len(session.RequestHistory) {
		return ""
	}
	return session.RequestHistory[index]
}

// AddToResponseHistory adds a response to the response history
func (sm *SessionManagement) AddToResponseHistory(session *ChatSession, response string) {
	if session == nil {
		return
	}

	session.ResponseHistory = append(session.ResponseHistory, response)

	// Maintain max size
	if len(session.ResponseHistory) > session.ContextConfig.MaxResponseDepth {
		session.ResponseHistory = session.ResponseHistory[1:]
	}

	session.LastActivity = time.Now().Format(time.RFC3339)
}

// GetResponseHistory gets the response history
func (sm *SessionManagement) GetResponseHistory(session *ChatSession) []string {
	if session == nil {
		return []string{}
	}
	return session.ResponseHistory
}

// GetResponseByIndex gets a response by index
func (sm *SessionManagement) GetResponseByIndex(session *ChatSession, index int) string {
	if session == nil || index < 0 || index >= len(session.ResponseHistory) {
		return ""
	}
	return session.ResponseHistory[index]
}

// GetThatByIndex gets a that by index
func (sm *SessionManagement) GetThatByIndex(session *ChatSession, index int) string {
	if session == nil || index < 0 || index >= len(session.ThatHistory) {
		return ""
	}
	return session.ThatHistory[index]
}

// GetThatHistoryStats gets statistics about that history
func (sm *SessionManagement) GetThatHistoryStats(session *ChatSession) map[string]interface{} {
	if session == nil {
		return map[string]interface{}{}
	}

	totalSize := 0
	for _, item := range session.ThatHistory {
		totalSize += len(item)
	}

	avgSize := 0.0
	if len(session.ThatHistory) > 0 {
		avgSize = float64(totalSize) / float64(len(session.ThatHistory))
	}

	return map[string]interface{}{
		"count":        len(session.ThatHistory),
		"total_size":   totalSize,
		"average_size": avgSize,
		"memory_usage": sm.calculateThatHistoryMemoryUsage(session),
	}
}

// calculateThatHistoryMemoryUsage calculates memory usage of that history
func (sm *SessionManagement) calculateThatHistoryMemoryUsage(session *ChatSession) int {
	totalSize := 0
	for _, item := range session.ThatHistory {
		totalSize += len([]byte(item))
	}
	return totalSize
}

// CompressThatHistory compresses the that history
func (sm *SessionManagement) CompressThatHistory(session *ChatSession) {
	if session == nil || !session.ContextConfig.EnableCompression {
		return
	}

	// Simple compression - remove oldest items if over threshold
	threshold := session.ContextConfig.CompressionThreshold
	if len(session.ThatHistory) > threshold {
		// Keep only the most recent items
		keepCount := threshold / 2
		startIdx := len(session.ThatHistory) - keepCount
		if startIdx < 0 {
			startIdx = 0
		}
		session.ThatHistory = session.ThatHistory[startIdx:]
	}
}

// ValidateThatHistory validates the that history
func (sm *SessionManagement) ValidateThatHistory(session *ChatSession) []string {
	if session == nil {
		return []string{}
	}

	var issues []string

	// Check for empty items
	for i, item := range session.ThatHistory {
		if strings.TrimSpace(item) == "" {
			issues = append(issues, fmt.Sprintf("Empty that item at index %d", i))
		}
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for i, item := range session.ThatHistory {
		if seen[item] {
			issues = append(issues, fmt.Sprintf("Duplicate that item at index %d", i))
		}
		seen[item] = true
	}

	return issues
}

// ClearThatHistory clears the that history
func (sm *SessionManagement) ClearThatHistory(session *ChatSession) {
	if session == nil {
		return
	}
	session.ThatHistory = make([]string, 0)
}

// GetThatHistoryDebugInfo gets debug information about that history
func (sm *SessionManagement) GetThatHistoryDebugInfo(session *ChatSession) map[string]interface{} {
	if session == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"session_id":    session.ID,
		"history_count": len(session.ThatHistory),
		"created_at":    session.CreatedAt,
		"last_activity": session.LastActivity,
		"config":        session.ContextConfig,
		"analytics":     nil, // ContextAnalytics not available in ChatSession
	}
}

// InitializeContextConfig initializes context configuration
func (sm *SessionManagement) InitializeContextConfig(session *ChatSession) {
	if session == nil {
		return
	}

	if session.ContextConfig == nil {
		session.ContextConfig = sm.getDefaultContextConfig()
	}

	// ContextAnalytics is not part of ChatSession, so we skip initialization
}

// SearchContext searches context items
func (sm *SessionManagement) SearchContext(session *ChatSession, query string, contextTypes []string) []ContextItem {
	if session == nil {
		return []ContextItem{}
	}

	var results []ContextItem

	// Search that history
	if sm.contains(contextTypes, "that") {
		for i, item := range session.ThatHistory {
			if strings.Contains(strings.ToLower(item), strings.ToLower(query)) {
				results = append(results, ContextItem{
					Content:    item,
					Type:       "that",
					Index:      i,
					Weight:     1.0,
					Tags:       []string{},
					Metadata:   map[string]interface{}{"index": i},
					CreatedAt:  session.LastActivity,
					LastUsed:   session.LastActivity,
					UsageCount: 1,
				})
			}
		}
	}

	// Search request history
	if sm.contains(contextTypes, "request") {
		for i, item := range session.RequestHistory {
			if strings.Contains(strings.ToLower(item), strings.ToLower(query)) {
				results = append(results, ContextItem{
					Content:    item,
					Type:       "request",
					Index:      i,
					Weight:     1.0,
					Tags:       []string{},
					Metadata:   map[string]interface{}{"index": i},
					CreatedAt:  session.LastActivity,
					LastUsed:   session.LastActivity,
					UsageCount: 1,
				})
			}
		}
	}

	// Search response history
	if sm.contains(contextTypes, "response") {
		for i, item := range session.ResponseHistory {
			if strings.Contains(strings.ToLower(item), strings.ToLower(query)) {
				results = append(results, ContextItem{
					Content:    item,
					Type:       "response",
					Index:      i,
					Weight:     1.0,
					Tags:       []string{},
					Metadata:   map[string]interface{}{"index": i},
					CreatedAt:  session.LastActivity,
					LastUsed:   session.LastActivity,
					UsageCount: 1,
				})
			}
		}
	}

	return results
}

// contains checks if a slice contains a string
func (sm *SessionManagement) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// CompressContext compresses context if needed
func (sm *SessionManagement) CompressContext(session *ChatSession) {
	if session == nil || !session.ContextConfig.EnableCompression {
		return
	}

	// Compress that history
	sm.CompressThatHistory(session)

	// Compress other histories if over threshold
	threshold := session.ContextConfig.CompressionThreshold

	if len(session.RequestHistory) > threshold {
		keepCount := threshold / 2
		startIdx := len(session.RequestHistory) - keepCount
		if startIdx < 0 {
			startIdx = 0
		}
		session.RequestHistory = session.RequestHistory[startIdx:]
	}

	if len(session.ResponseHistory) > threshold {
		keepCount := threshold / 2
		startIdx := len(session.ResponseHistory) - keepCount
		if startIdx < 0 {
			startIdx = 0
		}
		session.ResponseHistory = session.ResponseHistory[startIdx:]
	}
}
