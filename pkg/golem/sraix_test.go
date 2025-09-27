package golem

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestSRAIXManager tests the SRAIX manager functionality
func TestSRAIXManager(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	sm := NewSRAIXManager(logger, true)

	// Test adding a configuration
	config := &SRAIXConfig{
		Name:             "test_service",
		BaseURL:          "https://api.example.com/chat",
		Method:           "POST",
		Headers:          map[string]string{"Authorization": "Bearer token123"},
		Timeout:          10,
		ResponseFormat:   "json",
		ResponsePath:     "data.message",
		FallbackResponse: "Sorry, the service is unavailable.",
		IncludeWildcards: true,
	}

	err := sm.AddConfig(config)
	if err != nil {
		t.Errorf("Failed to add SRAIX config: %v", err)
	}

	// Test retrieving configuration
	retrievedConfig, exists := sm.GetConfig("test_service")
	if !exists {
		t.Error("Expected to find test_service config")
	}
	if retrievedConfig.Name != "test_service" {
		t.Errorf("Expected name 'test_service', got '%s'", retrievedConfig.Name)
	}

	// Test listing configurations
	configs := sm.ListConfigs()
	if len(configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(configs))
	}
}

// TestSRAIXProcessing tests SRAIX processing with a mock HTTP server
func TestSRAIXProcessing(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected application/json content type, got %s", contentType)
		}

		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer token123" {
			t.Errorf("Expected 'Bearer token123' authorization, got '%s'", auth)
		}

		// Parse request body
		var requestData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify input field
		input, ok := requestData["input"].(string)
		if !ok {
			t.Error("Expected 'input' field in request body")
		}
		if input != "Hello, world!" {
			t.Errorf("Expected input 'Hello, world!', got '%s'", input)
		}

		// Send JSON response
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"message": "Hello from external service!",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create SRAIX manager and add configuration
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	sm := NewSRAIXManager(logger, true)

	config := &SRAIXConfig{
		Name:             "test_service",
		BaseURL:          server.URL,
		Method:           "POST",
		Headers:          map[string]string{"Authorization": "Bearer token123"},
		Timeout:          5,
		ResponseFormat:   "json",
		ResponsePath:     "data.message",
		FallbackResponse: "Service unavailable",
		IncludeWildcards: false,
	}

	err := sm.AddConfig(config)
	if err != nil {
		t.Fatalf("Failed to add SRAIX config: %v", err)
	}

	// Test SRAIX processing
	response, err := sm.ProcessSRAIX("test_service", "Hello, world!", make(map[string]string))
	if err != nil {
		t.Errorf("SRAIX processing failed: %v", err)
	}

	expected := "Hello from external service!"
	if response != expected {
		t.Errorf("Expected response '%s', got '%s'", expected, response)
	}
}

// TestSRAIXErrorHandling tests error handling in SRAIX processing
func TestSRAIXErrorHandling(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	sm := NewSRAIXManager(logger, true)

	// Test with non-existent service
	_, err := sm.ProcessSRAIX("nonexistent", "test", make(map[string]string))
	if err == nil {
		t.Error("Expected error for non-existent service")
	}

	// Test with server returning error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	config := &SRAIXConfig{
		Name:             "error_service",
		BaseURL:          server.URL,
		Method:           "POST",
		Timeout:          5,
		FallbackResponse: "Service error occurred",
	}

	err = sm.AddConfig(config)
	if err != nil {
		t.Fatalf("Failed to add SRAIX config: %v", err)
	}

	// Test with fallback response
	response, err := sm.ProcessSRAIX("error_service", "test", make(map[string]string))
	if err != nil {
		t.Errorf("Expected fallback response, got error: %v", err)
	}
	if response != "Service error occurred" {
		t.Errorf("Expected fallback response, got '%s'", response)
	}
}

// TestSRAIXTimeout tests timeout handling
func TestSRAIXTimeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than timeout
		w.Write([]byte("Response"))
	}))
	defer server.Close()

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	sm := NewSRAIXManager(logger, true)

	config := &SRAIXConfig{
		Name:             "slow_service",
		BaseURL:          server.URL,
		Method:           "POST",
		Timeout:          1, // 1 second timeout
		FallbackResponse: "Request timeout",
	}

	err := sm.AddConfig(config)
	if err != nil {
		t.Fatalf("Failed to add SRAIX config: %v", err)
	}

	// Test timeout handling
	response, err := sm.ProcessSRAIX("slow_service", "test", make(map[string]string))
	if err != nil {
		t.Errorf("Expected fallback response, got error: %v", err)
	}
	if response != "Request timeout" {
		t.Errorf("Expected fallback response, got '%s'", response)
	}
}

// TestSRAIXWithWildcards tests SRAIX with wildcard inclusion
func TestSRAIXWithWildcards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestData)

		// Verify wildcards are included
		wildcards, ok := requestData["wildcards"].(map[string]interface{})
		if !ok {
			t.Error("Expected 'wildcards' field in request body")
		}

		star1, ok := wildcards["star1"].(string)
		if !ok || star1 != "world" {
			t.Errorf("Expected wildcard 'star1' to be 'world', got '%v'", wildcards["star1"])
		}

		response := map[string]interface{}{
			"message": "Hello world from external service!",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	sm := NewSRAIXManager(logger, true)

	config := &SRAIXConfig{
		Name:             "wildcard_service",
		BaseURL:          server.URL,
		Method:           "POST",
		Timeout:          5,
		ResponseFormat:   "json",
		ResponsePath:     "message",
		IncludeWildcards: true,
	}

	err := sm.AddConfig(config)
	if err != nil {
		t.Fatalf("Failed to add SRAIX config: %v", err)
	}

	// Test with wildcards
	wildcards := map[string]string{
		"star1": "world",
	}

	response, err := sm.ProcessSRAIX("wildcard_service", "Hello *", wildcards)
	if err != nil {
		t.Errorf("SRAIX processing failed: %v", err)
	}

	expected := "Hello world from external service!"
	if response != expected {
		t.Errorf("Expected response '%s', got '%s'", expected, response)
	}
}

