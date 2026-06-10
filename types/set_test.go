package types

import (
	"sort"
	"testing"
)

func TestNewSetValue(t *testing.T) {
	s := NewSetValue()
	if s == nil {
		t.Fatal("NewSetValue returned nil")
	}
	if s.Len() != 0 {
		t.Errorf("new set Len() = %d, want 0", s.Len())
	}
}

func TestSetValue_Add(t *testing.T) {
	s := NewSetValue()
	added := s.Add("a", "b", "c")
	if added != 3 {
		t.Errorf("Add(a,b,c) = %d, want 3", added)
	}
}

func TestSetValue_AddDuplicate(t *testing.T) {
	s := NewSetValue()
	s.Add("a", "b")
	added := s.Add("b", "c")
	if added != 1 {
		t.Errorf("Add(b,c) = %d, want 1", added)
	}
	if s.Len() != 3 {
		t.Errorf("Len() = %d, want 3", s.Len())
	}
}

func TestSetValue_Remove(t *testing.T) {
	s := NewSetValue()
	s.Add("a", "b", "c")

	removed := s.Remove("b", "d")
	if removed != 1 {
		t.Errorf("Remove(b,d) = %d, want 1", removed)
	}
	if s.Exists("b") {
		t.Error("b should be removed")
	}
}

func TestSetValue_Exists(t *testing.T) {
	s := NewSetValue()
	s.Add("a")

	if !s.Exists("a") {
		t.Error("Exists(a) = false, want true")
	}
	if s.Exists("z") {
		t.Error("Exists(z) = true, want false")
	}
}

func TestSetValue_MembersList(t *testing.T) {
	s := NewSetValue()
	s.Add("a", "b", "c")

	members := s.MembersList()
	sort.Strings(members)
	expected := []string{"a", "b", "c"}
	if len(members) != len(expected) {
		t.Fatalf("MembersList() len = %d, want %d", len(members), len(expected))
	}
	for i, v := range expected {
		if members[i] != v {
			t.Errorf("MembersList()[%d] = %q, want %q", i, members[i], v)
		}
	}
}

func TestSetValue_RandomMembers(t *testing.T) {
	s := NewSetValue()
	s.Add("a", "b", "c")

	// 正数：可重复
	members := s.RandomMembers(5)
	if len(members) != 5 {
		t.Errorf("RandomMembers(5) len = %d, want 5", len(members))
	}

	// 负数：不可重复
	members = s.RandomMembers(-2)
	if len(members) != 2 {
		t.Errorf("RandomMembers(-2) len = %d, want 2", len(members))
	}
}

func TestSetValue_RandomMembers_Empty(t *testing.T) {
	s := NewSetValue()
	members := s.RandomMembers(3)
	if members != nil {
		t.Errorf("RandomMembers on empty set = %v, want nil", members)
	}
}
