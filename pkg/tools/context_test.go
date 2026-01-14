package tools

import (
	"testing"
	"time"
)

func TestNewToolContext(t *testing.T) {
	input := SearchInput{Query: "test"}
	tc := NewToolContext(ToolSearch, input)

	if tc.ToolName != ToolSearch {
		t.Errorf("NewToolContext() ToolName = %v, want %v", tc.ToolName, ToolSearch)
	}

	if tc.Input != input {
		t.Errorf("NewToolContext() Input mismatch")
	}

	if tc.Extra == nil {
		t.Error("NewToolContext() Extra should be initialized")
	}

	if tc.StartTime.IsZero() {
		t.Error("NewToolContext() StartTime should be set")
	}
}

func TestToolContextDuration(t *testing.T) {
	tc := NewToolContext(ToolSearch, nil)

	// Sleep briefly
	time.Sleep(10 * time.Millisecond)

	duration := tc.Duration()
	if duration < 10*time.Millisecond {
		t.Errorf("Duration() = %v, expected at least 10ms", duration)
	}
}

func TestToolContextSetGet(t *testing.T) {
	tc := NewToolContext(ToolSearch, nil)

	// Test Set and Get
	tc.Set("key1", "value1")
	tc.Set("key2", 42)
	tc.Set("key3", []string{"a", "b"})

	// Get existing keys
	v1, ok := tc.Get("key1")
	if !ok {
		t.Error("Get() should return true for existing key")
	}
	if v1 != "value1" {
		t.Errorf("Get() = %v, want value1", v1)
	}

	v2, ok := tc.Get("key2")
	if !ok {
		t.Error("Get() should return true for existing key")
	}
	if v2 != 42 {
		t.Errorf("Get() = %v, want 42", v2)
	}

	// Get non-existent key
	_, ok = tc.Get("nonexistent")
	if ok {
		t.Error("Get() should return false for non-existent key")
	}
}

func TestToolContextOverwrite(t *testing.T) {
	tc := NewToolContext(ToolSearch, nil)

	tc.Set("key", "original")
	tc.Set("key", "updated")

	v, ok := tc.Get("key")
	if !ok {
		t.Error("Get() should return true")
	}
	if v != "updated" {
		t.Errorf("Get() = %v, want updated", v)
	}
}
