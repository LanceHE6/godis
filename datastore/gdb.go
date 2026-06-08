package datastore

import (
	"encoding/gob"
	"io"
	"time"
)

// GobItem 专门用于二进制传输的结构体
// 因为 time.Time 默认支持 Gob 序列化，我们直接复用
type GobItem struct {
	Value      string
	ExpiresAt  time.Time
	IsNeverDie bool
}

// AllDbsSnapshot 用于序列化所有数据库的快照结构
type AllDbsSnapshot struct {
	DBs []map[string]GobItem
}

// SaveAllToBinary 将所有数据库序列化为二进制 Gob 格式
func SaveAllToBinary(w io.Writer, dbs []*GodisDB) error {
	snapshot := AllDbsSnapshot{
		DBs: make([]map[string]GobItem, len(dbs)),
	}
	now := time.Now()

	// 复制整个数据库
	for i, db := range dbs {
		db.mu.RLock()
		cleanData := make(map[string]GobItem)
		for k, v := range db.data {
			if !v.IsNeverDie && now.After(v.ExpiresAt) {
				continue
			}
			cleanData[k] = GobItem{
				Value:      v.Value,
				ExpiresAt:  v.ExpiresAt,
				IsNeverDie: v.IsNeverDie,
			}
		}
		db.mu.RUnlock()
		snapshot.DBs[i] = cleanData
	}

	// 转换为二进制
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

	// 逐一恢复数据库数据
	for i, data := range snapshot.DBs {
		if i >= len(dbs) {
			break
		}
		dbs[i].mu.Lock()
		for k, v := range data {
			dbs[i].data[k] = Item{
				Value:      v.Value,
				ExpiresAt:  v.ExpiresAt,
				IsNeverDie: v.IsNeverDie,
			}
		}
		dbs[i].mu.Unlock()
	}
	return nil
}
