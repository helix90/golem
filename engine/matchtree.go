package engine

import (
	"fmt"
	"golem/parser"
	"os"
	"regexp"
	"strings"
)

// MatchNode represents a node in the match tree
// Each node can have children for words, '*' and '_', and may store a category
// The tree is 3-level: pattern -> that -> topic
// (AIML spec: topic and that are optional, default to "*" if not present)
type MatchNode struct {
	children map[string]*MatchNode
	star     *MatchNode // '*' wildcard
	under    *MatchNode // '_' wildcard
	category *parser.Category
}

// MatchTree is the root of the match tree
// It stores the root node for all patterns
// Insert and Match operate on pattern, that, topic
// All keys are uppercased and trimmed
type MatchTree struct {
	root *MatchNode
}

// Helper to select the most specific match: exact > set > '_' > '*'
type matchType int

const (
	starMatch matchType = iota // lowest specificity
	underMatch
	setMatch
	exactMatch // highest specificity
)

type matchResultWithType struct {
	res   *MatchResult
	mtype matchType
}

func selectBestMatch(matches []matchResultWithType) (*MatchResult, bool) {
	best := -1
	var bestRes *MatchResult
	for _, m := range matches {
		if m.res != nil && int(m.mtype) > best {
			best = int(m.mtype)
			bestRes = m.res
		}
	}
	return bestRes, bestRes != nil
}

func NewMatchTree() *MatchTree {
	return &MatchTree{root: newMatchNode()}
}

func newMatchNode() *MatchNode {
	return &MatchNode{children: make(map[string]*MatchNode)}
}

// Insert adds a category to the tree, indexed by pattern, that, topic
func (t *MatchTree) Insert(cat parser.Category) {
	// Normalize pattern, that, topic to uppercase and trimmed
	cat.Pattern = strings.ToUpper(strings.TrimSpace(cat.Pattern))
	cat.That = strings.ToUpper(strings.TrimSpace(cat.That))
	cat.Topic = strings.ToUpper(strings.TrimSpace(cat.Topic))

	if cat.That == "" {
		cat.That = "*"
	}
	if cat.Topic == "" {
		cat.Topic = "*"
	}

	patternWords := splitAIML(cat.Pattern)
	thatWords := splitAIML(cat.That)
	topicWords := splitAIML(cat.Topic)

	fmt.Fprintf(os.Stderr, "[DEBUG] Insert: Pattern=%q That=%q Topic=%q\n", cat.Pattern, cat.That, cat.Topic)
	fmt.Fprintf(os.Stderr, "[DEBUG] Insert tokens: patternWords=%q thatWords=%q topicWords=%q\n", patternWords, thatWords, topicWords)

	if len(thatWords) == 0 {
		thatWords = []string{"*"}
	}
	if len(topicWords) == 0 {
		topicWords = []string{"*"}
	}

	// Insert into tree: pattern -> that -> topic
	pnode := t.root
	for _, w := range patternWords {
		pnode = pnode.childFor(w)
	}
	tnode := pnode
	for _, w := range thatWords {
		tnode = tnode.childFor(w)
	}
	topicNode := tnode
	for _, w := range topicWords {
		topicNode = topicNode.childFor(w)
	}
	// Instead of newCat := cat; topicNode.category = &newCat
	topicNode.category = new(parser.Category)
	*topicNode.category = cat
}

// childFor returns the child node for a word, creating it if necessary
func (n *MatchNode) childFor(word string) *MatchNode {
	word = strings.ToUpper(strings.TrimSpace(word))
	switch word {
	case "*":
		if n.star == nil {
			n.star = newMatchNode()
		}
		return n.star
	case "_":
		if n.under == nil {
			n.under = newMatchNode()
		}
		return n.under
	default:
		if n.children[word] == nil {
			n.children[word] = newMatchNode()
		}
		return n.children[word]
	}
}

// splitAIML splits a pattern/that/topic into words, uppercased, and converts <set>NAME</set> (any case) to __SET_NAME__ tokens
func splitAIML(s string) []string {
	s = strings.TrimSpace(s)
	if strings.Contains(strings.ToLower(s), "<set>") {
		// Replace <set>NAME</set> (any case) with __SET_NAME__
		re := regexp.MustCompile(`(?i)<set>([^<]+)</set>`) // case-insensitive
		s = re.ReplaceAllStringFunc(s, func(m string) string {
			matches := re.FindStringSubmatch(m)
			if len(matches) == 2 {
				return " __SET_" + strings.ToUpper(strings.TrimSpace(matches[1])) + "__ "
			}
			return m
		})
	}
	words := strings.Fields(strings.ToUpper(s))
	return words
}

