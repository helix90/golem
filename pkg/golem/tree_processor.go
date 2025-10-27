package golem

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TreeProcessor handles processing of AST nodes for AIML tag processing
type TreeProcessor struct {
	golem *Golem
	ctx   *VariableContext
}

// NewTreeProcessor creates a new tree processor
func NewTreeProcessor(golem *Golem) *TreeProcessor {
	return &TreeProcessor{
		golem: golem,
	}
}

// ProcessTemplate processes a template using tree-based approach
func (tp *TreeProcessor) ProcessTemplate(template string, wildcards map[string]string, ctx *VariableContext) (string, error) {
	// Parse template into AST
	parser := NewASTParser(template)
	ast, err := parser.Parse()
	if err != nil {
		return template, err
	}

	// Store wildcards in context so they can be accessed by learn tag processing
	if ctx != nil {
		// Save current wildcards to restore them later
		oldWildcards := ctx.Wildcards
		ctx.Wildcards = wildcards

		defer func() {
			ctx.Wildcards = oldWildcards
		}()
	}

	// Store wildcards in session variables so they can be accessed by <star/> tags
	if ctx != nil && ctx.Session != nil && len(wildcards) > 0 {
		// Save current wildcards to restore them later
		oldSessionWildcards := make(map[string]string)
		for key := range wildcards {
			if value, exists := ctx.Session.Variables[key]; exists {
				oldSessionWildcards[key] = value
			}
		}

		// Set new wildcards
		for key, value := range wildcards {
			ctx.Session.Variables[key] = value
		}

		// Restore old wildcards after processing
		defer func() {
			// Remove current wildcards
			for key := range wildcards {
				delete(ctx.Session.Variables, key)
			}
			// Restore old wildcards
			for key, value := range oldSessionWildcards {
				ctx.Session.Variables[key] = value
			}
		}()
	}

	// Process the AST
	tp.ctx = ctx
	result := tp.processNode(ast)

	// Smart trimming: preserve intentional indentation; collapse whitespace-only to empty
	// This matches the behavior of the consolidated template processor
	finalResult := result
	if len(result) > 0 {
		// Collapse whitespace-only output to empty
		if strings.TrimSpace(result) == "" {
			finalResult = ""
		} else if result[0] != ' ' && result[0] != '\t' {
			// If it doesn't start with intentional indentation, trim normally
			finalResult = strings.TrimSpace(result)
		}
		// If it starts with space/tab and has non-whitespace content, preserve it (intentional indentation)
	}

	return finalResult, nil
}

// processNode processes a single AST node
func (tp *TreeProcessor) processNode(node *ASTNode) string {
	switch node.Type {
	case NodeTypeText:
		// If this is a text node with children, process children
		if len(node.Children) > 0 {
			children := ""
			for _, child := range node.Children {
				children += tp.processNode(child)
			}
			return children
		}
		return node.Content
	case NodeTypeComment:
		return "" // Comments are not output
	case NodeTypeCDATA:
		return node.Content // CDATA is output as-is
	case NodeTypeSelfClosingTag:
		return tp.processSelfClosingTag(node)
	case NodeTypeTag:
		return tp.processTag(node)
	default:
		return ""
	}
}

// processTag processes a tag node
func (tp *TreeProcessor) processTag(node *ASTNode) string {
	// Some tags need to process their children selectively (not all at once)
	// For those tags, skip pre-processing children
	skipChildProcessing := false
	switch node.TagName {
	case "random", "condition", "learn", "learnf":
		skipChildProcessing = true
	}

	// Process children first to handle nested tags (unless tag handles its own children)
	var content string
	if !skipChildProcessing {
		processedChildren := make([]string, len(node.Children))
		for i, child := range node.Children {
			processedChildren[i] = tp.processNode(child)
		}
		// Join processed children
		content = strings.Join(processedChildren, "")
	}

	// Process the tag based on its name
	switch node.TagName {
	case "srai":
		return tp.processSRAITag(node, content)
	case "sraix":
		return tp.processSRAIXTag(node, content)
	case "think":
		return tp.processThinkTag(node, content)
	case "set":
		return tp.processSetTag(node, content)
	case "get":
		return tp.processGetTag(node, content)
	case "bot":
		return tp.processBotTag(node, content)
	case "star":
		return tp.processStarTag(node, content)
	case "sr":
		return tp.processSRTag(node, content)
	case "that":
		return tp.processThatTag(node, content)
	case "topic":
		return tp.processTopicTag(node, content)
	case "random":
		return tp.processRandomTag(node, content)
	case "li":
		return tp.processListItemTag(node, content)
	case "condition":
		return tp.processConditionTag(node, content)
	case "map":
		return tp.processMapTag(node, content)
	case "list":
		return tp.processListTag(node, content)
	case "array":
		return tp.processArrayTag(node, content)
	case "learn":
		return tp.processLearnTag(node, content)
	case "learnf":
		return tp.processLearnfTag(node, content)
	case "uppercase":
		return tp.processUppercaseTag(node, content)
	case "lowercase":
		return tp.processLowercaseTag(node, content)
	case "formal":
		return tp.processFormalTag(node, content)
	case "capitalize":
		return tp.processCapitalizeTag(node, content)
	case "explode":
		return tp.processExplodeTag(node, content)
	case "reverse":
		return tp.processReverseTag(node, content)
	case "acronym":
		return tp.processAcronymTag(node, content)
	case "trim":
		return tp.processTrimTag(node, content)
	case "substring":
		return tp.processSubstringTag(node, content)
	case "replace":
		return tp.processReplaceTag(node, content)
	case "pluralize":
		return tp.processPluralizeTag(node, content)
	case "shuffle":
		return tp.processShuffleTag(node, content)
	case "length":
		return tp.processLengthTag(node, content)
	case "count":
		return tp.processCountTag(node, content)
	case "split":
		return tp.processSplitTag(node, content)
	case "join":
		return tp.processJoinTag(node, content)
	case "unique":
		return tp.processUniqueTag(node, content)
	case "indent":
		return tp.processIndentTag(node, content)
	case "dedent":
		return tp.processDedentTag(node, content)
	case "repeat":
		return tp.processRepeatTag(node, content)
	case "first":
		return tp.processFirstTag(node, content)
	case "rest":
		return tp.processRestTag(node, content)
	case "loop":
		return tp.processLoopTag(node, content)
	case "input":
		return tp.processInputTag(node, content)
	case "eval":
		return tp.processEvalTag(node, content)
	case "person":
		return tp.processPersonTag(node, content)
	case "person2":
		return tp.processPerson2Tag(node, content)
	case "gender":
		return tp.processGenderTag(node, content)
	case "sentence":
		return tp.processSentenceTag(node, content)
	case "word":
		return tp.processWordTag(node, content)
	case "date":
		return tp.processDateTag(node, content)
	case "time":
		return tp.processTimeTag(node, content)
	case "subj":
		return tp.processSubjTag(node, content)
	case "pred":
		return tp.processPredTag(node, content)
	case "obj":
		return tp.processObjTag(node, content)
	case "uniq":
		return tp.processUniqTag(node, content)
	case "size":
		return tp.processSizeTag(node, content)
	case "version":
		return tp.processVersionTag(node, content)
	case "id":
		return tp.processIdTag(node, content)
	case "request":
		return tp.processRequestTag(node, content)
	case "response":
		return tp.processResponseTag(node, content)
	case "normalize":
		return tp.processNormalizeTag(node, content)
	case "denormalize":
		return tp.processDenormalizeTag(node, content)
	case "unlearn":
		return tp.processUnlearnTag(node, content)
	case "unlearnf":
		return tp.processUnlearnfTag(node, content)
	case "var":
		return tp.processVarTag(node, content)
	case "gossip":
		return tp.processGossipTag(node, content)
	case "javascript":
		return tp.processJavascriptTag(node, content)
	case "system":
		return tp.processSystemTag(node, content)
	default:
		// Unknown tag, return as-is with processed content
		return fmt.Sprintf("<%s>%s</%s>", node.TagName, content, node.TagName)
	}
}

// processSelfClosingTag processes self-closing tags
func (tp *TreeProcessor) processSelfClosingTag(node *ASTNode) string {
	switch node.TagName {
	case "star":
		return tp.processStarTag(node, "")
	case "sr":
		return tp.processSRTag(node, "")
	case "input":
		return tp.processInputTag(node, "")
	case "loop":
		return tp.processLoopTag(node, "")
	case "date":
		return tp.processDateTag(node, "")
	case "time":
		return tp.processTimeTag(node, "")
	case "subj":
		return tp.processSubjTag(node, "")
	case "pred":
		return tp.processPredTag(node, "")
	case "obj":
		return tp.processObjTag(node, "")
	case "uniq":
		return tp.processUniqTag(node, "")
	case "size":
		return tp.processSizeTag(node, "")
	case "version":
		return tp.processVersionTag(node, "")
	case "id":
		return tp.processIdTag(node, "")
	case "request":
		return tp.processRequestTag(node, "")
	case "response":
		return tp.processResponseTag(node, "")
	case "get":
		return tp.processGetTag(node, "")
	case "that":
		return tp.processThatTag(node, "")
	case "bot":
		return tp.processBotTag(node, "")
	default:
		// Unknown self-closing tag, return as-is
		attrStr := ""
		if len(node.Attributes) > 0 {
			var attrs []string
			for k, v := range node.Attributes {
				if v == "" {
					attrs = append(attrs, k)
				} else {
					attrs = append(attrs, fmt.Sprintf(`%s="%s"`, k, v))
				}
			}
			attrStr = " " + strings.Join(attrs, " ")
		}
		return fmt.Sprintf("<%s%s/>", node.TagName, attrStr)
	}
}

