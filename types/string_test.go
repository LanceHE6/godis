package types

import "testing"

func TestNewStringValue(t *testing.T) {
	s := NewStringValue("hello")
	if s == nil {
		t.Fatal("NewStringValue returned nil")
	}
	if s.Value != "hello" {
		t.Errorf("Value = %q, want %q", s.Value, "hello")
	}
}

func TestNewStringValue_Empty(t *testing.T) {
	s := NewStringValue("")
	if s.Value != "" {
		t.Errorf("Value = %q, want %q", s.Value, "")
	}
}
