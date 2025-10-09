package golem

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace
)

// ContextConfig represents configuration for context management
type ContextConfig struct {
	MaxThatDepth         int     // Maximum depth for that history (default: 20)
	MaxRequestDepth      int     // Maximum depth for request history (default: 20)
	MaxResponseDepth     int     // Maximum depth for response history (default: 20)
	MaxTotalContext      int     // Maximum total context items (default: 100)
	CompressionThreshold int     // Threshold for context compression (default: 50)
	WeightDecay          float64 // Weight decay factor for older context (default: 0.9)
	EnableCompression    bool    // Enable context compression (default: true)
	EnableAnalytics      bool    // Enable context analytics (default: true)
	EnablePruning        bool    // Enable smart context pruning (default: true)
}

// ContextItem represents a single context item with metadata
type ContextItem struct {
	Content    string                 `json:"content"`
	Type       string                 `json:"type"` // "that", "request", "response"
	Index      int                    `json:"index"`
	Weight     float64                `json:"weight"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  string                 `json:"created_at"`
	LastUsed   string                 `json:"last_used"`
	UsageCount int                    `json:"usage_count"`
}

// ContextAnalytics represents analytics data for context usage
type ContextAnalytics struct {
	TotalItems       int            `json:"total_items"`
	ThatItems        int            `json:"that_items"`
	RequestItems     int            `json:"request_items"`
	ResponseItems    int            `json:"response_items"`
	AverageWeight    float64        `json:"average_weight"`
	MostUsedItems    []string       `json:"most_used_items"`
	LeastUsedItems   []string       `json:"least_used_items"`
	TagDistribution  map[string]int `json:"tag_distribution"`
	MemoryUsage      int            `json:"memory_usage_bytes"`
	CompressionRatio float64        `json:"compression_ratio"`
	LastPruned       string         `json:"last_pruned"`
	PruningCount     int            `json:"pruning_count"`
}

// TemplateProcessingMetrics represents metrics for template processing
type TemplateProcessingMetrics struct {
	TotalProcessed     int                `json:"total_processed"`
	AverageProcessTime float64            `json:"average_process_time_ms"`
	CacheHits          int                `json:"cache_hits"`
	CacheMisses        int                `json:"cache_misses"`
	CacheHitRate       float64            `json:"cache_hit_rate"`
	TagProcessingTimes map[string]float64 `json:"tag_processing_times"`
	ErrorCount         int                `json:"error_count"`
	LastProcessed      string             `json:"last_processed"`
	MemoryPeak         int                `json:"memory_peak_bytes"`
	ParallelOps        int                `json:"parallel_operations"`
}

// TemplateCache represents a cache for processed templates
type TemplateCache struct {
	Cache      map[string]string `json:"cache"`
	Timestamps map[string]string `json:"timestamps"`
	Hits       map[string]int    `json:"hits"`
	MaxSize    int               `json:"max_size"`
	TTL        int64             `json:"ttl_seconds"`
}

// TemplateProcessingConfig represents configuration for template processing
type TemplateProcessingConfig struct {
	EnableCaching     bool  `json:"enable_caching"`
	CacheSize         int   `json:"cache_size"`
	CacheTTL          int64 `json:"cache_ttl_seconds"`
	EnableParallel    bool  `json:"enable_parallel"`
	MaxParallelOps    int   `json:"max_parallel_operations"`
	EnableMetrics     bool  `json:"enable_metrics"`
	EnableValidation  bool  `json:"enable_validation"`
	EnableDebugging   bool  `json:"enable_debugging"`
	MemoryLimit       int   `json:"memory_limit_bytes"`
	ProcessingTimeout int64 `json:"processing_timeout_ms"`
}

// ChatSession represents a single chat session
type ChatSession struct {
	ID              string
	Variables       map[string]string
	History         []string
	CreatedAt       string
	LastActivity    string
	Topic           string   // Current conversation topic
	ThatHistory     []string // History of bot responses for that matching
	RequestHistory  []string // History of user requests for <request> tag
	ResponseHistory []string // History of bot responses for <response> tag

	// Enhanced context management
	ContextConfig   *ContextConfig         // Context configuration
	ContextWeights  map[string]float64     // Weights for different context levels
	ContextUsage    map[string]int         // Usage count for each context item
	ContextTags     map[string][]string    // Tags for context categorization
	ContextMetadata map[string]interface{} // Additional context metadata
}

// Golem represents the main library instance
//
// CRITICAL ARCHITECTURAL NOTE:
// This struct maintains state across multiple operations:
// - aimlKB: Loaded AIML knowledge base (persists across commands)
// - sessions: Active chat sessions (persists across commands)
// - currentID: Currently active session (persists across commands)
//
// CLI USAGE PATTERN:
// - Single command mode: Creates new instance per command (state lost)
// - Interactive mode: Single persistent instance (state preserved)
// - Library mode: User manages instance lifecycle (state controlled by user)
//
// DO NOT modify the state management without understanding the implications
// for all three usage patterns.
type Golem struct {
	verbose   bool
	logLevel  LogLevel
	logger    *log.Logger
	aimlKB    *AIMLKnowledgeBase
	sessions  map[string]*ChatSession
	currentID string
	sessionID int
	oobMgr    *OOBManager
	sraixMgr  *SRAIXManager
	// Text processing components
	sentenceSplitter     *SentenceSplitter
	wordBoundaryDetector *WordBoundaryDetector
	// Template processing components
	templateCache   *TemplateCache
	templateConfig  *TemplateProcessingConfig
	templateMetrics *TemplateProcessingMetrics
	// Persistent learning components
	persistentLearning *PersistentLearningManager
}

// New creates a new Golem instance
func New(verbose bool) *Golem {
	logger := log.New(os.Stdout, "[GOLEM] ", log.LstdFlags)

	// Set log level based on verbose flag
	// When verbose is enabled, show Info level and above (Info, Warn, Error)
	// When verbose is disabled, show only Error level
	logLevel := LogLevelError
	if verbose {
		logLevel = LogLevelInfo
	}

	// Create OOB manager and register built-in handlers
	oobMgr := NewOOBManager(verbose, logger)

	// Register built-in OOB handlers
	oobMgr.RegisterHandler(&SystemInfoHandler{})
	oobMgr.RegisterHandler(&SessionInfoHandler{})

	// Properties handler will be registered when AIML is loaded
	// since it needs access to the knowledge base

	// Create SRAIX manager
	sraixMgr := NewSRAIXManager(logger, verbose)

	// Create text processing components
	sentenceSplitter := NewSentenceSplitter()
	wordBoundaryDetector := NewWordBoundaryDetector()

	// Create template processing components
	templateCache := &TemplateCache{
		Cache:      make(map[string]string),
		Timestamps: make(map[string]string),
		Hits:       make(map[string]int),
		MaxSize:    1000,
		TTL:        3600, // 1 hour
	}

	templateConfig := &TemplateProcessingConfig{
		EnableCaching:     true,
		CacheSize:         1000,
		CacheTTL:          3600,
		EnableParallel:    true,
		MaxParallelOps:    4,
		EnableMetrics:     true,
		EnableValidation:  true,
		EnableDebugging:   verbose,
		MemoryLimit:       50 * 1024 * 1024, // 50MB
		ProcessingTimeout: 5000,             // 5 seconds
	}

	templateMetrics := &TemplateProcessingMetrics{
		TotalProcessed:     0,
		AverageProcessTime: 0.0,
		CacheHits:          0,
		CacheMisses:        0,
		CacheHitRate:       0.0,
		TagProcessingTimes: make(map[string]float64),
		ErrorCount:         0,
		LastProcessed:      "",
		MemoryPeak:         0,
		ParallelOps:        0,
	}

	// Create persistent learning manager with default storage path
	persistentLearning := NewPersistentLearningManager("./learned_categories")

	return &Golem{
		verbose:              verbose,
		logLevel:             logLevel,
		logger:               logger,
		sessions:             make(map[string]*ChatSession),
		sessionID:            1,
		oobMgr:               oobMgr,
		sraixMgr:             sraixMgr,
		sentenceSplitter:     sentenceSplitter,
		wordBoundaryDetector: wordBoundaryDetector,
		templateCache:        templateCache,
		templateConfig:       templateConfig,
		templateMetrics:      templateMetrics,
		persistentLearning:   persistentLearning,
	}
}

// LogError logs an error message
func (g *Golem) LogError(format string, args ...interface{}) {
	if g.logLevel >= LogLevelError {
		g.logger.Printf("[ERROR] "+format, args...)
	}
}

// LogWarn logs a warning message
func (g *Golem) LogWarn(format string, args ...interface{}) {
	if g.logLevel >= LogLevelWarn {
		g.logger.Printf("[WARN] "+format, args...)
	}
}

// LogInfo logs an info message
func (g *Golem) LogInfo(format string, args ...interface{}) {
	if g.logLevel >= LogLevelInfo {
		g.logger.Printf("[INFO] "+format, args...)
	}
}

// LogDebug logs a debug message
func (g *Golem) LogDebug(format string, args ...interface{}) {
	if g.logLevel >= LogLevelDebug {
		g.logger.Printf("[DEBUG] "+format, args...)
	}
}

// LogTrace logs a trace message
func (g *Golem) LogTrace(format string, args ...interface{}) {
	if g.logLevel >= LogLevelTrace {
		g.logger.Printf("[TRACE] "+format, args...)
	}
}

// SetPersistentLearningPath sets the path for persistent learning storage
func (g *Golem) SetPersistentLearningPath(path string) {
	if g.persistentLearning != nil {
		g.persistentLearning.SetStoragePath(path)
	}
}

// GetPersistentLearningInfo returns information about persistent learning
func (g *Golem) GetPersistentLearningInfo() (map[string]interface{}, error) {
	if g.persistentLearning == nil {
		return nil, fmt.Errorf("persistent learning not initialized")
	}
	return g.persistentLearning.GetPersistentCategoryInfo()
}

// LoadPersistentCategories loads categories from persistent storage
func (g *Golem) LoadPersistentCategories() error {
	if g.persistentLearning == nil {
		return fmt.Errorf("persistent learning not initialized")
	}

	if g.aimlKB == nil {
		return fmt.Errorf("no knowledge base available")
	}

	categories, err := g.persistentLearning.LoadPersistentCategories()
	if err != nil {
		return fmt.Errorf("failed to load persistent categories: %v", err)
	}

	// Add categories to the knowledge base
	for _, category := range categories {
		normalizedPattern := NormalizePattern(category.Pattern)

		// Check if category already exists
		if existingCategory, exists := g.aimlKB.Patterns[normalizedPattern]; exists {
			// Update existing category
			*existingCategory = category
		} else {
			// Add new category
			g.aimlKB.Categories = append(g.aimlKB.Categories, category)
			g.aimlKB.Patterns[normalizedPattern] = &g.aimlKB.Categories[len(g.aimlKB.Categories)-1]
		}
	}

	g.LogInfo("Loaded %d persistent categories", len(categories))
	return nil
}

// SavePersistentCategories saves all current categories to persistent storage
func (g *Golem) SavePersistentCategories(source string) error {
	if g.persistentLearning == nil {
		return fmt.Errorf("persistent learning not initialized")
	}

	if g.aimlKB == nil {
		return fmt.Errorf("no knowledge base available")
	}

	return g.persistentLearning.SavePersistentCategories(g.aimlKB.Categories, source)
}

// SetLogLevel sets the logging level
func (g *Golem) SetLogLevel(level LogLevel) {
	g.logLevel = level
}

// GetLogLevel returns the current logging level
func (g *Golem) GetLogLevel() LogLevel {
	return g.logLevel
}

// LogVerbose logs a message only if verbose mode is enabled (for backward compatibility)
// This is a convenience function that maps to LogDebug for backward compatibility
func (g *Golem) LogVerbose(format string, args ...interface{}) {
	if g.verbose {
		g.LogDebug(format, args...)
	}
}

/*
Logging Usage Examples:

Replace verbose logging patterns like this:

OLD PATTERN:
	if g.verbose {
		g.logger.Printf("Loading AIML from string")
	}

NEW PATTERN (using level-based logging):
	g.LogDebug("Loading AIML from string")

OLD PATTERN:
	if g.verbose {
		g.logger.Printf("Total categories: %d", len(g.aimlKB.Categories))
	}

NEW PATTERN:
	g.LogDebug("Total categories: %d", len(g.aimlKB.Categories))

OLD PATTERN:
	if g.verbose {
		g.logger.Printf("Failed to parse learnf content: %v", err)
	}

NEW PATTERN (for errors):
	g.LogError("Failed to parse learnf content: %v", err)

Available log levels:
- LogError: Error messages (always shown)
- LogWarn: Warning messages (shown when verbose enabled)
- LogInfo: Informational messages (shown when verbose enabled)
- LogDebug: Debug messages (shown when log level set to Debug or Trace)
- LogTrace: Very detailed trace messages (shown when log level set to Trace)

Verbose flag behavior:
- --verbose: Shows Info, Warn, and Error messages
- No --verbose: Shows only Error messages

Set log level manually:
	g.SetLogLevel(LogLevelDebug)
*/

// Execute runs the specified command with arguments
//
// IMPORTANT: This method operates on the current Golem instance state.
// - In CLI single-command mode: New instance created per command (state lost)
// - In CLI interactive mode: Same instance used across commands (state preserved)
// - In library mode: User controls instance lifecycle (state managed by user)
//
// Commands that modify state (load, session create/switch, properties set):
// - Will persist in interactive mode and library mode
// - Will be lost in single-command mode
func (g *Golem) Execute(command string, args []string) error {
	if g.verbose {
		g.logger.Printf("Executing command: %s with args: %v", command, args)
	}

	switch command {
	case "load":
		return g.loadCommand(args)
	case "chat":
		return g.chatCommand(args)
	case "session":
		return g.sessionCommand(args)
	case "properties":
		return g.propertiesCommand(args)
	case "oob":
		return g.oobCommand(args)
	case "sraix":
		return g.sraixCommand(args)
	case "process":
		return g.processCommand(args)
	case "analyze":
		return g.analyzeCommand(args)
	case "generate":
		return g.generateCommand(args)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// LoadCommand handles the load command
// loadAllRelatedFiles loads all .aiml, .map, and .set files from the same directory as the given file
func (g *Golem) loadAllRelatedFiles(filePath string) error {
	dir := filepath.Dir(filePath)

	if g.verbose {
		g.logger.Printf("Loading all related files from directory: %s", dir)
	}

	// Load AIML files from directory
	aimlKB, err := g.LoadAIMLFromDirectory(dir)
	if err != nil {
		// If no AIML files found, create an empty knowledge base
		if strings.Contains(err.Error(), "no AIML files found") {
			aimlKB = NewAIMLKnowledgeBase()
			// Load default properties
			err = g.loadDefaultProperties(aimlKB)
			if err != nil {
				return fmt.Errorf("failed to load default properties: %v", err)
			}
		} else {
			return fmt.Errorf("failed to load AIML files from directory: %v", err)
		}
	}

	// Load maps from directory
	maps, err := g.LoadMapsFromDirectory(dir)
	if err != nil {
		return fmt.Errorf("failed to load map files from directory: %v", err)
	}

	// Load sets from directory
	sets, err := g.LoadSetsFromDirectory(dir)
	if err != nil {
		return fmt.Errorf("failed to load set files from directory: %v", err)
	}

	// Merge maps into knowledge base
	for mapName, mapData := range maps {
		aimlKB.Maps[mapName] = mapData
	}

	// Merge sets into knowledge base
	for setName, setMembers := range sets {
		aimlKB.AddSetMembers(setName, setMembers)
	}

	// Set the knowledge base
	g.aimlKB = aimlKB

	// Print summary
	fmt.Printf("Successfully loaded all related files from directory: %s\n", dir)
	fmt.Printf("Loaded %d categories\n", len(aimlKB.Categories))
	fmt.Printf("Loaded %d maps\n", len(maps))
	fmt.Printf("Loaded %d sets\n", len(sets))

	return nil
}

func (g *Golem) loadCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("load command requires a filename or directory path")
	}

	path := args[0]
	if g.verbose {
		g.logger.Printf("Loading: %s", path)
	}

	// Check if path exists and get absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %v", path, err)
	}

	// Check if path exists
	fileInfo, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Check if it's a directory
	if fileInfo.IsDir() {
		// Load all AIML files from directory
		kb, err := g.LoadAIMLFromDirectory(absPath)
		if err != nil {
			return fmt.Errorf("failed to load AIML files from directory: %v", err)
		}
		g.aimlKB = kb
		fmt.Printf("Successfully loaded AIML files from directory: %s\n", absPath)
		fmt.Printf("Loaded %d categories\n", len(kb.Categories))
	} else if strings.HasSuffix(strings.ToLower(absPath), ".aiml") {
		// Load single AIML file and all related files from the same directory
		err := g.loadAllRelatedFiles(absPath)
		if err != nil {
			return fmt.Errorf("failed to load AIML file and related files: %v", err)
		}
	} else if strings.HasSuffix(strings.ToLower(absPath), ".map") {
		// Load single map file and all related files from the same directory
		err := g.loadAllRelatedFiles(absPath)
		if err != nil {
			return fmt.Errorf("failed to load map file and related files: %v", err)
		}
	} else if strings.HasSuffix(strings.ToLower(absPath), ".set") {
		// Load single set file and all related files from the same directory
		err := g.loadAllRelatedFiles(absPath)
		if err != nil {
			return fmt.Errorf("failed to load set file and related files: %v", err)
		}
	} else {
		// Read file contents (non-AIML file)
		content, err := g.LoadFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to load file %s: %v", absPath, err)
		}

		fmt.Printf("Successfully loaded file: %s\n", absPath)
		fmt.Printf("File size: %d bytes\n", len(content))

		if g.verbose {
			// Show first 200 characters of content
			preview := content
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			fmt.Printf("Content preview: %s\n", preview)
		}
	}

	return nil
}

// ChatCommand handles the chat command
func (g *Golem) chatCommand(args []string) error {
	if g.aimlKB == nil {
		return fmt.Errorf("no AIML knowledge base loaded. Use 'load' command first")
	}

	if len(args) == 0 {
		return fmt.Errorf("chat command requires input text")
	}

	// Get or create current session
	session := g.getCurrentSession()
	if session == nil {
		session = g.createSession("")
	}

	input := strings.Join(args, " ")
	if g.verbose {
		g.logger.Printf("Processing chat input in session %s: %s", session.ID, input)
	}

	// Check for OOB messages first
	if oobMsg, isOOB := ParseOOBMessage(input); isOOB {
		response, err := g.oobMgr.ProcessOOB(oobMsg.Raw, session)
		if err != nil {
			fmt.Printf("OOB Error: %v\n", err)
			session.History = append(session.History, "User: "+input)
			session.History = append(session.History, "Golem: OOB Error: "+err.Error())
			return nil
		}
		fmt.Printf("OOB: %s\n", response)
		session.History = append(session.History, "User: "+input)
		session.History = append(session.History, "Golem: OOB: "+response)
		return nil
	}

	// Add to history
	session.History = append(session.History, "User: "+input)

	// Add to request history for <request> tag support
	session.AddToRequestHistory(input)

	// Match pattern and get response
	category, wildcards, err := g.aimlKB.MatchPattern(input)
	if err != nil {
		response := g.aimlKB.GetProperty("default_response")
		if response == "" {
			response = "I don't understand: " + input
		}
		fmt.Printf("Golem: %s\n", response)
		session.History = append(session.History, "Golem: "+response)
		return nil
	}

	// Process template with session context
	response := g.ProcessTemplateWithSession(category.Template, wildcards, session)
	fmt.Printf("Golem: %s\n", response)
	session.History = append(session.History, "Golem: "+response)

	// Add to response history for <response> tag support
	session.AddToResponseHistory(response)

	return nil
}

// PropertiesCommand handles the properties command
func (g *Golem) propertiesCommand(args []string) error {
	if g.aimlKB == nil {
		return fmt.Errorf("no AIML knowledge base loaded. Use 'load' command first")
	}

	if len(args) == 0 {
		// Show all properties
		fmt.Println("Bot Properties:")
		fmt.Println(strings.Repeat("=", 50))
		for key, value := range g.aimlKB.Properties {
			fmt.Printf("%-20s: %s\n", key, value)
		}
		return nil
	}

	if len(args) == 1 {
		// Show specific property
		key := args[0]
		value := g.aimlKB.GetProperty(key)
		if value == "" {
			fmt.Printf("Property '%s' not found\n", key)
		} else {
			fmt.Printf("%s: %s\n", key, value)
		}
		return nil
	}

	if len(args) == 2 {
		// Set property
		key := args[0]
		value := args[1]
		g.aimlKB.SetProperty(key, value)
		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	}

	return fmt.Errorf("usage: properties [key] [value]")
}

// ProcessCommand handles the process command
func (g *Golem) processCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("process command requires input file")
	}

	inputFile := args[0]
	if g.verbose {
		g.logger.Printf("Processing file: %s", inputFile)
	}

	// TODO: Implement actual processing logic
	fmt.Printf("Processing file: %s\n", inputFile)
	return nil
}

// AnalyzeCommand handles the analyze command
func (g *Golem) analyzeCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("analyze command requires input file")
	}

	inputFile := args[0]
	if g.verbose {
		g.logger.Printf("Analyzing file: %s", inputFile)
	}

	// TODO: Implement actual analysis logic
	fmt.Printf("Analyzing file: %s\n", inputFile)
	return nil
}

// GenerateCommand handles the generate command
func (g *Golem) generateCommand(args []string) error {
	outputFile := "output.txt"

	// Parse optional output file argument
	if len(args) > 0 && args[0] == "--output" && len(args) > 1 {
		outputFile = args[1]
	}

	if g.verbose {
		g.logger.Printf("Generating output to: %s", outputFile)
	}

	// TODO: Implement actual generation logic
	fmt.Printf("Generating output to: %s\n", outputFile)
	return nil
}

// ProcessData is a library function that can be used by other programs
func (g *Golem) ProcessData(input string) (string, error) {
	if g.verbose {
		g.logger.Printf("Processing data: %s", input)
	}

	// TODO: Implement actual data processing logic
	result := fmt.Sprintf("Processed: %s", input)
	return result, nil
}

// ProcessInput processes user input with full context support
func (g *Golem) ProcessInput(input string, session *ChatSession) (string, error) {
	if g.aimlKB == nil {
		return "", fmt.Errorf("no AIML knowledge base loaded")
	}

	if g.verbose {
		g.logger.Printf("Processing input: %s", input)
	}

	// Normalize input
	normalizedInput := NormalizePattern(input)

	// Get current topic and that context
	currentTopic := session.GetSessionTopic()
	lastThat := session.GetLastThat()

	// Normalize the that context for matching using enhanced that normalization
	normalizedThat := ""
	if lastThat != "" {
		normalizedThat = NormalizeThatPattern(lastThat)
	}

	// Try to match pattern with full context (using index 0 for last response)
	category, wildcards, err := g.aimlKB.MatchPatternWithTopicAndThatIndexOriginal(normalizedInput, input, currentTopic, normalizedThat, 0)
	if err != nil {
		return "", err
	}

	// Capture that context from template before processing (for next input)
	// This needs to be done before the template is processed because <set> tags might change the content
	nextThatContext := g.extractThatContextFromTemplate(category.Template)

	// Process template with context
	response := g.ProcessTemplateWithContext(category.Template, wildcards, session)

	// Add to history
	session.History = append(session.History, input)
	session.LastActivity = time.Now().Format(time.RFC3339)

	// Add to request history for <request> tag support
	session.AddToRequestHistory(input)

	// Add the extracted that context to history for future context matching
	if nextThatContext != "" {
		session.AddToThatHistory(nextThatContext)
	}

	// Add to response history for <response> tag support
	session.AddToResponseHistory(response)

	return response, nil
}

// ProcessInputWithThatIndex processes user input with specific that context index
func (g *Golem) ProcessInputWithThatIndex(input string, session *ChatSession, thatIndex int) (string, error) {
	if g.aimlKB == nil {
		return "", fmt.Errorf("no AIML knowledge base loaded")
	}

	if g.verbose {
		g.logger.Printf("Processing input with that index %d: %s", thatIndex, input)
	}

	// Normalize input
	normalizedInput := NormalizePattern(input)

	// Get current topic and that context by index
	currentTopic := session.GetSessionTopic()
	thatContext := session.GetThatByIndex(thatIndex)

	if g.verbose {
		g.logger.Printf("That context for index %d: '%s'", thatIndex, thatContext)
		g.logger.Printf("That history: %v", session.ThatHistory)
	}

	// Normalize the that context for matching using enhanced that normalization
	normalizedThat := ""
	if thatContext != "" {
		normalizedThat = NormalizeThatPattern(thatContext)
	}

	// Try to match pattern with full context and specific that index
	category, wildcards, err := g.aimlKB.MatchPatternWithTopicAndThatIndexOriginal(normalizedInput, input, currentTopic, normalizedThat, thatIndex)
	if err != nil {
		return "", err
	}

	// Capture that context from template before processing (for next input)
	// This needs to be done before the template is processed because <set> tags might change the content
	nextThatContext := g.extractThatContextFromTemplate(category.Template)

	// Process template with context
	response := g.ProcessTemplateWithContext(category.Template, wildcards, session)

	// Add to history
	session.History = append(session.History, input)
	session.LastActivity = time.Now().Format(time.RFC3339)

	// Add to request history for <request> tag support
	session.AddToRequestHistory(input)

	// Add the extracted that context to history for future context matching
	if nextThatContext != "" {
		session.AddToThatHistory(nextThatContext)
	}

	// Add to response history for <response> tag support
	session.AddToResponseHistory(response)

	return response, nil
}

// extractThatContextFromTemplate extracts the that context from a template
// This is used to capture the that context before <set> tags are processed
func (g *Golem) extractThatContextFromTemplate(template string) string {
	// For that context, we need to extract only the content that comes after <set> tags
	// This is because <set> tags are processed and removed, but the that context
	// should only include the content that remains after processing

	// Find <set name="topic"> tags and extract content after them
	topicSetRegex := regexp.MustCompile(`<set\s+name="topic">(.*?)</set>`)
	matches := topicSetRegex.FindAllStringSubmatch(template, -1)

	if len(matches) > 0 {
		// If there are <set> tags, extract only the content after the last one
		lastMatch := matches[len(matches)-1]
		lastMatchEnd := strings.Index(template, lastMatch[0]) + len(lastMatch[0])

		// Get the content after the last <set> tag
		thatContext := strings.TrimSpace(template[lastMatchEnd:])

		return thatContext
	}

	// If no <set> tags, return the entire template
	processedTemplate := strings.TrimSpace(template)

	return processedTemplate
}

// AnalyzeData is a library function that can be used by other programs
func (g *Golem) AnalyzeData(input string) (map[string]interface{}, error) {
	if g.verbose {
		g.logger.Printf("Analyzing data: %s", input)
	}

	// TODO: Implement actual analysis logic
	result := map[string]interface{}{
		"input":  input,
		"status": "analyzed",
		"length": len(input),
	}
	return result, nil
}

// GenerateOutput is a library function that can be used by other programs
func (g *Golem) GenerateOutput(data interface{}) (string, error) {
	if g.verbose {
		g.logger.Printf("Generating output for data: %v", data)
	}

	// TODO: Implement actual generation logic
	result := fmt.Sprintf("Generated output for: %v", data)
	return result, nil
}

// LoadFile is a library function that loads a file and returns its contents
func (g *Golem) LoadFile(filename string) (string, error) {
	if g.verbose {
		g.logger.Printf("Loading file: %s", filename)
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read the file contents
	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	return string(content), nil
}

// SetKnowledgeBase sets the AIML knowledge base
func (g *Golem) SetKnowledgeBase(kb *AIMLKnowledgeBase) {
	g.aimlKB = kb

	// Register properties handler now that we have a knowledge base
	propertiesHandler := &PropertiesHandler{aimlKB: kb}
	g.oobMgr.RegisterHandler(propertiesHandler)
}

// GetKnowledgeBase returns the current AIML knowledge base
func (g *Golem) GetKnowledgeBase() *AIMLKnowledgeBase {
	return g.aimlKB
}

// SessionCommand handles session management
func (g *Golem) sessionCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("session command requires subcommand: create, list, switch, delete, current")
	}

	subcommand := args[0]
	switch subcommand {
	case "create":
		return g.createSessionCommand(args[1:])
	case "list":
		return g.listSessionsCommand()
	case "switch":
		return g.switchSessionCommand(args[1:])
	case "delete":
		return g.deleteSessionCommand(args[1:])
	case "current":
		return g.currentSessionCommand()
	default:
		return fmt.Errorf("unknown session subcommand: %s", subcommand)
	}
}

// createSessionCommand creates a new chat session
func (g *Golem) createSessionCommand(args []string) error {
	var sessionID string
	if len(args) > 0 {
		sessionID = args[0]
	}

	session := g.createSession(sessionID)
	fmt.Printf("Created session: %s\n", session.ID)
	return nil
}

// listSessionsCommand lists all active sessions
func (g *Golem) listSessionsCommand() error {
	if len(g.sessions) == 0 {
		fmt.Println("No active sessions")
		return nil
	}

	fmt.Println("Active Sessions:")
	fmt.Println(strings.Repeat("=", 50))
	for id, session := range g.sessions {
		marker := ""
		if id == g.currentID {
			marker = " (current)"
		}
		fmt.Printf("%-10s: Created %s, %d messages%s\n",
			id, session.CreatedAt, len(session.History), marker)
	}
	return nil
}

// switchSessionCommand switches to a different session
func (g *Golem) switchSessionCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("session switch requires session ID")
	}

	sessionID := args[0]
	if _, exists := g.sessions[sessionID]; !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	g.currentID = sessionID
	fmt.Printf("Switched to session: %s\n", sessionID)
	return nil
}

// deleteSessionCommand deletes a session
func (g *Golem) deleteSessionCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("session delete requires session ID")
	}

	sessionID := args[0]
	if _, exists := g.sessions[sessionID]; !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	delete(g.sessions, sessionID)
	if g.currentID == sessionID {
		g.currentID = ""
	}
	fmt.Printf("Deleted session: %s\n", sessionID)
	return nil
}

// currentSessionCommand shows current session info
func (g *Golem) currentSessionCommand() error {
	if g.currentID == "" {
		fmt.Println("No current session")
		return nil
	}

	session := g.sessions[g.currentID]
	fmt.Printf("Current session: %s\n", session.ID)
	fmt.Printf("Created: %s\n", session.CreatedAt)
	fmt.Printf("Messages: %d\n", len(session.History))
	return nil
}

// createSession creates a new chat session
// CreateSession creates a new chat session with the given ID
func (g *Golem) CreateSession(sessionID string) *ChatSession {
	return g.createSession(sessionID)
}

func (g *Golem) createSession(sessionID string) *ChatSession {
	if sessionID == "" {
		sessionID = fmt.Sprintf("session_%d", g.sessionID)
		g.sessionID++
	}

	now := "now" // In a real implementation, use time.Now().Format()
	session := &ChatSession{
		ID:              sessionID,
		Variables:       make(map[string]string),
		History:         []string{},
		CreatedAt:       now,
		LastActivity:    now,
		RequestHistory:  []string{},
		ResponseHistory: []string{},
		ThatHistory:     []string{},
	}

	// Initialize enhanced context management
	session.InitializeContextConfig()

	g.sessions[sessionID] = session
	g.currentID = sessionID
	return session
}

// getCurrentSession returns the current session
func (g *Golem) getCurrentSession() *ChatSession {
	if g.currentID == "" {
		return nil
	}
	return g.sessions[g.currentID]
}

// ProcessTemplateWithSession processes a template with session context
func (g *Golem) ProcessTemplateWithSession(template string, wildcards map[string]string, session *ChatSession) string {
	// Create variable context for template processing with session
	ctx := &VariableContext{
		LocalVars:     make(map[string]string),
		Session:       session,
		Topic:         "", // TODO: Implement topic tracking
		KnowledgeBase: g.aimlKB,
	}

	return g.processTemplateWithContext(template, wildcards, ctx)
}

// replaceSessionVariableTags replaces <get name="var"/> tags with session variables
func (g *Golem) replaceSessionVariableTags(template string, session *ChatSession) string {
	// Find all <get name="var"/> tags
	getTagRegex := regexp.MustCompile(`<get name="([^"]+)"/>`)
	matches := getTagRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			varValue := session.Variables[varName]
			if varValue != "" {
				template = strings.ReplaceAll(template, match[0], varValue)
			}
		}
	}

	return template
}

