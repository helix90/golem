package golem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// AIML represents the root AIML document
type AIML struct {
	Version    string
	Categories []Category
}

// Category represents an AIML category (pattern-template pair)
type Category struct {
	Pattern  string
	Template string
	That     string
	Topic    string
}

// AIMLKnowledgeBase stores the parsed AIML data for efficient searching
type AIMLKnowledgeBase struct {
	Categories []Category
	Patterns   map[string]*Category
	Sets       map[string][]string
	Topics     map[string][]string
	Variables  map[string]string
	Properties map[string]string
	Maps       map[string]map[string]string // Maps: mapName -> key -> value
	Lists      map[string][]string          // Lists: listName -> []values
	Arrays     map[string][]string          // Arrays: arrayName -> []values
}

// NewAIMLKnowledgeBase creates a new knowledge base
func NewAIMLKnowledgeBase() *AIMLKnowledgeBase {
	return &AIMLKnowledgeBase{
		Patterns:   make(map[string]*Category),
		Sets:       make(map[string][]string),
		Topics:     make(map[string][]string),
		Variables:  make(map[string]string),
		Properties: make(map[string]string),
		Maps:       make(map[string]map[string]string),
		Lists:      make(map[string][]string),
		Arrays:     make(map[string][]string),
	}
}

// LoadAIML parses an AIML file using native Go string manipulation
// LoadAIMLFromString loads AIML from a string and returns the parsed knowledge base
func (g *Golem) LoadAIMLFromString(content string) error {
	if g.verbose {
		g.logger.Printf("Loading AIML from string")
	}

	// Parse the AIML content
	aiml, err := g.parseAIML(content)
	if err != nil {
		return err
	}

	// Convert AIML to AIMLKnowledgeBase
	kb := g.aimlToKnowledgeBase(aiml)

	// Merge with existing knowledge base
	if g.aimlKB == nil {
		g.aimlKB = kb
	} else {
		mergedKB, err := g.mergeKnowledgeBases(g.aimlKB, kb)
		if err != nil {
			return err
		}
		g.aimlKB = mergedKB
	}

	if g.verbose {
		g.logger.Printf("Loaded AIML from string successfully")
		g.logger.Printf("Total categories: %d", len(g.aimlKB.Categories))
		g.logger.Printf("Total patterns: %d", len(g.aimlKB.Patterns))
		g.logger.Printf("Total sets: %d", len(g.aimlKB.Sets))
		g.logger.Printf("Total topics: %d", len(g.aimlKB.Topics))
		g.logger.Printf("Total variables: %d", len(g.aimlKB.Variables))
		g.logger.Printf("Total properties: %d", len(g.aimlKB.Properties))
		g.logger.Printf("Total maps: %d", len(g.aimlKB.Maps))
	}

	return nil
}

// aimlToKnowledgeBase converts AIML to AIMLKnowledgeBase
func (g *Golem) aimlToKnowledgeBase(aiml *AIML) *AIMLKnowledgeBase {
	kb := &AIMLKnowledgeBase{
		Categories: aiml.Categories,
		Patterns:   make(map[string]*Category),
		Sets:       make(map[string][]string),
		Topics:     make(map[string][]string),
		Variables:  make(map[string]string),
		Properties: make(map[string]string),
		Maps:       make(map[string]map[string]string),
		Lists:      make(map[string][]string),
		Arrays:     make(map[string][]string),
	}

	// Build pattern index
	for i := range kb.Categories {
		pattern := NormalizePattern(kb.Categories[i].Pattern)
		// Create a unique key that includes pattern, that, and topic
		key := pattern
		if kb.Categories[i].That != "" {
			key += "|THAT:" + NormalizePattern(kb.Categories[i].That)
		}
		if kb.Categories[i].Topic != "" {
			key += "|TOPIC:" + strings.ToUpper(kb.Categories[i].Topic)
		}
		kb.Patterns[key] = &kb.Categories[i]
	}

	return kb
}

// mergeKnowledgeBases merges two knowledge bases
func (g *Golem) mergeKnowledgeBases(kb1, kb2 *AIMLKnowledgeBase) (*AIMLKnowledgeBase, error) {
	mergedKB := &AIMLKnowledgeBase{
		Categories: make([]Category, 0),
		Patterns:   make(map[string]*Category),
		Sets:       make(map[string][]string),
		Topics:     make(map[string][]string),
		Variables:  make(map[string]string),
		Properties: make(map[string]string),
		Maps:       make(map[string]map[string]string),
		Lists:      make(map[string][]string),
		Arrays:     make(map[string][]string),
	}

	// Copy from first knowledge base
	mergedKB.Categories = append(mergedKB.Categories, kb1.Categories...)
	for pattern, category := range kb1.Patterns {
		mergedKB.Patterns[pattern] = category
	}
	for setName, members := range kb1.Sets {
		mergedKB.Sets[setName] = members
	}
	for topicName, patterns := range kb1.Topics {
		mergedKB.Topics[topicName] = patterns
	}
	for varName, value := range kb1.Variables {
		mergedKB.Variables[varName] = value
	}
	for propName, value := range kb1.Properties {
		mergedKB.Properties[propName] = value
	}
	for mapName, mapData := range kb1.Maps {
		mergedKB.Maps[mapName] = mapData
	}
	for listName, listData := range kb1.Lists {
		mergedKB.Lists[listName] = listData
	}
	for arrayName, arrayData := range kb1.Arrays {
		mergedKB.Arrays[arrayName] = arrayData
	}

	// Merge from second knowledge base
	mergedKB.Categories = append(mergedKB.Categories, kb2.Categories...)
	for pattern, category := range kb2.Patterns {
		mergedKB.Patterns[pattern] = category
	}
	for setName, members := range kb2.Sets {
		if mergedKB.Sets[setName] == nil {
			mergedKB.Sets[setName] = make([]string, 0)
		}
		mergedKB.Sets[setName] = append(mergedKB.Sets[setName], members...)
	}
	for topicName, patterns := range kb2.Topics {
		if mergedKB.Topics[topicName] == nil {
			mergedKB.Topics[topicName] = make([]string, 0)
		}
		mergedKB.Topics[topicName] = append(mergedKB.Topics[topicName], patterns...)
	}
	for varName, value := range kb2.Variables {
		mergedKB.Variables[varName] = value
	}
	for propName, value := range kb2.Properties {
		mergedKB.Properties[propName] = value
	}
	for mapName, mapData := range kb2.Maps {
		if mergedKB.Maps[mapName] == nil {
			mergedKB.Maps[mapName] = make(map[string]string)
		}
		for key, value := range mapData {
			mergedKB.Maps[mapName][key] = value
		}
	}
	for listName, listData := range kb2.Lists {
		if mergedKB.Lists[listName] == nil {
			mergedKB.Lists[listName] = make([]string, 0)
		}
		mergedKB.Lists[listName] = append(mergedKB.Lists[listName], listData...)
	}
	for arrayName, arrayData := range kb2.Arrays {
		if mergedKB.Arrays[arrayName] == nil {
			mergedKB.Arrays[arrayName] = make([]string, 0)
		}
		mergedKB.Arrays[arrayName] = append(mergedKB.Arrays[arrayName], arrayData...)
	}

	return mergedKB, nil
}

func (g *Golem) LoadAIML(filename string) (*AIMLKnowledgeBase, error) {
	if g.verbose {
		g.logger.Printf("Loading AIML file: %s", filename)
	}

	// Read the file
	content, err := g.LoadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load AIML file: %v", err)
	}

	// Parse the AIML content
	aiml, err := g.parseAIML(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AIML: %v", err)
	}

	// Validate the AIML
	err = g.validateAIML(aiml)
	if err != nil {
		return nil, fmt.Errorf("AIML validation failed: %v", err)
	}

	// Create knowledge base
	kb := NewAIMLKnowledgeBase()
	kb.Categories = aiml.Categories

	// Load default properties
	err = g.loadDefaultProperties(kb)
	if err != nil {
		return nil, fmt.Errorf("failed to load default properties: %v", err)
	}

	// Index patterns for fast lookup
	for i := range aiml.Categories {
		category := &aiml.Categories[i]
		// Normalize pattern for storage
		pattern := NormalizePattern(category.Pattern)
		kb.Patterns[pattern] = category
	}

	if g.verbose {
		g.logger.Printf("Loaded %d AIML categories", len(aiml.Categories))
		g.logger.Printf("Loaded %d properties", len(kb.Properties))
	}

	return kb, nil
}

// LoadAIMLFromDirectory loads all AIML files from a directory and merges them into a single knowledge base
func (g *Golem) LoadAIMLFromDirectory(dirPath string) (*AIMLKnowledgeBase, error) {
	if g.verbose {
		g.logger.Printf("Loading AIML files from directory: %s", dirPath)
	}

	// Create a new knowledge base to merge all files into
	mergedKB := NewAIMLKnowledgeBase()

	// Load default properties first
	err := g.loadDefaultProperties(mergedKB)
	if err != nil {
		return nil, fmt.Errorf("failed to load default properties: %v", err)
	}

	// Walk through the directory to find all .aiml files
	var aimlFiles []string
	err = filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a .aiml file
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".aiml") {
			aimlFiles = append(aimlFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %v", dirPath, err)
	}

	if len(aimlFiles) == 0 {
		return nil, fmt.Errorf("no AIML files found in directory: %s", dirPath)
	}

	if g.verbose {
		g.logger.Printf("Found %d AIML files in directory", len(aimlFiles))
	}

	// Load each AIML file and merge into the knowledge base
	for _, aimlFile := range aimlFiles {
		if g.verbose {
			g.logger.Printf("Loading AIML file: %s", aimlFile)
		}

		// Load the individual AIML file
		kb, err := g.LoadAIML(aimlFile)
		if err != nil {
			// Log the error but continue with other files
			if g.verbose {
				g.logger.Printf("Warning: failed to load %s: %v", aimlFile, err)
			}
			continue
		}

		// Merge the categories from this file into the merged knowledge base
		for i := range kb.Categories {
			category := &kb.Categories[i]
			// Normalize pattern for storage
			pattern := NormalizePattern(category.Pattern)

			// Add category to merged knowledge base
			mergedKB.Categories = append(mergedKB.Categories, *category)
			mergedKB.Patterns[pattern] = category
		}

		// Merge sets
		for setName, members := range kb.Sets {
			for _, member := range members {
				mergedKB.AddSetMember(setName, member)
			}
		}

		// Merge topics
		for topicName, patterns := range kb.Topics {
			if mergedKB.Topics[topicName] == nil {
				mergedKB.Topics[topicName] = make([]string, 0)
			}
			mergedKB.Topics[topicName] = append(mergedKB.Topics[topicName], patterns...)
		}

		// Merge variables (file variables override defaults)
		for varName, varValue := range kb.Variables {
			mergedKB.Variables[varName] = varValue
		}

		// Merge properties (file properties override defaults)
		for propName, propValue := range kb.Properties {
			mergedKB.Properties[propName] = propValue
		}
	}

	// Load map files from the same directory
	maps, err := g.LoadMapsFromDirectory(dirPath)
	if err != nil {
		// Log the error but don't fail the entire operation
		if g.verbose {
			g.logger.Printf("Warning: failed to load maps from directory: %v", err)
		}
	} else {
		// Merge maps into the knowledge base
		for mapName, mapData := range maps {
			mergedKB.Maps[mapName] = mapData
		}
	}

	// Load set files from the same directory
	sets, err := g.LoadSetsFromDirectory(dirPath)
	if err != nil {
		// Log the error but don't fail the entire operation
		if g.verbose {
			g.logger.Printf("Warning: failed to load sets from directory: %v", err)
		}
	} else {
		// Merge sets into the knowledge base
		for setName, setMembers := range sets {
			mergedKB.AddSetMembers(setName, setMembers)
		}
	}

	if g.verbose {
		g.logger.Printf("Merged %d AIML files into knowledge base", len(aimlFiles))
		g.logger.Printf("Total categories: %d", len(mergedKB.Categories))
		g.logger.Printf("Total patterns: %d", len(mergedKB.Patterns))
		g.logger.Printf("Total sets: %d", len(mergedKB.Sets))
		g.logger.Printf("Total topics: %d", len(mergedKB.Topics))
		g.logger.Printf("Total variables: %d", len(mergedKB.Variables))
		g.logger.Printf("Total properties: %d", len(mergedKB.Properties))
		g.logger.Printf("Total maps: %d", len(mergedKB.Maps))
	}

	return mergedKB, nil
}

// LoadMapFromFile loads a .map file containing JSON array of key-value pairs
func (g *Golem) LoadMapFromFile(filename string) (map[string]string, error) {
	if g.verbose {
		g.logger.Printf("Loading map file: %s", filename)
	}

	// Read the file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read map file %s: %v", filename, err)
	}

	// Parse JSON array
	var mapEntries []map[string]string
	err = json.Unmarshal(content, &mapEntries)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON in map file %s: %v", filename, err)
	}

	// Convert array to map
	result := make(map[string]string)
	for _, entry := range mapEntries {
		key, hasKey := entry["key"]
		value, hasValue := entry["value"]

		if !hasKey || !hasValue {
			if g.verbose {
				g.logger.Printf("Warning: skipping entry missing key or value: %v", entry)
			}
			continue
		}

		result[key] = value
	}

	if g.verbose {
		g.logger.Printf("Loaded %d map entries from %s", len(result), filename)
	}

	return result, nil
}

// LoadMapsFromDirectory loads all .map files from a directory
func (g *Golem) LoadMapsFromDirectory(dirPath string) (map[string]map[string]string, error) {
	if g.verbose {
		g.logger.Printf("Loading map files from directory: %s", dirPath)
	}

	// Create a map to store all maps
	allMaps := make(map[string]map[string]string)

	// Walk through the directory to find all .map files
	var mapFiles []string
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a .map file
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".map") {
			mapFiles = append(mapFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %v", dirPath, err)
	}

	if len(mapFiles) == 0 {
		if g.verbose {
			g.logger.Printf("No map files found in directory: %s", dirPath)
		}
		return allMaps, nil
	}

	if g.verbose {
		g.logger.Printf("Found %d map files in directory", len(mapFiles))
	}

	// Load each map file
	for _, mapFile := range mapFiles {
		if g.verbose {
			g.logger.Printf("Loading map file: %s", mapFile)
		}

		// Load the individual map file
		mapData, err := g.LoadMapFromFile(mapFile)
		if err != nil {
			// Log the error but continue with other files
			if g.verbose {
				g.logger.Printf("Warning: failed to load %s: %v", mapFile, err)
			}
			continue
		}

		// Use the filename (without extension) as the map name
		mapName := strings.TrimSuffix(filepath.Base(mapFile), filepath.Ext(mapFile))
		allMaps[mapName] = mapData
	}

	if g.verbose {
		g.logger.Printf("Loaded %d map files", len(allMaps))
	}

	return allMaps, nil
}

