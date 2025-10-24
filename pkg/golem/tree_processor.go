package golem

import (
	"fmt"
	"regexp"
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

	return result, nil
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
	// Process children first to handle nested tags
	processedChildren := make([]string, len(node.Children))
	for i, child := range node.Children {
		processedChildren[i] = tp.processNode(child)
	}

	// Join processed children
	content := strings.Join(processedChildren, "")

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
	// Process SRAI tag - recursive AIML processing
	// Use the existing SRAI processing method
	return tp.golem.processSRAITagsWithContext(fmt.Sprintf("<srai>%s</srai>", content), tp.ctx)
}

func (tp *TreeProcessor) processSRAIXTag(node *ASTNode, content string) string {
	// Process SRAIX tag - external service integration
	// Use the existing SRAIX processing method
	return tp.golem.processSRAIXTagsWithContext(fmt.Sprintf("<sraix>%s</sraix>", content), tp.ctx)
}

func (tp *TreeProcessor) processThinkTag(node *ASTNode, content string) string {
	// Think tags don't output content, but may set variables
	// Use the existing think processing method
	return tp.golem.processThinkTagsWithContext(fmt.Sprintf("<think>%s</think>", content), tp.ctx)
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
	// Process set tag - variable assignment
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	// Evaluate the name if it contains AIML tags (like <star/>)
	name = tp.evaluateAttributeValue(name)

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
		} else {
			// Fallback to local variables
			if tp.ctx.LocalVars == nil {
				tp.ctx.LocalVars = make(map[string]string)
			}
			tp.ctx.LocalVars[name] = value
		}
	}

	// Set tags don't output content
	return ""
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
		// Check global variables
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

	// Get wildcard value
	if tp.ctx != nil && tp.ctx.Session != nil {
		key := fmt.Sprintf("star%d", index)
		if value, exists := tp.ctx.Session.Variables[key]; exists {
			return value
		}
	}

	return ""
}

