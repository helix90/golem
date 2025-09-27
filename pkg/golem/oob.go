package golem

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

// OOBHandler defines the interface for Out-of-Band message handlers
type OOBHandler interface {
	// CanHandle returns true if this handler can process the given OOB message
	CanHandle(message string) bool

	// Process handles the OOB message and returns a response
	Process(message string, session *ChatSession) (string, error)

	// GetName returns the handler name for identification
	GetName() string

	// GetDescription returns a description of what this handler does
	GetDescription() string
}

// OOBManager manages OOB handlers and processing
type OOBManager struct {
	handlers map[string]OOBHandler
	verbose  bool
	logger   *log.Logger
}

// NewOOBManager creates a new OOB manager
func NewOOBManager(verbose bool, logger *log.Logger) *OOBManager {
	return &OOBManager{
		handlers: make(map[string]OOBHandler),
		verbose:  verbose,
		logger:   logger,
	}
}

// RegisterHandler registers a new OOB handler
func (om *OOBManager) RegisterHandler(handler OOBHandler) {
	om.handlers[handler.GetName()] = handler
	if om.verbose {
		om.logger.Printf("Registered OOB handler: %s - %s", handler.GetName(), handler.GetDescription())
	}
}

// ProcessOOB processes an OOB message by finding the appropriate handler
func (om *OOBManager) ProcessOOB(message string, session *ChatSession) (string, error) {
	if om.verbose {
		om.logger.Printf("Processing OOB message: %s", message)
	}

	// Try each handler to see if it can handle the message
	for name, handler := range om.handlers {
		if handler.CanHandle(message) {
			if om.verbose {
				om.logger.Printf("Using OOB handler: %s", name)
			}
			return handler.Process(message, session)
		}
	}

	return "", fmt.Errorf("no OOB handler found for message: %s", message)
}

// ListHandlers returns a list of registered OOB handlers
func (om *OOBManager) ListHandlers() []string {
	var names []string
	for name := range om.handlers {
		names = append(names, name)
	}
	return names
}

// GetHandler returns a specific handler by name
func (om *OOBManager) GetHandler(name string) (OOBHandler, bool) {
	handler, exists := om.handlers[name]
	return handler, exists
}

// Built-in OOB Handlers

// SystemInfoHandler handles system information requests
type SystemInfoHandler struct{}

func (h *SystemInfoHandler) CanHandle(message string) bool {
	return strings.HasPrefix(strings.ToUpper(message), "SYSTEM INFO") ||
		strings.HasPrefix(strings.ToUpper(message), "SYSTEMINFO")
}

func (h *SystemInfoHandler) Process(message string, session *ChatSession) (string, error) {
	// Extract specific info request if any
	parts := strings.Fields(strings.ToUpper(message))
	if len(parts) > 2 {
		infoType := strings.Join(parts[2:], " ")
		switch infoType {
		case "VERSION":
			return "Golem v1.0.0", nil
		case "STATUS":
			return "Running", nil
		case "HANDLERS":
			return "Available OOB handlers: system_info, session_info, properties", nil
		default:
			return fmt.Sprintf("Unknown system info request: %s", infoType), nil
		}
	}
	return "System Info: Golem v1.0.0, Status: Running", nil
}

func (h *SystemInfoHandler) GetName() string {
	return "system_info"
}

func (h *SystemInfoHandler) GetDescription() string {
	return "Handles system information requests (version, status, handlers)"
}

// SessionInfoHandler handles session information requests
type SessionInfoHandler struct{}

func (h *SessionInfoHandler) CanHandle(message string) bool {
	return strings.HasPrefix(strings.ToUpper(message), "SESSION INFO") ||
		strings.HasPrefix(strings.ToUpper(message), "SESSIONINFO")
}