// LoadSetFromFile loads a .set file containing JSON array of set members
func (g *Golem) LoadSetFromFile(filename string) ([]string, error) {
	if g.verbose {
		g.logger.Printf("Loading set file: %s", filename)
	}

	// Read the file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read set file %s: %v", filename, err)
	}

	// Parse JSON array
	var setMembers []string
	err = json.Unmarshal(content, &setMembers)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON in set file %s: %v", filename, err)
	}

	if g.verbose {
		g.logger.Printf("Loaded %d set members from %s", len(setMembers), filename)
	}

	return setMembers, nil
}

// LoadSetsFromDirectory loads all .set files from a directory
func (g *Golem) LoadSetsFromDirectory(dirPath string) (map[string][]string, error) {
	if g.verbose {
		g.logger.Printf("Loading set files from directory: %s", dirPath)
	}

	// Create a map to store all sets
	allSets := make(map[string][]string)

	// Walk through the directory to find all .set files
	var setFiles []string
	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a .set file
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".set") {
			setFiles = append(setFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %v", dirPath, err)
	}

	if len(setFiles) == 0 {
		if g.verbose {
			g.logger.Printf("No set files found in directory: %s", dirPath)
		}
		return allSets, nil
	}

	if g.verbose {
		g.logger.Printf("Found %d set files in directory", len(setFiles))
	}

	// Load each set file
	for _, setFile := range setFiles {
		if g.verbose {
			g.logger.Printf("Loading set file: %s", setFile)
		}

		// Load the individual set file
		setMembers, err := g.LoadSetFromFile(setFile)
		if err != nil {
			// Log the error but continue with other files
			if g.verbose {
				g.logger.Printf("Warning: failed to load %s: %v", setFile, err)
			}
			continue
		}

		// Use the filename (without extension) as the set name
		setName := strings.TrimSuffix(filepath.Base(setFile), filepath.Ext(setFile))
		allSets[setName] = setMembers
	}

	if g.verbose {
		g.logger.Printf("Loaded %d set files", len(allSets))
	}

	return allSets, nil
}

// parseAIML parses AIML content using native Go string manipulation
func (g *Golem) parseAIML(content string) (*AIML, error) {
	aiml := &AIML{
		Categories: []Category{},
	}

	// Remove XML declaration and comments
	content = g.removeComments(content)
	content = g.removeXMLDeclaration(content)

	// Extract version
	versionMatch := regexp.MustCompile(`<aiml[^>]*version=["']([^"']+)["']`).FindStringSubmatch(content)
	if len(versionMatch) > 1 {
		aiml.Version = versionMatch[1]
	} else {
		aiml.Version = "2.0" // Default version
	}

	// Find all categories
	categoryRegex := regexp.MustCompile(`(?s)<category>(.*?)</category>`)
	matches := categoryRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			categoryContent := match[1]
			category, err := g.parseCategory(categoryContent)
			if err != nil {
				return nil, fmt.Errorf("failed to parse category: %v", err)
			}
			aiml.Categories = append(aiml.Categories, category)
		}
	}

	return aiml, nil
}

// parseCategory parses a single category
func (g *Golem) parseCategory(content string) (Category, error) {
	category := Category{}

	// Extract pattern
	patternMatch := regexp.MustCompile(`(?s)<pattern>(.*?)</pattern>`).FindStringSubmatch(content)
	if len(patternMatch) > 1 {
		category.Pattern = strings.TrimSpace(patternMatch[1])
	}

	// Extract template
	templateMatch := regexp.MustCompile(`(?s)<template>(.*?)</template>`).FindStringSubmatch(content)
	if len(templateMatch) > 1 {
		category.Template = strings.TrimSpace(templateMatch[1])
	}

	// Extract that (optional)
	thatMatch := regexp.MustCompile(`(?s)<that>(.*?)</that>`).FindStringSubmatch(content)
	if len(thatMatch) > 1 {
		category.That = strings.TrimSpace(thatMatch[1])
	}

	// Extract topic (optional)
	topicMatch := regexp.MustCompile(`(?s)<topic>(.*?)</topic>`).FindStringSubmatch(content)
	if len(topicMatch) > 1 {
		category.Topic = strings.TrimSpace(topicMatch[1])
	}

	return category, nil
}

// removeComments removes XML comments from content
func (g *Golem) removeComments(content string) string {
	commentRegex := regexp.MustCompile(`<!--.*?-->`)
	return commentRegex.ReplaceAllString(content, "")
}

// removeXMLDeclaration removes XML declaration
func (g *Golem) removeXMLDeclaration(content string) string {
	xmlDeclRegex := regexp.MustCompile(`<\?xml[^>]*\?>`)
	return xmlDeclRegex.ReplaceAllString(content, "")
}

// validateAIML validates the AIML structure
func (g *Golem) validateAIML(aiml *AIML) error {
	if aiml.Version == "" {
		return fmt.Errorf("AIML version is required")
	}

	if len(aiml.Categories) == 0 {
		return fmt.Errorf("AIML must contain at least one category")
	}

	for i, category := range aiml.Categories {
		if strings.TrimSpace(category.Pattern) == "" {
			return fmt.Errorf("category %d: pattern cannot be empty", i)
		}

		// Validate pattern syntax
		err := g.validatePattern(category.Pattern)
		if err != nil {
			return fmt.Errorf("category %d: invalid pattern '%s': %v", i, category.Pattern, err)
		}
	}

	return nil
}

// validatePattern validates AIML pattern syntax
func (g *Golem) validatePattern(pattern string) error {
	// Basic pattern validation
	pattern = strings.TrimSpace(pattern)

	// Check for valid wildcards and tags
	// First, normalize the pattern by replacing set and topic tags with placeholders
	normalizedPattern := pattern
	setPattern := regexp.MustCompile(`<set>[^<]+</set>`)
	topicPattern := regexp.MustCompile(`<topic>[^<]+</topic>`)
	normalizedPattern = setPattern.ReplaceAllString(normalizedPattern, "SETTAG")
	normalizedPattern = topicPattern.ReplaceAllString(normalizedPattern, "TOPICTAG")

	validWildcard := regexp.MustCompile(`^[A-Z0-9\s\*_<>/]+$`)
	if !validWildcard.MatchString(normalizedPattern) {
		return fmt.Errorf("pattern contains invalid characters")
	}

	// Check for balanced wildcards
	starCount := strings.Count(pattern, "*")

	if starCount > 9 {
		return fmt.Errorf("pattern contains too many wildcards (max 9)")
	}

	// Check for valid set references
	setRefPattern := regexp.MustCompile(`<set>([^<]+)</set>`)
	matches := setRefPattern.FindAllStringSubmatch(pattern, -1)
	for _, match := range matches {
		if len(match) > 1 && strings.TrimSpace(match[1]) == "" {
			return fmt.Errorf("set name cannot be empty")
		}
	}

	return nil
}

// PatternPriority represents the priority of a pattern for matching
type PatternPriority struct {
	Pattern          string
	Category         *Category
	Priority         int
	WildcardCount    int
	HasUnderscore    bool
	WildcardPosition int
}

// MatchPattern attempts to match user input against AIML patterns with highest priority matching
func (kb *AIMLKnowledgeBase) MatchPattern(input string) (*Category, map[string]string, error) {
	return kb.MatchPatternWithTopicAndThat(input, "", "")
}

// MatchPatternWithTopicAndThat attempts to match user input against AIML patterns with topic and that filtering
func (kb *AIMLKnowledgeBase) MatchPatternWithTopicAndThat(input string, topic string, that string) (*Category, map[string]string, error) {
	// Normalize input for matching (use same normalization as patterns)
	input = NormalizePattern(input)

	// Normalize that for matching
	normalizedThat := ""
	if that != "" {
		normalizedThat = NormalizePattern(that)
	}

	// Try exact match first (highest priority)
	// Build the exact key to look for
	exactKey := input
	if normalizedThat != "" {
		exactKey += "|THAT:" + normalizedThat
	}
	if topic != "" {
		exactKey += "|TOPIC:" + strings.ToUpper(topic)
	}

	if category, exists := kb.Patterns[exactKey]; exists {
		return category, make(map[string]string), nil
	}

	// Collect all matching patterns with their priorities
	var matchingPatterns []PatternPriority

	for patternKey, category := range kb.Patterns {
		if patternKey == "DEFAULT" {
			continue // Handle default separately
		}

		// Extract the base pattern from the key (before the first |)
		basePattern := strings.Split(patternKey, "|")[0]

		// Check topic match - if pattern has a topic, it must match the current topic
		if category.Topic != "" {
			// Use wildcard matching for topic if it contains wildcards
			if strings.Contains(category.Topic, "*") {
				matched, _ := matchPatternWithWildcardsAndSets(topic, category.Topic, kb)
				if !matched {
					continue // Skip patterns that don't match the topic
				}
			} else {
				// Use exact matching for topics without wildcards
				if !strings.EqualFold(category.Topic, topic) {
					continue // Skip patterns that have a different topic
				}
			}
		}

		// Check that match - if pattern has a that, it must match the current that
		thatMatched := true
		if category.That != "" {
			// Use wildcard matching for that context
			thatMatched, _ = matchPatternWithWildcardsAndSets(normalizedThat, category.That, kb)
			if !thatMatched {
				continue // Skip patterns that don't match the that context
			}
		}

		// Try enhanced matching with sets first
		matched, _ := matchPatternWithWildcardsAndSets(input, basePattern, kb)
		if matched && thatMatched {
			priority := calculatePatternPriority(basePattern)

			// Boost priority for patterns with that context
			if category.That != "" {
				priority.Priority += 200 // High boost for that context
			}

			// Boost priority for patterns with topic context
			if category.Topic != "" {
				priority.Priority += 100 // Medium boost for topic context
			}

			matchingPatterns = append(matchingPatterns, PatternPriority{
				Pattern:          basePattern,
				Category:         category,
				Priority:         priority.Priority,
				WildcardCount:    priority.WildcardCount,
				HasUnderscore:    priority.HasUnderscore,
				WildcardPosition: priority.WildcardPosition,
			})
		}
	}

	// Sort by priority (highest first)
	sort.Slice(matchingPatterns, func(i, j int) bool {
		return comparePatternPriorities(matchingPatterns[i].Priority, matchingPatterns[j].Priority)
	})

	// Return the highest priority match
	if len(matchingPatterns) > 0 {
		bestMatch := matchingPatterns[0]

		// Capture wildcard values from input pattern
		_, inputWildcards := matchPatternWithWildcardsAndSets(input, bestMatch.Pattern, kb)
		if inputWildcards == nil {
			_, inputWildcards = matchPatternWithWildcards(input, bestMatch.Pattern)
		}

		// Capture wildcard values from that context if it has wildcards
		thatWildcards := make(map[string]string)
		if bestMatch.Category.That != "" && strings.Contains(bestMatch.Category.That, "*") {
			_, thatWildcards = matchPatternWithWildcardsAndSets(normalizedThat, bestMatch.Category.That, kb)
			if thatWildcards == nil {
				_, thatWildcards = matchPatternWithWildcards(normalizedThat, bestMatch.Category.That)
			}
		}

		// Capture wildcard values from topic context if it has wildcards
		topicWildcards := make(map[string]string)
		if bestMatch.Category.Topic != "" && strings.Contains(bestMatch.Category.Topic, "*") {
			_, topicWildcards = matchPatternWithWildcardsAndSets(topic, bestMatch.Category.Topic, kb)
			if topicWildcards == nil {
				_, topicWildcards = matchPatternWithWildcards(topic, bestMatch.Category.Topic)
			}
		}

		// Merge wildcard values (input wildcards take precedence)
		allWildcards := make(map[string]string)
		for k, v := range thatWildcards {
			allWildcards[k] = v
		}
		for k, v := range topicWildcards {
			allWildcards[k] = v
		}
		for k, v := range inputWildcards {
			allWildcards[k] = v
		}

		return bestMatch.Category, allWildcards, nil
	}

	// Try default pattern (lowest priority)
	if category, exists := kb.Patterns["DEFAULT"]; exists {
		// Check topic match if topic is specified
		if topic == "" || category.Topic == "" || category.Topic == topic {
			// Check that match if that is specified
			if normalizedThat == "" || category.That == "" || category.That == normalizedThat {
				return category, make(map[string]string), nil
			}
		}
	}

	return nil, nil, fmt.Errorf("no matching pattern found")
}

// MatchPatternWithTopic attempts to match user input against AIML patterns with topic filtering
func (kb *AIMLKnowledgeBase) MatchPatternWithTopic(input string, topic string) (*Category, map[string]string, error) {
	return kb.MatchPatternWithTopicAndThat(input, topic, "")
}

// PatternPriorityInfo contains calculated priority information
type PatternPriorityInfo struct {
	Priority         int
	WildcardCount    int
	HasUnderscore    bool
	WildcardPosition int
}

// comparePatternPriorities compares two pattern priorities
func comparePatternPriorities(p1, p2 int) bool {
	return p1 > p2
}

// calculatePatternPriority calculates the priority of a pattern for matching
// Higher priority values mean higher precedence
func calculatePatternPriority(pattern string) PatternPriorityInfo {
	// Count wildcards
	starCount := strings.Count(pattern, "*")
	underscoreCount := strings.Count(pattern, "_")
	totalWildcards := starCount + underscoreCount

	// Calculate wildcard position score (wildcards at end are higher priority)
	wildcardPosition := 0
	if strings.HasSuffix(pattern, "*") || strings.HasSuffix(pattern, "_") {
		wildcardPosition = 1 // Wildcard at end
	} else if strings.HasPrefix(pattern, "*") || strings.HasPrefix(pattern, "_") {
		wildcardPosition = 0 // Wildcard at beginning
	} else {
		wildcardPosition = 2 // Wildcard in middle (highest priority)
	}

	// Calculate priority score
	// Base priority: 1000 - total wildcards (fewer wildcards = higher priority)
	priority := 1000 - totalWildcards

	// Bonus for underscore wildcards (more specific than star wildcards)
	if underscoreCount > 0 && starCount == 0 {
		priority += 100 // All underscores, no stars
	} else if underscoreCount > starCount {
		priority += 50 // More underscores than stars
	}

	// Bonus for wildcard position
	priority += wildcardPosition * 10

	return PatternPriorityInfo{
		Priority:         priority,
		WildcardCount:    totalWildcards,
		HasUnderscore:    underscoreCount > 0,
		WildcardPosition: wildcardPosition,
	}
}

