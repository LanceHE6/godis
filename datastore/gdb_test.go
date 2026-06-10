package datastore

import (
	"bytes"
	"godis/types"
	"testing"
)

func TestSaveLoadAllFromBinary_String(t *testing.T) {
	dbs := []*GodisDB{NewGodisDB(), NewGodisDB()}
	dbs[0].Set("hello", "world", 0)

	var buf bytes.Buffer
	if err := SaveAllToBinary(&buf, dbs); err != nil {
		t.Fatalf("SaveAllToBinary: %v", err)
	}

	restored := []*GodisDB{NewGodisDB(), NewGodisDB()}
	if err := LoadAllFromBinary(&buf, restored); err != nil {
		t.Fatalf("LoadAllFromBinary: %v", err)
	}

	val, ok := restored[0].Get("hello")
	if !ok || val != "world" {
		t.Errorf("restored Get(hello) = (%q, %v), want (world, true)", val, ok)
	}
}

func TestSaveLoadAllFromBinary_MultiDB(t *testing.T) {
	dbs := []*GodisDB{NewGodisDB(), NewGodisDB()}
	dbs[0].Set("db0-key", "db0-val", 0)
	dbs[1].Set("db1-key", "db1-val", 0)

	var buf bytes.Buffer
	if err := SaveAllToBinary(&buf, dbs); err != nil {
		t.Fatalf("SaveAllToBinary: %v", err)
	}

	restored := []*GodisDB{NewGodisDB(), NewGodisDB()}
	if err := LoadAllFromBinary(&buf, restored); err != nil {
		t.Fatalf("LoadAllFromBinary: %v", err)
	}

	val0, ok0 := restored[0].Get("db0-key")
	if !ok0 || val0 != "db0-val" {
		t.Errorf("db0: Get = (%q, %v)", val0, ok0)
	}
	val1, ok1 := restored[1].Get("db1-key")
	if !ok1 || val1 != "db1-val" {
		t.Errorf("db1: Get = (%q, %v)", val1, ok1)
	}
}

func TestSaveLoadAllFromBinary_Hash(t *testing.T) {
	dbs := []*GodisDB{NewGodisDB()}

	hv := types.NewHashValue()
	hv.Set("f1", "v1")
	hv.Set("f2", "v2")
	dbs[0].putItem("myhash", Item{
		Type:       types.TypeHash,
		Value:      hv,
		IsNeverDie: true,
	})

	var buf bytes.Buffer
	if err := SaveAllToBinary(&buf, dbs); err != nil {
		t.Fatalf("SaveAllToBinary: %v", err)
	}

	restored := []*GodisDB{NewGodisDB()}
	if err := LoadAllFromBinary(&buf, restored); err != nil {
		t.Fatalf("LoadAllFromBinary: %v", err)
	}

	item, ok := restored[0].GetItem("myhash")
	if !ok || item.Type != types.TypeHash {
		t.Fatal("restored item not found or wrong type")
	}
	rh := item.Value.(*types.HashValue)
	v, _ := rh.Get("f1")
	if v != "v1" {
		t.Errorf("hash f1 = %q, want v1", v)
	}
}

func TestSaveLoadAllFromBinary_List(t *testing.T) {
	dbs := []*GodisDB{NewGodisDB()}

	lv := types.NewListValue()
	lv.PushRight("a", "b", "c")
	dbs[0].putItem("mylist", Item{
		Type:       types.TypeList,
		Value:      lv,
		IsNeverDie: true,
	})

	var buf bytes.Buffer
	SaveAllToBinary(&buf, dbs)

	restored := []*GodisDB{NewGodisDB()}
	LoadAllFromBinary(&buf, restored)

	item, _ := restored[0].GetItem("mylist")
	rl := item.Value.(*types.ListValue)
	if rl.Len() != 3 {
		t.Errorf("list len = %d, want 3", rl.Len())
	}
	val, _ := rl.Index(0)
	if val != "a" {
		t.Errorf("list[0] = %q, want a", val)
	}
}

func TestSaveLoadAllFromBinary_Set(t *testing.T) {
	dbs := []*GodisDB{NewGodisDB()}

	sv := types.NewSetValue()
	sv.Add("x", "y", "z")
	dbs[0].putItem("myset", Item{
		Type:       types.TypeSet,
		Value:      sv,
		IsNeverDie: true,
	})

	var buf bytes.Buffer
	SaveAllToBinary(&buf, dbs)

	restored := []*GodisDB{NewGodisDB()}
	LoadAllFromBinary(&buf, restored)

	item, _ := restored[0].GetItem("myset")
	rs := item.Value.(*types.SetValue)
	if rs.Len() != 3 {
		t.Errorf("set len = %d, want 3", rs.Len())
	}
}

func TestSaveLoadAllFromBinary_ZSet(t *testing.T) {
	dbs := []*GodisDB{NewGodisDB()}

	zv := types.NewZSetValue()
	zv.Add(1.0, "a")
	zv.Add(2.0, "b")
	dbs[0].putItem("myzset", Item{
		Type:       types.TypeZSet,
		Value:      zv,
		IsNeverDie: true,
	})

	var buf bytes.Buffer
	SaveAllToBinary(&buf, dbs)

	restored := []*GodisDB{NewGodisDB()}
	LoadAllFromBinary(&buf, restored)

	item, _ := restored[0].GetItem("myzset")
	rz := item.Value.(*types.ZSetValue)
	if rz.Len() != 2 {
		t.Errorf("zset len = %d, want 2", rz.Len())
	}
	score, ok := rz.Score("a")
	if !ok || score != 1.0 {
		t.Errorf("zset score(a) = (%f, %v), want (1.0, true)", score, ok)
	}
}
