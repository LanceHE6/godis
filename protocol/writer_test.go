package protocol

import "testing"

func TestMakeSimpleString(t *testing.T) {
	got := MakeSimpleString("OK")
	want := "+OK\r\n"
	if got != want {
		t.Errorf("MakeSimpleString(%q) = %q, want %q", "OK", got, want)
	}
}

func TestMakeError(t *testing.T) {
	got := MakeError("ERR unknown command")
	want := "-ERR unknown command\r\n"
	if got != want {
		t.Errorf("MakeError = %q, want %q", got, want)
	}
}

func TestMakeInt(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, ":0\r\n"},
		{100, ":100\r\n"},
		{-1, ":-1\r\n"},
	}
	for _, tt := range tests {
		got := MakeInt(tt.input)
		if got != tt.want {
			t.Errorf("MakeInt(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMakeBulkString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "$5\r\nhello\r\n"},
		{"", "$0\r\n\r\n"},
		{"hello world", "$11\r\nhello world\r\n"},
	}
	for _, tt := range tests {
		got := MakeBulkString(tt.input)
		if got != tt.want {
			t.Errorf("MakeBulkString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMakeNull(t *testing.T) {
	got := MakeNull()
	want := "$-1\r\n"
	if got != want {
		t.Errorf("MakeNull() = %q, want %q", got, want)
	}
}

func TestMakeArray(t *testing.T) {
	elements := []string{
		MakeBulkString("hello"),
		MakeBulkString("world"),
	}
	got := MakeArray(elements)
	want := "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n"
	if got != want {
		t.Errorf("MakeArray = %q, want %q", got, want)
	}
}

func TestMakeArray_Empty(t *testing.T) {
	got := MakeArray([]string{})
	want := "*0\r\n"
	if got != want {
		t.Errorf("MakeArray(empty) = %q, want %q", got, want)
	}
}