// OOBCommand handles OOB-related commands
func (g *Golem) oobCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("oob command requires subcommand: list, test, register")
	}

	subcommand := args[0]
	switch subcommand {
	case "list":
		return g.listOOBHandlers()
	case "test":
		return g.testOOBCommand(args[1:])
	case "register":
		return g.registerOOBHandler(args[1:])
	default:
		return fmt.Errorf("unknown oob subcommand: %s", subcommand)
	}
}

// listOOBHandlers lists all registered OOB handlers
func (g *Golem) listOOBHandlers() error {
	handlers := g.oobMgr.ListHandlers()
	if len(handlers) == 0 {
		fmt.Println("No OOB handlers registered")
		return nil
	}

	fmt.Println("Registered OOB Handlers:")
	fmt.Println(strings.Repeat("=", 40))
	for _, name := range handlers {
		if handler, exists := g.oobMgr.GetHandler(name); exists {
			fmt.Printf("%-20s: %s\n", name, handler.GetDescription())
		}
	}
	return nil
}

// testOOBCommand tests an OOB message
func (g *Golem) testOOBCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("oob test requires a message")
	}

	message := strings.Join(args, " ")
	session := g.getCurrentSession()
	if session == nil {
		session = g.createSession("")
	}

	response, err := g.oobMgr.ProcessOOB(message, session)
	if err != nil {
		fmt.Printf("OOB Error: %v\n", err)
		return nil
	}

	fmt.Printf("OOB Response: %s\n", response)
	return nil
}