// MatchResult contains metadata about a successful match
// WildcardCaptures: map from wildcard position (pattern/that/topic) to the captured words
// MatchedPattern/That/Topic: the actual pattern/that/topic that matched (with wildcards expanded)
type MatchResult struct {
	Category         *parser.Category
	MatchedPattern   string
	MatchedThat      string
	MatchedTopic     string
	Template         string
	WildcardCaptures map[string][]string // keys: pattern, that, topic
}

const matchRecursionLimit = 20

// MatchWithMeta now takes both pattern and input tokens separately for proper set matching
func (t *MatchTree) MatchWithMeta(input, that, topic string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	inputWords := splitAIML(input)
	thatWords := splitAIML(that)
	topicWords := splitAIML(topic)
	if len(thatWords) == 0 {
		thatWords = []string{"*"}
	}
	if len(topicWords) == 0 {
		topicWords = []string{"*"}
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Match: Input=%q That=%q Topic=%q\n", strings.Join(inputWords, " "), strings.Join(thatWords, " "), strings.Join(topicWords, " "))
	fmt.Fprintf(os.Stderr, "[DEBUG] Match tokens: inputWords=%q thatWords=%q topicWords=%q\n", inputWords, thatWords, topicWords)

	// Start recursive match from the root node with all tokens
	return matchNodeMetaPattern(t.root, inputWords, inputWords, thatWords, topicWords, 0, map[string][]string{"pattern": {}, "that": {}, "topic": {}}, sets)
}

func matchNodeMetaPattern(node *MatchNode, pattern, input, that, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	if depth > matchRecursionLimit {
		return nil, false
	}
	if len(pattern) > 0 {
		w := pattern[0]
		matches := []matchResultWithType{}
		// 1. Exact match
		if len(input) > 0 {
			if child, ok := node.children[w]; ok && w == input[0] {
				if res, found := matchNodeMetaPattern(child, pattern[1:], input[1:], that, topic, depth+1, captures, sets); found {
					matches = append(matches, matchResultWithType{res, exactMatch})
				}
			}
		}
		// 2. Set match (try all set children)
		if len(input) > 0 {
			for childKey, childNode := range node.children {
				if strings.HasPrefix(childKey, "__SET_") && strings.HasSuffix(childKey, "__") {
					setName := strings.TrimSuffix(strings.TrimPrefix(childKey, "__SET_"), "__")
					if set, exists := sets[setName]; exists {
						_, isMember := set[input[0]]
						fmt.Fprintf(os.Stderr, "[DEBUG] Set match: setName=%q inputWord=%q exists=%v isMember=%v\n", setName, input[0], exists, isMember)
						if isMember {
							if res, found := matchNodeMetaPattern(childNode, pattern[1:], input[1:], that, topic, depth+1, captures, sets); found {
								matches = append(matches, matchResultWithType{res, setMatch})
							}
						}
					}
				}
			}
		}
		// 3. '_' wildcard (single word)
		if node.under != nil && len(input) > 0 {
			capCopy := copyCaptures(captures)
			capCopy["pattern"] = append([]string{}, captures["pattern"]...)
			capCopy["pattern"] = append(capCopy["pattern"], input[0])
			if res, found := matchNodeMetaPattern(node.under, pattern[1:], input[1:], that, topic, depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, underMatch})
			}
		}
		// 4. '*' wildcard (zero or more words, all splits)
		if node.star != nil {
			for i := 0; i <= len(input); i++ {
				capCopy := copyCaptures(captures)
				capCopy["pattern"] = append([]string{}, captures["pattern"]...)
				capCopy["pattern"] = append(capCopy["pattern"], input[:i]...)
				if res, found := matchNodeMetaPattern(node.star, pattern[1:], input[i:], that, topic, depth+1, capCopy, sets); found {
					matches = append(matches, matchResultWithType{res, starMatch})
				}
			}
		}
		return selectBestMatch(matches)
	}
	// Pattern exhausted, now match that
	return matchNodeMetaThatSection(node, that, topic, depth, captures, sets)
}

func matchNodeMetaThatSection(node *MatchNode, that, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	capCopy := copyCaptures(captures)
	capCopy["that"] = []string{}
	return matchNodeMetaThatSectionRecurse(node, that, topic, depth, capCopy, sets)
}