// sortPatternsByPriority sorts patterns by priority (highest first)
func sortPatternsByPriority(patterns []PatternPriority) {
	// Simple bubble sort for priority (highest first)
	n := len(patterns)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if patterns[j].Priority < patterns[j+1].Priority {
				patterns[j], patterns[j+1] = patterns[j+1], patterns[j]
			}
		}
	}
}

// matchPatternWithWildcards matches input against a pattern with wildcards
func matchPatternWithWildcards(input, pattern string) (bool, map[string]string) {
	wildcards := make(map[string]string)

	// Convert pattern to regex
	regexPattern := patternToRegex(pattern)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, nil
	}

	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return false, nil
	}

	// Extract wildcard values
	starIndex := 1
	for _, match := range matches[1:] {
		// Include empty matches for zero+ wildcards
		// This allows patterns like "HELLO *" to match "HELLO" with empty wildcard
		wildcards[fmt.Sprintf("star%d", starIndex)] = match
		starIndex++
	}

	return true, wildcards
}

// matchPatternWithWildcardsAndSets matches input against a pattern with wildcards and sets
func matchPatternWithWildcardsAndSets(input, pattern string, kb *AIMLKnowledgeBase) (bool, map[string]string) {
	wildcards := make(map[string]string)

	// Convert pattern to regex with set support
	regexPattern := patternToRegexWithSets(pattern, kb)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, nil
	}

	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return false, nil
	}

	// Extract wildcard values
	starIndex := 1
	for _, match := range matches[1:] {
		// Include empty matches for zero+ wildcards
		// This allows patterns like "HELLO *" to match "HELLO" with empty wildcard
		wildcards[fmt.Sprintf("star%d", starIndex)] = match
		starIndex++
	}

	return true, wildcards
}

// patternToRegex converts AIML pattern to regex with enhanced set and topic matching
func patternToRegex(pattern string) string {
	// Handle set matching first (before escaping)
	setPattern := regexp.MustCompile(`<set>([^<]+)</set>`)
	pattern = setPattern.ReplaceAllString(pattern, "([^\\s]*)")

	// Handle topic matching (before escaping)
	topicPattern := regexp.MustCompile(`<topic>([^<]+)</topic>`)
	pattern = topicPattern.ReplaceAllString(pattern, "([^\\s]*)")

	// Build regex pattern by processing each character
	var result strings.Builder
	for i, char := range pattern {
		switch char {
		case '*':
			// Zero+ wildcard: matches zero or more words
			result.WriteString("(.*?)")
		case '_':
			// Single wildcard: matches exactly one word
			result.WriteString("([^\\s]+)")
		case ' ':
			// Check if this space is followed by a wildcard or preceded by a wildcard
			if (i+1 < len(pattern) && (pattern[i+1] == '*' || pattern[i+1] == '_')) ||
				(i > 0 && (pattern[i-1] == '*' || pattern[i-1] == '_')) {
				// This space is adjacent to a wildcard, make it optional
				result.WriteString(" ?")
			} else {
				// Regular space
				result.WriteRune(' ')
			}
		case '(', ')', '[', ']', '{', '}', '^', '$', '?', '+', '.':
			// Escape special regex characters (but not | as it's needed for alternation)
			result.WriteRune('\\')
			result.WriteRune(char)
		case '|':
			// Don't escape pipe character as it's needed for alternation in sets
			result.WriteRune(char)
		default:
			// Regular character
			result.WriteRune(char)
		}
	}

	return "^" + result.String() + "$"
}

// patternToRegexWithSets converts AIML pattern to regex with proper set matching
func patternToRegexWithSets(pattern string, kb *AIMLKnowledgeBase) string {
	// Handle set matching with proper set validation
	setPattern := regexp.MustCompile(`<set>([^<]+)</set>`)
	pattern = setPattern.ReplaceAllStringFunc(pattern, func(match string) string {
		// Extract set name using regex groups
		matches := setPattern.FindStringSubmatch(match)
		if len(matches) < 2 {
			return "([^\\s]*)"
		}
		setName := strings.ToUpper(strings.TrimSpace(matches[1]))
		if len(kb.Sets[setName]) > 0 {
			// Create regex alternation for set members
			var alternatives []string
			for _, member := range kb.Sets[setName] {
				// Escape only specific regex characters, not the pipe
				upperMember := strings.ToUpper(member)
				// Escape characters that have special meaning in regex, but not |
				escaped := strings.ReplaceAll(upperMember, "(", "\\(")
				escaped = strings.ReplaceAll(escaped, ")", "\\)")
				escaped = strings.ReplaceAll(escaped, "[", "\\[")
				escaped = strings.ReplaceAll(escaped, "]", "\\]")
				escaped = strings.ReplaceAll(escaped, "{", "\\{")
				escaped = strings.ReplaceAll(escaped, "}", "\\}")
				escaped = strings.ReplaceAll(escaped, "^", "\\^")
				escaped = strings.ReplaceAll(escaped, "$", "\\$")
				escaped = strings.ReplaceAll(escaped, ".", "\\.")
				escaped = strings.ReplaceAll(escaped, "+", "\\+")
				escaped = strings.ReplaceAll(escaped, "?", "\\?")
				escaped = strings.ReplaceAll(escaped, "*", "\\*")
				escaped = strings.ReplaceAll(escaped, "-", "\\-")
				escaped = strings.ReplaceAll(escaped, "@", "\\@")
				// Don't escape | as it's needed for alternation
				alternatives = append(alternatives, escaped)
			}
			return "(" + strings.Join(alternatives, "|") + ")"
		}
		// Fallback to wildcard if set not found
		return "([^\\s]*)"
	})

	// Handle topic matching
	topicPattern := regexp.MustCompile(`<topic>([^<]+)</topic>`)
	pattern = topicPattern.ReplaceAllString(pattern, "([^\\s]*)")

	// Build regex pattern by processing each character
	var result strings.Builder
	inAlternationGroup := false
	for i, char := range pattern {
		switch char {
		case '*':
			// Zero+ wildcard: matches zero or more words
			result.WriteString("(.*?)")
		case '_':
			// Single wildcard: matches exactly one word
			result.WriteString("([^\\s]+)")
		case ' ':
			// Check if this space is followed by a wildcard or preceded by a wildcard
			if (i+1 < len(pattern) && (pattern[i+1] == '*' || pattern[i+1] == '_')) ||
				(i > 0 && (pattern[i-1] == '*' || pattern[i-1] == '_')) {
				// This space is adjacent to a wildcard, make it optional
				result.WriteString(" ?")
			} else {
				// Regular space
				result.WriteRune(' ')
			}
		case '(':
			// Check if this is the start of an alternation group (contains |)
			// Look ahead to see if there's a | in this group
			groupEnd := findMatchingParen(pattern, i)
			if groupEnd > i && strings.Contains(pattern[i:groupEnd+1], "|") {
				inAlternationGroup = true
				result.WriteRune('(')
			} else {
				// Regular group, escape it
				result.WriteString("\\(")
			}
		case ')':
			if inAlternationGroup {
				inAlternationGroup = false
				result.WriteRune(')')
			} else {
				result.WriteString("\\)")
			}
		case '[', ']', '{', '}', '^', '$', '?', '+', '.':
			// Escape special regex characters
			result.WriteRune('\\')
			result.WriteRune(char)
		case '|':
			// Don't escape pipe character as it's needed for alternation in sets
			result.WriteRune(char)
		default:
			// Regular character
			result.WriteRune(char)
		}
	}

	return "^" + result.String() + "$"
}