// registerOOBHandler registers a new OOB handler (for advanced users)
func (g *Golem) registerOOBHandler(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("oob register requires handler name and description")
	}

	name := args[0]
	description := strings.Join(args[1:], " ")

	// Create a simple test handler
	handler := &TestOOBHandler{
		name:        name,
		description: description,
	}

	g.oobMgr.RegisterHandler(handler)
	fmt.Printf("Registered custom OOB handler: %s\n", name)
	return nil
}

// TestOOBHandler is a simple test handler for demonstration
type TestOOBHandler struct {
	name        string
	description string
}

func (h *TestOOBHandler) CanHandle(message string) bool {
	return strings.HasPrefix(strings.ToUpper(message), strings.ToUpper(h.name))
}

func (h *TestOOBHandler) Process(message string, session *ChatSession) (string, error) {
	return fmt.Sprintf("Test handler '%s' processed: %s", h.name, message), nil
}

func (h *TestOOBHandler) GetName() string {
	return h.name
}

func (h *TestOOBHandler) GetDescription() string {
	return h.description
}

// SRAIX Management Methods

// AddSRAIXConfig adds a new SRAIX service configuration
func (g *Golem) AddSRAIXConfig(config *SRAIXConfig) error {
	return g.sraixMgr.AddConfig(config)
}

