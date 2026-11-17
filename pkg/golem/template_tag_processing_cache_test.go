package golem

import (
	"testing"
	"time"
)

func TestTemplateTagProcessingCacheBasicOperations(t *testing.T) {
	// Create a new template tag processing cache
	cache := NewTemplateTagProcessingCache(10, 60)

	// Create a test context
	ctx := &VariableContext{
		LocalVars: map[string]string{"test_var": "local_value"},
		Session:   &ChatSession{ID: "test_session", Variables: map[string]string{"session_var": "session_value"}},
		Topic:     "",
		KnowledgeBase: &AIMLKnowledgeBase{
			Variables:  map[string]string{"global_var": "global_value"},
			Properties: map[string]string{"property_var": "property_value"},
		},
	}

	// Test getting a non-existent tag
	result, found := cache.GetProcessedTag("person", "hello world", ctx)
	if found {
		t.Error("Expected tag not to be found in cache")
	}
	if result != "" {
		t.Error("Expected empty result for non-existent tag")
	}

	// Test setting and getting a tag
	cache.SetProcessedTag("person", "hello world", "HELLO WORLD", ctx)
	result, found = cache.GetProcessedTag("person", "hello world", ctx)
	if !found {
		t.Error("Expected tag to be found in cache")
	}
	if result != "HELLO WORLD" {
		t.Errorf("Expected 'HELLO WORLD', got '%s'", result)
	}

	// Check cache stats
	stats := cache.GetCacheStats()
	if stats["results"].(int) != 1 {
		t.Errorf("Expected 1 result in cache, got %v", stats["results"])
	}
	if stats["misses"].(int) != 1 {
		t.Errorf("Expected 1 miss, got %v", stats["misses"])
	}
}

func TestTemplateTagProcessingCacheContextInvalidation(t *testing.T) {
	cache := NewTemplateTagProcessingCache(10, 60)

	// Create initial context
	ctx1 := &VariableContext{
		LocalVars:     map[string]string{"test_var": "value1"},
		Session:       &ChatSession{ID: "session1", Variables: map[string]string{}},
		Topic:         "",
		KnowledgeBase: &AIMLKnowledgeBase{Variables: map[string]string{}},
	}

	// Set a tag in cache
	cache.SetProcessedTag("person", "hello", "HELLO", ctx1)
	result, found := cache.GetProcessedTag("person", "hello", ctx1)
	if !found || result != "HELLO" {
		t.Error("Expected tag to be found with correct value")
	}

	// Create context with different variables
	ctx2 := &VariableContext{
		LocalVars:     map[string]string{"test_var": "value2"}, // Different value
		Session:       &ChatSession{ID: "session1", Variables: map[string]string{}},
		Topic:         "",
		KnowledgeBase: &AIMLKnowledgeBase{Variables: map[string]string{}},
	}

	// Should not find cached value due to context change
	_, found = cache.GetProcessedTag("person", "hello", ctx2)
	if found {
		t.Error("Expected tag not to be found due to context change")
	}

	// Set new value for new context
	cache.SetProcessedTag("person", "hello", "HELLO2", ctx2)
	result, found = cache.GetProcessedTag("person", "hello", ctx2)
	if !found || result != "HELLO2" {
		t.Error("Expected tag to be found with new value")
	}
}

func TestTemplateTagProcessingCacheLRUEviction(t *testing.T) {
	// Create a small cache
	cache := NewTemplateTagProcessingCache(2, 60)

	ctx := &VariableContext{
		Session: &ChatSession{ID: "test_session"},
	}

	// Add tags to fill the cache
	cache.SetProcessedTag("person", "hello", "HELLO", ctx)
	cache.SetProcessedTag("gender", "world", "WORLD", ctx)

	// Add a third tag to trigger eviction
	cache.SetProcessedTag("uppercase", "test", "TEST", ctx)

	// Check that cache size is still 2
	stats := cache.GetCacheStats()
	if stats["results"].(int) != 2 {
		t.Errorf("Expected 2 results in cache after eviction, got %v", stats["results"])
	}

	// person tag should be evicted (LRU)
	_, found := cache.GetProcessedTag("person", "hello", ctx)
	if found {
		t.Error("Expected person tag to be evicted")
	}

	// gender and uppercase tags should still be there
	_, found = cache.GetProcessedTag("gender", "world", ctx)
	if !found {
		t.Error("Expected gender tag to still be in cache")
	}
	_, found = cache.GetProcessedTag("uppercase", "test", ctx)
	if !found {
		t.Error("Expected uppercase tag to still be in cache")
	}
}

func TestTemplateTagProcessingCacheTTL(t *testing.T) {
	// Create a cache with very short TTL
	cache := NewTemplateTagProcessingCache(10, 1) // 1 second TTL

	ctx := &VariableContext{
		Session: &ChatSession{ID: "test_session"},
	}

	// Add a tag
	cache.SetProcessedTag("person", "hello", "HELLO", ctx)

	// Verify it's in cache
	stats := cache.GetCacheStats()
	if stats["results"].(int) != 1 {
		t.Errorf("Expected 1 result in cache, got %v", stats["results"])
	}

	// Wait for TTL to expire
	time.Sleep(2 * time.Second)

	// Try to get the tag again - should not be found due to TTL expiry
	_, found := cache.GetProcessedTag("person", "hello", ctx)
	if found {
		t.Error("Expected tag not to be found after TTL expiry")
	}
}