func (tp *TreeProcessor) processSRTag(node *ASTNode, content string) string {
	// Process SR tag - shorthand for <srai><star/></srai>
	// SR recursively processes the first wildcard (star1)

	if tp.ctx == nil || tp.ctx.Session == nil {
		tp.golem.LogDebug("SR tag: no context or session available")
		return ""
	}

	// Get the first wildcard (star1) from session variables
	starContent := ""
	if value, exists := tp.ctx.Session.Variables["star1"]; exists {
		starContent = value
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
	if tp.ctx.Session.Variables != nil {
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
	}

	// Process the matched template recursively
	result := tp.golem.processTemplateWithContext(category.Template, wildcards, tp.ctx)

	// Restore old wildcards
	if tp.ctx.Session.Variables != nil {
		// Remove wildcards from the matched pattern
		for k := range wildcards {
			delete(tp.ctx.Session.Variables, k)
		}
		// Restore original wildcards
		for k, v := range oldWildcards {
			tp.ctx.Session.Variables[k] = v
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
	// Process list item tag - just return processed content
	return tp.processNode(&ASTNode{
		Type:     NodeTypeText,
		Content:  content,
		Children: []*ASTNode{},
	})
}

func (tp *TreeProcessor) processConditionTag(node *ASTNode, content string) string {
	// Process condition tag - conditional logic
	// Build the condition tag with attributes for the regex-based processing
	var conditionTag string
	if name, exists := node.Attributes["name"]; exists {
		conditionTag = fmt.Sprintf(`<condition name="%s"`, name)
		if value, exists := node.Attributes["value"]; exists {
			conditionTag += fmt.Sprintf(` value="%s"`, value)
		}
		conditionTag += fmt.Sprintf(`>%s</condition>`, content)
	} else {
		conditionTag = fmt.Sprintf("<condition>%s</condition>", content)
	}

	return tp.golem.processConditionTagsWithContext(conditionTag, tp.ctx)
}

func (tp *TreeProcessor) processMapTag(node *ASTNode, content string) string {
	// Process map tag - mapping operations
	return tp.golem.processMapTagsWithContext(fmt.Sprintf("<map>%s</map>", content), tp.ctx)
}

func (tp *TreeProcessor) processListTag(node *ASTNode, content string) string {
	// Process list tag - list operations
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

	operation := node.Attributes["operation"]

	// If no knowledge base, just return empty string (operation performed)
	if tp.ctx == nil || tp.ctx.KnowledgeBase == nil || tp.ctx.KnowledgeBase.Lists == nil {
		return ""
	}

	// Get or create the list
	if tp.ctx.KnowledgeBase.Lists[name] == nil {
		tp.ctx.KnowledgeBase.Lists[name] = make([]string, 0)
	}
	list := tp.ctx.KnowledgeBase.Lists[name]

	switch operation {
	case "add", "append":
		// Add item to the end of the list
		list = append(list, content)
		tp.ctx.KnowledgeBase.Lists[name] = list
		return "" // List operations don't return content
	case "get":
		// Return all items joined
		return strings.Join(list, " ")
	case "size":
		// Return size of list
		return strconv.Itoa(len(list))
	default:
		// For other operations, just return empty string
		return ""
	}
}

func (tp *TreeProcessor) processArrayTag(node *ASTNode, content string) string {
	// Process array tag - array operations
	return tp.golem.processArrayTagsWithContext(fmt.Sprintf("<array>%s</array>", content), tp.ctx)
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
	// Process content directly - convert to acronym
	words := strings.Fields(content)
	var acronym strings.Builder
	for _, word := range words {
		if len(word) > 0 {
			acronym.WriteRune(rune(word[0]))
		}
	}
	return acronym.String()
}

func (tp *TreeProcessor) processTrimTag(node *ASTNode, content string) string {
	// Process content directly - trim whitespace
	return strings.TrimSpace(content)
}

func (tp *TreeProcessor) processSubstringTag(node *ASTNode, content string) string {
	// Use the existing substring processing method
	startStr, startExists := node.Attributes["start"]
	endStr, endExists := node.Attributes["end"]

	if !startExists || !endExists {
		return content
	}

	return tp.golem.processSubstringTagsWithContext(fmt.Sprintf(`<substring start="%s" end="%s">%s</substring>`, startStr, endStr, content), tp.ctx)
}

func (tp *TreeProcessor) processReplaceTag(node *ASTNode, content string) string {
	// Use the existing replace processing method
	search, searchExists := node.Attributes["search"]
	replace, replaceExists := node.Attributes["replace"]

	if !searchExists || !replaceExists {
		return content
	}

	return tp.golem.processReplaceTagsWithContext(fmt.Sprintf(`<replace search="%s" replace="%s">%s</replace>`, search, replace, content), tp.ctx)
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
	// Use the existing length processing method
	return tp.golem.processLengthTagsWithContext(fmt.Sprintf("<length>%s</length>", content), tp.ctx)
}

func (tp *TreeProcessor) processCountTag(node *ASTNode, content string) string {
	// Use the existing count processing method
	return tp.golem.processCountTagsWithContext(fmt.Sprintf("<count>%s</count>", content), tp.ctx)
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
	// Use the existing person processing method
	return tp.golem.processPersonTagsWithContext(fmt.Sprintf("<person>%s</person>", content), tp.ctx)
}

func (tp *TreeProcessor) processPerson2Tag(node *ASTNode, content string) string {
	// Use the existing person2 processing method
	return tp.golem.processPerson2TagsWithContext(fmt.Sprintf("<person2>%s</person2>", content), tp.ctx)
}

func (tp *TreeProcessor) processGenderTag(node *ASTNode, content string) string {
	// Use the existing gender processing method
	return tp.golem.processGenderTagsWithContext(fmt.Sprintf("<gender>%s</gender>", content), tp.ctx)
}

func (tp *TreeProcessor) processSentenceTag(node *ASTNode, content string) string {
	// Use the existing sentence processing method
	return tp.golem.processSentenceTagsWithContext(fmt.Sprintf("<sentence>%s</sentence>", content), tp.ctx)
}

func (tp *TreeProcessor) processWordTag(node *ASTNode, content string) string {
	// Use the existing word processing method
	return tp.golem.processWordTagsWithContext(fmt.Sprintf("<word>%s</word>", content), tp.ctx)
}

func (tp *TreeProcessor) processDateTag(node *ASTNode, content string) string {
	// Date tag - current date
	format := "Monday, January 2, 2006"
	if f, exists := node.Attributes["format"]; exists {
		format = f
	}
	return time.Now().Format(format)
}

func (tp *TreeProcessor) processTimeTag(node *ASTNode, content string) string {
	// Time tag - current time
	format := "3:04 PM"
	if f, exists := node.Attributes["format"]; exists {
		format = f
	}
	return time.Now().Format(format)
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
	// Unlearnf tag - remove learned files
	// For now, return empty string as this functionality needs to be implemented
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