// GetSRAIXConfig retrieves a SRAIX service configuration
func (g *Golem) GetSRAIXConfig(name string) (*SRAIXConfig, bool) {
	return g.sraixMgr.GetConfig(name)
}

// ListSRAIXConfigs returns all configured SRAIX services
func (g *Golem) ListSRAIXConfigs() map[string]*SRAIXConfig {
	return g.sraixMgr.ListConfigs()
}

// LoadSRAIXConfigsFromFile loads SRAIX configurations from a JSON file
func (g *Golem) LoadSRAIXConfigsFromFile(filename string) error {
	return g.sraixMgr.LoadSRAIXConfigsFromFile(filename)
}

// LoadSRAIXConfigsFromDirectory loads all SRAIX configuration files from a directory
func (g *Golem) LoadSRAIXConfigsFromDirectory(dirPath string) error {
	return g.sraixMgr.LoadSRAIXConfigsFromDirectory(dirPath)
}

// sraixCommand handles SRAIX-related CLI commands
func (g *Golem) sraixCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("sraix command requires subcommand: load, list, test")
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "load":
		return g.sraixLoadCommand(subArgs)
	case "list":
		return g.sraixListCommand()
	case "test":
		return g.sraixTestCommand(subArgs)
	default:
		return fmt.Errorf("unknown sraix subcommand: %s", subcommand)
	}
}