// Tag processing methods

func (tp *TreeProcessor) processSRAITag(node *ASTNode, content string) string {
	// Process SRAI tag - recursive AIML processing (Symbolic Reduction and Inference)
	// Check recursion depth to prevent infinite recursion
	if tp.ctx == nil || tp.ctx.RecursionDepth >= MaxSRAIRecursionDepth {
		tp.golem.LogWarn("SRAI recursion depth limit reached (%d), stopping recursion", MaxSRAIRecursionDepth)
		return content
	}

	// Trim and normalize the content
	sraiContent := strings.TrimSpace(content)

	tp.golem.LogInfo("Processing SRAI: '%s' (depth: %d)", sraiContent, tp.ctx.RecursionDepth)

	// Try to match the SRAI content as a new AIML pattern
	if tp.golem.aimlKB != nil {
		category, wildcards, err := tp.golem.aimlKB.MatchPattern(sraiContent)
		tp.golem.LogInfo("SRAI pattern match: content='%s', err=%v, category=%v, wildcards=%v",
			sraiContent, err, category != nil, wildcards)

		if err == nil && category != nil {
			// Create a new context with incremented recursion depth
			// Preserve all context except increment recursion depth
			newCtx := &VariableContext{
				LocalVars:      tp.ctx.LocalVars,
				Session:        tp.ctx.Session,
				Topic:          tp.ctx.Topic,
				KnowledgeBase:  tp.ctx.KnowledgeBase,
				RecursionDepth: tp.ctx.RecursionDepth + 1,
				Wildcards:      tp.ctx.Wildcards, // Preserve parent wildcards
			}

			// Process the matched template with the new context
			response := tp.golem.processTemplateWithContext(category.Template, wildcards, newCtx)
			tp.golem.LogInfo("SRAI result: '%s' -> '%s'", sraiContent, response)
			return response
		} else {
			// No match found, return the content as-is
			tp.golem.LogInfo("SRAI no match for: '%s'", sraiContent)
			return sraiContent
		}
	}

	// No knowledge base, return content as-is
	return sraiContent
}

func (tp *TreeProcessor) processSRAIXTag(node *ASTNode, content string) string {
	// Process SRAIX tag - external service integration (SRAI eXtended)
	// Check if SRAIX manager is configured
	if tp.golem.sraixMgr == nil {
		tp.golem.LogInfo("SRAIX manager not configured, returning content")
		return content
	}

	// Get and evaluate attributes
	serviceName := ""
	if val, exists := node.Attributes["service"]; exists {
		serviceName = strings.TrimSpace(tp.evaluateAttributeValue(val))
	}

	botName := ""
	if val, exists := node.Attributes["bot"]; exists {
		botName = strings.TrimSpace(tp.evaluateAttributeValue(val))
	}

	botID := ""
	if val, exists := node.Attributes["botid"]; exists {
		botID = strings.TrimSpace(tp.evaluateAttributeValue(val))
	}

	hostName := ""
	if val, exists := node.Attributes["host"]; exists {
		hostName = strings.TrimSpace(tp.evaluateAttributeValue(val))
	}

	defaultResponse := ""
	if val, exists := node.Attributes["default"]; exists {
		defaultResponse = strings.TrimSpace(tp.evaluateAttributeValue(val))
	}

	hintText := ""
	if val, exists := node.Attributes["hint"]; exists {
		hintText = strings.TrimSpace(tp.evaluateAttributeValue(val))
	}

	// The content is already processed by the AST
	sraixContent := strings.TrimSpace(content)

	tp.golem.LogInfo("Processing SRAIX: service='%s', bot='%s', botid='%s', host='%s', default='%s', hint='%s', content='%s'",
		serviceName, botName, botID, hostName, defaultResponse, hintText, sraixContent)

	// Determine which service to use based on available attributes
	var targetService string
	if serviceName != "" {
		targetService = serviceName
	} else if botName != "" {
		// Use bot name as service identifier
		targetService = botName
	} else {
		tp.golem.LogInfo("SRAIX tag missing service or bot attribute")
		// Use default response if available
		if defaultResponse != "" {
			return defaultResponse
		}
		return content
	}

	// Build request parameters
	requestParams := make(map[string]string)
	if botID != "" {
		requestParams["botid"] = botID
	}
	if hostName != "" {
		requestParams["host"] = hostName
	}
	if hintText != "" {
		requestParams["hint"] = hintText
	}

	// Make the external service request
	response, err := tp.golem.sraixMgr.ProcessSRAIX(targetService, sraixContent, requestParams)
	if err != nil {
		tp.golem.LogInfo("SRAIX request failed: %v", err)
		// Use default response if available, otherwise return content
		if defaultResponse != "" {
			return defaultResponse
		}
		return content
	}

	tp.golem.LogInfo("SRAIX result: service='%s', input='%s' -> '%s'", targetService, sraixContent, response)
	return response
}

func (tp *TreeProcessor) processThinkTag(node *ASTNode, content string) string {
	// Process think tag - evaluates content but produces no output
	// The content parameter already contains the fully processed result of all child nodes
	// (variables have been set, operations performed, etc.)
	// We simply return empty string to suppress output

	tp.golem.LogInfo("Think tag: processed '%s' (no output)", content)

	// Think tags don't output anything
	return ""
}

// evaluateAttributeValue evaluates an attribute value if it contains AIML tags
// For example, name="<star/>" will be evaluated to the actual wildcard value
func (tp *TreeProcessor) evaluateAttributeValue(value string) string {
	// Quick check: if it doesn't contain '<', it's a plain string
	if !strings.Contains(value, "<") {
		return value
	}

	// Check if it contains AIML tags
	if strings.Contains(value, "<star") || strings.Contains(value, "<get") ||
		strings.Contains(value, "<bot") || strings.Contains(value, "<that") ||
		strings.Contains(value, "<input") || strings.Contains(value, "<id") {
		// Parse and evaluate the attribute value as AIML
		parser := NewASTParser(value)
		root, err := parser.Parse()
		if err != nil {
			// If parsing fails, return the original value
			return value
		}

		// Process the parsed tree
		var result strings.Builder
		for _, node := range root.Children {
			result.WriteString(tp.processNode(node))
		}
		return strings.TrimSpace(result.String())
	}

	return value
}

func (tp *TreeProcessor) processSetTag(node *ASTNode, content string) string {
	// Process set tag - can be either variable assignment OR collection operations
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	// Evaluate the name if it contains AIML tags (like <star/>)
	name = tp.evaluateAttributeValue(name)

	// Check if this is a collection operation (has operation attribute)
	operation, hasOperation := node.Attributes["operation"]

	if hasOperation {
		// This is a Set collection operation, not a variable assignment
		return tp.processSetCollectionTag(node, name, operation, content)
	}

	// No operation attribute - check if a Set collection with this name already exists
	// If yes, treat it as "get" operation; if no, treat as variable assignment
	if tp.ctx != nil && tp.ctx.KnowledgeBase != nil && tp.ctx.KnowledgeBase.SetCollections != nil {
		if _, exists := tp.ctx.KnowledgeBase.SetCollections[name]; exists {
			// Set collection exists, treat this as a "get" operation
			return tp.processSetCollectionTag(node, name, "get", content)
		}
	}

	// No operation and no existing Set collection - this is variable assignment (original behavior)
	// Process the content to get the value
	value := content // Content is already processed by processNode

	// Set the variable in context
	if tp.ctx != nil {
		// Set in session variables if session exists
		if tp.ctx.Session != nil {
			if tp.ctx.Session.Variables == nil {
				tp.ctx.Session.Variables = make(map[string]string)
			}
			tp.ctx.Session.Variables[name] = value
		} else if tp.ctx.KnowledgeBase != nil {
			// No session - set in knowledge base variables (global)
			if tp.ctx.KnowledgeBase.Variables == nil {
				tp.ctx.KnowledgeBase.Variables = make(map[string]string)
			}
			tp.ctx.KnowledgeBase.Variables[name] = value
		} else {
			// Fallback to local variables as last resort
			if tp.ctx.LocalVars == nil {
				tp.ctx.LocalVars = make(map[string]string)
			}
			tp.ctx.LocalVars[name] = value
		}
	}

	// Set tags don't output content
	return ""
}

