package golem

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SRAIXConfig represents configuration for external SRAIX services
type SRAIXConfig struct {
	// Service name identifier
	Name string `json:"name"`
	// Base URL for the service
	BaseURL string `json:"base_url"`
	// HTTP method (GET, POST, etc.)
	Method string `json:"method"`
	// Headers to include in requests
	Headers map[string]string `json:"headers"`
	// Request timeout in seconds
	Timeout int `json:"timeout"`
	// Response format (json, xml, text)
	ResponseFormat string `json:"response_format"`
	// JSON path to extract response (for JSON responses)
	ResponsePath string `json:"response_path"`
	// Fallback response when service is unavailable
	FallbackResponse string `json:"fallback_response"`
	// Whether to include wildcards in the request
	IncludeWildcards bool `json:"include_wildcards"`
}

// SRAIXManager manages external service configurations and HTTP client
type SRAIXManager struct {
	configs map[string]*SRAIXConfig
	client  *http.Client
	logger  *log.Logger
	verbose bool
}

// NewSRAIXManager creates a new SRAIX manager
func NewSRAIXManager(logger *log.Logger, verbose bool) *SRAIXManager {
	return &SRAIXManager{
		configs: make(map[string]*SRAIXConfig),
		client: &http.Client{
			Timeout: 30 * time.Second, // Default timeout
		},
		logger:  logger,
		verbose: verbose,
	}
}

// AddConfig adds a new SRAIX service configuration
func (sm *SRAIXManager) AddConfig(config *SRAIXConfig) error {
	if config.Name == "" {
		return fmt.Errorf("SRAIX config name cannot be empty")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("SRAIX config base URL cannot be empty")
	}
	if config.Method == "" {
		config.Method = "POST" // Default to POST
	}
	if config.Timeout == 0 {
		config.Timeout = 30 // Default 30 seconds
	}
	if config.ResponseFormat == "" {
		config.ResponseFormat = "text" // Default to text
	}
	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}

	sm.configs[config.Name] = config
	if sm.verbose {
		sm.logger.Printf("Added SRAIX config: %s -> %s", config.Name, config.BaseURL)
	}
	return nil
}

// GetConfig retrieves a SRAIX service configuration
func (sm *SRAIXManager) GetConfig(name string) (*SRAIXConfig, bool) {
	config, exists := sm.configs[name]
	return config, exists
}

// ListConfigs returns all configured SRAIX services
func (sm *SRAIXManager) ListConfigs() map[string]*SRAIXConfig {
	return sm.configs
}

// ProcessSRAIX processes a SRAIX tag by making an external HTTP request
func (sm *SRAIXManager) ProcessSRAIX(serviceName, input string, wildcards map[string]string) (string, error) {
	config, exists := sm.GetConfig(serviceName)
	if !exists {
		return "", fmt.Errorf("SRAIX service '%s' not configured", serviceName)
	}

	// Prepare the request
	url := config.BaseURL
	var body io.Reader
	var contentType string

	// Build request body based on method and configuration
	if config.Method == "GET" {
		// For GET requests, append input as query parameter
		if strings.Contains(url, "?") {
			url += "&input=" + strings.ReplaceAll(input, " ", "+")
		} else {
			url += "?input=" + strings.ReplaceAll(input, " ", "+")
		}
	} else {
		// For POST/PUT requests, create JSON body
		requestData := map[string]interface{}{
			"input": input,
		}

		// Include wildcards if configured
		if config.IncludeWildcards && len(wildcards) > 0 {
			requestData["wildcards"] = wildcards
		}

		// Include additional SRAIX parameters
		if botid, exists := wildcards["botid"]; exists && botid != "" {
			requestData["botid"] = botid
		}
		if host, exists := wildcards["host"]; exists && host != "" {
			requestData["host"] = host
		}
		if hint, exists := wildcards["hint"]; exists && hint != "" {
			requestData["hint"] = hint
		}

		jsonData, err := json.Marshal(requestData)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request data: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
		contentType = "application/json"
	}

	// Create HTTP request
	req, err := http.NewRequest(config.Method, url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	// Set headers
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	// Make the request
	if sm.verbose {
		sm.logger.Printf("SRAIX request to %s: %s %s", serviceName, config.Method, url)
	}

	resp, err := sm.client.Do(req)
	if err != nil {
		if sm.verbose {
			sm.logger.Printf("SRAIX request failed: %v", err)
		}
		// Return fallback response if configured
		if config.FallbackResponse != "" {
			return config.FallbackResponse, nil
		}
		return "", fmt.Errorf("SRAIX request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		if sm.verbose {
			sm.logger.Printf("SRAIX request returned status %d: %s", resp.StatusCode, string(responseBody))
		}
		// Return fallback response if configured
		if config.FallbackResponse != "" {
			return config.FallbackResponse, nil
		}
		return "", fmt.Errorf("SRAIX request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Process response based on format
	response := string(responseBody)
	switch config.ResponseFormat {
	case "json":
		if config.ResponsePath != "" {
			// Extract specific field from JSON response
			var jsonData map[string]interface{}
			if err := json.Unmarshal(responseBody, &jsonData); err != nil {
				return "", fmt.Errorf("failed to parse JSON response: %v", err)
			}
			// Simple JSON path extraction (supports dot notation like "data.message")
			response = sm.extractJSONPath(jsonData, config.ResponsePath)
		}
	case "xml":
		// For XML, we'll return the raw response for now
		// Could be enhanced to parse XML and extract specific elements
	case "text":
		// Return raw text response
	default:
		// Default to text
	}

	if sm.verbose {
		sm.logger.Printf("SRAIX response from %s: %s", serviceName, response)
	}

	return strings.TrimSpace(response), nil
}

// extractJSONPath extracts a value from JSON data using dot notation
func (sm *SRAIXManager) extractJSONPath(data map[string]interface{}, path string) string {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part, return the value
			if val, ok := current[part]; ok {
				if str, ok := val.(string); ok {
					return str
				}
				return fmt.Sprintf("%v", val)
			}
			return ""
		}

		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}

	return ""
}

// LoadSRAIXConfigsFromFile loads SRAIX configurations from a JSON file
func (sm *SRAIXManager) LoadSRAIXConfigsFromFile(filename string) error {
	data, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read SRAIX config file: %v", err)
	}

	var configs []*SRAIXConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse SRAIX config file: %v", err)
	}

	for _, config := range configs {
		if err := sm.AddConfig(config); err != nil {
			return fmt.Errorf("failed to add SRAIX config %s: %v", config.Name, err)
		}
	}

	return nil
}

// LoadSRAIXConfigsFromDirectory loads all SRAIX configuration files from a directory
func (sm *SRAIXManager) LoadSRAIXConfigsFromDirectory(dirPath string) error {
	files, err := listFiles(dirPath, ".sraix.json")
	if err != nil {
		return fmt.Errorf("failed to list SRAIX config files: %v", err)
	}

	for _, file := range files {
		if err := sm.LoadSRAIXConfigsFromFile(file); err != nil {
			sm.logger.Printf("Warning: Failed to load SRAIX config file %s: %v", file, err)
		}
	}

	return nil
}

// readFile reads the contents of a file
func readFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// listFiles lists files in a directory with a specific extension
func listFiles(dirPath, extension string) ([]string, error) {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), strings.ToLower(extension)) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}