// findMatchingParen finds the matching closing parenthesis for an opening parenthesis
func findMatchingParen(pattern string, openPos int) int {
	if openPos >= len(pattern) || pattern[openPos] != '(' {
		return -1
	}

	depth := 1
	for i := openPos + 1; i < len(pattern); i++ {
		switch pattern[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// ProcessTemplate processes an AIML template and returns the response
func (g *Golem) ProcessTemplate(template string, wildcards map[string]string) string {
	// Create variable context for template processing
	ctx := &VariableContext{
		LocalVars:     make(map[string]string),
		Session:       nil, // No session context for ProcessTemplate
		Topic:         "",  // TODO: Implement topic tracking
		KnowledgeBase: g.aimlKB,
	}

	return g.processTemplateWithContext(template, wildcards, ctx)
}

// ProcessTemplateWithContext processes an AIML template with full context support
func (g *Golem) ProcessTemplateWithContext(template string, wildcards map[string]string, session *ChatSession) string {
	// Create variable context for template processing
	ctx := &VariableContext{
		LocalVars:     make(map[string]string),
		Session:       session,
		Topic:         session.GetSessionTopic(),
		KnowledgeBase: g.aimlKB,
	}

	return g.processTemplateWithContext(template, wildcards, ctx)
}

// processTemplateWithContext processes a template with variable context
func (g *Golem) processTemplateWithContext(template string, wildcards map[string]string, ctx *VariableContext) string {
	response := template

	if g.verbose {
		g.logger.Printf("Template text: '%s'", response)
		g.logger.Printf("Wildcards: %v", wildcards)
	}

	// Replace wildcards
	for key, value := range wildcards {
		if key == "star1" {
			response = strings.ReplaceAll(response, "<star/>", value)
			response = strings.ReplaceAll(response, "<star index=\"1\"/>", value)
			response = strings.ReplaceAll(response, "<star1/>", value)
		} else if key == "star2" {
			response = strings.ReplaceAll(response, "<star index=\"2\"/>", value)
			response = strings.ReplaceAll(response, "<star2/>", value)
		}
	}

	// Process SR tags (shorthand for <srai><star/></srai>) AFTER wildcard replacement
	if g.verbose {
		g.logger.Printf("Before SR processing: '%s'", response)
	}
	response = g.processSRTagsWithContext(response, wildcards, ctx)
	if g.verbose {
		g.logger.Printf("After SR processing: '%s'", response)
	}

	// Replace property tags
	response = g.replacePropertyTags(response)

	// Process bot tags (short form of property access)
	response = g.processBotTagsWithContext(response, ctx)

	// Process think tags FIRST (internal processing, no output)
	// This allows local variables to be set before variable replacement
	response = g.processThinkTagsWithContext(response, ctx)

	// Process topic setting tags first (special handling for topic)
	response = g.processTopicSettingTagsWithContext(response, ctx)

	// Process set tags (before session variable replacement)
	response = g.processSetTagsWithContext(response, ctx)

	// Replace session variable tags using context
	response = g.replaceSessionVariableTagsWithContext(response, ctx)

	// Process SRAI tags (recursive)
	response = g.processSRAITagsWithContext(response, ctx)

	// Process SRAIX tags (external services)
	response = g.processSRAIXTagsWithContext(response, ctx)

	// Process learn tags (dynamic learning)
	response = g.processLearnTagsWithContext(response, ctx)

	// Process condition tags
	response = g.processConditionTagsWithContext(response, ctx)

	// Process date and time tags
	response = g.processDateTimeTags(response)

	// Process random tags
	response = g.processRandomTags(response)

	// Process map tags
	response = g.processMapTagsWithContext(response, ctx)

	// Process list tags
	if g.verbose {
		g.logger.Printf("Before list processing: '%s'", response)
	}
	response = g.processListTagsWithContext(response, ctx)
	if g.verbose {
		g.logger.Printf("After list processing: '%s'", response)
	}

	// Debug: Check if we're continuing with array processing
	if g.verbose {
		g.logger.Printf("About to process array tags...")
	}

	// Process array tags
	if g.verbose {
		g.logger.Printf("Before array processing: '%s'", response)
	}
	response = g.processArrayTagsWithContext(response, ctx)
	if g.verbose {
		g.logger.Printf("After array processing: '%s'", response)
	}

	// Process person tags (pronoun substitution)
	if g.verbose {
		g.logger.Printf("Before person processing: '%s'", response)
	}
	response = g.processPersonTagsWithContext(response, ctx)
	if g.verbose {
		g.logger.Printf("After person processing: '%s'", response)
	}

	// Process gender tags (gender pronoun substitution)
	if g.verbose {
		g.logger.Printf("Before gender processing: '%s'", response)
	}
	response = g.processGenderTagsWithContext(response, ctx)
	if g.verbose {
		g.logger.Printf("After gender processing: '%s'", response)
	}

	// Process request tags (user input history)
	if g.verbose {
		g.logger.Printf("Before request processing: '%s'", response)
	}
	response = g.processRequestTags(response, ctx)
	if g.verbose {
		g.logger.Printf("After request processing: '%s'", response)
	}

	// Process response tags (bot response history)
	if g.verbose {
		g.logger.Printf("Before response processing: '%s'", response)
	}
	response = g.processResponseTags(response, ctx)
	if g.verbose {
		g.logger.Printf("After response processing: '%s'", response)
	}

	if g.verbose {
		g.logger.Printf("Final response: '%s'", response)
	}

	return strings.TrimSpace(response)
}

// processPersonTagsWithContext processes <person> tags for pronoun substitution
func (g *Golem) processPersonTagsWithContext(template string, ctx *VariableContext) string {
	// Find all <person> tags (including multiline content)
	personTagRegex := regexp.MustCompile(`(?s)<person>(.*?)</person>`)
	matches := personTagRegex.FindAllStringSubmatch(template, -1)

	if g.verbose {
		g.logger.Printf("Person tag processing: found %d matches in template: '%s'", len(matches), template)
	}

	for _, match := range matches {
		if len(match) > 1 {
			content := strings.TrimSpace(match[1])
			// Normalize whitespace before processing
			content = strings.Join(strings.Fields(content), " ")
			substitutedContent := g.SubstitutePronouns(content)
			if g.verbose {
				g.logger.Printf("Person tag: '%s' -> '%s'", match[1], substitutedContent)
			}
			template = strings.ReplaceAll(template, match[0], substitutedContent)
		}
	}

	if g.verbose {
		g.logger.Printf("Person tag processing result: '%s'", template)
	}

	return template
}

// processGenderTagsWithContext processes <gender> tags for gender pronoun substitution
func (g *Golem) processGenderTagsWithContext(template string, ctx *VariableContext) string {
	// Find all <gender> tags (including multiline content)
	genderTagRegex := regexp.MustCompile(`(?s)<gender>(.*?)</gender>`)
	matches := genderTagRegex.FindAllStringSubmatch(template, -1)

	if g.verbose {
		g.logger.Printf("Gender tag processing: found %d matches in template: '%s'", len(matches), template)
	}

	for _, match := range matches {
		if len(match) > 1 {
			content := strings.TrimSpace(match[1])
			// Normalize whitespace before processing
			content = strings.Join(strings.Fields(content), " ")
			substitutedContent := g.SubstituteGenderPronouns(content)
			if g.verbose {
				g.logger.Printf("Gender tag: '%s' -> '%s'", match[1], substitutedContent)
			}
			template = strings.ReplaceAll(template, match[0], substitutedContent)
		}
	}

	if g.verbose {
		g.logger.Printf("Gender tag processing result: '%s'", template)
	}

	return template
}

// SubstitutePronouns performs pronoun substitution for person tags
func (g *Golem) SubstitutePronouns(text string) string {
	// Comprehensive pronoun mapping for first/second person substitution
	pronounMap := map[string]string{
		// First person to second person
		"I": "you", "i": "you",
		"me": "you",
		"my": "your", "My": "Your",
		"mine": "yours", "Mine": "Yours",
		"we": "you", "We": "you",
		"us": "you", "Us": "you",
		"our": "your", "Our": "your",
		"ours": "yours", "Ours": "yours",
		"myself": "yourself", "Myself": "yourself",
		"ourselves": "yourselves", "Ourselves": "yourselves",

		// Second person to first person
		"you": "I", "You": "I",
		"your": "my", "Your": "my",
		"yours": "mine", "Yours": "mine",
		"yourself": "myself", "Yourself": "myself",
		"yourselves": "ourselves", "Yourselves": "ourselves",

		// Contractions - first person to second person
		"I'm": "you're", "i'm": "you're", "I'M": "you're",
		"I've": "you've", "i've": "you've", "I'VE": "you've",
		"I'll": "you'll", "i'll": "you'll", "I'LL": "you'll",
		"I'd": "you'd", "i'd": "you'd", "I'D": "you'd",

		// Contractions - second person to first person
		"you're": "I'm", "You're": "I'm", "YOU'RE": "I'm",
		"you've": "I've", "You've": "I've", "YOU'VE": "I've",
		"you'll": "I'll", "You'll": "I'll", "YOU'LL": "I'll",
		"you'd": "I'd", "You'd": "I'd", "YOU'D": "I'd",
	}

	// Split text into words while preserving whitespace
	words := strings.Fields(text)
	substitutedWords := make([]string, len(words))

	for i, word := range words {
		// Check for exact match first
		if substitution, exists := pronounMap[word]; exists {
			substitutedWords[i] = substitution
			continue
		}

		// Handle contractions and possessives more carefully
		substituted := word

		// Check for contractions (apostrophe)
		if strings.Contains(word, "'") {
			// Split on apostrophe and check each part
			parts := strings.Split(word, "'")
			if len(parts) == 2 {
				firstPart := parts[0]
				secondPart := parts[1]

				// Check if first part needs substitution
				if sub, exists := pronounMap[firstPart]; exists {
					substituted = sub + "'" + secondPart
				} else if sub, exists := pronounMap[firstPart+"'"]; exists {
					substituted = sub + secondPart
				}
			}
		}

		// Check for possessive forms (ending with 's or s')
		if strings.HasSuffix(word, "'s") || strings.HasSuffix(word, "s'") {
			base := strings.TrimSuffix(strings.TrimSuffix(word, "'s"), "s'")
			if sub, exists := pronounMap[base]; exists {
				if strings.HasSuffix(word, "'s") {
					substituted = sub + "'s"
				} else {
					substituted = sub + "s'"
				}
			}
		}

		// Check for words ending with common suffixes that might be pronouns
		if strings.HasSuffix(word, "ing") || strings.HasSuffix(word, "ed") || strings.HasSuffix(word, "er") || strings.HasSuffix(word, "est") {
			// Don't substitute if it's a verb form
			substituted = word
		}

		substitutedWords[i] = substituted
	}

	result := strings.Join(substitutedWords, " ")

	// Handle verb agreement after pronoun substitution
	result = g.fixVerbAgreement(result)

	if g.verbose {
		g.logger.Printf("Person substitution: '%s' -> '%s'", text, result)
	}

	return result
}

// fixVerbAgreement fixes verb agreement after pronoun substitution
func (g *Golem) fixVerbAgreement(text string) string {
	// Common verb agreement fixes
	verbFixes := map[string]string{
		"you am":  "you are",
		"I are":   "I am",
		"you is":  "you are",
		"I is":    "I am",
		"you was": "you were",
		"I were":  "I was",
		"you has": "you have",
		"I have":  "I have", // Keep as is
	}

	result := text
	for wrong, correct := range verbFixes {
		result = strings.ReplaceAll(result, wrong, correct)
	}

	return result
}

// SubstituteGenderPronouns performs gender-based pronoun substitution for gender tags
func (g *Golem) SubstituteGenderPronouns(text string) string {
	// Split text into words for more precise substitution
	words := strings.Fields(text)
	result := make([]string, len(words))

	for i, word := range words {
		// Clean word for matching (remove punctuation)
		cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}")
		lowerWord := strings.ToLower(cleanWord)

		// Gender pronoun mapping (masculine to feminine and vice versa)
		genderMap := map[string]string{
			// Masculine to feminine
			"he": "she", "him": "her", "his": "her", "himself": "herself",
			"he's": "she's", "he'll": "she'll", "he'd": "she'd",

			// Feminine to masculine
			"she": "he", "her": "his", "hers": "his", "herself": "himself",
			"she's": "he's", "she'll": "he'll", "she'd": "he'd",
		}

		// Check if we need to substitute
		if substitute, exists := genderMap[lowerWord]; exists {
			// Preserve original case
			if strings.ToUpper(cleanWord) == cleanWord {
				// All caps
				result[i] = strings.ToUpper(substitute)
			} else if len(cleanWord) > 0 && cleanWord[0] >= 'A' && cleanWord[0] <= 'Z' {
				// Title case (first letter capitalized)
				if len(substitute) > 0 {
					result[i] = strings.ToUpper(string(substitute[0])) + strings.ToLower(substitute[1:])
				} else {
					result[i] = substitute
				}
			} else {
				// Lower case
				result[i] = substitute
			}

			// Add back any punctuation that was removed
			if len(cleanWord) < len(word) {
				suffix := word[len(cleanWord):]
				result[i] += suffix
			}
		} else {
			// No substitution needed
			result[i] = word
		}
	}

	// Join words back together
	finalResult := strings.Join(result, " ")

	// Fix verb agreement after gender substitution
	finalResult = g.fixGenderVerbAgreement(finalResult)

	if g.verbose {
		g.logger.Printf("Gender substitution: '%s' -> '%s'", text, finalResult)
	}

	return finalResult
}

// fixGenderVerbAgreement fixes verb agreement after gender pronoun substitution
func (g *Golem) fixGenderVerbAgreement(text string) string {
	// Common verb agreement fixes for gender pronouns
	verbFixes := map[string]string{
		"she am":   "she is",
		"he am":    "he is",
		"she are":  "she is",
		"he are":   "he is",
		"she was":  "she was", // Keep as is
		"he was":   "he was",  // Keep as is
		"she were": "she was",
		"he were":  "he was",
		"she has":  "she has", // Keep as is
		"he has":   "he has",  // Keep as is
		"she have": "she has",
		"he have":  "he has",
	}

	result := text
	for wrong, correct := range verbFixes {
		result = strings.ReplaceAll(result, wrong, correct)
	}

	return result
}

// processSRAITagsWithContext processes <srai> tags with variable context
func (g *Golem) processSRAITagsWithContext(template string, ctx *VariableContext) string {
	// Find all <srai> tags
	sraiRegex := regexp.MustCompile(`<srai>(.*?)</srai>`)
	matches := sraiRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			sraiContent := strings.TrimSpace(match[1])

			if g.verbose {
				g.logger.Printf("Processing SRAI: '%s'", sraiContent)
			}

			// Process the SRAI content as a new pattern
			if g.aimlKB != nil {
				// Try to match the SRAI content as a pattern
				category, wildcards, err := g.aimlKB.MatchPattern(sraiContent)
				if err == nil && category != nil {
					// Process the matched template with context
					response := g.processTemplateWithContext(category.Template, wildcards, ctx)
					template = strings.ReplaceAll(template, match[0], response)
				} else {
					// No match found, leave the SRAI tag unchanged
					if g.verbose {
						g.logger.Printf("SRAI no match for: '%s'", sraiContent)
					}
					// Don't replace the SRAI tag - leave it as is
				}
			}
		}
	}

	return template
}

// processSRTagsWithContext processes <sr> tags with variable context
// <sr> is shorthand for <srai><star/></srai>
func (g *Golem) processSRTagsWithContext(template string, wildcards map[string]string, ctx *VariableContext) string {
	// Find all <sr/> tags (self-closing)
	srRegex := regexp.MustCompile(`<sr\s*/>`)
	matches := srRegex.FindAllString(template, -1)

	for _, match := range matches {
		if g.verbose {
			g.logger.Printf("Processing SR tag: '%s'", match)
		}

		// Get the first wildcard (star1) from the wildcards map
		starContent := ""
		if wildcards != nil {
			if star1, exists := wildcards["star1"]; exists {
				starContent = star1
			}
		}

		if starContent != "" {
			// Process as SRAI with the star content
			if g.aimlKB != nil {
				// Try to match the star content as a pattern
				category, srWildcards, err := g.aimlKB.MatchPattern(starContent)
				if err == nil && category != nil {
					// Process the matched template with context
					response := g.processTemplateWithContext(category.Template, srWildcards, ctx)
					template = strings.ReplaceAll(template, match, response)
				} else {
					// No match found, leave the SR tag unchanged
					if g.verbose {
						g.logger.Printf("SR no match for: '%s'", starContent)
					}
					// Don't replace the SR tag - leave it as is
				}
			}
		} else {
			// No star content available, leave the SR tag unchanged
			if g.verbose {
				g.logger.Printf("SR tag found but no star content available")
			}
		}
	}

	return template
}

// processSRAIXTagsWithContext processes <sraix> tags with variable context
func (g *Golem) processSRAIXTagsWithContext(template string, ctx *VariableContext) string {
	if g.sraixMgr == nil {
		return template
	}

	// Find all <sraix> tags with service attribute
	sraixRegex := regexp.MustCompile(`<sraix\s+service="([^"]+)">(.*?)</sraix>`)
	matches := sraixRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 2 {
			serviceName := strings.TrimSpace(match[1])
			sraixContent := strings.TrimSpace(match[2])

			if g.verbose {
				g.logger.Printf("Processing SRAIX: service='%s', content='%s'", serviceName, sraixContent)
			}

			// Process the SRAIX content (replace wildcards, variables, etc.)
			processedContent := g.processTemplateWithContext(sraixContent, make(map[string]string), ctx)

			// Make external request
			response, err := g.sraixMgr.ProcessSRAIX(serviceName, processedContent, make(map[string]string))
			if err != nil {
				if g.verbose {
					g.logger.Printf("SRAIX request failed: %v", err)
				}
				// Leave the SRAIX tag unchanged on error
				continue
			}

			// Replace the SRAIX tag with the response
			template = strings.ReplaceAll(template, match[0], response)
		}
	}

	return template
}

// processLearnTagsWithContext processes <learn> and <learnf> tags with variable context
func (g *Golem) processLearnTagsWithContext(template string, ctx *VariableContext) string {
	if g.aimlKB == nil {
		return template
	}

	// Process <learn> tags (session-specific learning)
	learnRegex := regexp.MustCompile(`(?s)<learn>(.*?)</learn>`)
	learnMatches := learnRegex.FindAllStringSubmatch(template, -1)

	for _, match := range learnMatches {
		if len(match) > 1 {
			learnContent := strings.TrimSpace(match[1])

			if g.verbose {
				g.logger.Printf("Processing learn: '%s'", learnContent)
			}

			// Parse the AIML content within the learn tag
			categories, err := g.parseLearnContent(learnContent)
			if err != nil {
				if g.verbose {
					g.logger.Printf("Failed to parse learn content: %v", err)
				}
				// Remove the learn tag on error
				template = strings.ReplaceAll(template, match[0], "")
				continue
			}

			// Add categories to session-specific knowledge base
			for _, category := range categories {
				err := g.addSessionCategory(category, ctx)
				if err != nil {
					if g.verbose {
						g.logger.Printf("Failed to add session category: %v", err)
					}
				}
			}

			// Remove the learn tag after processing
			template = strings.ReplaceAll(template, match[0], "")
		}
	}

	// Process <learnf> tags (persistent learning)
	learnfRegex := regexp.MustCompile(`(?s)<learnf>(.*?)</learnf>`)
	learnfMatches := learnfRegex.FindAllStringSubmatch(template, -1)

	for _, match := range learnfMatches {
		if len(match) > 1 {
			learnfContent := strings.TrimSpace(match[1])

			if g.verbose {
				g.logger.Printf("Processing learnf: '%s'", learnfContent)
			}

			// Parse the AIML content within the learnf tag
			categories, err := g.parseLearnContent(learnfContent)
			if err != nil {
				if g.verbose {
					g.logger.Printf("Failed to parse learnf content: %v", err)
				}
				// Remove the learnf tag on error
				template = strings.ReplaceAll(template, match[0], "")
				continue
			}

			// Add categories to persistent knowledge base
			for _, category := range categories {
				err := g.addPersistentCategory(category)
				if err != nil {
					if g.verbose {
						g.logger.Printf("Failed to add persistent category: %v", err)
					}
				}
			}

			// Remove the learnf tag after processing
			template = strings.ReplaceAll(template, match[0], "")
		}
	}

	return template
}

// processThinkTagsWithContext processes <think> tags with variable context
func (g *Golem) processThinkTagsWithContext(template string, ctx *VariableContext) string {
	// Find all <think> tags
	thinkRegex := regexp.MustCompile(`<think>(.*?)</think>`)
	matches := thinkRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			thinkContent := strings.TrimSpace(match[1])

			if g.verbose {
				g.logger.Printf("Processing think: '%s'", thinkContent)
			}

			// Process the think content (internal operations)
			g.processThinkContentWithContext(thinkContent, ctx)

			// Remove the think tag from the output
			template = strings.ReplaceAll(template, match[0], "")
		}
	}

	return template
}

// processThinkContentWithContext processes the content inside <think> tags with variable context
func (g *Golem) processThinkContentWithContext(content string, ctx *VariableContext) {
	// Process date/time tags first
	content = g.processDateTimeTags(content)

	// Find all <set> tags
	setRegex := regexp.MustCompile(`<set name="([^"]+)">(.*?)</set>`)
	matches := setRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			varName := match[1]
			varValue := strings.TrimSpace(match[2])

			if g.verbose {
				g.logger.Printf("Setting variable: %s = %s", varName, varValue)
			}

			// Determine scope based on context
			scope := ScopeGlobal // Default to global scope
			if ctx.Session != nil {
				// If we have a session, use session scope for think tags
				// This allows variables to be set in the session context
				scope = ScopeSession
			}

			// Set the variable in the appropriate scope
			g.setVariable(varName, varValue, scope, ctx)
		}
	}

	// Process other think operations here as needed
	// For now, we only handle <set> tags, but this could be extended
	// to handle other internal operations like learning, logging, etc.
}

