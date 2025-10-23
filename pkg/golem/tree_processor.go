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

func (tp *TreeProcessor) processSetTag(node *ASTNode, content string) string {
	// Process set tag - variable assignment
	name, exists := node.Attributes["name"]
	if !exists {
		return content
	}

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

func (tp *TreeProcessor) processThatTag(node *ASTNode, content string) string {
	// Process that tag - previous response reference
	// This would need to be implemented based on the actual ChatSession structure
	// For now, return empty string
	return ""
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
	// This is a complex tag that would need full implementation
	return tp.golem.processConditionTagsWithContext(fmt.Sprintf("<condition>%s</condition>", content), tp.ctx)
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

	operation, _ := node.Attributes["operation"]

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
	// Process learn tag - dynamic learning
	return tp.golem.processLearnTagsWithContext(fmt.Sprintf("<learn>%s</learn>", content), tp.ctx)
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
	// Get user input
	if tp.ctx != nil && tp.ctx.Session != nil {
		// This would need to be implemented based on the actual ChatSession structure
		// For now, return empty string
		return ""
	}
	return ""
}

func (tp *TreeProcessor) processEvalTag(node *ASTNode, content string) string {
	// Use the existing eval processing method
	// Use the existing eval processing method from consolidated processor
	// For now, return content as eval processing is complex
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
	index := 1
	if idx, exists := node.Attributes["index"]; exists {
		if parsed, err := strconv.Atoi(idx); err == nil {
			index = parsed
		}
	}

	if tp.ctx != nil && tp.ctx.Session != nil {
		if tp.ctx.Session.RequestHistory != nil && index <= len(tp.ctx.Session.RequestHistory) {
			return tp.ctx.Session.RequestHistory[index-1]
		}
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
	// Use the existing normalize processing method
	return tp.golem.processNormalizeTagsWithContext(fmt.Sprintf("<normalize>%s</normalize>", content), tp.ctx)
}

func (tp *TreeProcessor) processDenormalizeTag(node *ASTNode, content string) string {
	// Denormalize tag - text denormalization
	// Use the existing denormalize processing method
	return tp.golem.processDenormalizeTagsWithContext(fmt.Sprintf("<denormalize>%s</denormalize>", content), tp.ctx)
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
	// This would integrate with the existing RDF implementation
	return content
}

func (tp *TreeProcessor) processPredTag(node *ASTNode, content string) string {
	// Pred tag - RDF predicate
	// This would integrate with the existing RDF implementation
	return content
}

func (tp *TreeProcessor) processObjTag(node *ASTNode, content string) string {
	// Obj tag - RDF object
	// This would integrate with the existing RDF implementation
	return content
}

func (tp *TreeProcessor) processUniqTag(node *ASTNode, content string) string {
	// Uniq tag - RDF unique
	// This would integrate with the existing RDF implementation
	return content
}

// Helper method for random number generation
func (g *Golem) randomIntTree(max int) int {
	// This would use the existing random number generation from the Golem instance
	// For now, return a simple implementation
	return int(time.Now().UnixNano() % int64(max))
}
