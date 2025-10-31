package golem

import (
	"strings"
	"testing"
)

// TestSRAIXFallbackResponses tests intelligent fallback responses when SRAIX services are unavailable
func TestSRAIXFallbackResponses(t *testing.T) {
	g := New(true)

	testCases := []struct {
		name             string
		template         string
		expectedContains string
		shouldNotContain string
	}{
		{
			name:             "FAVORITE pattern with no service",
			template:         `<sraix service="pannous">FAVORITE ANIMAL</sraix>`,
			expectedContains: "favorite animal",
			shouldNotContain: "<sraix",
		},
		{
			name:             "WHO IS pattern with no service",
			template:         `<sraix service="pannous">WHO IS george washington</sraix>`,
			expectedContains: "information about that person",
			shouldNotContain: "<sraix",
		},
		{
			name:             "WHO WAS pattern with no service",
			template:         `<sraix service="pannous">WHO WAS shakespeare</sraix>`,
			expectedContains: "information about that person",
			shouldNotContain: "<sraix",
		},
		{
			name:             "WHAT IS pattern with no service",
			template:         `<sraix service="pannous">WHAT IS quantum physics</sraix>`,
			expectedContains: "information available",
			shouldNotContain: "<sraix",
		},
		{
			name:             "WHERE IS pattern with no service",
			template:         `<sraix service="pannous">WHERE IS paris</sraix>`,
			expectedContains: "location information",
			shouldNotContain: "<sraix",
		},
		{
			name:             "WHEN IS pattern with no service",
			template:         `<sraix service="pannous">WHEN IS christmas</sraix>`,
			expectedContains: "date or time information",
			shouldNotContain: "<sraix",
		},
		{
			name:             "WHY pattern with no service",
			template:         `<sraix service="pannous">WHY is the sky blue</sraix>`,
			expectedContains: "interesting question",
			shouldNotContain: "<sraix",
		},
		{
			name:             "HOW pattern with no service",
			template:         `<sraix service="pannous">HOW do I bake a cake</sraix>`,
			expectedContains: "detailed instructions",
			shouldNotContain: "<sraix",
		},
		{
			name:             "DEFINE pattern with no service",
			template:         `<sraix service="pannous">DEFINE algorithm</sraix>`,
			expectedContains: "definition for 'algorithm'",
			shouldNotContain: "<sraix",
		},
		{
			name:             "WEATHER pattern with no service",
			template:         `<sraix service="pannous">WEATHER in seattle</sraix>`,
			expectedContains: "weather information",
			shouldNotContain: "<sraix",
		},
		{
			name:             "JOKE pattern with no service",
			template:         `<sraix service="pannous">JOKE</sraix>`,
			expectedContains: "jokes",
			shouldNotContain: "<sraix",
		},
		{
			name:             "RECOMMEND pattern with no service",
			template:         `<sraix service="pannous">RECOMMEND a book</sraix>`,
			expectedContains: "recommendation services",
			shouldNotContain: "<sraix",
		},
		{
			name:             "Generic query with service name",
			template:         `<sraix service="myservice">some query</sraix>`,
			expectedContains: "myservice",
			shouldNotContain: "<sraix",
		},
		{
			name:             "Generic query with bot name",
			template:         `<sraix bot="mybot">some query</sraix>`,
			expectedContains: "mybot",
			shouldNotContain: "<sraix",
		},
		{
			name:             "Generic query with no identifiers",
			template:         `<sraix>some random query</sraix>`,
			expectedContains: "external services",
			shouldNotContain: "<sraix",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := g.ProcessTemplate(tc.template, map[string]string{})

			// Verify expected content is present
			if !strings.Contains(strings.ToLower(result), strings.ToLower(tc.expectedContains)) {
				t.Errorf("Expected result to contain '%s', got: %q", tc.expectedContains, result)
			}

			// Verify unwanted content is not present
			if strings.Contains(result, tc.shouldNotContain) {
				t.Errorf("Result should not contain '%s', got: %q", tc.shouldNotContain, result)
			}

			// Verify no XML tags in output
			if strings.Contains(result, "<") && strings.Contains(result, ">") {
				t.Errorf("Result contains XML tags: %q", result)
			}
		})
	}
}

// TestSRAIXWithDefaultAttribute tests that default attribute is used when service unavailable
func TestSRAIXWithDefaultAttribute(t *testing.T) {
	g := New(true)

	testCases := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "SRAIX with default attribute",
			template: `<sraix service="pannous" default="I don't know that.">WHO IS bob</sraix>`,
			expected: "I don't know that.",
		},
		{
			name:     "SRAIX with default containing variables",
			template: `<sraix service="test" default="Sorry, I can't help with that right now.">QUERY</sraix>`,
			expected: "Sorry, I can't help with that right now.",
		},
		{
			name:     "SRAIX with empty default falls back to intelligent response",
			template: `<sraix service="test" default="">FAVORITE COLOR</sraix>`,
			expected: "I don't have a specific favorite color, but I appreciate many things!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := g.ProcessTemplate(tc.template, map[string]string{})

			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}

			// Verify no XML tags in output
			if strings.Contains(result, "<") && strings.Contains(result, ">") {
				t.Errorf("Result contains XML tags: %q", result)
			}
		})
	}
}

// TestSRAIXFallbackCaseSensitivity tests that pattern matching is case-insensitive
func TestSRAIXFallbackCaseSensitivity(t *testing.T) {
	g := New(true)

	testCases := []struct {
		query            string
		expectedContains string
	}{
		{"favorite animal", "favorite animal"},
		{"FAVORITE ANIMAL", "favorite animal"},
		{"Favorite Animal", "favorite animal"},
		{"FaVoRiTe AnImAl", "favorite animal"},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			template := `<sraix service="test">` + tc.query + `</sraix>`
			result := g.ProcessTemplate(template, map[string]string{})

			if !strings.Contains(strings.ToLower(result), tc.expectedContains) {
				t.Errorf("Expected result to contain '%s', got: %q", tc.expectedContains, result)
			}
		})
	}
}

// TestSRAIXFallbackWithWildcards tests SRAIX with star tag content
func TestSRAIXFallbackWithWildcards(t *testing.T) {
	g := New(true)

	testCases := []struct {
		name             string
		template         string
		wildcards        map[string]string
		expectedContains string
	}{
		{
			name:             "FAVORITE with star",
			template:         `<sraix service="test">FAVORITE <star/></sraix>`,
			wildcards:        map[string]string{"star1": "FOOD"},
			expectedContains: "favorite food",
		},
		{
			name:             "WHO IS with star",
			template:         `<sraix service="test">WHO IS <star/></sraix>`,
			wildcards:        map[string]string{"star1": "EINSTEIN"},
			expectedContains: "information about that person",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := g.ProcessTemplate(tc.template, tc.wildcards)

			if !strings.Contains(strings.ToLower(result), tc.expectedContains) {
				t.Errorf("Expected result to contain '%s', got: %q", tc.expectedContains, result)
			}
		})
	}
}
