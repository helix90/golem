package golem

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

// sraixRecorder is a test server that records the requests it receives
type sraixRecorder struct {
	mu       sync.Mutex
	requests []*http.Request
	body     string
	status   int
}

func newSRAIXRecorder(t *testing.T, body string) (*sraixRecorder, *httptest.Server) {
	rec := &sraixRecorder{body: body, status: http.StatusOK}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.mu.Lock()
		rec.requests = append(rec.requests, r)
		status, body := rec.status, rec.body
		rec.mu.Unlock()
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return rec, srv
}

func (rec *sraixRecorder) last(t *testing.T) *http.Request {
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.requests) == 0 {
		t.Fatal("Expected a request to the SRAIX service")
	}
	return rec.requests[len(rec.requests)-1]
}

func newSRAIXGolem(t *testing.T, config *SRAIXConfig, aiml string) *Golem {
	g := New(false)
	if err := g.AddSRAIXConfig(config); err != nil {
		t.Fatalf("Failed to add SRAIX config: %v", err)
	}
	if err := g.LoadAIMLFromString(aiml); err != nil {
		t.Fatalf("Failed to load AIML: %v", err)
	}
	return g
}

const askSRAIXAIML = `<aiml version="2.0">
<category><pattern>ASK *</pattern><template><sraix service="svc"><star/></sraix></template></category>
</aiml>`

// User input must be query-escaped, so it cannot break the URL or add
// parameters of its own.
func TestSRAIXEscapesInput(t *testing.T) {
	rec, srv := newSRAIXRecorder(t, "answer")

	for _, config := range []*SRAIXConfig{
		{Name: "svc", URLTemplate: srv.URL + "/q?text={input}&key=SECRET", Method: "GET"},
		{Name: "svc", BaseURL: srv.URL + "/q?key=SECRET", QueryParam: "text", Method: "GET"},
	} {
		g := newSRAIXGolem(t, config, askSRAIXAIML)
		if _, err := g.ProcessInput("ask at&t key=EVIL #1 café?", g.CreateSession("test")); err != nil {
			t.Fatalf("Failed to process input: %v", err)
		}
		query := rec.last(t).URL.Query()
		if got := query.Get("text"); got != "at&t key=EVIL #1 café?" {
			t.Errorf("Expected the whole input in text, got %q (query %v)", got, query)
		}
		if got := query["key"]; len(got) != 1 || got[0] != "SECRET" {
			t.Errorf("Expected only the configured key, got %v", got)
		}
	}
}

// Placeholder text in user input must not be substituted.
func TestSRAIXInputPlaceholdersNotSubstituted(t *testing.T) {
	rec, srv := newSRAIXRecorder(t, "answer")
	g := newSRAIXGolem(t, &SRAIXConfig{
		Name:        "svc",
		URLTemplate: srv.URL + "/q?text={input}",
		Method:      "GET",
		Headers:     map[string]string{"Authorization": "TOKEN123"},
	}, askSRAIXAIML)

	if _, err := g.ProcessInput("ask {apikey} {lat} {BOTID}", g.CreateSession("test")); err != nil {
		t.Fatalf("Failed to process input: %v", err)
	}
	req := rec.last(t)
	if got := req.URL.Query().Get("text"); got != "{apikey} {lat} {BOTID}" {
		t.Errorf("Expected the input unchanged, got %q", got)
	}
	if strings.Contains(req.URL.String(), "TOKEN123") {
		t.Errorf("API key leaked into URL: %s", req.URL)
	}
}

// Pattern wildcards and the location predicate must be available to URL
// templates.
func TestSRAIXTemplateWildcardsAndLocation(t *testing.T) {
	rec, srv := newSRAIXRecorder(t, "sunny")
	g := newSRAIXGolem(t, &SRAIXConfig{
		Name:        "svc",
		URLTemplate: srv.URL + "/weather?q={location}&day={STAR2}&city={STAR1}",
		Method:      "GET",
	}, `<aiml version="2.0">
<category><pattern>WEATHER IN * ON *</pattern><template><think><set name="location">New York</set></think><sraix service="svc">weather</sraix></template></category>
</aiml>`)

	response, err := g.ProcessInput("weather in paris on monday", g.CreateSession("test"))
	if err != nil {
		t.Fatalf("Failed to process input: %v", err)
	}
	if response != "sunny" {
		t.Errorf("Expected 'sunny', got %q", response)
	}
	query := rec.last(t).URL.Query()
	want := url.Values{"q": {"New York"}, "day": {"monday"}, "city": {"paris"}}
	for key := range want {
		if !strings.EqualFold(query.Get(key), want.Get(key)) {
			t.Errorf("Expected %s=%q, got %q", key, want.Get(key), query.Get(key))
		}
	}
}

// Secrets from environment variables and the API key must not be logged.
func TestSRAIXVerboseLogMasksSecrets(t *testing.T) {
	t.Setenv("SRAIX_TEST_SECRET", "ENVSECRET")
	_, srv := newSRAIXRecorder(t, "answer")

	var logs bytes.Buffer
	sm := NewSRAIXManager(log.New(&logs, "", 0), true)
	if err := sm.AddConfig(&SRAIXConfig{
		Name:        "svc",
		URLTemplate: srv.URL + "/q/${SRAIX_TEST_SECRET}?key={apikey}&text={input}",
		Method:      "GET",
		Headers:     map[string]string{"Authorization": "TOKEN123"},
	}); err != nil {
		t.Fatalf("Failed to add config: %v", err)
	}
	if _, err := sm.ProcessSRAIX("svc", "hello", map[string]string{}); err != nil {
		t.Fatalf("SRAIX failed: %v", err)
	}
	if strings.Contains(logs.String(), "ENVSECRET") || strings.Contains(logs.String(), "TOKEN123") {
		t.Errorf("Secrets leaked into log:\n%s", logs.String())
	}
}