// sraixLoadCommand loads SRAIX configurations from file or directory
func (g *Golem) sraixLoadCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("sraix load requires a filename or directory path")
	}

	path := args[0]

	// Check if it's a file or directory
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access path %s: %v", path, err)
	}

	if info.IsDir() {
		err = g.LoadSRAIXConfigsFromDirectory(path)
		if err != nil {
			return fmt.Errorf("failed to load SRAIX configs from directory: %v", err)
		}
		fmt.Printf("Successfully loaded SRAIX configurations from directory: %s\n", path)
	} else {
		err = g.LoadSRAIXConfigsFromFile(path)
		if err != nil {
			return fmt.Errorf("failed to load SRAIX config file: %v", err)
		}
		fmt.Printf("Successfully loaded SRAIX configuration file: %s\n", path)
	}

	// Show loaded configurations
	configs := g.ListSRAIXConfigs()
	fmt.Printf("Loaded %d SRAIX service(s)\n", len(configs))
	for name, config := range configs {
		fmt.Printf("  %s: %s %s\n", name, config.Method, config.BaseURL)
	}

	return nil
}

// sraixListCommand lists all configured SRAIX services
func (g *Golem) sraixListCommand() error {
	configs := g.ListSRAIXConfigs()

	if len(configs) == 0 {
		fmt.Println("No SRAIX services configured")
		return nil
	}

	fmt.Println("Configured SRAIX Services:")
	fmt.Println("==========================================")
	for name, config := range configs {
		fmt.Printf("Name: %s\n", name)
		fmt.Printf("  URL: %s %s\n", config.Method, config.BaseURL)
		fmt.Printf("  Timeout: %ds\n", config.Timeout)
		fmt.Printf("  Format: %s\n", config.ResponseFormat)
		if config.ResponsePath != "" {
			fmt.Printf("  Path: %s\n", config.ResponsePath)
		}
		if config.FallbackResponse != "" {
			fmt.Printf("  Fallback: %s\n", config.FallbackResponse)
		}
		fmt.Printf("  Wildcards: %t\n", config.IncludeWildcards)
		fmt.Println()
	}

	return nil
}