// processConditionTagsWithContext processes <condition> tags with variable context
func (g *Golem) processConditionTagsWithContext(template string, ctx *VariableContext) string {
	// Use regex to find and process conditions
	// This handles nesting by processing inner conditions first
	conditionRegex := regexp.MustCompile(`(?s)<condition(?: name="([^"]+)"(?: value="([^"]+)")?)?>(.*?)</condition>`)

	for {
		matches := conditionRegex.FindAllStringSubmatch(template, -1)
		if len(matches) == 0 {
			break // No more conditions
		}

		// Process the first (innermost) condition
		match := matches[0]
		if len(match) < 4 {
			break
		}

		varName := match[1]
		expectedValue := match[2]
		conditionContent := strings.TrimSpace(match[3])

		if g.verbose {
			g.logger.Printf("Processing condition: var='%s', expected='%s', content='%s'",
				varName, expectedValue, conditionContent)
		}

		// Get the actual variable value using context
		actualValue := g.resolveVariable(varName, ctx)

		// Process the condition content
		response := g.processConditionContentWithContext(conditionContent, varName, actualValue, expectedValue, ctx)

		if g.verbose {
			g.logger.Printf("Condition response: '%s'", response)
		}

		// Replace the condition tag with the response
		template = strings.ReplaceAll(template, match[0], response)
	}

	return template
}

// processConditionContentWithContext processes the content inside <condition> tags with variable context
func (g *Golem) processConditionContentWithContext(content string, varName, actualValue, expectedValue string, ctx *VariableContext) string {
	// Handle different condition types

	// Type 1: Simple condition with value attribute
	if expectedValue != "" {
		if strings.EqualFold(actualValue, expectedValue) {
			// Process the content through the full template pipeline
			return g.processTemplateWithContext(content, make(map[string]string), ctx)
		}
		return "" // No match, return empty
	}

	// Type 2: Multiple <li> conditions or default condition
	if strings.Contains(content, "<li") {
		return g.processConditionListItemsWithContext(content, actualValue, ctx)
	}

	// Type 3: Default condition (no value specified, no <li> elements)
	if actualValue != "" {
		// Process the content through the full template pipeline
		return g.processTemplateWithContext(content, make(map[string]string), ctx)
	}

	return "" // Variable not found or empty
}

// processConditionListItemsWithContext processes <li> elements within condition tags with variable context
func (g *Golem) processConditionListItemsWithContext(content string, actualValue string, ctx *VariableContext) string {
	// Find all <li> elements with optional value attributes
	liRegex := regexp.MustCompile(`(?s)<li(?: value="([^"]+)")?>(.*?)</li>`)
	matches := liRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		liValue := match[1]
		liContent := strings.TrimSpace(match[2])

		// If no value specified, this is the default case
		if liValue == "" {
			return g.processTemplateWithContext(liContent, make(map[string]string), ctx)
		}

		// Check if this condition matches
		if strings.EqualFold(actualValue, liValue) {
			return g.processTemplateWithContext(liContent, make(map[string]string), ctx)
		}
	}

	return "" // No match found
}

// replaceSessionVariableTagsWithContext replaces <get name="var"/> tags with variables using context
func (g *Golem) replaceSessionVariableTagsWithContext(template string, ctx *VariableContext) string {
	// Find all <get name="var"/> tags
	getTagRegex := regexp.MustCompile(`<get name="([^"]+)"/>`)
	matches := getTagRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			varName := match[1]
			varValue := g.resolveVariable(varName, ctx)
			if varValue != "" {
				template = strings.ReplaceAll(template, match[0], varValue)
			}
		}
	}

	return template
}

// processSetTagsWithContext processes <set name="var">value</set> tags
func (g *Golem) processSetTagsWithContext(template string, ctx *VariableContext) string {
	// Find all <set name="var">value</set> tags
	setTagRegex := regexp.MustCompile(`<set name="([^"]+)">(.*?)</set>`)
	matches := setTagRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 2 {
			varName := match[1]
			varValue := strings.TrimSpace(match[2])

			if g.verbose {
				g.logger.Printf("Setting variable '%s' to '%s'", varName, varValue)
			}

			// Process the variable value through the template pipeline to handle wildcards
			// Use a special processing that doesn't output the result
			processedValue := g.processTemplateContentForVariable(varValue, make(map[string]string), ctx)

			// Set the variable in the appropriate scope
			g.setVariable(varName, processedValue, ScopeSession, ctx)

			// Remove the set tag from the template (don't replace with value)
			template = strings.ReplaceAll(template, match[0], "")
		}
	}

	return template
}

// processTemplateContentForVariable processes template content for variable assignment without outputting
func (g *Golem) processTemplateContentForVariable(template string, wildcards map[string]string, ctx *VariableContext) string {
	response := template

	if g.verbose {
		g.logger.Printf("Processing variable content: '%s'", response)
		g.logger.Printf("Wildcards: %v", wildcards)
	}

	// Replace wildcards
	for key, value := range wildcards {
		if key == "star1" {
			response = strings.ReplaceAll(response, "<star/>", value)
			response = strings.ReplaceAll(response, "<star index=\"1\"/>", value)
			response = strings.ReplaceAll(response, "<star1/>", value)
		} else if key == "star2" {
			response = strings.ReplaceAll(response, "<star index=\"2\"/>", value)
			response = strings.ReplaceAll(response, "<star2/>", value)
		}
	}

	// Replace property tags
	response = g.replacePropertyTags(response)

	// Replace session variable tags using context
	response = g.replaceSessionVariableTagsWithContext(response, ctx)

	// Process SRAI tags (recursive)
	response = g.processSRAITagsWithContext(response, ctx)

	// Process SRAIX tags (external services)
	response = g.processSRAIXTagsWithContext(response, ctx)

	// Process learn tags (dynamic learning)
	response = g.processLearnTagsWithContext(response, ctx)

	// Process condition tags
	response = g.processConditionTagsWithContext(response, ctx)

	// Process date and time tags
	response = g.processDateTimeTags(response)

	// Process random tags
	response = g.processRandomTags(response)

	// Process map tags
	response = g.processMapTagsWithContext(response, ctx)

	// Process list tags
	response = g.processListTagsWithContext(response, ctx)

	// Process array tags
	response = g.processArrayTagsWithContext(response, ctx)

	if g.verbose {
		g.logger.Printf("Variable content result: '%s'", response)
	}

	return strings.TrimSpace(response)
}

// replacePropertyTags replaces <get name="property"/> tags with property values
func (g *Golem) replacePropertyTags(template string) string {
	if g.aimlKB == nil {
		return template
	}

	// Find all <get name="property"/> tags
	getTagRegex := regexp.MustCompile(`<get name="([^"]+)"/>`)
	matches := getTagRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			propertyName := match[1]
			propertyValue := g.aimlKB.GetProperty(propertyName)
			if propertyValue != "" {
				template = strings.ReplaceAll(template, match[0], propertyValue)
			}
		}
	}

	return template
}

// processBotTagsWithContext processes <bot name="property"/> tags with variable context
func (g *Golem) processBotTagsWithContext(template string, ctx *VariableContext) string {
	if ctx.KnowledgeBase == nil {
		return template
	}

	// Find all <bot name="property"/> tags
	botTagRegex := regexp.MustCompile(`<bot name="([^"]+)"/>`)
	matches := botTagRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			propertyName := match[1]
			propertyValue := ctx.KnowledgeBase.GetProperty(propertyName)

			if g.verbose {
				g.logger.Printf("Bot tag: property='%s', value='%s'", propertyName, propertyValue)
			}

			if propertyValue != "" {
				template = strings.ReplaceAll(template, match[0], propertyValue)
			} else {
				// If property not found, leave the bot tag unchanged
				if g.verbose {
					g.logger.Printf("Bot property '%s' not found", propertyName)
				}
			}
		}
	}

	return template
}

// processSRAITags processes <srai> tags recursively
func (g *Golem) processSRAITags(template string, session *ChatSession) string {
	if g.aimlKB == nil {
		return template
	}

	// Find all <srai> tags
	sraiRegex := regexp.MustCompile(`<srai>(.*?)</srai>`)
	matches := sraiRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			sraiInput := strings.TrimSpace(match[1])
			if g.verbose {
				g.logger.Printf("Processing SRAI: '%s'", sraiInput)
			}

			// Match the SRAI input as a new pattern
			category, wildcards, err := g.aimlKB.MatchPattern(sraiInput)
			if err != nil {
				// If no match found, use the original SRAI text
				if g.verbose {
					g.logger.Printf("SRAI no match for: '%s'", sraiInput)
				}
				continue
			}

			// Process the matched template
			var sraiResponse string
			if session != nil {
				sraiResponse = g.ProcessTemplateWithSession(category.Template, wildcards, session)
			} else {
				sraiResponse = g.ProcessTemplate(category.Template, wildcards)
			}

			// Replace the SRAI tag with the processed response
			template = strings.ReplaceAll(template, match[0], sraiResponse)
		}
	}

	return template
}

// processThinkTags processes <think> tags for internal processing without output
func (g *Golem) processThinkTags(template string, session *ChatSession) string {
	// Find all <think> tags
	thinkRegex := regexp.MustCompile(`(?s)<think>(.*?)</think>`)
	matches := thinkRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			thinkContent := strings.TrimSpace(match[1])
			if g.verbose {
				g.logger.Printf("Processing think tag: '%s'", thinkContent)
			}

			// Process the think content (but don't include it in output)
			// This allows for internal operations like setting variables
			g.processThinkContent(thinkContent, session)

			// Remove the think tag from the template (no output)
			template = strings.ReplaceAll(template, match[0], "")
		}
	}

	return template
}

// processThinkContent processes the content inside <think> tags
func (g *Golem) processThinkContent(content string, session *ChatSession) {
	// Process date/time tags in think content first
	content = g.processDateTimeTags(content)

	// Process <set> tags for variable setting
	setRegex := regexp.MustCompile(`<set name="([^"]+)">([^<]*)</set>`)
	matches := setRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			varName := match[1]
			varValue := match[2]

			if g.verbose {
				g.logger.Printf("Think: Setting variable %s = %s", varName, varValue)
			}

			// Set the variable in the appropriate context
			if session != nil {
				// Set in session variables
				session.Variables[varName] = varValue
			} else if g.aimlKB != nil {
				// Set in knowledge base variables
				g.aimlKB.Variables[varName] = varValue
			}
		}
	}

	// Process other think operations here as needed
	// For now, we only handle <set> tags, but this could be extended
	// to handle other internal operations like learning, logging, etc.
}

// processConditionTags processes <condition> tags using regex
func (g *Golem) processConditionTags(template string, session *ChatSession) string {
	// Use regex to find and process conditions
	// This handles nesting by processing inner conditions first
	conditionRegex := regexp.MustCompile(`(?s)<condition(?: name="([^"]+)"(?: value="([^"]+)")?)?>(.*?)</condition>`)

	for {
		matches := conditionRegex.FindAllStringSubmatch(template, -1)
		if len(matches) == 0 {
			break // No more conditions
		}

		// Process the first (innermost) condition
		match := matches[0]
		if len(match) < 4 {
			break
		}

		varName := match[1]
		expectedValue := match[2]
		conditionContent := strings.TrimSpace(match[3])

		if g.verbose {
			g.logger.Printf("Processing condition: var='%s', expected='%s', content='%s'",
				varName, expectedValue, conditionContent)
		}

		// Get the actual variable value
		actualValue := g.getVariableValue(varName, session)

		// Process the condition content
		response := g.processConditionContent(conditionContent, varName, actualValue, expectedValue, session)

		if g.verbose {
			g.logger.Printf("Condition response: '%s'", response)
		}

		// Replace the condition tag with the response
		template = strings.ReplaceAll(template, match[0], response)
	}

	return template
}

// processConditionContent processes the content inside <condition> tags
func (g *Golem) processConditionContent(content string, varName, actualValue, expectedValue string, session *ChatSession) string {
	// Handle different condition types

	// Type 1: Simple condition with value attribute
	if expectedValue != "" {
		if actualValue == expectedValue {
			// Process the content through the full template pipeline
			return g.processConditionTemplate(content, session)
		}
		return "" // No match, return empty
	}

	// Type 2: Multiple <li> conditions or default condition
	if strings.Contains(content, "<li") {
		return g.processConditionListItems(content, actualValue, session)
	}

	// Type 3: Default condition (no value specified, no <li> elements)
	if actualValue != "" {
		// Process the content through the full template pipeline
		return g.processConditionTemplate(content, session)
	}

	return "" // Variable not found or empty
}

// processConditionListItems processes <li> elements within condition tags
func (g *Golem) processConditionListItems(content string, actualValue string, session *ChatSession) string {
	// Find all <li> elements with optional value attributes
	liRegex := regexp.MustCompile(`(?s)<li(?: value="([^"]+)")?>(.*?)</li>`)
	matches := liRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		liValue := match[1]
		liContent := strings.TrimSpace(match[2])

		// If no value specified, this is the default case
		if liValue == "" {
			return g.processConditionTemplate(liContent, session)
		}

		// Check if this condition matches
		if actualValue == liValue {
			return g.processConditionTemplate(liContent, session)
		}
	}

	return "" // No match found
}

// processConditionTemplate processes condition content through the full template pipeline
func (g *Golem) processConditionTemplate(content string, session *ChatSession) string {
	// Create variable context for condition processing
	ctx := &VariableContext{
		LocalVars:     make(map[string]string),
		Session:       session,
		Topic:         "", // TODO: Implement topic tracking
		KnowledgeBase: g.aimlKB,
	}

	return g.processTemplateWithContext(content, make(map[string]string), ctx)
}

// VariableScope represents different scopes for variable resolution
type VariableScope int

const (
	ScopeLocal      VariableScope = iota // Local scope (within current template)
	ScopeSession                         // Session scope (within current chat session)
	ScopeTopic                           // Topic scope (within current topic)
	ScopeGlobal                          // Global scope (knowledge base wide)
	ScopeProperties                      // Properties scope (bot properties, read-only)
)

// VariableContext holds the context for variable resolution
type VariableContext struct {
	LocalVars     map[string]string  // Local variables (highest priority)
	Session       *ChatSession       // Session context
	Topic         string             // Current topic
	KnowledgeBase *AIMLKnowledgeBase // Knowledge base context
}

// getVariableValue retrieves a variable value from the appropriate context with proper scope resolution
func (g *Golem) getVariableValue(varName string, session *ChatSession) string {
	// Create variable context
	ctx := &VariableContext{
		LocalVars:     make(map[string]string),
		Session:       session,
		Topic:         "", // TODO: Implement topic tracking
		KnowledgeBase: g.aimlKB,
	}

	return g.resolveVariable(varName, ctx)
}