func (tp *TreeProcessor) processSetCollectionTag(node *ASTNode, name string, operation string, content string) string {
	// Process Set collection operations (unique values)
	// Check for required knowledge base
	if tp.ctx == nil || tp.ctx.KnowledgeBase == nil || tp.ctx.KnowledgeBase.SetCollections == nil {
		tp.golem.LogInfo("Set collection: no knowledge base available")
		return ""
	}

	// Get or create the set
	if tp.ctx.KnowledgeBase.SetCollections[name] == nil {
		tp.ctx.KnowledgeBase.SetCollections[name] = make(map[string]bool)
		tp.golem.LogInfo("Created new set collection '%s'", name)
	}

	setData := tp.ctx.KnowledgeBase.SetCollections[name]
	item := strings.TrimSpace(content)

	tp.golem.LogInfo("Set collection tag: name='%s', operation='%s', item='%s'", name, operation, item)
	tp.golem.LogInfo("Before operation: set '%s' has %d items", name, len(setData))

	switch operation {
	case "add":
		// Add item to set (only if not already present)
		if item != "" {
			setData[item] = true
			tp.golem.LogInfo("Added '%s' to set '%s'", item, name)
		}
		return "" // Add operations don't return content

	case "remove", "delete":
		// Remove item from set
		if item != "" {
			if _, exists := setData[item]; exists {
				delete(setData, item)
				tp.golem.LogInfo("Removed '%s' from set '%s'", item, name)
			} else {
				tp.golem.LogInfo("Item '%s' not found in set '%s'", item, name)
			}
		}
		return "" // Remove operations don't return content

	case "clear":
		// Clear all items from set
		tp.ctx.KnowledgeBase.SetCollections[name] = make(map[string]bool)
		tp.golem.LogInfo("Cleared set '%s'", name)
		return "" // Clear operations don't return content

	case "size", "length":
		// Return the size of the set
		size := strconv.Itoa(len(setData))
		tp.golem.LogInfo("Set '%s' size: %s", name, size)
		return size

	case "contains", "has":
		// Check if set contains item
		_, contains := setData[item]
		result := "false"
		if contains {
			result = "true"
		}
		tp.golem.LogInfo("Set '%s' contains '%s': %s", name, item, result)
		return result

	case "get", "":
		// Return all items in set (space-separated, sorted for consistency)
		items := make([]string, 0, len(setData))
		for item := range setData {
			items = append(items, item)
		}
		sort.Strings(items) // Sort for consistent output
		result := strings.Join(items, " ")
		tp.golem.LogInfo("Got all items from set '%s': '%s'", name, result)
		return result

	default:
		// Unknown operation, return all items
		tp.golem.LogInfo("Unknown operation '%s', returning all items", operation)
		items := make([]string, 0, len(setData))
		for item := range setData {
			items = append(items, item)
		}
		sort.Strings(items)
		return strings.Join(items, " ")
	}
}

func (tp *TreeProcessor) processGetTag(node *ASTNode, content string) string {
	// Process get tag - variable retrieval
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	// Evaluate the name if it contains AIML tags (like <star/>)
	name = tp.evaluateAttributeValue(name)

	// Get the variable value from context
	if tp.ctx != nil {
		// Check local variables first
		if tp.ctx.LocalVars != nil {
			if value, exists := tp.ctx.LocalVars[name]; exists {
				return value
			}
		}
		// Check session variables
		if tp.ctx.Session != nil && tp.ctx.Session.Variables != nil {
			if value, exists := tp.ctx.Session.Variables[name]; exists {
				return value
			}
		}
		// Check topic variables
		if tp.ctx.Topic != "" && tp.ctx.KnowledgeBase != nil && tp.ctx.KnowledgeBase.TopicVars != nil {
			if topicVars, exists := tp.ctx.KnowledgeBase.TopicVars[tp.ctx.Topic]; exists {
				if value, exists := topicVars[name]; exists {
					return value
				}
			}
		}
		// Check global variables (from knowledge base)
		if tp.ctx.KnowledgeBase != nil && tp.ctx.KnowledgeBase.Variables != nil {
			if value, exists := tp.ctx.KnowledgeBase.Variables[name]; exists {
				return value
			}
		}
		// Check bot properties
		if tp.ctx.KnowledgeBase != nil && tp.ctx.KnowledgeBase.Properties != nil {
			if value, exists := tp.ctx.KnowledgeBase.Properties[name]; exists {
				return value
			}
		}
	}

	// If variable not found, return the processed content as default
	return content
}

func (tp *TreeProcessor) processBotTag(node *ASTNode, content string) string {
	// Process bot tag - bot property access
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	// Get bot property from knowledge base
	if tp.ctx != nil && tp.ctx.KnowledgeBase != nil {
		if value, exists := tp.ctx.KnowledgeBase.Properties[name]; exists {
			return value
		}
	}

	// Property not found
	return ""
}

func (tp *TreeProcessor) processStarTag(node *ASTNode, content string) string {
	// Process star tag - wildcard reference
	index := 1
	if idx, exists := node.Attributes["index"]; exists {
		if parsed, err := strconv.Atoi(idx); err == nil {
			index = parsed
		}
	}

	key := fmt.Sprintf("star%d", index)

	// Get wildcard value - check session variables first, then wildcards in context
	if tp.ctx != nil {
		if tp.ctx.Session != nil {
			if value, exists := tp.ctx.Session.Variables[key]; exists {
				return value
			}
		}
		// Also check the Wildcards map directly (for cases without a session)
		if tp.ctx.Wildcards != nil {
			if value, exists := tp.ctx.Wildcards[key]; exists {
				return value
			}
		}
	}

	return ""
}

func (tp *TreeProcessor) processSRTag(node *ASTNode, content string) string {
	// Process SR tag - shorthand for <srai><star/></srai>
	// SR recursively processes the first wildcard (star1)

	if tp.ctx == nil {
		tp.golem.LogDebug("SR tag: no context available")
		return ""
	}

	// Get the first wildcard (star1) - check session variables first, then wildcards in context
	starContent := ""
	if tp.ctx.Session != nil {
		if value, exists := tp.ctx.Session.Variables["star1"]; exists {
			starContent = value
		}
	}
	// Also check the Wildcards map directly (for cases without a session)
	if starContent == "" && tp.ctx.Wildcards != nil {
		if value, exists := tp.ctx.Wildcards["star1"]; exists {
			starContent = value
		}
	}

	tp.golem.LogDebug("SR tag: star1 content='%s'", starContent)

	// If no star content, return empty
	if starContent == "" {
		tp.golem.LogDebug("SR tag: no star content available")
		return ""
	}

	// If no knowledge base, return empty
	if tp.ctx.KnowledgeBase == nil {
		tp.golem.LogDebug("SR tag: no knowledge base available")
		return ""
	}

	// Try to match the star content as a pattern in the knowledge base
	category, wildcards, err := tp.ctx.KnowledgeBase.MatchPattern(starContent)
	if err != nil || category == nil {
		tp.golem.LogDebug("SR tag: no matching pattern for '%s'", starContent)
		return ""
	}

	tp.golem.LogDebug("SR tag: found matching pattern for '%s'", starContent)

	// Check recursion depth to prevent infinite loops
	if tp.ctx.RecursionDepth >= 100 {
		tp.golem.LogDebug("SR tag: max recursion depth reached")
		return ""
	}

	// Increment recursion depth
	oldDepth := tp.ctx.RecursionDepth
	tp.ctx.RecursionDepth++
	defer func() {
		tp.ctx.RecursionDepth = oldDepth
	}()

	// Store old wildcards and restore them after processing
	oldWildcards := make(map[string]string)
	if tp.ctx.Session != nil && tp.ctx.Session.Variables != nil {
		// Save current wildcards
		for k, v := range tp.ctx.Session.Variables {
			if strings.HasPrefix(k, "star") {
				oldWildcards[k] = v
			}
		}

		// Set new wildcards from the matched pattern
		for k, v := range wildcards {
			tp.ctx.Session.Variables[k] = v
		}
	} else if tp.ctx.Wildcards != nil {
		// No session - use context wildcards instead
		// Save current wildcards
		for k, v := range tp.ctx.Wildcards {
			if strings.HasPrefix(k, "star") {
				oldWildcards[k] = v
			}
		}

		// Set new wildcards from the matched pattern
		for k, v := range wildcards {
			tp.ctx.Wildcards[k] = v
		}
	}

	// Process the matched template recursively
	result := tp.golem.processTemplateWithContext(category.Template, wildcards, tp.ctx)

	// Restore old wildcards
	if tp.ctx.Session != nil && tp.ctx.Session.Variables != nil {
		// Remove wildcards from the matched pattern
		for k := range wildcards {
			delete(tp.ctx.Session.Variables, k)
		}
		// Restore original wildcards
		for k, v := range oldWildcards {
			tp.ctx.Session.Variables[k] = v
		}
	} else if tp.ctx.Wildcards != nil {
		// Remove wildcards from the matched pattern
		for k := range wildcards {
			delete(tp.ctx.Wildcards, k)
		}
		// Restore original wildcards
		for k, v := range oldWildcards {
			tp.ctx.Wildcards[k] = v
		}
	}

	tp.golem.LogDebug("SR tag: result='%s'", result)

	return result
}

func (tp *TreeProcessor) processThatTag(node *ASTNode, content string) string {
	// Process that tag - previous response reference
	// <that/> or <that> with no index returns the most recent response (index 1)
	// <that index="N"/> returns the Nth most recent response

	if tp.ctx == nil || tp.ctx.Session == nil {
		return ""
	}

	// Get the index attribute, default to 1 (most recent)
	index := 1
	if indexStr, exists := node.Attributes["index"]; exists {
		if parsed, err := strconv.Atoi(indexStr); err == nil && parsed > 0 {
			index = parsed
		}
	}

	// Get the response by index
	response := tp.ctx.Session.GetResponseByIndex(index)

	tp.golem.LogDebug("That tag: index=%d, response='%s'", index, response)

	return response
}

