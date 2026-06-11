package integration

import (
	"context"
	"testing"
)

func TestPing(t *testing.T) {
	cleanDB(t)
	t.Log("sending PING...")
	result, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		t.Fatalf("PING failed: %v", err)
	}
	t.Logf("PING response: %q", result)
	if result != "PONG" {
		t.Errorf("PING = %q, want PONG", result)
	}
}

func TestPing_WithMessage(t *testing.T) {
	cleanDB(t)
	t.Log("sending PING hello...")
	result, err := rdb.Do(context.Background(), "PING", "hello").Text()
	if err != nil {
		t.Fatalf("PING hello failed: %v", err)
	}
	t.Logf("PING hello response: %q", result)
	if result != "hello" {
		t.Errorf("PING hello = %q, want hello", result)
	}
}

func TestTime(t *testing.T) {
	cleanDB(t)
	t.Log("sending TIME...")
	vals, err := rdb.Do(context.Background(), "TIME").Slice()
	if err != nil {
		t.Fatalf("TIME failed: %v", err)
	}
	t.Logf("TIME response: %v (count=%d)", vals, len(vals))
	if len(vals) != 2 {
		t.Errorf("TIME returned %d elements, want 2", len(vals))
	}
}