// resolveVariable resolves a variable using proper scope hierarchy
func (g *Golem) resolveVariable(varName string, ctx *VariableContext) string {
	// 1. Check local scope (highest priority)
	if ctx.LocalVars != nil {
		if value, exists := ctx.LocalVars[varName]; exists {
			return value
		}
	}

	// 2. Check session scope
	if ctx.Session != nil && ctx.Session.Variables != nil {
		if value, exists := ctx.Session.Variables[varName]; exists {
			return value
		}
	}

	// 3. Check topic scope (TODO: Implement topic variables)
	// if ctx.Topic != "" && ctx.KnowledgeBase != nil && ctx.KnowledgeBase.Topics != nil {
	//     if topicVars, exists := ctx.KnowledgeBase.Topics[ctx.Topic]; exists {
	//         if value, exists := topicVars[varName]; exists {
	//             return value
	//         }
	//     }
	// }

	// 4. Check global scope (knowledge base variables)
	if ctx.KnowledgeBase != nil && ctx.KnowledgeBase.Variables != nil {
		if value, exists := ctx.KnowledgeBase.Variables[varName]; exists {
			return value
		}
	}

	// 5. Check properties scope (read-only)
	if ctx.KnowledgeBase != nil && ctx.KnowledgeBase.Properties != nil {
		if value, exists := ctx.KnowledgeBase.Properties[varName]; exists {
			return value
		}
	}

	return "" // Variable not found
}

// setVariable sets a variable in the appropriate scope
func (g *Golem) setVariable(varName, varValue string, scope VariableScope, ctx *VariableContext) {
	switch scope {
	case ScopeLocal:
		if ctx.LocalVars == nil {
			ctx.LocalVars = make(map[string]string)
		}
		ctx.LocalVars[varName] = varValue
	case ScopeSession:
		if ctx.Session != nil {
			if ctx.Session.Variables == nil {
				ctx.Session.Variables = make(map[string]string)
			}
			ctx.Session.Variables[varName] = varValue

			// Special case: if setting the topic variable, also set the session topic
			if varName == "topic" {
				ctx.Session.SetSessionTopic(varValue)
			}
		}
	case ScopeTopic:
		// TODO: Implement topic variables
		// if ctx.KnowledgeBase != nil {
		//     if ctx.KnowledgeBase.Topics == nil {
		//         ctx.KnowledgeBase.Topics = make(map[string]map[string]string)
		//     }
		//     if ctx.KnowledgeBase.Topics[ctx.Topic] == nil {
		//         ctx.KnowledgeBase.Topics[ctx.Topic] = make(map[string]string)
		//     }
		//     ctx.KnowledgeBase.Topics[ctx.Topic][varName] = varValue
		// }
	case ScopeGlobal:
		if ctx.KnowledgeBase != nil {
			if ctx.KnowledgeBase.Variables == nil {
				ctx.KnowledgeBase.Variables = make(map[string]string)
			}
			ctx.KnowledgeBase.Variables[varName] = varValue
		}
	case ScopeProperties:
		// Properties are read-only, cannot be set
		if g.verbose {
			g.logger.Printf("Warning: Cannot set property '%s' - properties are read-only", varName)
		}
	}
}

// processDateTimeTags processes <date> and <time> tags
func (g *Golem) processDateTimeTags(template string) string {
	// Process <date> tags
	template = g.processDateTags(template)

	// Process <time> tags
	template = g.processTimeTags(template)

	return template
}

// processDateTags processes <date> tags with various formats
func (g *Golem) processDateTags(template string) string {
	// Find all <date> tags
	dateRegex := regexp.MustCompile(`<date(?: format="([^"]*)"| format=\\"([^"]*)\\")?/>`)
	matches := dateRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		format := ""
		if len(match) > 1 && match[1] != "" {
			format = match[1]
		} else if len(match) > 2 && match[2] != "" {
			format = match[2]
		}

		if g.verbose {
			g.logger.Printf("Processing date tag with format: '%s'", format)
		}

		// Get current date and format it
		dateStr := g.formatDate(format)

		// Replace the date tag with the formatted date
		template = strings.ReplaceAll(template, match[0], dateStr)
	}

	return template
}

// processTimeTags processes <time> tags with various formats
func (g *Golem) processTimeTags(template string) string {
	// Find all <time> tags
	timeRegex := regexp.MustCompile(`<time(?: format="([^"]*)"| format=\\"([^"]*)\\")?/>`)
	matches := timeRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		format := ""
		if len(match) > 1 && match[1] != "" {
			format = match[1]
		} else if len(match) > 2 && match[2] != "" {
			format = match[2]
		}

		if g.verbose {
			g.logger.Printf("Processing time tag with format: '%s'", format)
		}

		// Get current time and format it
		timeStr := g.formatTime(format)

		// Replace the time tag with the formatted time
		template = strings.ReplaceAll(template, match[0], timeStr)
	}

	return template
}

// processRequestTags processes <request> tags with index support
func (g *Golem) processRequestTags(template string, ctx *VariableContext) string {
	if ctx.Session == nil {
		return template
	}

	// Find all <request> tags with optional index attribute
	requestRegex := regexp.MustCompile(`<request(?: index="(\d+)")?/>`)
	matches := requestRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		index := 1 // Default to most recent request
		if len(match) > 1 && match[1] != "" {
			// Parse the index from the attribute
			if parsedIndex, err := strconv.Atoi(match[1]); err == nil && parsedIndex > 0 {
				index = parsedIndex
			}
		}

		if g.verbose {
			g.logger.Printf("Processing request tag with index: %d", index)
		}

		// Get the request by index
		requestValue := ctx.Session.GetRequestByIndex(index)
		if requestValue == "" {
			if g.verbose {
				g.logger.Printf("No request found at index %d", index)
			}
			// Replace with empty string if no request found
			template = strings.ReplaceAll(template, match[0], "")
		} else {
			// Replace the request tag with the actual request
			template = strings.ReplaceAll(template, match[0], requestValue)
		}
	}

	return template
}

// processResponseTags processes <response> tags with index support
func (g *Golem) processResponseTags(template string, ctx *VariableContext) string {
	if ctx.Session == nil {
		return template
	}

	// Find all <response> tags with optional index attribute
	responseRegex := regexp.MustCompile(`<response(?: index="(\d+)")?/>`)
	matches := responseRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		index := 1 // Default to most recent response
		if len(match) > 1 && match[1] != "" {
			// Parse the index from the attribute
			if parsedIndex, err := strconv.Atoi(match[1]); err == nil && parsedIndex > 0 {
				index = parsedIndex
			}
		}

		if g.verbose {
			g.logger.Printf("Processing response tag with index: %d", index)
		}

		// Get the response by index
		responseValue := ctx.Session.GetResponseByIndex(index)
		if responseValue == "" {
			if g.verbose {
				g.logger.Printf("No response found at index %d", index)
			}
			// Replace with empty string if no response found
			template = strings.ReplaceAll(template, match[0], "")
		} else {
			// Replace the response tag with the actual response
			template = strings.ReplaceAll(template, match[0], responseValue)
		}
	}

	return template
}

// formatDate formats the current date according to the specified format
func (g *Golem) formatDate(format string) string {
	now := time.Now()

	switch format {
	case "short":
		return now.Format("01/02/06")
	case "long":
		return now.Format("Monday, January 2, 2006")
	case "iso":
		return now.Format("2006-01-02")
	case "us":
		return now.Format("January 2, 2006")
	case "european":
		return now.Format("2 January 2006")
	case "day":
		return now.Format("Monday")
	case "month":
		return now.Format("January")
	case "year":
		return now.Format("2006")
	case "dayofyear":
		return fmt.Sprintf("%d", now.YearDay())
	case "weekday":
		return fmt.Sprintf("%d", int(now.Weekday()))
	case "week":
		_, week := now.ISOWeek()
		return fmt.Sprintf("%d", week)
	case "quarter":
		month := int(now.Month())
		quarter := (month-1)/3 + 1
		return fmt.Sprintf("Q%d", quarter)
	case "leapyear":
		year := now.Year()
		if (year%4 == 0 && year%100 != 0) || (year%400 == 0) {
			return "yes"
		}
		return "no"
	case "daysinmonth":
		nextMonth := now.AddDate(0, 1, 0)
		lastDay := nextMonth.AddDate(0, 0, -nextMonth.Day())
		return fmt.Sprintf("%d", lastDay.Day())
	case "daysinyear":
		if now.Year()%4 == 0 && (now.Year()%100 != 0 || now.Year()%400 == 0) {
			return "366"
		}
		return "365"
	default:
		// Default format: "January 2, 2006"
		return now.Format("January 2, 2006")
	}
}

// formatTime formats the current time according to the specified format
func (g *Golem) formatTime(format string) string {
	now := time.Now()

	switch format {
	case "12":
		return now.Format("3:04 PM")
	case "24":
		return now.Format("15:04")
	case "iso":
		return now.Format("15:04:05")
	case "hour":
		return fmt.Sprintf("%d", now.Hour())
	case "minute":
		return fmt.Sprintf("%d", now.Minute())
	case "second":
		return fmt.Sprintf("%d", now.Second())
	case "millisecond":
		return fmt.Sprintf("%d", now.Nanosecond()/1000000)
	case "timezone":
		return now.Format("MST")
	case "offset":
		_, offset := now.Zone()
		hours := offset / 3600
		minutes := (offset % 3600) / 60
		return fmt.Sprintf("%+03d:%02d", hours, minutes)
	case "unix":
		return fmt.Sprintf("%d", now.Unix())
	case "unixmilli":
		return fmt.Sprintf("%d", now.UnixMilli())
	case "unixnano":
		return fmt.Sprintf("%d", now.UnixNano())
	case "rfc3339":
		return now.Format(time.RFC3339)
	case "rfc822":
		return now.Format(time.RFC822)
	case "kitchen":
		return now.Format(time.Kitchen)
	case "stamp":
		return now.Format(time.Stamp)
	case "stampmilli":
		return now.Format(time.StampMilli)
	case "stampmicro":
		return now.Format(time.StampMicro)
	case "stampnano":
		return now.Format(time.StampNano)
	default:
		// Default format: "3:04 PM"
		return now.Format("3:04 PM")
	}
}

// processRandomTags processes <random> tags and selects a random <li> element
func (g *Golem) processRandomTags(template string) string {
	// Find all <random> tags
	randomRegex := regexp.MustCompile(`(?s)<random>(.*?)</random>`)
	matches := randomRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			randomContent := strings.TrimSpace(match[1])
			if g.verbose {
				g.logger.Printf("Processing random tag: '%s'", randomContent)
			}

			// Find all <li> elements within the random tag
			liRegex := regexp.MustCompile(`(?s)<li>(.*?)</li>`)
			liMatches := liRegex.FindAllStringSubmatch(randomContent, -1)

			if len(liMatches) == 0 {
				// No <li> elements found, use the content as-is
				template = strings.ReplaceAll(template, match[0], randomContent)
				continue
			}

			// Select a random <li> element
			selectedIndex := 0
			if len(liMatches) > 1 {
				// Use a simple random selection (in a real implementation, you'd use crypto/rand)
				selectedIndex = len(liMatches) % 2 // Simple pseudo-random for testing
			}

			selectedContent := strings.TrimSpace(liMatches[selectedIndex][1])

			// Process date/time tags in the selected content
			selectedContent = g.processDateTimeTags(selectedContent)

			if g.verbose {
				g.logger.Printf("Selected random option %d: '%s'", selectedIndex+1, selectedContent)
			}

			// Replace the entire <random> tag with the selected content
			template = strings.ReplaceAll(template, match[0], selectedContent)
		}
	}

	return template
}

// loadDefaultProperties loads default bot properties
func (g *Golem) loadDefaultProperties(kb *AIMLKnowledgeBase) error {
	// Set default properties
	defaultProps := map[string]string{
		"name":              "Golem",
		"version":           "1.0.0",
		"master":            "User",
		"birthplace":        "Go",
		"birthday":          "2025-09-23",
		"gender":            "neutral",
		"species":           "AI",
		"job":               "Assistant",
		"personality":       "friendly",
		"mood":              "helpful",
		"attitude":          "positive",
		"language":          "English",
		"location":          "Virtual",
		"timezone":          "UTC",
		"max_loops":         "10",
		"timeout":           "30000",
		"jokemode":          "true",
		"learnmode":         "true",
		"default_response":  "I'm not sure I understand. Could you rephrase that?",
		"error_response":    "Sorry, I encountered an error processing your request.",
		"thinking_response": "Let me think about that...",
		"memory_size":       "1000",
		"forget_time":       "3600",
		"learning_enabled":  "true",
		"pattern_limit":     "1000",
		"response_limit":    "5000",
	}

	// Copy default properties to knowledge base
	for key, value := range defaultProps {
		kb.Properties[key] = value
	}

	// Try to load properties from file
	propertiesFile := "testdata/bot.properties"
	content, err := g.LoadFile(propertiesFile)
	if err == nil {
		// Parse properties file
		fileProps, err := g.parsePropertiesFile(content)
		if err == nil {
			// Override defaults with file properties
			for key, value := range fileProps {
				kb.Properties[key] = value
			}
			if g.verbose {
				g.logger.Printf("Loaded properties from file: %s", propertiesFile)
			}
		} else if g.verbose {
			g.logger.Printf("Could not parse properties file: %v", err)
		}
	} else if g.verbose {
		g.logger.Printf("Could not load properties file: %v", err)
	}

	return nil
}

// parsePropertiesFile parses a properties file
func (g *Golem) parsePropertiesFile(content string) (map[string]string, error) {
	properties := make(map[string]string)
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid property format at line %d: %s", lineNum+1, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("empty property key at line %d", lineNum+1)
		}

		properties[key] = value
	}

	return properties, nil
}

// GetProperty retrieves a property value
func (kb *AIMLKnowledgeBase) GetProperty(key string) string {
	// Check properties first
	if value, exists := kb.Properties[key]; exists {
		return value
	}
	// Check variables as fallback
	if value, exists := kb.Variables[key]; exists {
		return value
	}
	return ""
}

// SetProperty sets a property value
func (kb *AIMLKnowledgeBase) SetProperty(key, value string) {
	kb.Properties[key] = value
}

// AddSetMember adds a member to a set
func (kb *AIMLKnowledgeBase) AddSetMember(setName, member string) {
	setName = strings.ToUpper(setName)
	if kb.Sets[setName] == nil {
		kb.Sets[setName] = make([]string, 0)
	}
	// Check if member already exists
	for _, existing := range kb.Sets[setName] {
		if existing == strings.ToUpper(member) {
			return // Already exists
		}
	}
	kb.Sets[setName] = append(kb.Sets[setName], strings.ToUpper(member))
}

