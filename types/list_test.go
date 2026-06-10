package types

import "testing"

func TestNewListValue(t *testing.T) {
	l := NewListValue()
	if l == nil {
		t.Fatal("NewListValue returned nil")
	}
	if l.Len() != 0 {
		t.Errorf("new list Len() = %d, want 0", l.Len())
	}
}

func TestListValue_PushRight(t *testing.T) {
	l := NewListValue()
	n := l.PushRight("a", "b", "c")
	if n != 3 {
		t.Errorf("PushRight returned %d, want 3", n)
	}
	data := l.Data()
	expected := []string{"a", "b", "c"}
	for i, v := range expected {
		if data[i] != v {
			t.Errorf("Data()[%d] = %q, want %q", i, data[i], v)
		}
	}
}

func TestListValue_PushLeft(t *testing.T) {
	l := NewListValue()
	l.PushRight("b")
	l.PushLeft("a")

	data := l.Data()
	if data[0] != "a" || data[1] != "b" {
		t.Errorf("Data() = %v, want [a b]", data)
	}
}

func TestListValue_PopLeft(t *testing.T) {
	l := NewListValue()
	l.PushRight("a", "b", "c")

	val, ok := l.PopLeft()
	if !ok || val != "a" {
		t.Errorf("PopLeft() = (%q, %v), want (a, true)", val, ok)
	}
	if l.Len() != 2 {
		t.Errorf("Len() = %d, want 2", l.Len())
	}
}

func TestListValue_PopLeft_Empty(t *testing.T) {
	l := NewListValue()
	_, ok := l.PopLeft()
	if ok {
		t.Error("PopLeft() on empty list should return false")
	}
}

func TestListValue_PopRight(t *testing.T) {
	l := NewListValue()
	l.PushRight("a", "b", "c")

	val, ok := l.PopRight()
	if !ok || val != "c" {
		t.Errorf("PopRight() = (%q, %v), want (c, true)", val, ok)
	}
}

func TestListValue_PopRight_Empty(t *testing.T) {
	l := NewListValue()
	_, ok := l.PopRight()
	if ok {
		t.Error("PopRight() on empty list should return false")
	}
}

func TestListValue_Index(t *testing.T) {
	l := NewListValue()
	l.PushRight("a", "b", "c")

	tests := []struct {
		index int
		want  string
		ok    bool
	}{
		{0, "a", true},
		{1, "b", true},
		{2, "c", true},
		{-1, "c", true},
		{-3, "a", true},
		{5, "", false},
	}
	for _, tt := range tests {
		val, ok := l.Index(tt.index)
		if ok != tt.ok || (ok && val != tt.want) {
			t.Errorf("Index(%d) = (%q, %v), want (%q, %v)", tt.index, val, ok, tt.want, tt.ok)
		}
	}
}

func TestListValue_Range(t *testing.T) {
	l := NewListValue()
	l.PushRight("a", "b", "c", "d")

	tests := []struct {
		start, stop int
		want        []string
	}{
		{0, -1, []string{"a", "b", "c", "d"}},
		{1, 2, []string{"b", "c"}},
		{-2, -1, []string{"c", "d"}},
		{0, 10, []string{"a", "b", "c", "d"}},
		{3, 1, nil},
	}
	for _, tt := range tests {
		got := l.Range(tt.start, tt.stop)
		if len(got) != len(tt.want) {
			t.Errorf("Range(%d, %d) len = %d, want %d", tt.start, tt.stop, len(got), len(tt.want))
			continue
		}
		for i, v := range tt.want {
			if got[i] != v {
				t.Errorf("Range(%d, %d)[%d] = %q, want %q", tt.start, tt.stop, i, got[i], v)
			}
		}
	}
}

func TestListValue_LoadData(t *testing.T) {
	l := NewListValue()
	l.Load([]string{"x", "y", "z"})

	if l.Len() != 3 {
		t.Errorf("Len() = %d, want 3", l.Len())
	}
	data := l.Data()
	expected := []string{"x", "y", "z"}
	for i, v := range expected {
		if data[i] != v {
			t.Errorf("Data()[%d] = %q, want %q", i, data[i], v)
		}
	}
}