// sraixTestCommand tests a SRAIX service
func (g *Golem) sraixTestCommand(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("sraix test requires service name and test input")
	}

	serviceName := args[0]
	testInput := strings.Join(args[1:], " ")

	config, exists := g.GetSRAIXConfig(serviceName)
	if !exists {
		return fmt.Errorf("SRAIX service '%s' not found", serviceName)
	}

	fmt.Printf("Testing SRAIX service '%s' with input: '%s'\n", serviceName, testInput)
	fmt.Printf("Service URL: %s %s\n", config.Method, config.BaseURL)
	fmt.Println("Making request...")

	// Make the SRAIX request
	response, err := g.sraixMgr.ProcessSRAIX(serviceName, testInput, make(map[string]string))
	if err != nil {
		fmt.Printf("SRAIX request failed: %v\n", err)
		return nil
	}

	fmt.Printf("Response: %s\n", response)
	return nil
}

// GetTemplateProcessingMetrics returns current template processing metrics
func (g *Golem) GetTemplateProcessingMetrics() *TemplateProcessingMetrics {
	return g.templateMetrics
}

// GetTemplateProcessingConfig returns current template processing configuration
func (g *Golem) GetTemplateProcessingConfig() *TemplateProcessingConfig {
	return g.templateConfig
}

