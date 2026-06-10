package types

import (
	"sort"
	"testing"
)

func TestNewHashValue(t *testing.T) {
	h := NewHashValue()
	if h == nil {
		t.Fatal("NewHashValue returned nil")
	}
	if h.Len() != 0 {
		t.Errorf("new hash Len() = %d, want 0", h.Len())
	}
}

func TestHashValue_SetGet(t *testing.T) {
	h := NewHashValue()
	h.Set("name", "alice")

	val, ok := h.Get("name")
	if !ok || val != "alice" {
		t.Errorf("Get(name) = (%q, %v), want (alice, true)", val, ok)
	}
}

func TestHashValue_GetNotExist(t *testing.T) {
	h := NewHashValue()
	val, ok := h.Get("missing")
	if ok || val != "" {
		t.Errorf("Get(missing) = (%q, %v), want (, false)", val, ok)
	}
}

func TestHashValue_Overwrite(t *testing.T) {
	h := NewHashValue()
	h.Set("k", "v1")
	h.Set("k", "v2")

	val, _ := h.Get("k")
	if val != "v2" {
		t.Errorf("Get(k) = %q, want v2", val)
	}
}

func TestHashValue_Del(t *testing.T) {
	h := NewHashValue()
	h.Set("k", "v")

	if !h.Del("k") {
		t.Error("Del(k) = false, want true")
	}
	if h.Del("k") {
		t.Error("Del(k) again = true, want false")
	}
	if _, ok := h.Get("k"); ok {
		t.Error("Get(k) after del should not exist")
	}
}

func TestHashValue_Len(t *testing.T) {
	h := NewHashValue()
	h.Set("a", "1")
	h.Set("b", "2")
	h.Set("c", "3")

	if h.Len() != 3 {
		t.Errorf("Len() = %d, want 3", h.Len())
	}
}

func TestHashValue_Keys(t *testing.T) {
	h := NewHashValue()
	h.Set("a", "1")
	h.Set("b", "2")

	keys := h.Keys()
	sort.Strings(keys)
	expected := []string{"a", "b"}
	if len(keys) != len(expected) {
		t.Fatalf("Keys() len = %d, want %d", len(keys), len(expected))
	}
	for i, k := range expected {
		if keys[i] != k {
			t.Errorf("Keys()[%d] = %q, want %q", i, keys[i], k)
		}
	}
}

func TestHashValue_Values(t *testing.T) {
	h := NewHashValue()
	h.Set("a", "1")
	h.Set("b", "2")

	vals := h.Values()
	sort.Strings(vals)
	expected := []string{"1", "2"}
	if len(vals) != len(expected) {
		t.Fatalf("Values() len = %d, want %d", len(vals), len(expected))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("Values()[%d] = %q, want %q", i, vals[i], v)
		}
	}
}

func TestHashValue_GetAll(t *testing.T) {
	h := NewHashValue()
	h.Set("a", "1")
	h.Set("b", "2")

	all := h.GetAll()
	if len(all) != 2 || all["a"] != "1" || all["b"] != "2" {
		t.Errorf("GetAll() = %v, want {a:1, b:2}", all)
	}
}

func TestHashValue_Exists(t *testing.T) {
	h := NewHashValue()
	h.Set("k", "v")

	if !h.Exists("k") {
		t.Error("Exists(k) = false, want true")
	}
	if h.Exists("missing") {
		t.Error("Exists(missing) = true, want false")
	}
}