func (tp *TreeProcessor) processTopicTag(node *ASTNode, content string) string {
	// Process topic tag - topic reference
	index := 1
	if idx, exists := node.Attributes["index"]; exists {
		if parsed, err := strconv.Atoi(idx); err == nil {
			index = parsed
		}
	}

	// Get topic value
	if tp.ctx != nil {
		if index == 1 {
			return tp.ctx.Topic
		}
	}

	return ""
}

func (tp *TreeProcessor) processRandomTag(node *ASTNode, content string) string {
	// Process random tag - random selection from list items
	var items []string
	for _, child := range node.Children {
		if child.Type == NodeTypeTag && child.TagName == "li" {
			item := tp.processNode(child)
			// Trim whitespace from each item
			item = strings.TrimSpace(item)
			if item != "" {
				items = append(items, item)
			}
		}
	}

	if len(items) == 0 {
		return content
	}

	// Select random item
	index := tp.golem.randomIntTree(len(items))
	return items[index]
}

func (tp *TreeProcessor) processListItemTag(node *ASTNode, content string) string {
	// Process list item tag - process and return children
	var result strings.Builder
	for _, child := range node.Children {
		result.WriteString(tp.processNode(child))
	}
	return result.String()
}

func (tp *TreeProcessor) processConditionTag(node *ASTNode, content string) string {
	// Process condition tag - conditional logic (native implementation)

	// Get the variable name and expected value from attributes
	varName, hasName := node.Attributes["name"]
	expectedValue, hasExpectedValue := node.Attributes["value"]

	// Get the actual variable value
	var actualValue string
	if hasName {
		actualValue = tp.golem.resolveVariable(varName, tp.ctx)
	}

	// Type 1: Simple condition with value attribute
	if hasExpectedValue {
		if strings.EqualFold(actualValue, expectedValue) {
			// Process children
			var result strings.Builder
			for _, child := range node.Children {
				result.WriteString(tp.processNode(child))
			}
			return result.String()
		}
		return "" // No match
	}

	// Type 2: Multiple <li> conditions
	var defaultLi *ASTNode
	for _, child := range node.Children {
		if child.Type == NodeTypeTag && child.TagName == "li" {
			liValue, hasValue := child.Attributes["value"]

			// If no value, this is the default case - save it for later
			if !hasValue || liValue == "" {
				defaultLi = child
				continue
			}

			// Check if this condition matches
			if strings.EqualFold(actualValue, liValue) {
				// Process this li's children
				var result strings.Builder
				for _, liChild := range child.Children {
					result.WriteString(tp.processNode(liChild))
				}
				return strings.TrimSpace(result.String())
			}
		}
	}

	// No match found, use default <li> if available
	if defaultLi != nil {
		var result strings.Builder
		for _, liChild := range defaultLi.Children {
			result.WriteString(tp.processNode(liChild))
		}
		return strings.TrimSpace(result.String())
	}

	// Type 3: No <li> elements and no value - just check if variable has a value
	if hasName && actualValue != "" {
		var result strings.Builder
		for _, child := range node.Children {
			result.WriteString(tp.processNode(child))
		}
		return result.String()
	}

	return "" // No match
}