// UpdateTemplateProcessingConfig updates template processing configuration
func (g *Golem) UpdateTemplateProcessingConfig(config *TemplateProcessingConfig) {
	g.templateConfig = config
	// Update cache settings
	if g.templateCache != nil {
		g.templateCache.MaxSize = config.CacheSize
		g.templateCache.TTL = config.CacheTTL
	}
}

// ClearTemplateCache clears the template cache
func (g *Golem) ClearTemplateCache() {
	if g.templateCache != nil {
		g.templateCache.Cache = make(map[string]string)
		g.templateCache.Timestamps = make(map[string]string)
		g.templateCache.Hits = make(map[string]int)
	}
}

// GetTemplateCacheStats returns template cache statistics
func (g *Golem) GetTemplateCacheStats() map[string]interface{} {
	if g.templateCache == nil {
		return map[string]interface{}{
			"cache_size": 0,
			"hits":       0,
			"misses":     0,
			"hit_rate":   0.0,
		}
	}

	totalRequests := g.templateMetrics.CacheHits + g.templateMetrics.CacheMisses
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(g.templateMetrics.CacheHits) / float64(totalRequests)
	}

	return map[string]interface{}{
		"cache_size":      len(g.templateCache.Cache),
		"max_size":        g.templateCache.MaxSize,
		"ttl_seconds":     g.templateCache.TTL,
		"hits":            g.templateMetrics.CacheHits,
		"misses":          g.templateMetrics.CacheMisses,
		"hit_rate":        hitRate,
		"total_processed": g.templateMetrics.TotalProcessed,
		"average_time_ms": g.templateMetrics.AverageProcessTime,
		"error_count":     g.templateMetrics.ErrorCount,
	}
}