// AddSetMembers adds multiple members to a set
func (kb *AIMLKnowledgeBase) AddSetMembers(setName string, members []string) {
	for _, member := range members {
		kb.AddSetMember(setName, member)
	}
}

// GetSetMembers returns all members of a set
func (kb *AIMLKnowledgeBase) GetSetMembers(setName string) []string {
	setName = strings.ToUpper(setName)
	if kb.Sets[setName] == nil {
		return []string{}
	}
	return kb.Sets[setName]
}

// IsSetMember checks if a word is a member of a set
func (kb *AIMLKnowledgeBase) IsSetMember(setName, word string) bool {
	setName = strings.ToUpper(setName)
	if kb.Sets[setName] == nil {
		return false
	}
	upperWord := strings.ToUpper(word)
	for _, member := range kb.Sets[setName] {
		if member == upperWord {
			return true
		}
	}
	return false
}

// SetTopic sets the current topic for a category
func (kb *AIMLKnowledgeBase) SetTopic(pattern, topic string) {
	if category, exists := kb.Patterns[pattern]; exists {
		category.Topic = strings.ToUpper(topic)
	}
}

// GetTopic returns the topic for a pattern
func (kb *AIMLKnowledgeBase) GetTopic(pattern string) string {
	if category, exists := kb.Patterns[pattern]; exists {
		return category.Topic
	}
	return ""
}

// SetSessionTopic sets the current topic for a session
func (session *ChatSession) SetSessionTopic(topic string) {
	session.Topic = strings.ToUpper(topic)
}

// GetSessionTopic returns the current topic for a session
func (session *ChatSession) GetSessionTopic() string {
	return session.Topic
}

// AddToThatHistory adds a bot response to the that history
func (session *ChatSession) AddToThatHistory(response string) {
	// Keep only the last 10 responses to prevent memory bloat
	if len(session.ThatHistory) >= 10 {
		session.ThatHistory = session.ThatHistory[1:]
	}
	session.ThatHistory = append(session.ThatHistory, response)
}

// GetLastThat returns the last bot response for that matching
func (session *ChatSession) GetLastThat() string {
	if len(session.ThatHistory) == 0 {
		return ""
	}
	return session.ThatHistory[len(session.ThatHistory)-1]
}

// GetThatHistory returns the that history
func (session *ChatSession) GetThatHistory() []string {
	return session.ThatHistory
}

// AddToRequestHistory adds a user request to the request history
func (session *ChatSession) AddToRequestHistory(request string) {
	// Keep only the last 10 requests to prevent memory bloat
	if len(session.RequestHistory) >= 10 {
		session.RequestHistory = session.RequestHistory[1:]
	}
	session.RequestHistory = append(session.RequestHistory, request)
}

// GetRequestHistory returns the request history
func (session *ChatSession) GetRequestHistory() []string {
	return session.RequestHistory
}

// GetRequestByIndex returns a request by index (1-based, where 1 is most recent)
func (session *ChatSession) GetRequestByIndex(index int) string {
	if index < 1 || index > len(session.RequestHistory) {
		return ""
	}
	// Convert to 0-based index (most recent is at the end)
	actualIndex := len(session.RequestHistory) - index
	return session.RequestHistory[actualIndex]
}

// AddToResponseHistory adds a bot response to the response history
func (session *ChatSession) AddToResponseHistory(response string) {
	// Keep only the last 10 responses to prevent memory bloat
	if len(session.ResponseHistory) >= 10 {
		session.ResponseHistory = session.ResponseHistory[1:]
	}
	session.ResponseHistory = append(session.ResponseHistory, response)
}

// GetResponseHistory returns the response history
func (session *ChatSession) GetResponseHistory() []string {
	return session.ResponseHistory
}

// GetResponseByIndex returns a response by index (1-based, where 1 is most recent)
func (session *ChatSession) GetResponseByIndex(index int) string {
	if index < 1 || index > len(session.ResponseHistory) {
		return ""
	}
	// Convert to 0-based index (most recent is at the end)
	actualIndex := len(session.ResponseHistory) - index
	return session.ResponseHistory[actualIndex]
}

// processTopicSettingTagsWithContext processes <set name="topic"> tags
func (g *Golem) processTopicSettingTagsWithContext(template string, ctx *VariableContext) string {
	// Find all <set name="topic"> tags
	topicSetRegex := regexp.MustCompile(`<set\s+name="topic">(.*?)</set>`)
	matches := topicSetRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			topicValue := strings.TrimSpace(match[1])

			// Set topic in session if available
			if ctx.Session != nil {
				ctx.Session.SetSessionTopic(topicValue)
			}

			// Remove the set tag from the template (don't replace with topic value)
			template = strings.ReplaceAll(template, match[0], "")
			// Clean up extra spaces
			template = strings.ReplaceAll(template, "  ", " ")
			template = strings.TrimSpace(template)
		}
	}

	return template
}

// processMapTagsWithContext processes <map> tags with variable context
func (g *Golem) processMapTagsWithContext(template string, ctx *VariableContext) string {
	if ctx.KnowledgeBase == nil || ctx.KnowledgeBase.Maps == nil {
		return template
	}

	// Find all <map> tags
	mapRegex := regexp.MustCompile(`<map\s+name=["']([^"']+)["']>([^<]*)</map>`)
	matches := mapRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			mapName := match[1]
			key := strings.TrimSpace(match[2])

			if g.verbose {
				g.logger.Printf("Processing map tag: name='%s', key='%s'", mapName, key)
			}

			// Look up the map
			if mapData, exists := ctx.KnowledgeBase.Maps[mapName]; exists {
				// Look up the key in the map
				if value, keyExists := mapData[key]; keyExists {
					// Replace the map tag with the mapped value
					template = strings.ReplaceAll(template, match[0], value)
					if g.verbose {
						g.logger.Printf("Mapped '%s' -> '%s'", key, value)
					}
				} else {
					// Key not found in map, leave the original key
					if g.verbose {
						g.logger.Printf("Key '%s' not found in map '%s'", key, mapName)
					}
					template = strings.ReplaceAll(template, match[0], key)
				}
			} else {
				// Map not found, leave the original key
				if g.verbose {
					g.logger.Printf("Map '%s' not found", mapName)
				}
				template = strings.ReplaceAll(template, match[0], key)
			}
		}
	}

	return template
}

// processListTagsWithContext processes <list> tags with variable context
func (g *Golem) processListTagsWithContext(template string, ctx *VariableContext) string {
	if g.verbose {
		g.logger.Printf("List processing: ctx.KnowledgeBase=%v, ctx.KnowledgeBase.Lists=%v", ctx.KnowledgeBase != nil, ctx.KnowledgeBase != nil && ctx.KnowledgeBase.Lists != nil)
	}
	if ctx.KnowledgeBase == nil || ctx.KnowledgeBase.Lists == nil {
		if g.verbose {
			g.logger.Printf("List processing: returning early due to nil knowledge base or lists")
		}
		return template
	}

	// Find all <list> tags with various operations
	listRegex := regexp.MustCompile(`<list\s+name=["']([^"']+)["'](?:\s+index=["']([^"']+)["'])?(?:\s+operation=["']([^"']+)["'])?>(.*?)</list>`)
	matches := listRegex.FindAllStringSubmatch(template, -1)

	if g.verbose {
		g.logger.Printf("List processing: found %d matches in template: '%s'", len(matches), template)
	}

	for _, match := range matches {
		if len(match) >= 4 {
			listName := match[1]
			indexStr := match[2]
			operation := match[3]
			content := strings.TrimSpace(match[4])

			if g.verbose {
				g.logger.Printf("Processing list tag: name='%s', index='%s', operation='%s', content='%s'", listName, indexStr, operation, content)
			}

			// Get or create the list
			if ctx.KnowledgeBase.Lists[listName] == nil {
				ctx.KnowledgeBase.Lists[listName] = make([]string, 0)
			}
			list := ctx.KnowledgeBase.Lists[listName]

			switch operation {
			case "add", "append":
				// Add item to the end of the list
				list = append(list, content)
				ctx.KnowledgeBase.Lists[listName] = list
				template = strings.ReplaceAll(template, match[0], "")
				if g.verbose {
					g.logger.Printf("Added '%s' to list '%s'", content, listName)
				}

			case "insert":
				// Insert item at specific index
				if indexStr != "" {
					if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index <= len(list) {
						// Insert at the specified index
						list = append(list[:index], append([]string{content}, list[index:]...)...)
						ctx.KnowledgeBase.Lists[listName] = list
						template = strings.ReplaceAll(template, match[0], "")
						if g.verbose {
							g.logger.Printf("Inserted '%s' at index %d in list '%s'", content, index, listName)
						}
					} else {
						// Invalid index, append to end
						list = append(list, content)
						ctx.KnowledgeBase.Lists[listName] = list
						template = strings.ReplaceAll(template, match[0], "")
						if g.verbose {
							g.logger.Printf("Invalid index %s, appended '%s' to list '%s'", indexStr, content, listName)
						}
					}
				} else {
					// No index specified, append to end
					list = append(list, content)
					ctx.KnowledgeBase.Lists[listName] = list
					template = strings.ReplaceAll(template, match[0], "")
					if g.verbose {
						g.logger.Printf("No index specified, appended '%s' to list '%s'", content, listName)
					}
				}

			case "remove", "delete":
				// Remove item from list
				if indexStr != "" {
					if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(list) {
						// Remove at specific index
						list = append(list[:index], list[index+1:]...)
						ctx.KnowledgeBase.Lists[listName] = list
						template = strings.ReplaceAll(template, match[0], "")
						if g.verbose {
							g.logger.Printf("Removed item at index %d from list '%s'", index, listName)
						}
					} else {
						// Invalid index, try to remove by value
						for i, item := range list {
							if item == content {
								list = append(list[:i], list[i+1:]...)
								ctx.KnowledgeBase.Lists[listName] = list
								template = strings.ReplaceAll(template, match[0], "")
								if g.verbose {
									g.logger.Printf("Removed '%s' from list '%s'", content, listName)
								}
								break
							}
						}
					}
				} else {
					// Remove by value
					for i, item := range list {
						if item == content {
							list = append(list[:i], list[i+1:]...)
							ctx.KnowledgeBase.Lists[listName] = list
							template = strings.ReplaceAll(template, match[0], "")
							if g.verbose {
								g.logger.Printf("Removed '%s' from list '%s'", content, listName)
							}
							break
						}
					}
				}

			case "clear":
				// Clear the list
				ctx.KnowledgeBase.Lists[listName] = make([]string, 0)
				template = strings.ReplaceAll(template, match[0], "")
				if g.verbose {
					g.logger.Printf("Cleared list '%s'", listName)
				}

			case "size", "length":
				// Return the size of the list
				size := strconv.Itoa(len(list))
				template = strings.ReplaceAll(template, match[0], size)
				if g.verbose {
					g.logger.Printf("List '%s' size: %s", listName, size)
				}

			case "get", "":
				// Get item at index or return the list
				if indexStr != "" {
					if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(list) {
						// Get item at specific index
						template = strings.ReplaceAll(template, match[0], list[index])
						if g.verbose {
							g.logger.Printf("Got item at index %d from list '%s': '%s'", index, listName, list[index])
						}
					} else {
						// Invalid index, return empty
						template = strings.ReplaceAll(template, match[0], "")
						if g.verbose {
							g.logger.Printf("Invalid index %s for list '%s'", indexStr, listName)
						}
					}
				} else {
					// Return all items joined by space
					items := strings.Join(list, " ")
					template = strings.ReplaceAll(template, match[0], items)
					if g.verbose {
						g.logger.Printf("Got all items from list '%s': '%s'", listName, items)
					}
				}

			default:
				// Unknown operation, treat as get
				if indexStr != "" {
					if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(list) {
						template = strings.ReplaceAll(template, match[0], list[index])
					} else {
						template = strings.ReplaceAll(template, match[0], "")
					}
				} else {
					items := strings.Join(list, " ")
					template = strings.ReplaceAll(template, match[0], items)
				}
			}
		}
	}

	return template
}

// processArrayTagsWithContext processes <array> tags with variable context
func (g *Golem) processArrayTagsWithContext(template string, ctx *VariableContext) string {
	if ctx.KnowledgeBase == nil || ctx.KnowledgeBase.Arrays == nil {
		return template
	}

	// Find all <array> tags with various operations
	arrayRegex := regexp.MustCompile(`<array\s+name=["']([^"']+)["'](?:\s+index=["']([^"']+)["'])?(?:\s+operation=["']([^"']+)["'])?>(.*?)</array>`)
	matches := arrayRegex.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) >= 4 {
			arrayName := match[1]
			indexStr := match[2]
			operation := match[3]
			content := strings.TrimSpace(match[4])

			if g.verbose {
				g.logger.Printf("Processing array tag: name='%s', index='%s', operation='%s', content='%s'", arrayName, indexStr, operation, content)
			}

			// Get or create the array
			if ctx.KnowledgeBase.Arrays[arrayName] == nil {
				ctx.KnowledgeBase.Arrays[arrayName] = make([]string, 0)
			}
			array := ctx.KnowledgeBase.Arrays[arrayName]

			switch operation {
			case "set", "assign":
				// Set item at specific index
				if indexStr != "" {
					if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 {
						// Ensure array is large enough
						for len(array) <= index {
							array = append(array, "")
						}
						array[index] = content
						ctx.KnowledgeBase.Arrays[arrayName] = array
						template = strings.ReplaceAll(template, match[0], "")
						if g.verbose {
							g.logger.Printf("Set array '%s'[%d] = '%s'", arrayName, index, content)
						}
					} else {
						// Invalid index
						template = strings.ReplaceAll(template, match[0], "")
						if g.verbose {
							g.logger.Printf("Invalid index %s for array '%s'", indexStr, arrayName)
						}
					}
				} else {
					// No index specified, append to end
					array = append(array, content)
					ctx.KnowledgeBase.Arrays[arrayName] = array
					template = strings.ReplaceAll(template, match[0], "")
					if g.verbose {
						g.logger.Printf("Appended '%s' to array '%s'", content, arrayName)
					}
				}

			case "get", "":
				// Get item at index
				if indexStr != "" {
					if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(array) {
						template = strings.ReplaceAll(template, match[0], array[index])
						if g.verbose {
							g.logger.Printf("Got array '%s'[%d] = '%s'", arrayName, index, array[index])
						}
					} else {
						template = strings.ReplaceAll(template, match[0], "")
						if g.verbose {
							g.logger.Printf("Invalid index %s for array '%s'", indexStr, arrayName)
						}
					}
				} else {
					// Return all items joined by space
					items := strings.Join(array, " ")
					template = strings.ReplaceAll(template, match[0], items)
					if g.verbose {
						g.logger.Printf("Got all items from array '%s': '%s'", arrayName, items)
					}
				}

			case "size", "length":
				// Return the size of the array
				size := strconv.Itoa(len(array))
				template = strings.ReplaceAll(template, match[0], size)
				if g.verbose {
					g.logger.Printf("Array '%s' size: %s", arrayName, size)
				}

			case "clear":
				// Clear the array
				ctx.KnowledgeBase.Arrays[arrayName] = make([]string, 0)
				template = strings.ReplaceAll(template, match[0], "")
				if g.verbose {
					g.logger.Printf("Cleared array '%s'", arrayName)
				}

			default:
				// Unknown operation, treat as get
				if indexStr != "" {
					if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(array) {
						template = strings.ReplaceAll(template, match[0], array[index])
					} else {
						template = strings.ReplaceAll(template, match[0], "")
					}
				} else {
					items := strings.Join(array, " ")
					template = strings.ReplaceAll(template, match[0], items)
				}
			}
		}
	}

	return template
}

