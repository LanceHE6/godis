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

// SaveToBinary 将当前内存数据库的所有数据，以二进制 Gob 格式写入指定的写入器（如文件）
func (db *GodisDB) SaveToBinary(w io.Writer) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// 过滤掉已经过期的数据，只打包活着的键值对
	now := time.Now()
	cleanData := make(map[string]GobItem)

	for k, v := range db.data {
		if !v.IsNeverDie && now.After(v.ExpiresAt) {
			continue // 已过期的跳过，不写入快照
		}
		cleanData[k] = GobItem{
			Value:      v.Value,
			ExpiresAt:  v.ExpiresAt,
			IsNeverDie: v.IsNeverDie,
		}
	}

	// 使用 Gob 编码器将 map 整体转为二进制流写入
	encoder := gob.NewEncoder(w)
	return encoder.Encode(cleanData)
}

// LoadFromBinary 从二进制 Gob 流中恢复数据到内存
func (db *GodisDB) LoadFromBinary(r io.Reader) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	var decodedData map[string]GobItem
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&decodedData); err != nil {
		return err
	}

	// 还原到原生的数据库 map 中
	for k, v := range decodedData {
		db.data[k] = Item{
			Value:      v.Value,
			ExpiresAt:  v.ExpiresAt,
			IsNeverDie: v.IsNeverDie,
		}
	}
	return nil
}