// ResetTemplateMetrics resets template processing metrics
func (g *Golem) ResetTemplateMetrics() {
	if g.templateMetrics != nil {
		g.templateMetrics.TotalProcessed = 0
		g.templateMetrics.AverageProcessTime = 0.0
		g.templateMetrics.CacheHits = 0
		g.templateMetrics.CacheMisses = 0
		g.templateMetrics.CacheHitRate = 0.0
		g.templateMetrics.TagProcessingTimes = make(map[string]float64)
		g.templateMetrics.ErrorCount = 0
		g.templateMetrics.LastProcessed = ""
		g.templateMetrics.MemoryPeak = 0
		g.templateMetrics.ParallelOps = 0
	}
}

// generateTemplateCacheKey creates a cache key from template, wildcards and minimal ctx
func (g *Golem) generateTemplateCacheKey(template string, wildcards map[string]string, ctx *VariableContext) string {
	// Build a deterministic key: template + sorted wildcards + session/topic markers
	var b strings.Builder
	b.WriteString("tpl:")
	b.WriteString(template)
	b.WriteString("|wc:")
	// sort wildcards for deterministic order
	keys := make([]string, 0, len(wildcards))
	for k := range wildcards {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(wildcards[k])
		b.WriteString(";")
	}
	if ctx != nil && ctx.Session != nil {
		b.WriteString("|sid:")
		b.WriteString(ctx.Session.ID)
	}
	if ctx != nil {
		b.WriteString("|topic:")
		b.WriteString(ctx.Topic)
	}
	return b.String()
}

// getFromTemplateCache fetches a cached response if present
func (g *Golem) getFromTemplateCache(key string) (string, bool) {
	if g.templateCache == nil {
		return "", false
	}
	v, ok := g.templateCache.Cache[key]
	if ok {
		g.templateCache.Hits[key] = g.templateCache.Hits[key] + 1
	}
	return v, ok
}

// storeInTemplateCache stores a processed template result
func (g *Golem) storeInTemplateCache(key, value string) {
	if g.templateCache == nil {
		return
	}
	// Evict if over capacity (simple FIFO by timestamps if needed)
	if len(g.templateCache.Cache) >= g.templateCache.MaxSize {
		// naive eviction: remove an arbitrary oldest by timestamp string comparison
		var oldestKey string
		var oldestTs string
		for k, ts := range g.templateCache.Timestamps {
			if oldestTs == "" || ts < oldestTs {
				oldestTs = ts
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(g.templateCache.Cache, oldestKey)
			delete(g.templateCache.Timestamps, oldestKey)
			delete(g.templateCache.Hits, oldestKey)
		}
	}
	g.templateCache.Cache[key] = value
	// store a simple increasing timestamp using TotalProcessed to keep it consistent
	g.templateMetrics.TotalProcessed++
	g.templateCache.Timestamps[key] = strconv.Itoa(g.templateMetrics.TotalProcessed)
	if _, exists := g.templateCache.Hits[key]; !exists {
		g.templateCache.Hits[key] = 0
	}
}

// updateCacheHitRate recomputes cache hit rate metric
func (g *Golem) updateCacheHitRate() {
	total := g.templateMetrics.CacheHits + g.templateMetrics.CacheMisses
	if total > 0 {
		g.templateMetrics.CacheHitRate = float64(g.templateMetrics.CacheHits) / float64(total)
	} else {
		g.templateMetrics.CacheHitRate = 0
	}
}