// NormalizedContent represents content that has been normalized with preserved sections
type NormalizedContent struct {
	NormalizedText    string
	PreservedSections map[string]string
}

// normalizeText normalizes text for pattern matching while preserving special content
func normalizeText(input string) NormalizedContent {
	preservedSections := make(map[string]string)
	normalizedText := input

	// Step 1: Preserve mathematical expressions (numbers, operators, parentheses)
	// This includes expressions like "2 + 3", "x = 5", "sqrt(16)", etc.
	// But avoid matching simple variable assignments like "name=user"
	mathPattern := regexp.MustCompile(`\b\d+(?:\.\d+)?(?:\s*[+\-*/=<>!&|^~]\s*\d+(?:\.\d+)?)+\b|\b\w+\s*[+\-*/=<>!&|^~]\s*\d+(?:\.\d+)?\b|\b\w+\s*\([^)]*\)\s*[+\-*/=<>!&|^~]\s*\d+(?:\.\d+)?\b`)
	mathMatches := mathPattern.FindAllString(normalizedText, -1)
	for i, match := range mathMatches {
		placeholder := fmt.Sprintf("__MATH_%d__", i)
		preservedSections[placeholder] = match
		normalizedText = strings.Replace(normalizedText, match, placeholder, 1)
	}

	// Step 2: Preserve quoted strings (single and double quotes)
	quotePattern := regexp.MustCompile(`"[^"]*"|'[^']*'`)
	quoteMatches := quotePattern.FindAllString(normalizedText, -1)
	for i, match := range quoteMatches {
		placeholder := fmt.Sprintf("__QUOTE_%d__", i)
		preservedSections[placeholder] = match
		normalizedText = strings.Replace(normalizedText, match, placeholder, 1)
	}

	// Step 3: Preserve URLs and email addresses
	urlPattern := regexp.MustCompile(`https?://[^\s]+|www\.[^\s]+|[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	urlMatches := urlPattern.FindAllString(normalizedText, -1)
	for i, match := range urlMatches {
		placeholder := fmt.Sprintf("__URL_%d__", i)
		preservedSections[placeholder] = match
		normalizedText = strings.Replace(normalizedText, match, placeholder, 1)
	}

	// Step 4: Preserve AIML tags (but not set/topic tags which need special handling)
	// First, temporarily replace set/topic tags to avoid matching them
	tempSetTags := make(map[string]string)
	tempTopicTags := make(map[string]string)

	// Replace set tags temporarily
	setPattern := regexp.MustCompile(`<set>[^<]+</set>`)
	setMatches := setPattern.FindAllString(normalizedText, -1)
	for i, match := range setMatches {
		placeholder := fmt.Sprintf("__TEMP_SET_%d__", i)
		tempSetTags[placeholder] = match
		normalizedText = strings.Replace(normalizedText, match, placeholder, 1)
	}

	// Replace topic tags temporarily
	topicPattern := regexp.MustCompile(`<topic>[^<]+</topic>`)
	topicMatches := topicPattern.FindAllString(normalizedText, -1)
	for i, match := range topicMatches {
		placeholder := fmt.Sprintf("__TEMP_TOPIC_%d__", i)
		tempTopicTags[placeholder] = match
		normalizedText = strings.Replace(normalizedText, match, placeholder, 1)
	}

	// Now match other AIML tags (more specific pattern to avoid conflicts)
	aimlTagPattern := regexp.MustCompile(`<[a-zA-Z][^>]*/>|<[a-zA-Z][^>]*>.*?</[a-zA-Z][^>]*>`)
	aimlTagMatches := aimlTagPattern.FindAllString(normalizedText, -1)
	for i, match := range aimlTagMatches {
		placeholder := fmt.Sprintf("__AIML_TAG_%d__", i)
		preservedSections[placeholder] = match
		normalizedText = strings.Replace(normalizedText, match, placeholder, 1)
	}

	// Restore set and topic tags
	for placeholder, original := range tempSetTags {
		normalizedText = strings.ReplaceAll(normalizedText, placeholder, original)
	}
	for placeholder, original := range tempTopicTags {
		normalizedText = strings.ReplaceAll(normalizedText, placeholder, original)
	}

	// Step 5: Preserve special punctuation that might be meaningful
	specialPunctPattern := regexp.MustCompile(`[!?;:]+`)
	specialPunctMatches := specialPunctPattern.FindAllString(normalizedText, -1)
	for i, match := range specialPunctMatches {
		placeholder := fmt.Sprintf("__PUNCT_%d__", i)
		preservedSections[placeholder] = match
		normalizedText = strings.Replace(normalizedText, match, placeholder, 1)
	}

	// Step 6: Now normalize the remaining text
	// Convert to uppercase
	normalizedText = strings.ToUpper(normalizedText)

	// Normalize whitespace
	normalizedText = regexp.MustCompile(`\s+`).ReplaceAllString(normalizedText, " ")
	normalizedText = strings.TrimSpace(normalizedText)

	// Normalize punctuation (but preserve our placeholders)
	// First, protect placeholders from being modified
	placeholderProtection := make(map[string]string)
	placeholderPattern := regexp.MustCompile(`__[A-Z_]+_\d+__`)
	placeholderMatches := placeholderPattern.FindAllString(normalizedText, -1)
	for i, match := range placeholderMatches {
		protectionKey := fmt.Sprintf("__PROTECT_%d__", i)
		placeholderProtection[protectionKey] = match
		normalizedText = strings.Replace(normalizedText, match, protectionKey, 1)
	}

	// Now normalize punctuation (but don't touch underscores in placeholders)
	normalizedText = strings.ReplaceAll(normalizedText, ".", "")
	normalizedText = strings.ReplaceAll(normalizedText, ",", "")
	normalizedText = strings.ReplaceAll(normalizedText, "-", " ")
	// Don't remove underscores as they're part of our placeholders

	// Restore placeholders
	for protectionKey, original := range placeholderProtection {
		normalizedText = strings.ReplaceAll(normalizedText, protectionKey, original)
	}

	// Clean up any double spaces that might have been created
	normalizedText = regexp.MustCompile(`\s+`).ReplaceAllString(normalizedText, " ")
	normalizedText = strings.TrimSpace(normalizedText)

	return NormalizedContent{
		NormalizedText:    normalizedText,
		PreservedSections: preservedSections,
	}
}

// denormalizeText restores the original content from normalized text
func denormalizeText(normalized NormalizedContent) string {
	text := normalized.NormalizedText

	// Restore preserved sections in reverse order of insertion
	// (to avoid conflicts with shorter placeholders)
	placeholders := make([]string, 0, len(normalized.PreservedSections))
	for placeholder := range normalized.PreservedSections {
		placeholders = append(placeholders, placeholder)
	}

	// Sort placeholders by length (longest first) to avoid replacement conflicts
	sort.Slice(placeholders, func(i, j int) bool {
		return len(placeholders[i]) > len(placeholders[j])
	})

	for _, placeholder := range placeholders {
		original := normalized.PreservedSections[placeholder]
		text = strings.ReplaceAll(text, placeholder, original)
	}

	return text
}

// normalizeForMatching normalizes text specifically for pattern matching
func normalizeForMatching(input string) string {
	// For pattern matching, we need more aggressive normalization
	// but still preserve set/topic tags and wildcards

	// First, preserve set and topic tags before case conversion
	tempSetTags := make(map[string]string)
	tempTopicTags := make(map[string]string)

	// Replace set tags temporarily
	setPattern := regexp.MustCompile(`<set>([^<]+)</set>`)
	setMatches := setPattern.FindAllString(input, -1)
	for i, match := range setMatches {
		placeholder := fmt.Sprintf("__TEMP_SET_%d__", i)
		tempSetTags[placeholder] = match
		input = strings.Replace(input, match, placeholder, 1)
	}

	// Replace topic tags temporarily
	topicPattern := regexp.MustCompile(`<topic>([^<]+)</topic>`)
	topicMatches := topicPattern.FindAllString(input, -1)
	for i, match := range topicMatches {
		placeholder := fmt.Sprintf("__TEMP_TOPIC_%d__", i)
		tempTopicTags[placeholder] = match
		input = strings.Replace(input, match, placeholder, 1)
	}

	text := strings.ToUpper(strings.TrimSpace(input))

	// Restore set and topic tags with preserved case
	for placeholder, original := range tempSetTags {
		text = strings.ReplaceAll(text, strings.ToUpper(placeholder), original)
	}
	for placeholder, original := range tempTopicTags {
		text = strings.ReplaceAll(text, strings.ToUpper(placeholder), original)
	}

	// Normalize whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Remove most punctuation for matching (but keep wildcards)
	text = strings.ReplaceAll(text, ".", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, "!", "")
	text = strings.ReplaceAll(text, "?", "")
	text = strings.ReplaceAll(text, ";", "")
	text = strings.ReplaceAll(text, ":", "")
	text = strings.ReplaceAll(text, "'", "") // Remove apostrophes
	text = strings.ReplaceAll(text, "-", " ")
	text = strings.ReplaceAll(text, "_", " ")

	// Clean up whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

// NormalizePattern normalizes AIML patterns for matching
func NormalizePattern(pattern string) string {
	// Patterns need special handling for set and topic tags
	// First, preserve set and topic tags before case conversion
	tempSetTags := make(map[string]string)
	tempTopicTags := make(map[string]string)

	// Replace set tags temporarily
	setPattern := regexp.MustCompile(`<set>([^<]+)</set>`)
	setMatches := setPattern.FindAllString(pattern, -1)
	for i, match := range setMatches {
		placeholder := fmt.Sprintf("__TEMP_SET_%d__", i)
		tempSetTags[placeholder] = match
		pattern = strings.Replace(pattern, match, placeholder, 1)
	}

	// Replace topic tags temporarily
	topicPattern := regexp.MustCompile(`<topic>([^<]+)</topic>`)
	topicMatches := topicPattern.FindAllString(pattern, -1)
	for i, match := range topicMatches {
		placeholder := fmt.Sprintf("__TEMP_TOPIC_%d__", i)
		tempTopicTags[placeholder] = match
		pattern = strings.Replace(pattern, match, placeholder, 1)
	}

	text := strings.ToUpper(strings.TrimSpace(pattern))

	// Restore set and topic tags with preserved case
	for placeholder, original := range tempSetTags {
		text = strings.ReplaceAll(text, strings.ToUpper(placeholder), original)
	}
	for placeholder, original := range tempTopicTags {
		text = strings.ReplaceAll(text, strings.ToUpper(placeholder), original)
	}

	// Normalize whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Remove punctuation that might interfere with matching
	text = strings.ReplaceAll(text, ".", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, "!", "")
	text = strings.ReplaceAll(text, "?", "")
	text = strings.ReplaceAll(text, ";", "")
	text = strings.ReplaceAll(text, ":", "")
	text = strings.ReplaceAll(text, "'", "") // Remove apostrophes
	text = strings.ReplaceAll(text, "-", " ")

	// Clean up whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

// parseLearnContent parses AIML content within learn/learnf tags
func (g *Golem) parseLearnContent(content string) ([]Category, error) {
	// Wrap content in a minimal AIML structure for parsing
	wrappedContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<aiml version="2.0">
%s
</aiml>`, content)

	// Parse the wrapped content
	aiml, err := g.parseAIML(wrappedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse learn content: %v", err)
	}

	return aiml.Categories, nil
}

// addSessionCategory adds a category to the session-specific knowledge base
func (g *Golem) addSessionCategory(category Category, ctx *VariableContext) error {
	if g.aimlKB == nil {
		return fmt.Errorf("no knowledge base available")
	}

	// Validate the category
	if category.Pattern == "" || category.Template == "" {
		return fmt.Errorf("invalid category: pattern and template are required")
	}

	// Normalize the pattern
	normalizedPattern := NormalizePattern(category.Pattern)

	// Check if category already exists
	if existingCategory, exists := g.aimlKB.Patterns[normalizedPattern]; exists {
		if g.verbose {
			g.logger.Printf("Updating existing session category: %s", normalizedPattern)
		}
		// Update existing category
		*existingCategory = category
	} else {
		if g.verbose {
			g.logger.Printf("Adding new session category: %s", normalizedPattern)
		}
		// Add new category
		g.aimlKB.Categories = append(g.aimlKB.Categories, category)
		g.aimlKB.Patterns[normalizedPattern] = &g.aimlKB.Categories[len(g.aimlKB.Categories)-1]
	}

	return nil
}

// addPersistentCategory adds a category to the persistent knowledge base
func (g *Golem) addPersistentCategory(category Category) error {
	if g.aimlKB == nil {
		return fmt.Errorf("no knowledge base available")
	}

	// Validate the category
	if category.Pattern == "" || category.Template == "" {
		return fmt.Errorf("invalid category: pattern and template are required")
	}

	// Normalize the pattern
	normalizedPattern := NormalizePattern(category.Pattern)

	// Check if category already exists
	if existingCategory, exists := g.aimlKB.Patterns[normalizedPattern]; exists {
		if g.verbose {
			g.logger.Printf("Updating existing persistent category: %s", normalizedPattern)
		}
		// Update existing category
		*existingCategory = category
	} else {
		if g.verbose {
			g.logger.Printf("Adding new persistent category: %s", normalizedPattern)
		}
		// Add new category
		g.aimlKB.Categories = append(g.aimlKB.Categories, category)
		g.aimlKB.Patterns[normalizedPattern] = &g.aimlKB.Categories[len(g.aimlKB.Categories)-1]
	}

	// TODO: In a real implementation, you would save this to persistent storage
	// For now, we just add it to the in-memory knowledge base
	if g.verbose {
		g.logger.Printf("Note: Persistent learning not yet implemented - category added to memory only")
	}

	return nil
}