// TestSRAIXIntegration tests SRAIX integration with the main Golem system
func TestSRAIXIntegration(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var requestData map[string]interface{}
		json.NewDecoder(r.Body).Decode(&requestData)

		input := requestData["input"].(string)
		response := map[string]interface{}{
			"message": "External response to: " + input,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create Golem instance
	g := New(false)

	// Add SRAIX configuration
	config := &SRAIXConfig{
		Name:           "external_chat",
		BaseURL:        server.URL,
		Method:         "POST",
		Timeout:        5,
		ResponseFormat: "json",
		ResponsePath:   "message",
	}

	err := g.AddSRAIXConfig(config)
	if err != nil {
		t.Fatalf("Failed to add SRAIX config: %v", err)
	}

	// Create AIML knowledge base with SRAIX
	kb := NewAIMLKnowledgeBase()
	kb.Categories = []Category{
		{
			Pattern:  "ASK EXTERNAL *",
			Template: "I'll ask the external service: <sraix service=\"external_chat\">ASK EXTERNAL <star/></sraix>",
		},
	}

	g.SetKnowledgeBase(kb)

	// Test SRAIX processing by directly processing the template
	template := "I'll ask the external service: <sraix service=\"external_chat\">ASK EXTERNAL hello</sraix>"
	response := g.ProcessTemplate(template, make(map[string]string))
	if !strings.Contains(response, "External response to:") {
		t.Errorf("Expected SRAIX response, got: %s", response)
	}
}

// TestSRAIXConfigLoading tests loading SRAIX configurations from files
func TestSRAIXConfigLoading(t *testing.T) {
	// Create temporary config file
	configData := []*SRAIXConfig{
		{
			Name:           "service1",
			BaseURL:        "https://api1.example.com",
			Method:         "POST",
			ResponseFormat: "json",
		},
		{
			Name:           "service2",
			BaseURL:        "https://api2.example.com",
			Method:         "GET",
			ResponseFormat: "text",
		},
	}

	// Write config to temporary file
	configJSON, _ := json.Marshal(configData)
	tempFile := createTempFile(t, "sraix_config.json", string(configJSON))
	defer deleteTempFile(t, tempFile)

	// Test loading from file
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	sm := NewSRAIXManager(logger, true)

	err := sm.LoadSRAIXConfigsFromFile(tempFile)
	if err != nil {
		t.Errorf("Failed to load SRAIX configs: %v", err)
	}

	// Verify configurations were loaded
	configs := sm.ListConfigs()
	if len(configs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(configs))
	}

	// Verify specific configs
	service1, exists := sm.GetConfig("service1")
	if !exists {
		t.Error("Expected service1 config")
	}
	if service1.BaseURL != "https://api1.example.com" {
		t.Errorf("Expected service1 URL 'https://api1.example.com', got '%s'", service1.BaseURL)
	}
}

// TestSRAIXWithDifferentFormats tests different response formats
func TestSRAIXWithDifferentFormats(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	sm := NewSRAIXManager(logger, true)

	// Test JSON response with path extraction
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"result": "JSON response",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &SRAIXConfig{
		Name:           "json_service",
		BaseURL:        server.URL,
		Method:         "POST",
		Timeout:        5,
		ResponseFormat: "json",
		ResponsePath:   "data.result",
	}

	err := sm.AddConfig(config)
	if err != nil {
		t.Fatalf("Failed to add SRAIX config: %v", err)
	}

	response, err := sm.ProcessSRAIX("json_service", "test", make(map[string]string))
	if err != nil {
		t.Errorf("SRAIX processing failed: %v", err)
	}

	if response != "JSON response" {
		t.Errorf("Expected 'JSON response', got '%s'", response)
	}

	// Test plain text response
	textServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Plain text response"))
	}))
	defer textServer.Close()

	textConfig := &SRAIXConfig{
		Name:           "text_service",
		BaseURL:        textServer.URL,
		Method:         "POST",
		Timeout:        5,
		ResponseFormat: "text",
	}

	err = sm.AddConfig(textConfig)
	if err != nil {
		t.Fatalf("Failed to add text SRAIX config: %v", err)
	}

	textResponse, err := sm.ProcessSRAIX("text_service", "test", make(map[string]string))
	if err != nil {
		t.Errorf("Text SRAIX processing failed: %v", err)
	}

	if textResponse != "Plain text response" {
		t.Errorf("Expected 'Plain text response', got '%s'", textResponse)
	}
}

// Helper functions for testing

func createTempFile(t *testing.T, name, content string) string {
	file, err := os.CreateTemp("", name)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	file.Close()
	return file.Name()
}

func deleteTempFile(t *testing.T, filename string) {
	err := os.Remove(filename)
	if err != nil {
		t.Fatalf("Failed to delete temp file: %v", err)
	}
}