func (h *SessionInfoHandler) Process(message string, session *ChatSession) (string, error) {
	if session == nil {
		return "No active session", nil
	}

	info := fmt.Sprintf("Session ID: %s, Messages: %d, Variables: %d",
		session.ID, len(session.History), len(session.Variables))

	// Add variable details if requested
	if strings.Contains(strings.ToUpper(message), "DETAILS") {
		varDetails := make([]string, 0, len(session.Variables))
		for k, v := range session.Variables {
			varDetails = append(varDetails, fmt.Sprintf("%s=%s", k, v))
		}
		if len(varDetails) > 0 {
			info += fmt.Sprintf(", Variables: [%s]", strings.Join(varDetails, ", "))
		}
	}

	return info, nil
}

func (h *SessionInfoHandler) GetName() string {
	return "session_info"
}

func (h *SessionInfoHandler) GetDescription() string {
	return "Handles session information requests"
}

// PropertiesHandler handles property-related OOB requests
type PropertiesHandler struct {
	aimlKB *AIMLKnowledgeBase
}

func (h *PropertiesHandler) CanHandle(message string) bool {
	return strings.HasPrefix(strings.ToUpper(message), "PROPERTIES") ||
		strings.HasPrefix(strings.ToUpper(message), "GET PROPERTY") ||
		strings.HasPrefix(strings.ToUpper(message), "SET PROPERTY")
}

func (h *PropertiesHandler) Process(message string, session *ChatSession) (string, error) {
	if h.aimlKB == nil {
		return "No knowledge base loaded", nil
	}

	parts := strings.Fields(strings.ToUpper(message))
	if len(parts) < 2 {
		return "Usage: PROPERTIES [GET|SET] [key] [value]", nil
	}

	switch parts[1] {
	case "GET":
		if len(parts) < 3 {
			// List all properties
			var props []string
			for k, v := range h.aimlKB.Properties {
				props = append(props, fmt.Sprintf("%s=%s", k, v))
			}
			return fmt.Sprintf("Properties: [%s]", strings.Join(props, ", ")), nil
		}
		key := strings.ToLower(parts[2]) // Convert to lowercase to match property keys
		value := h.aimlKB.GetProperty(key)
		if value == "" {
			return fmt.Sprintf("%s=undefined", key), nil
		}
		return fmt.Sprintf("%s=%s", key, value), nil

	case "SET":
		if len(parts) < 4 {
			return "Usage: PROPERTIES SET <key> <value>", nil
		}
		key := strings.ToLower(parts[2]) // Convert to lowercase to match property keys
		value := strings.Join(parts[3:], " ")
		h.aimlKB.SetProperty(key, value)
		return fmt.Sprintf("Set %s=%s", key, value), nil

	default:
		return "Usage: PROPERTIES [GET|SET] [key] [value]", nil
	}
}

func (h *PropertiesHandler) GetName() string {
	return "properties"
}

func (h *PropertiesHandler) GetDescription() string {
	return "Handles property get/set operations"
}

// OOBMessage represents a parsed OOB message
type OOBMessage struct {
	Type    string
	Content string
	Raw     string
}

// ParseOOBMessage parses an OOB message from input
func ParseOOBMessage(input string) (*OOBMessage, bool) {
	// Look for OOB markers: <oob>...</oob> or [OOB]...[/OOB]
	oobRegex := regexp.MustCompile(`(?i)<oob>(.*?)</oob>|\[OOB\](.*?)\[/OOB\]`)
	matches := oobRegex.FindStringSubmatch(input)

	if len(matches) == 0 {
		return nil, false
	}

	// Extract content from either match group
	content := matches[1]
	if content == "" {
		content = matches[2]
	}

	// Parse type and content
	parts := strings.Fields(strings.TrimSpace(content))
	if len(parts) == 0 {
		return nil, false
	}

	contentStr := ""
	if len(parts) > 1 {
		contentStr = strings.Join(parts[1:], " ")
	}

	return &OOBMessage{
		Type:    strings.ToUpper(parts[0]),
		Content: contentStr,
		Raw:     strings.TrimSpace(content),
	}, true
}