func matchNodeMetaThatSectionRecurse(node *MatchNode, that, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	if depth > matchRecursionLimit {
		return nil, false
	}
	if len(that) > 0 {
		w := that[0]
		matches := []matchResultWithType{}
		// 1. Exact match
		if child, ok := node.children[w]; ok {
			if res, found := matchNodeMetaThatSectionRecurse(child, that[1:], topic, depth+1, captures, sets); found {
				matches = append(matches, matchResultWithType{res, exactMatch})
			}
		}
		// 2. Set match (try all set children)
		for childKey, childNode := range node.children {
			if strings.HasPrefix(childKey, "__SET_") && strings.HasSuffix(childKey, "__") {
				setName := strings.TrimSuffix(strings.TrimPrefix(childKey, "__SET_"), "__")
				if set, exists := sets[setName]; exists {
					_, isMember := set[that[0]]
					if isMember {
						if res, found := matchNodeMetaThatSectionRecurse(childNode, that[1:], topic, depth+1, captures, sets); found {
							matches = append(matches, matchResultWithType{res, setMatch})
						}
					}
				}
			}
		}
		// 3. '_' wildcard (single word)
		if node.under != nil {
			capCopy := copyCaptures(captures)
			capCopy["that"] = append([]string{}, captures["that"]...)
			capCopy["that"] = append(capCopy["that"], that[0])
			if res, found := matchNodeMetaThatSectionRecurse(node.under, that[1:], topic, depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, underMatch})
			}
		}
		// 4. '*' wildcard (zero or more words, all splits)
		if node.star != nil {
			for i := 0; i <= len(that); i++ {
				capCopy := copyCaptures(captures)
				capCopy["that"] = append([]string{}, captures["that"]...)
				capCopy["that"] = append(capCopy["that"], that[:i]...)
				if res, found := matchNodeMetaThatSectionRecurse(node.star, that[i:], topic, depth+1, capCopy, sets); found {
					matches = append(matches, matchResultWithType{res, starMatch})
				}
			}
		}
		return selectBestMatch(matches)
	}
	// That exhausted, now match topic
	return matchNodeMetaTopicSection(node, topic, depth, captures, sets)
}

func matchNodeMetaTopicSection(node *MatchNode, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	capCopy := copyCaptures(captures)
	capCopy["topic"] = []string{}
	return matchNodeMetaTopicSectionRecurse(node, topic, depth, capCopy, sets)
}

func matchNodeMetaTopicSectionRecurse(node *MatchNode, topic []string, depth int, captures map[string][]string, sets map[string]map[string]struct{}) (*MatchResult, bool) {
	if depth > matchRecursionLimit {
		return nil, false
	}
	if len(topic) > 0 {
		w := topic[0]
		matches := []matchResultWithType{}
		// 1. Exact match
		if child, ok := node.children[w]; ok {
			if res, found := matchNodeMetaTopicSectionRecurse(child, topic[1:], depth+1, captures, sets); found {
				matches = append(matches, matchResultWithType{res, exactMatch})
			}
		}
		// 2. Set match (try all set children)
		for childKey, childNode := range node.children {
			if strings.HasPrefix(childKey, "__SET_") && strings.HasSuffix(childKey, "__") {
				setName := strings.TrimSuffix(strings.TrimPrefix(childKey, "__SET_"), "__")
				if set, exists := sets[setName]; exists {
					_, isMember := set[topic[0]]
					if isMember {
						if res, found := matchNodeMetaTopicSectionRecurse(childNode, topic[1:], depth+1, captures, sets); found {
							matches = append(matches, matchResultWithType{res, setMatch})
						}
					}
				}
			}
		}
		// 3. '_' wildcard (single word)
		if node.under != nil {
			capCopy := copyCaptures(captures)
			capCopy["topic"] = append([]string{}, captures["topic"]...)
			capCopy["topic"] = append(capCopy["topic"], topic[0])
			if res, found := matchNodeMetaTopicSectionRecurse(node.under, topic[1:], depth+1, capCopy, sets); found {
				matches = append(matches, matchResultWithType{res, underMatch})
			}
		}
		// 4. '*' wildcard (zero or more words, all splits)
		if node.star != nil {
			for i := 0; i <= len(topic); i++ {
				capCopy := copyCaptures(captures)
				capCopy["topic"] = append([]string{}, captures["topic"]...)
				capCopy["topic"] = append(capCopy["topic"], topic[:i]...)
				if res, found := matchNodeMetaTopicSectionRecurse(node.star, topic[i:], depth+1, capCopy, sets); found {
					matches = append(matches, matchResultWithType{res, starMatch})
				}
			}
		}
		return selectBestMatch(matches)
	}
	// All exhausted, return category if present
	if node.category != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] Matched: Pattern=%q That=%q Topic=%q Template=%q\n", node.category.Pattern, node.category.That, node.category.Topic, node.category.Template)
		return &MatchResult{
			Category:         node.category,
			MatchedPattern:   node.category.Pattern,
			MatchedThat:      node.category.That,
			MatchedTopic:     node.category.Topic,
			Template:         node.category.Template,
			WildcardCaptures: captures,
		}, true
	}
	return nil, false
}

func copyCaptures(src map[string][]string) map[string][]string {
	out := make(map[string][]string, len(src))
	for k, v := range src {
		out[k] = append([]string{}, v...)
	}
	return out
}

// Match is the legacy interface, returns only the template
func (t *MatchTree) Match(input, that, topic string) (*parser.Category, bool) {
	res, found := t.MatchWithMeta(input, that, topic, nil)
	if !found {
		return nil, false
	}
	return res.Category, true
}