func (tp *TreeProcessor) processMapTag(node *ASTNode, content string) string {
	// Process map tag - mapping operations
	// Check for required knowledge base
	if tp.ctx == nil || tp.ctx.KnowledgeBase == nil || tp.ctx.KnowledgeBase.Maps == nil {
		return content
	}

	// Get the map name
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	// Evaluate the map name if it contains tags
	name = tp.evaluateAttributeValue(name)

	// Get the key attribute (optional)
	keyAttr, hasKeyAttr := node.Attributes["key"]
	if hasKeyAttr {
		keyAttr = tp.evaluateAttributeValue(keyAttr)
	}

	// Get the operation (optional, default is "get")
	operation, hasOperation := node.Attributes["operation"]
	if !hasOperation {
		operation = "get"
	}

	// Determine the key: if key attribute is provided, use it; otherwise use content
	key := keyAttr
	if !hasKeyAttr || key == "" {
		key = strings.TrimSpace(content)
	}

	tp.golem.LogInfo("Map tag: name='%s', key='%s', operation='%s', content='%s'", name, key, operation, content)

	// Get or create the map
	if tp.ctx.KnowledgeBase.Maps[name] == nil {
		tp.ctx.KnowledgeBase.Maps[name] = make(map[string]string)
		tp.golem.LogInfo("Created new map '%s'", name)
	}

	mapData := tp.ctx.KnowledgeBase.Maps[name]
	tp.golem.LogInfo("Before operation: map '%s' = %v", name, mapData)

	switch operation {
	case "set", "assign":
		// Set a key-value pair
		if key != "" {
			// Use content as value
			value := strings.TrimSpace(content)
			if hasKeyAttr {
				// Key was in attribute, content is the value
				value = content
			} else {
				// Key was in content, value is also content (for now)
				value = content
			}
			tp.ctx.KnowledgeBase.Maps[name][key] = strings.TrimSpace(value)
			tp.golem.LogInfo("Set map '%s'['%s'] = '%s'", name, key, strings.TrimSpace(value))
			tp.golem.LogInfo("After set: map '%s' = %v", name, tp.ctx.KnowledgeBase.Maps[name])
			return "" // Set operations don't return content
		}
		return ""

	case "remove", "delete":
		// Remove a key-value pair
		if key != "" {
			if _, exists := tp.ctx.KnowledgeBase.Maps[name][key]; exists {
				delete(tp.ctx.KnowledgeBase.Maps[name], key)
				tp.golem.LogInfo("Removed key '%s' from map '%s'", key, name)
				tp.golem.LogInfo("After remove: map '%s' = %v", name, tp.ctx.KnowledgeBase.Maps[name])
			} else {
				tp.golem.LogInfo("Key '%s' not found in map '%s'", key, name)
			}
		}
		return "" // Remove operations don't return content

	case "clear":
		// Clear all entries
		tp.ctx.KnowledgeBase.Maps[name] = make(map[string]string)
		tp.golem.LogInfo("Cleared map '%s'", name)
		return "" // Clear operations don't return content

	case "size", "length":
		// Return the size of the map
		size := strconv.Itoa(len(tp.ctx.KnowledgeBase.Maps[name]))
		tp.golem.LogInfo("Map '%s' size: %s", name, size)
		return size

	case "contains", "has":
		// Check if map contains key
		contains := false
		if key != "" {
			_, contains = tp.ctx.KnowledgeBase.Maps[name][key]
		}
		result := "false"
		if contains {
			result = "true"
		}
		tp.golem.LogInfo("Map '%s' contains key '%s': %s", name, key, result)
		return result

	case "keys":
		// Return all keys
		keys := make([]string, 0, len(tp.ctx.KnowledgeBase.Maps[name]))
		for k := range tp.ctx.KnowledgeBase.Maps[name] {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Sort for consistent output
		keysString := strings.Join(keys, " ")
		tp.golem.LogInfo("Map '%s' keys: %s", name, keysString)
		return keysString

	case "values":
		// Return all values
		values := make([]string, 0, len(tp.ctx.KnowledgeBase.Maps[name]))
		for _, v := range tp.ctx.KnowledgeBase.Maps[name] {
			values = append(values, v)
		}
		sort.Strings(values) // Sort for consistent output
		valuesString := strings.Join(values, " ")
		tp.golem.LogInfo("Map '%s' values: %s", name, valuesString)
		return valuesString

	case "list":
		// Return all key-value pairs
		pairs := make([]string, 0, len(tp.ctx.KnowledgeBase.Maps[name]))
		for k, v := range tp.ctx.KnowledgeBase.Maps[name] {
			pairs = append(pairs, k+":"+v)
		}
		sort.Strings(pairs) // Sort for consistent output
		pairsString := strings.Join(pairs, " ")
		tp.golem.LogInfo("Map '%s' pairs: %s", name, pairsString)
		return pairsString

	case "get", "":
		// Get value by key (original functionality)
		if key != "" {
			if value, exists := tp.ctx.KnowledgeBase.Maps[name][key]; exists {
				tp.golem.LogInfo("Mapped '%s' -> '%s'", key, value)
				return value
			} else {
				// Key not found in map, return the original key
				tp.golem.LogInfo("Key '%s' not found in map '%s', returning key", key, name)
				return key
			}
		}
		return ""

	default:
		// Unknown operation, treat as get
		tp.golem.LogInfo("Unknown operation '%s', treating as get", operation)
		if key != "" {
			if value, exists := tp.ctx.KnowledgeBase.Maps[name][key]; exists {
				return value
			}
			return key
		}
		return ""
	}
}

func (tp *TreeProcessor) processListTag(node *ASTNode, content string) string {
	// Process list tag - list operations
	// Get the list name
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	// Evaluate the list name if it contains tags
	name = tp.evaluateAttributeValue(name)

	// Get the index attribute (optional)
	indexStr, hasIndex := node.Attributes["index"]
	if hasIndex {
		indexStr = tp.evaluateAttributeValue(indexStr)
	}

	// Get the operation (optional)
	operation, hasOperation := node.Attributes["operation"]
	if !hasOperation {
		operation = ""
	}

	// If no knowledge base, just return empty string for operations
	if tp.ctx == nil || tp.ctx.KnowledgeBase == nil || tp.ctx.KnowledgeBase.Lists == nil {
		tp.golem.LogInfo("List processing: no knowledge base available")
		return ""
	}

	// Get or create the list
	if tp.ctx.KnowledgeBase.Lists[name] == nil {
		tp.ctx.KnowledgeBase.Lists[name] = make([]string, 0)
		tp.golem.LogInfo("Created new list '%s'", name)
	}
	list := tp.ctx.KnowledgeBase.Lists[name]
	tp.golem.LogInfo("Processing list tag: name='%s', index='%s', operation='%s', content='%s'", name, indexStr, operation, content)
	tp.golem.LogInfo("Before operation: list '%s' = %v", name, list)

	switch operation {
	case "add", "append":
		// Add item to the end of the list
		list = append(list, content)
		tp.ctx.KnowledgeBase.Lists[name] = list
		tp.golem.LogInfo("Added '%s' to list '%s'", content, name)
		tp.golem.LogInfo("After add: list '%s' = %v", name, list)
		return "" // List operations don't return content

	case "insert":
		// Insert item at specific index
		if hasIndex {
			if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index <= len(list) {
				// Insert at the specified index
				list = append(list[:index], append([]string{content}, list[index:]...)...)
				tp.ctx.KnowledgeBase.Lists[name] = list
				tp.golem.LogInfo("Inserted '%s' at index %d in list '%s'", content, index, name)
				tp.golem.LogInfo("After insert: list '%s' = %v", name, list)
			} else {
				// Invalid index, append to end
				list = append(list, content)
				tp.ctx.KnowledgeBase.Lists[name] = list
				tp.golem.LogInfo("Invalid index %s, appended '%s' to list '%s'", indexStr, content, name)
				tp.golem.LogInfo("After append: list '%s' = %v", name, list)
			}
		} else {
			// No index specified, append to end
			list = append(list, content)
			tp.ctx.KnowledgeBase.Lists[name] = list
			tp.golem.LogInfo("No index specified, appended '%s' to list '%s'", content, name)
			tp.golem.LogInfo("After append: list '%s' = %v", name, list)
		}
		return ""

	case "remove", "delete":
		// Remove item from list
		if hasIndex {
			if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(list) {
				// Remove at specific index
				list = append(list[:index], list[index+1:]...)
				tp.ctx.KnowledgeBase.Lists[name] = list
				tp.golem.LogInfo("Removed item at index %d from list '%s'", index, name)
				tp.golem.LogInfo("After remove by index: list '%s' = %v", name, list)
			} else {
				// Invalid index, try to remove by value
				for i, item := range list {
					if item == content {
						list = append(list[:i], list[i+1:]...)
						tp.ctx.KnowledgeBase.Lists[name] = list
						tp.golem.LogInfo("Removed '%s' from list '%s'", content, name)
						tp.golem.LogInfo("After remove by value: list '%s' = %v", name, list)
						break
					}
				}
			}
		} else {
			// No index, remove by value
			for i, item := range list {
				if item == content {
					list = append(list[:i], list[i+1:]...)
					tp.ctx.KnowledgeBase.Lists[name] = list
					tp.golem.LogInfo("Removed '%s' from list '%s'", content, name)
					tp.golem.LogInfo("After remove by value: list '%s' = %v", name, list)
					break
				}
			}
		}
		return ""

	case "clear":
		// Clear the list
		tp.ctx.KnowledgeBase.Lists[name] = make([]string, 0)
		tp.golem.LogInfo("Cleared list '%s'", name)
		return ""

	case "size", "length":
		// Return size of list
		size := strconv.Itoa(len(list))
		tp.golem.LogInfo("List '%s' size: %s", name, size)
		return size

	case "get", "":
		// Get item(s) from list
		if hasIndex {
			// Get item at specific index
			if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(list) {
				tp.golem.LogInfo("Got list '%s'[%d] = '%s'", name, index, list[index])
				return list[index]
			}
			// Invalid index
			tp.golem.LogInfo("Invalid index %s for list '%s'", indexStr, name)
			return ""
		}
		// No index, return all items joined
		result := strings.Join(list, " ")
		tp.golem.LogInfo("Got all items from list '%s': '%s'", name, result)
		return result

	default:
		// Unknown operation, return all items
		tp.golem.LogInfo("Unknown operation '%s', returning all items", operation)
		return strings.Join(list, " ")
	}
}

func (tp *TreeProcessor) processArrayTag(node *ASTNode, content string) string {
	// Process array tag - array operations
	// Get the array name
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	// Evaluate the array name if it contains tags
	name = tp.evaluateAttributeValue(name)

	// Get the index attribute (optional)
	indexStr, hasIndex := node.Attributes["index"]
	if hasIndex {
		indexStr = tp.evaluateAttributeValue(indexStr)
	}

	// Get the operation (optional, default is "get")
	operation, hasOperation := node.Attributes["operation"]
	if !hasOperation {
		operation = "get"
	}

	// If no knowledge base, just return empty string
	if tp.ctx == nil || tp.ctx.KnowledgeBase == nil || tp.ctx.KnowledgeBase.Arrays == nil {
		tp.golem.LogInfo("Array processing: no knowledge base available")
		return ""
	}

	// Get or create the array
	if tp.ctx.KnowledgeBase.Arrays[name] == nil {
		tp.ctx.KnowledgeBase.Arrays[name] = make([]string, 0)
		tp.golem.LogInfo("Created new array '%s'", name)
	}
	array := tp.ctx.KnowledgeBase.Arrays[name]
	tp.golem.LogInfo("Processing array tag: name='%s', index='%s', operation='%s', content='%s'", name, indexStr, operation, content)
	tp.golem.LogInfo("Before operation: array '%s' = %v", name, array)

	switch operation {
	case "set", "assign":
		// Set item at specific index
		if hasIndex {
			if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 {
				// Ensure array is large enough
				for len(array) <= index {
					array = append(array, "")
				}
				array[index] = content
				tp.ctx.KnowledgeBase.Arrays[name] = array
				tp.golem.LogInfo("Set array '%s'[%d] = '%s'", name, index, content)
				tp.golem.LogInfo("After set: array '%s' = %v", name, array)
			} else {
				// Invalid index
				tp.golem.LogInfo("Invalid index %s for array '%s'", indexStr, name)
			}
		} else {
			// No index specified, append to end
			array = append(array, content)
			tp.ctx.KnowledgeBase.Arrays[name] = array
			tp.golem.LogInfo("Appended '%s' to array '%s'", content, name)
			tp.golem.LogInfo("After append: array '%s' = %v", name, array)
		}
		return "" // Set operations don't return content

	case "get", "":
		// Get item at index
		if hasIndex {
			if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(array) {
				tp.golem.LogInfo("Got array '%s'[%d] = '%s'", name, index, array[index])
				return array[index]
			}
			// Invalid index
			tp.golem.LogInfo("Invalid index %s for array '%s'", indexStr, name)
			return ""
		}
		// Return all items joined by space
		items := strings.Join(array, " ")
		tp.golem.LogInfo("Got all items from array '%s': '%s'", name, items)
		return items

	case "size", "length":
		// Return the size of the array
		size := strconv.Itoa(len(array))
		tp.golem.LogInfo("Array '%s' size: %s", name, size)
		return size

	case "clear":
		// Clear the array
		tp.ctx.KnowledgeBase.Arrays[name] = make([]string, 0)
		tp.golem.LogInfo("Cleared array '%s'", name)
		tp.golem.LogInfo("After clear: array '%s' = %v", name, tp.ctx.KnowledgeBase.Arrays[name])
		return ""

	default:
		// Unknown operation, treat as get
		tp.golem.LogInfo("Unknown operation '%s', treating as get", operation)
		if hasIndex {
			if index, err := strconv.Atoi(indexStr); err == nil && index >= 0 && index < len(array) {
				return array[index]
			}
			return ""
		}
		return strings.Join(array, " ")
	}
}

func (tp *TreeProcessor) processLearnTag(node *ASTNode, content string) string {
	// Process learn tag - dynamic learning (session-specific)
	// Process content while preserving wildcard/reference tags
	// This evaluates tags like <get>, <uppercase>, etc. but preserves <star/>, <that/>, etc.
	processedContent := tp.processNodePreservingReferences(node)

	// The underlying function processes both <learn> and <learnf> tags via regex
	return tp.golem.processLearnTagsWithContext(fmt.Sprintf("<learn>%s</learn>", processedContent), tp.ctx)
}

func (tp *TreeProcessor) processLearnfTag(node *ASTNode, content string) string {
	// Process learnf tag - persistent learning
	// The <learnf> tag adds categories to the persistent knowledge base
	// Unlike <learn>, these persist across sessions
	// Process content while preserving wildcard/reference tags
	// This evaluates tags like <get>, <uppercase>, etc. but preserves <star/>, <that/>, etc.
	processedContent := tp.processNodePreservingReferences(node)

	// The underlying function processes both <learn> and <learnf> tags via regex
	return tp.golem.processLearnTagsWithContext(fmt.Sprintf("<learnf>%s</learnf>", processedContent), tp.ctx)
}

// processNodePreservingReferences processes a node's children while preserving reference tags
// Reference tags (like <star/>, <that/>, <input/>, etc.) are output as their string representation
// Other tags (like <get/>, <uppercase/>, etc.) are processed normally
func (tp *TreeProcessor) processNodePreservingReferences(node *ASTNode) string {
	var result strings.Builder

	for _, child := range node.Children {
		result.WriteString(tp.processChildPreservingReferences(child))
	}

	return result.String()
}

// processChildPreservingReferences processes a single child node
// Returns the string representation for reference tags, processed content for others
func (tp *TreeProcessor) processChildPreservingReferences(node *ASTNode) string {
	// For text nodes, return content as-is
	if node.Type == NodeTypeText {
		if len(node.Children) > 0 {
			var result strings.Builder
			for _, child := range node.Children {
				result.WriteString(tp.processChildPreservingReferences(child))
			}
			return result.String()
		}
		return node.Content
	}

	// For comments and CDATA, return as-is
	if node.Type == NodeTypeComment || node.Type == NodeTypeCDATA {
		return node.String()
	}

	// For tags, check if they should be preserved as references
	if node.Type == NodeTypeSelfClosingTag || node.Type == NodeTypeTag {
		// List of tags that should be preserved (wildcards, history references, formatting, and variables)
		// These tags should not be processed during learning, but preserved for runtime
		preservedTags := map[string]bool{
			"star":      true, // Wildcard references
			"that":      true, // Response history
			"thatstar":  true, // That wildcard
			"topicstar": true, // Topic wildcard
			"input":     true, // Request history (alternative form)
			"request":   true, // Request history
			"response":  true, // Response history
			"sr":        true, // Shorthand SRAI - should be preserved for runtime
			// Formatting tags - preserve during learning
			"uppercase": true,
			"lowercase": true,
			"formal":    true,
			"sentence":  true,
			"explode":   true,
			"normalize": true,
			// Variable tags - preserve during learning so they evaluate at runtime
			"get":        true,
			"set":        true,
			"bot":        true,
			"name":       true,
			"id":         true,
			"size":       true,
			"version":    true,
			"date":       true,
			"vocabulary": true,
			// Recursive tags - preserve for runtime evaluation
			"srai":  true,
			"sraix": true,
			// Conditional tags - preserve for runtime evaluation
			"condition": true,
			"li":        true,
			// Random tags - preserve for runtime evaluation
			"random": true,
		}

		if preservedTags[node.TagName] {
			// Return the tag as its string representation
			return node.String()
		}

		// For non-preserved tags, process them normally
		// But we need to recursively preserve references in their children
		if node.Type == NodeTypeTag {
			// Process children while preserving references
			var processedChildren strings.Builder
			for _, child := range node.Children {
				processedChildren.WriteString(tp.processChildPreservingReferences(child))
			}

			// Now process this tag with the processed children content
			// We need to temporarily set up the node's processed content
			// and call the appropriate tag processor
			return tp.processTagWithContent(node, processedChildren.String())
		} else {
			// Self-closing tag - process it
			return tp.processSelfClosingTag(node)
		}
	}

	// For other node types, just process normally
	return tp.processNode(node)
}

// processTagWithContent processes a tag with given content
// This is similar to processTag but uses provided content instead of processing children
func (tp *TreeProcessor) processTagWithContent(node *ASTNode, content string) string {
	// Helper function to format tag with attributes
	formatTag := func(tagName string, attrs map[string]string, content string) string {
		if len(attrs) == 0 {
			return fmt.Sprintf("<%s>%s</%s>", tagName, content, tagName)
		}

		var attrStr strings.Builder
		for key, value := range attrs {
			if value == "" {
				attrStr.WriteString(fmt.Sprintf(" %s", key))
			} else {
				attrStr.WriteString(fmt.Sprintf(` %s="%s"`, key, value))
			}
		}

		return fmt.Sprintf("<%s%s>%s</%s>", tagName, attrStr.String(), content, tagName)
	}

	// Process the tag based on its name
	switch node.TagName {
	case "template", "think", "random", "li":
		// For structural tags, preserve with attributes if any
		return formatTag(node.TagName, node.Attributes, content)
	case "condition":
		// Condition tags need special handling - preserve the structure
		return formatTag("condition", node.Attributes, content)
	case "pattern", "that", "topic":
		// Pattern-related tags should be preserved with their content and attributes
		return formatTag(node.TagName, node.Attributes, content)
	case "category":
		// Category tag should be preserved
		return fmt.Sprintf("<category>%s</category>", content)
	case "get":
		return tp.processGetTag(node, content)
	case "set":
		return tp.processSetTag(node, content)
	case "bot":
		return tp.processBotTag(node, content)
	case "uppercase":
		return tp.processUppercaseTag(node, content)
	case "lowercase":
		return tp.processLowercaseTag(node, content)
	case "formal":
		return tp.processFormalTag(node, content)
	case "sentence":
		return tp.processSentenceTag(node, content)
	case "person":
		return tp.processPersonTag(node, content)
	case "person2":
		return tp.processPerson2Tag(node, content)
	case "gender":
		return tp.processGenderTag(node, content)
	case "srai":
		return tp.processSRAITag(node, content)
	case "eval":
		return tp.processEvalTag(node, content)
	default:
		// For unknown tags, return content wrapped in the tag with attributes
		return formatTag(node.TagName, node.Attributes, content)
	}
}

// Text processing tags

func (tp *TreeProcessor) processUppercaseTag(node *ASTNode, content string) string {
	// Process content directly - convert to uppercase
	processedContent := strings.ToUpper(content)

	// Normalize whitespace like the original method
	processedContent = strings.TrimSpace(processedContent)
	if processedContent == "" && len(content) > 0 {
		return content // Preserve whitespace-only content
	}

	// Normalize internal whitespace
	processedContent = regexp.MustCompile(`\s+`).ReplaceAllString(processedContent, " ")

	return processedContent
}

func (tp *TreeProcessor) processLowercaseTag(node *ASTNode, content string) string {
	// Process content directly - convert to lowercase
	processedContent := strings.ToLower(content)

	// Normalize whitespace like the original method
	processedContent = strings.TrimSpace(processedContent)
	if processedContent == "" && len(content) > 0 {
		return content // Preserve whitespace-only content
	}

	// Normalize internal whitespace
	processedContent = regexp.MustCompile(`\s+`).ReplaceAllString(processedContent, " ")

	return processedContent
}

func (tp *TreeProcessor) processFormalTag(node *ASTNode, content string) string {
	// Process content directly - capitalize first letter of each word
	words := strings.Fields(content)
	var result []string

	for _, word := range words {
		if len(word) > 0 {
			// Capitalize first letter, lowercase the rest
			capitalized := strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
			result = append(result, capitalized)
		}
	}

	return strings.Join(result, " ")
}

func (tp *TreeProcessor) processCapitalizeTag(node *ASTNode, content string) string {
	// Process content directly - capitalize first letter only
	if content == "" {
		return content
	}

	// Capitalize first letter, keep rest as-is
	return strings.ToUpper(string(content[0])) + content[1:]
}

func (tp *TreeProcessor) processExplodeTag(node *ASTNode, content string) string {
	// Process content directly - add spaces between characters
	var result strings.Builder
	for i, char := range content {
		if i > 0 {
			result.WriteRune(' ')
		}
		result.WriteRune(char)
	}
	return result.String()
}

func (tp *TreeProcessor) processReverseTag(node *ASTNode, content string) string {
	// Process content directly - reverse the string
	runes := []rune(content)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (tp *TreeProcessor) processAcronymTag(node *ASTNode, content string) string {
	// Process content directly - convert to acronym (first letter of each word, uppercase)
	words := strings.Fields(content)
	var acronym strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			// Take first letter and uppercase it
			firstLetter := strings.ToUpper(string(word[0]))
			acronym.WriteString(firstLetter)
		}
	}
	return acronym.String()
}

