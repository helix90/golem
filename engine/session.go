package engine

import "sync"

// UserSession holds per-user AIML context
// Includes topic, that, and variables
//
type UserSession struct {
	UserID   string
	Topic    string
	That     string
	Vars     map[string]string
	Wildcards map[string][]string // for completeness, if needed
	mu       sync.RWMutex
}

func newUserSession(userID string) *UserSession {
	return &UserSession{
		UserID: userID,
		Topic:  "*",
		That:   "*",
		Vars:   make(map[string]string),
		Wildcards: make(map[string][]string),
	}
}

// SessionManager manages all user sessions
//
type SessionManager struct {
	sessions map[string]*UserSession
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*UserSession),
	}
}

// GetOrCreateSession returns the session for a user, creating it if needed
func (sm *SessionManager) GetOrCreateSession(userID string) *UserSession {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sess, ok := sm.sessions[userID]
	if !ok {
		sess = newUserSession(userID)
		sm.sessions[userID] = sess
	}
	return sess
}

// UpdateTopic sets the topic for a user
func (sm *SessionManager) UpdateTopic(userID, topic string) {
	sess := sm.GetOrCreateSession(userID)
	sess.mu.Lock()
	defer sess.mu.Unlock()
	sess.Topic = topic
}

// UpdateThat sets the 'that' (last bot response) for a user
func (sm *SessionManager) UpdateThat(userID, that string) {
	sess := sm.GetOrCreateSession(userID)
	sess.mu.Lock()
	defer sess.mu.Unlock()
	sess.That = that
}

// SetVar sets a variable for a user
func (sm *SessionManager) SetVar(userID, key, value string) {
	sess := sm.GetOrCreateSession(userID)
	sess.mu.Lock()
	defer sess.mu.Unlock()
	sess.Vars[key] = value
}

// GetVar gets a variable for a user
func (sm *SessionManager) GetVar(userID, key string) string {
	sess := sm.GetOrCreateSession(userID)
	sess.mu.RLock()
	defer sess.mu.RUnlock()
	return sess.Vars[key]
}

// GetTopic returns the current topic for a user
func (sm *SessionManager) GetTopic(userID string) string {
	sess := sm.GetOrCreateSession(userID)
	sess.mu.RLock()
	defer sess.mu.RUnlock()
	return sess.Topic
}

// GetThat returns the last bot response for a user
func (sm *SessionManager) GetThat(userID string) string {
	sess := sm.GetOrCreateSession(userID)
	sess.mu.RLock()
	defer sess.mu.RUnlock()
	return sess.That
} 