func TestTemplateTagProcessingCacheTagTypeInvalidation(t *testing.T) {
	cache := NewTemplateTagProcessingCache(10, 60)

	ctx := &VariableContext{
		Session: &ChatSession{ID: "test_session"},
	}

	// Set up some tags with different types
	cache.SetProcessedTag("person", "hello", "HELLO", ctx)
	cache.SetProcessedTag("gender", "world", "WORLD", ctx)
	cache.SetProcessedTag("uppercase", "test", "TEST", ctx)

	// Invalidate person tags
	cache.InvalidateTagType("person")

	// person tag should be removed
	_, found := cache.GetProcessedTag("person", "hello", ctx)
	if found {
		t.Error("Expected person tag to be removed after tag type invalidation")
	}

	// gender and uppercase tags should still be there
	_, found = cache.GetProcessedTag("gender", "world", ctx)
	if !found {
		t.Error("Expected gender tag to still be in cache after tag type invalidation")
	}
	_, found = cache.GetProcessedTag("uppercase", "test", ctx)
	if !found {
		t.Error("Expected uppercase tag to still be in cache after tag type invalidation")
	}
}

func TestTemplateTagProcessingCacheAdvancedFeatures(t *testing.T) {
	cache := NewTemplateTagProcessingCache(3, 2) // Small cache, 2 second TTL

	ctx := &VariableContext{
		Session: &ChatSession{ID: "test_session"},
	}

	// Test TTL functionality
	cache.SetProcessedTag("person", "hello", "HELLO", ctx)

	// Should be found immediately
	_, found := cache.GetProcessedTag("person", "hello", ctx)
	if !found {
		t.Error("Expected tag to be found immediately")
	}

	// Wait for TTL to expire
	time.Sleep(3 * time.Second)

	// Should not be found after TTL expiry
	_, found = cache.GetProcessedTag("person", "hello", ctx)
	if found {
		t.Error("Expected tag not to be found after TTL expiry")
	}

	// Test access order tracking
	cache.SetProcessedTag("person", "hello", "HELLO", ctx)
	cache.SetProcessedTag("gender", "world", "WORLD", ctx)
	cache.SetProcessedTag("uppercase", "test", "TEST", ctx)

	// Access person tag to make it most recent
	cache.GetProcessedTag("person", "hello", ctx)

	// Add lowercase tag, should evict gender tag (least recently used)
	cache.SetProcessedTag("lowercase", "message", "MESSAGE", ctx)

	// gender tag should be evicted
	_, found = cache.GetProcessedTag("gender", "world", ctx)
	if found {
		t.Error("Expected gender tag to be evicted")
	}

	// person, uppercase, lowercase tags should still be there
	_, found = cache.GetProcessedTag("person", "hello", ctx)
	if !found {
		t.Error("Expected person tag to still be in cache")
	}
	_, found = cache.GetProcessedTag("uppercase", "test", ctx)
	if !found {
		t.Error("Expected uppercase tag to still be in cache")
	}
	_, found = cache.GetProcessedTag("lowercase", "message", ctx)
	if !found {
		t.Error("Expected lowercase tag to still be in cache")
	}
}

func TestTemplateTagProcessingCacheStatistics(t *testing.T) {
	cache := NewTemplateTagProcessingCache(10, 60)

	ctx := &VariableContext{
		Session: &ChatSession{ID: "test_session"},
	}

	// Add some tags
	cache.SetProcessedTag("person", "hello", "HELLO", ctx)
	cache.SetProcessedTag("gender", "world", "WORLD", ctx)

	// Access tags to generate hits
	cache.GetProcessedTag("person", "hello", ctx)
	cache.GetProcessedTag("person", "hello", ctx)
	cache.GetProcessedTag("gender", "world", ctx)

	stats := cache.GetCacheStats()

	// Verify statistics
	if stats["results"].(int) != 2 {
		t.Errorf("Expected 2 results, got %v", stats["results"])
	}
	if stats["max_size"].(int) != 10 {
		t.Errorf("Expected max_size 10, got %v", stats["max_size"])
	}
	if stats["ttl_seconds"].(int64) != 60 {
		t.Errorf("Expected ttl_seconds 60, got %v", stats["ttl_seconds"])
	}
	if stats["misses"].(int) != 0 {
		t.Errorf("Expected 0 misses, got %v", stats["misses"])
	}
	if stats["tag_types"].(int) != 2 {
		t.Errorf("Expected 2 tag types, got %v", stats["tag_types"])
	}

	// Check hit rate
	hitRate := stats["hit_rate"].(float64)
	if hitRate <= 0 {
		t.Error("Expected positive hit rate")
	}
}