func (tp *TreeProcessor) processTrimTag(node *ASTNode, content string) string {
	// Process content directly - trim whitespace
	return strings.TrimSpace(content)
}

func (tp *TreeProcessor) processSubstringTag(node *ASTNode, content string) string {
	// Process substring tag - extract substring based on start and end positions
	// Get and evaluate attributes
	startStr, startExists := node.Attributes["start"]
	endStr, endExists := node.Attributes["end"]

	if !startExists || !endExists {
		return content
	}

	// Evaluate attributes (they might contain AIML tags)
	startStr = tp.evaluateAttributeValue(startStr)
	endStr = tp.evaluateAttributeValue(endStr)

	// Trim content
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	// Check cache first
	var result string
	cacheKey := fmt.Sprintf("%s|%s|%s", startStr, endStr, content)
	if tp.golem.templateTagProcessingCache != nil {
		if cached, found := tp.golem.templateTagProcessingCache.GetProcessedTag("substring", cacheKey, tp.ctx); found {
			result = cached
		} else {
			result = tp.golem.extractSubstring(content, startStr, endStr)
			tp.golem.templateTagProcessingCache.SetProcessedTag("substring", cacheKey, result, tp.ctx)
		}
	} else {
		result = tp.golem.extractSubstring(content, startStr, endStr)
	}

	tp.golem.LogDebug("Substring tag: '%s' (start=%s, end=%s) -> '%s'", content, startStr, endStr, result)
	return result
}

