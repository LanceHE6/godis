package datastore

import (
	"encoding/gob"
	"godis/types"
	"io"
	"time"
)

// ── Gob 序列化用的快照结构（不含 mutex） ──

type GobStringSnapshot struct {
	Value string
}

type GobHashSnapshot struct {
	Fields map[string]string
}

type GobListSnapshot struct {
	Values []string
}

type GobSetSnapshot struct {
	Members []string
}

type GobZSetSnapshot struct {
	Members []types.ZSetMember
}

// GobItem 用于二进制传输的数据项
type GobItem struct {
	Type       types.DataType
	Value      interface{} // *GobStringSnapshot / *GobHashSnapshot / *GobListSnapshot / *GobSetSnapshot / *GobZSetSnapshot
	ExpiresAt  time.Time
	IsNeverDie bool
}

// AllDbsSnapshot 用于序列化所有数据库的快照结构
type AllDbsSnapshot struct {
	DBs []map[string]GobItem
}

func init() {
	gob.Register(&GobStringSnapshot{})
	gob.Register(&GobHashSnapshot{})
	gob.Register(&GobListSnapshot{})
	gob.Register(&GobSetSnapshot{})
	gob.Register(&GobZSetSnapshot{})
}

// toGobValue 将运行时 Value 转换为可序列化的快照
func toGobValue(t types.DataType, v interface{}) interface{} {
	switch t {
	case types.TypeString:
		return &GobStringSnapshot{Value: v.(*types.StringValue).Value}
	case types.TypeHash:
		hv := v.(*types.HashValue)
		hv.Mu.RLock()
		fields := make(map[string]string, len(hv.Fields))
		for k, val := range hv.Fields {
			fields[k] = val
		}
		hv.Mu.RUnlock()
		return &GobHashSnapshot{Fields: fields}
	case types.TypeList:
		lv := v.(*types.ListValue)
		return &GobListSnapshot{Values: lv.Data()}
	case types.TypeSet:
		sv := v.(*types.SetValue)
		return &GobSetSnapshot{Members: sv.MembersList()}
	case types.TypeZSet:
		zv := v.(*types.ZSetValue)
		return &GobZSetSnapshot{Members: zv.Data()}
	}
	return v
}

// fromGobValue 将序列化的快照还原为运行时 Value
func fromGobValue(t types.DataType, v interface{}) interface{} {
	switch t {
	case types.TypeString:
		return types.NewStringValue(v.(*GobStringSnapshot).Value)
	case types.TypeHash:
		snap := v.(*GobHashSnapshot)
		hv := types.NewHashValue()
		hv.Fields = snap.Fields
		return hv
	case types.TypeList:
		snap := v.(*GobListSnapshot)
		lv := types.NewListValue()
		if len(snap.Values) > 0 {
			lv.Load(snap.Values)
		}
		return lv
	case types.TypeSet:
		snap := v.(*GobSetSnapshot)
		sv := types.NewSetValue()
		if len(snap.Members) > 0 {
			sv.Add(snap.Members...)
		}
		return sv
	case types.TypeZSet:
		snap := v.(*GobZSetSnapshot)
		zv := types.NewZSetValue()
		zv.Load(snap.Members)
		return zv
	}
	return v
}

// SaveAllToBinary 将所有数据库序列化为二进制 Gob 格式
func SaveAllToBinary(w io.Writer, dbs []*GodisDB) error {
	snapshot := AllDbsSnapshot{
		DBs: make([]map[string]GobItem, len(dbs)),
	}
	now := time.Now()

	for i, db := range dbs {
		db.mu.RLock()
		cleanData := make(map[string]GobItem)
		for k, v := range db.data {
			if !v.IsNeverDie && now.After(v.ExpiresAt) {
				continue
			}
			cleanData[k] = GobItem{
				Type:       v.Type,
				Value:      toGobValue(v.Type, v.Value),
				ExpiresAt:  v.ExpiresAt,
				IsNeverDie: v.IsNeverDie,
			}
		}
		db.mu.RUnlock()
		snapshot.DBs[i] = cleanData
	}

	encoder := gob.NewEncoder(w)
	return encoder.Encode(snapshot)
}

// LoadAllFromBinary 从二进制 Gob 流中恢复所有数据库
func LoadAllFromBinary(r io.Reader, dbs []*GodisDB) error {
	var snapshot AllDbsSnapshot
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&snapshot); err != nil {
		return err
	}

	for i, data := range snapshot.DBs {
		if i >= len(dbs) {
			break
		}
		dbs[i].mu.Lock()
		now := time.Now()
		for k, v := range data {
			// 跳过已过期的数据，避免恢复后走 GC 清理
			if !v.IsNeverDie && now.After(v.ExpiresAt) {
				continue
			}
			dbs[i].data[k] = Item{
				Type:       v.Type,
				Value:      fromGobValue(v.Type, v.Value),
				ExpiresAt:  v.ExpiresAt,
				IsNeverDie: v.IsNeverDie,
			}
		}
		dbs[i].mu.Unlock()
	}
	return nil
}
