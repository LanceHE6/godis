package types

import (
	"testing"
)

func TestNewZSetValue(t *testing.T) {
	z := NewZSetValue()
	if z == nil {
		t.Fatal("NewZSetValue returned nil")
	}
	if z.Len() != 0 {
		t.Errorf("new zset Len() = %d, want 0", z.Len())
	}
}

func TestZSetValue_Add(t *testing.T) {
	z := NewZSetValue()
	added := z.Add(1.0, "a")
	if added != 1 {
		t.Errorf("Add(a) = %d, want 1", added)
	}
	// 更新已存在的 member
	added = z.Add(2.0, "a")
	if added != 0 {
		t.Errorf("Add(a) again = %d, want 0", added)
	}
}

func TestZSetValue_AddMultiple(t *testing.T) {
	z := NewZSetValue()
	added := z.AddMultiple(
		ZSetMember{Member: "a", Score: 1.0},
		ZSetMember{Member: "b", Score: 2.0},
		ZSetMember{Member: "c", Score: 3.0},
	)
	if added != 3 {
		t.Errorf("AddMultiple = %d, want 3", added)
	}
}

func TestZSetValue_Remove(t *testing.T) {
	z := NewZSetValue()
	z.Add(1.0, "a")

	if !z.Remove("a") {
		t.Error("Remove(a) = false, want true")
	}
	if z.Remove("a") {
		t.Error("Remove(a) again = true, want false")
	}
}

func TestZSetValue_Score(t *testing.T) {
	z := NewZSetValue()
	z.Add(3.14, "pi")

	score, ok := z.Score("pi")
	if !ok || score != 3.14 {
		t.Errorf("Score(pi) = (%f, %v), want (3.14, true)", score, ok)
	}

	_, ok = z.Score("missing")
	if ok {
		t.Error("Score(missing) should not exist")
	}
}

func TestZSetValue_Exists(t *testing.T) {
	z := NewZSetValue()
	z.Add(1.0, "a")

	if !z.Exists("a") {
		t.Error("Exists(a) = false, want true")
	}
	if z.Exists("z") {
		t.Error("Exists(z) = true, want false")
	}
}

func TestZSetValue_Range(t *testing.T) {
	z := NewZSetValue()
	z.Add(1.0, "a")
	z.Add(2.0, "b")
	z.Add(3.0, "c")
	z.Add(4.0, "d")

	// 全范围
	all := z.Range(0, -1)
	if len(all) != 4 {
		t.Fatalf("Range(0,-1) len = %d, want 4", len(all))
	}
	if all[0].Member != "a" || all[3].Member != "d" {
		t.Errorf("Range order wrong: %v", all)
	}

	// 子范围
	sub := z.Range(1, 2)
	if len(sub) != 2 || sub[0].Member != "b" || sub[1].Member != "c" {
		t.Errorf("Range(1,2) = %v, want [b c]", sub)
	}
}

func TestZSetValue_Rank(t *testing.T) {
	z := NewZSetValue()
	z.Add(10.0, "a")
	z.Add(20.0, "b")
	z.Add(30.0, "c")

	if r := z.Rank("a"); r != 0 {
		t.Errorf("Rank(a) = %d, want 0", r)
	}
	if r := z.Rank("c"); r != 2 {
		t.Errorf("Rank(c) = %d, want 2", r)
	}
	if r := z.Rank("z"); r != -1 {
		t.Errorf("Rank(z) = %d, want -1", r)
	}
}

func TestZSetValue_RangeByScore(t *testing.T) {
	z := NewZSetValue()
	z.Add(1.0, "a")
	z.Add(2.0, "b")
	z.Add(3.0, "c")
	z.Add(4.0, "d")

	result := z.RangeByScore(2.0, 3.0, 0, 0)
	if len(result) != 2 {
		t.Fatalf("RangeByScore(2,3) len = %d, want 2", len(result))
	}
	if result[0].Member != "b" || result[1].Member != "c" {
		t.Errorf("RangeByScore(2,3) = %v, want [b c]", result)
	}
}

func TestZSetValue_LoadData(t *testing.T) {
	z := NewZSetValue()
	z.Load([]ZSetMember{
		{Member: "a", Score: 1.0},
		{Member: "b", Score: 2.0},
	})

	if z.Len() != 2 {
		t.Errorf("Len() = %d, want 2", z.Len())
	}

	data := z.Data()
	if len(data) != 2 || data[0].Member != "a" || data[1].Member != "b" {
		t.Errorf("Data() = %v, want [a b]", data)
	}
}