func (tp *TreeProcessor) processReplaceTag(node *ASTNode, content string) string {
	// Process replace tag - replace occurrences of search string with replacement string
	// Get and evaluate attributes
	search, searchExists := node.Attributes["search"]
	replace, replaceExists := node.Attributes["replace"]

	if !searchExists || !replaceExists {
		return content
	}

	// Evaluate attributes (they might contain AIML tags)
	search = tp.evaluateAttributeValue(search)
	replace = tp.evaluateAttributeValue(replace)

	// Trim content
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	// Check cache first
	var result string
	cacheKey := fmt.Sprintf("%s|%s|%s", search, replace, content)
	if tp.golem.templateTagProcessingCache != nil {
		if cached, found := tp.golem.templateTagProcessingCache.GetProcessedTag("replace", cacheKey, tp.ctx); found {
			result = cached
		} else {
			result = strings.ReplaceAll(content, search, replace)
			tp.golem.templateTagProcessingCache.SetProcessedTag("replace", cacheKey, result, tp.ctx)
		}
	} else {
		result = strings.ReplaceAll(content, search, replace)
	}

	tp.golem.LogDebug("Replace tag: '%s' (search='%s', replace='%s') -> '%s'", content, search, replace, result)
	return result
}

func (tp *TreeProcessor) processPluralizeTag(node *ASTNode, content string) string {
	// Use the existing pluralize processing method
	return tp.golem.processPluralizeTagsWithContext(fmt.Sprintf("<pluralize>%s</pluralize>", content), tp.ctx)
}

func (tp *TreeProcessor) processShuffleTag(node *ASTNode, content string) string {
	// Use the existing shuffle processing method
	return tp.golem.processShuffleTagsWithContext(fmt.Sprintf("<shuffle>%s</shuffle>", content), tp.ctx)
}

func (tp *TreeProcessor) processLengthTag(node *ASTNode, content string) string {
	// Process length tag - calculate length of content
	// Get type attribute (optional)
	lengthType := ""
	if val, exists := node.Attributes["type"]; exists {
		lengthType = strings.TrimSpace(tp.evaluateAttributeValue(val))
	}

	// Trim content
	content = strings.TrimSpace(content)
	if content == "" {
		return "0"
	}

	// Check cache first
	var result string
	cacheKey := fmt.Sprintf("%s|%s", lengthType, content)
	if tp.golem.templateTagProcessingCache != nil {
		if cached, found := tp.golem.templateTagProcessingCache.GetProcessedTag("length", cacheKey, tp.ctx); found {
			result = cached
		} else {
			result = tp.golem.calculateLength(content, lengthType)
			tp.golem.templateTagProcessingCache.SetProcessedTag("length", cacheKey, result, tp.ctx)
		}
	} else {
		result = tp.golem.calculateLength(content, lengthType)
	}

	tp.golem.LogDebug("Length tag: '%s' (type='%s') -> '%s'", content, lengthType, result)
	return result
}

func (tp *TreeProcessor) processCountTag(node *ASTNode, content string) string {
	// Process count tag - count occurrences of search string in content
	// Get and evaluate search attribute
	search, searchExists := node.Attributes["search"]
	if !searchExists {
		return "0"
	}

	// Evaluate attribute (might contain AIML tags)
	search = tp.evaluateAttributeValue(search)

	// Trim content
	content = strings.TrimSpace(content)
	if content == "" || search == "" {
		return "0"
	}

	// Check cache first
	var result string
	cacheKey := fmt.Sprintf("%s|%s", search, content)
	if tp.golem.templateTagProcessingCache != nil {
		if cached, found := tp.golem.templateTagProcessingCache.GetProcessedTag("count", cacheKey, tp.ctx); found {
			result = cached
		} else {
			count := strings.Count(content, search)
			result = strconv.Itoa(count)
			tp.golem.templateTagProcessingCache.SetProcessedTag("count", cacheKey, result, tp.ctx)
		}
	} else {
		count := strings.Count(content, search)
		result = strconv.Itoa(count)
	}

	tp.golem.LogDebug("Count tag: '%s' (search='%s') -> '%s'", content, search, result)
	return result
}

func (tp *TreeProcessor) processSplitTag(node *ASTNode, content string) string {
	// Use the existing split processing method
	return tp.golem.processSplitTagsWithContext(fmt.Sprintf("<split>%s</split>", content), tp.ctx)
}

func (tp *TreeProcessor) processJoinTag(node *ASTNode, content string) string {
	// Use the existing join processing method
	return tp.golem.processJoinTagsWithContext(fmt.Sprintf("<join>%s</join>", content), tp.ctx)
}

func (tp *TreeProcessor) processUniqueTag(node *ASTNode, content string) string {
	// Use the existing unique processing method
	return tp.golem.processUniqueTagsWithContext(fmt.Sprintf("<unique>%s</unique>", content), tp.ctx)
}

func (tp *TreeProcessor) processIndentTag(node *ASTNode, content string) string {
	// Use the existing indent processing method
	return tp.golem.processIndentTagsWithContext(fmt.Sprintf("<indent>%s</indent>", content), tp.ctx)
}

func (tp *TreeProcessor) processDedentTag(node *ASTNode, content string) string {
	// Use the existing dedent processing method
	return tp.golem.processDedentTagsWithContext(fmt.Sprintf("<dedent>%s</dedent>", content), tp.ctx)
}

func (tp *TreeProcessor) processRepeatTag(node *ASTNode, content string) string {
	// Use the existing repeat processing method
	times := 1
	if t, exists := node.Attributes["times"]; exists {
		if parsed, err := strconv.Atoi(t); err == nil {
			times = parsed
		}
	}

	return tp.golem.processRepeatTagsWithContext(fmt.Sprintf(`<repeat times="%d">%s</repeat>`, times, content), tp.ctx)
}

func (tp *TreeProcessor) processFirstTag(node *ASTNode, content string) string {
	// Get first word
	words := strings.Fields(content)
	if len(words) == 0 {
		return ""
	}
	return words[0]
}

func (tp *TreeProcessor) processRestTag(node *ASTNode, content string) string {
	// Get all words except the first
	words := strings.Fields(content)
	if len(words) <= 1 {
		return ""
	}
	return strings.Join(words[1:], " ")
}

func (tp *TreeProcessor) processLoopTag(node *ASTNode, content string) string {
	// Loop tag - just return empty for now
	return ""
}

func (tp *TreeProcessor) processInputTag(node *ASTNode, content string) string {
	// Process input tag - returns the most recent user input
	// <input/> always returns the current/most recent user input (last item in RequestHistory)
	// This is different from <request> which can take an index attribute

	if tp.ctx == nil || tp.ctx.Session == nil {
		tp.golem.LogDebug("Input tag: no context or session available")
		return ""
	}

	// Get the most recent user input from request history
	if len(tp.ctx.Session.RequestHistory) == 0 {
		tp.golem.LogDebug("Input tag: no request history available")
		return ""
	}

	// Return the last (most recent) item from RequestHistory
	currentInput := tp.ctx.Session.RequestHistory[len(tp.ctx.Session.RequestHistory)-1]

	tp.golem.LogDebug("Input tag: returning '%s'", currentInput)

	return currentInput
}

func (tp *TreeProcessor) processEvalTag(node *ASTNode, content string) string {
	// Process eval tag - evaluates AIML code dynamically
	// The <eval> tag causes its content to be evaluated as AIML template code
	// In the AST, child nodes are already processed before reaching this point,
	// so the content parameter contains the fully evaluated result
	// This allows for dynamic tag construction and re-evaluation

	// Trim whitespace from the evaluated content
	content = strings.TrimSpace(content)

	// If empty after trimming, return empty string
	if content == "" {
		tp.golem.LogDebug("Eval tag: empty content after evaluation")
		return ""
	}

	tp.golem.LogDebug("Eval tag: evaluated content='%s'", content)

	// Return the evaluated content
	// Note: Unlike the regex processor which re-processes the content through
	// the full template pipeline, the AST naturally handles nested evaluation
	// through its tree traversal, so we simply return the already-evaluated content
	return content
}

func (tp *TreeProcessor) processPersonTag(node *ASTNode, content string) string {
	// Process person tag - pronoun substitution (1st/2nd person swap)
	// Normalize whitespace
	content = strings.TrimSpace(content)
	content = strings.Join(strings.Fields(content), " ")

	// Check cache first
	var substitutedContent string
	if tp.golem.templateTagProcessingCache != nil {
		if cached, found := tp.golem.templateTagProcessingCache.GetProcessedTag("person", content, tp.ctx); found {
			substitutedContent = cached
		} else {
			substitutedContent = tp.golem.SubstitutePronouns(content)
			tp.golem.templateTagProcessingCache.SetProcessedTag("person", content, substitutedContent, tp.ctx)
		}
	} else {
		substitutedContent = tp.golem.SubstitutePronouns(content)
	}

	tp.golem.LogInfo("Person tag: '%s' -> '%s'", content, substitutedContent)
	return substitutedContent
}

func (tp *TreeProcessor) processPerson2Tag(node *ASTNode, content string) string {
	// Process person2 tag - first-to-third person pronoun substitution
	// Normalize whitespace
	content = strings.TrimSpace(content)
	content = strings.Join(strings.Fields(content), " ")

	// Check cache first
	var substitutedContent string
	if tp.golem.templateTagProcessingCache != nil {
		if cached, found := tp.golem.templateTagProcessingCache.GetProcessedTag("person2", content, tp.ctx); found {
			substitutedContent = cached
		} else {
			substitutedContent = tp.golem.SubstitutePronouns2(content)
			tp.golem.templateTagProcessingCache.SetProcessedTag("person2", content, substitutedContent, tp.ctx)
		}
	} else {
		substitutedContent = tp.golem.SubstitutePronouns2(content)
	}

	tp.golem.LogInfo("Person2 tag: '%s' -> '%s'", content, substitutedContent)
	return substitutedContent
}

func (tp *TreeProcessor) processGenderTag(node *ASTNode, content string) string {
	// Process gender tag - swap masculine/feminine pronouns
	// Normalize whitespace
	content = strings.TrimSpace(content)
	content = strings.Join(strings.Fields(content), " ")

	// Check cache first
	var substitutedContent string
	if tp.golem.templateTagProcessingCache != nil {
		if cached, found := tp.golem.templateTagProcessingCache.GetProcessedTag("gender", content, tp.ctx); found {
			substitutedContent = cached
		} else {
			substitutedContent = tp.golem.SubstituteGenderPronouns(content)
			tp.golem.templateTagProcessingCache.SetProcessedTag("gender", content, substitutedContent, tp.ctx)
		}
	} else {
		substitutedContent = tp.golem.SubstituteGenderPronouns(content)
	}

	tp.golem.LogInfo("Gender tag: '%s' -> '%s'", content, substitutedContent)
	return substitutedContent
}

func (tp *TreeProcessor) processSentenceTag(node *ASTNode, content string) string {
	// Process sentence tag - capitalize first letter of each sentence
	content = strings.TrimSpace(content)

	if content == "" {
		return ""
	}

	// Capitalize sentences using the existing method
	processedContent := tp.golem.capitalizeSentences(content)

	tp.golem.LogDebug("Sentence tag: '%s' -> '%s'", content, processedContent)
	return processedContent
}

func (tp *TreeProcessor) processWordTag(node *ASTNode, content string) string {
	// Process word tag - capitalize first letter of each word (Title Case)
	content = strings.TrimSpace(content)

	if content == "" {
		return ""
	}

	// Capitalize words using the existing method
	processedContent := tp.golem.capitalizeWords(content)

	tp.golem.LogDebug("Word tag: '%s' -> '%s'", content, processedContent)
	return processedContent
}

func (tp *TreeProcessor) processDateTag(node *ASTNode, content string) string {
	// Date tag - current date
	format := "Monday, January 2, 2006"
	if f, exists := node.Attributes["format"]; exists {
		format = f
	}
	// Convert C-style or alternative formats to Go time format
	goFormat := tp.golem.convertToGoTimeFormat(format)
	return time.Now().Format(goFormat)
}

func (tp *TreeProcessor) processTimeTag(node *ASTNode, content string) string {
	// Time tag - current time
	defaultFormat := "3:04 PM"
	format := defaultFormat
	if f, exists := node.Attributes["format"]; exists {
		format = f
	}
	// Convert C-style or alternative formats to Go time format
	goFormat := tp.golem.convertToGoTimeFormat(format)

	// If format conversion didn't change anything and it's not a recognized Go format,
	// fall back to default
	if goFormat == format && !tp.golem.looksLikeGoTimeFormat(format) && format != defaultFormat {
		goFormat = defaultFormat
	}

	return time.Now().Format(goFormat)
}

// System tags

func (tp *TreeProcessor) processSizeTag(node *ASTNode, content string) string {
	// Size tag - knowledge base size
	if tp.ctx != nil && tp.ctx.KnowledgeBase != nil {
		return strconv.Itoa(len(tp.ctx.KnowledgeBase.Categories))
	}
	return "0"
}

func (tp *TreeProcessor) processVersionTag(node *ASTNode, content string) string {
	// Version tag - bot version
	if tp.ctx != nil && tp.ctx.KnowledgeBase != nil {
		if version, exists := tp.ctx.KnowledgeBase.Properties["version"]; exists {
			return version
		}
	}
	return "1.0"
}

func (tp *TreeProcessor) processIdTag(node *ASTNode, content string) string {
	// ID tag - bot ID
	if tp.ctx != nil && tp.ctx.KnowledgeBase != nil {
		if id, exists := tp.ctx.KnowledgeBase.Properties["id"]; exists {
			return id
		}
	}
	return "golem"
}

func (tp *TreeProcessor) processRequestTag(node *ASTNode, content string) string {
	// Request tag - previous request
	// Index 1 = most recent, index 2 = 2nd most recent, etc.
	index := 1
	if idx, exists := node.Attributes["index"]; exists {
		if parsed, err := strconv.Atoi(idx); err == nil {
			index = parsed
		}
	}

	if tp.ctx != nil && tp.ctx.Session != nil {
		// Use GetRequestByIndex which properly handles reverse indexing
		return tp.ctx.Session.GetRequestByIndex(index)
	}
	return ""
}

func (tp *TreeProcessor) processResponseTag(node *ASTNode, content string) string {
	// Response tag - previous response
	index := 1
	if idx, exists := node.Attributes["index"]; exists {
		if parsed, err := strconv.Atoi(idx); err == nil {
			index = parsed
		}
	}

	if tp.ctx != nil && tp.ctx.Session != nil {
		if tp.ctx.Session.ResponseHistory != nil && index <= len(tp.ctx.Session.ResponseHistory) {
			return tp.ctx.Session.ResponseHistory[index-1]
		}
	}
	return ""
}

// Text processing tags

func (tp *TreeProcessor) processNormalizeTag(node *ASTNode, content string) string {
	// Normalize tag - text normalization
	// Process the content directly using the normalization function
	return tp.golem.normalizeTextForOutput(content)
}

func (tp *TreeProcessor) processDenormalizeTag(node *ASTNode, content string) string {
	// Denormalize tag - text denormalization
	// Process the content directly using the denormalization function
	return tp.golem.denormalizeText(content)
}

// Learning tags

func (tp *TreeProcessor) processUnlearnTag(node *ASTNode, content string) string {
	// Unlearn tag - remove learned categories
	// Use the existing unlearn processing method
	return tp.golem.processUnlearnTagsWithContext(fmt.Sprintf("<unlearn>%s</unlearn>", content), tp.ctx)
}

func (tp *TreeProcessor) processUnlearnfTag(node *ASTNode, content string) string {
	// Unlearnf tag - remove categories from persistent storage
	if tp.golem.aimlKB == nil {
		tp.golem.LogWarn("Unlearnf: No knowledge base available")
		return ""
	}

	tp.golem.LogInfo("Processing unlearnf: '%s'", content)

	// Parse the AIML content within the unlearnf tag
	categories, err := tp.golem.parseLearnContent(content)
	if err != nil {
		tp.golem.LogError("Failed to parse unlearnf content: %v", err)
		return ""
	}

	// Remove categories from persistent knowledge base
	for _, category := range categories {
		err := tp.golem.removePersistentCategory(category)
		if err != nil {
			tp.golem.LogInfo("Failed to remove persistent category: %v", err)
		}
	}

	// Unlearnf tags don't output content
	return ""
}

// Advanced tags

func (tp *TreeProcessor) processVarTag(node *ASTNode, content string) string {
	// Var tag - variable declaration
	// Similar to set tag but for variable declaration
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	// Process the content to get the value
	value := content

	// Set the variable in context
	if tp.ctx != nil {
		if tp.ctx.LocalVars == nil {
			tp.ctx.LocalVars = make(map[string]string)
		}
		tp.ctx.LocalVars[name] = value
	}

	// Var tags don't output content
	return ""
}

func (tp *TreeProcessor) processGossipTag(node *ASTNode, content string) string {
	// Gossip tag - gossip processing
	// For now, return empty string as this functionality needs to be implemented
	return ""
}

func (tp *TreeProcessor) processJavascriptTag(node *ASTNode, content string) string {
	// Javascript tag - JavaScript execution
	// For now, return empty string as this functionality needs to be implemented
	return ""
}

func (tp *TreeProcessor) processSystemTag(node *ASTNode, content string) string {
	// System tag - system command execution
	// For now, return empty string as this functionality needs to be implemented
	return ""
}

func (tp *TreeProcessor) processSubjTag(node *ASTNode, content string) string {
	// Subj tag - RDF subject
	// Process content and add trailing space for RDF readability
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	return content + " "
}

func (tp *TreeProcessor) processPredTag(node *ASTNode, content string) string {
	// Pred tag - RDF predicate
	// Process content and add trailing space for RDF readability
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	return content + " "
}

func (tp *TreeProcessor) processObjTag(node *ASTNode, content string) string {
	// Obj tag - RDF object
	// Process content without trailing space (it's the last element)
	content = strings.TrimSpace(content)
	return content
}

func (tp *TreeProcessor) processUniqTag(node *ASTNode, content string) string {
	// Uniq tag - RDF unique/triple container
	// Process content and format with proper spacing
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}

	// Clean up multiple spaces and format for readability
	words := strings.Fields(content)
	if len(words) == 0 {
		return ""
	}

	return strings.Join(words, " ")
}

// Helper method for random number generation
func (g *Golem) randomIntTree(max int) int {
	// This would use the existing random number generation from the Golem instance
	// For now, return a simple implementation
	return int(time.Now().UnixNano() % int64(max))
}
