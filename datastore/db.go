package datastore

import (
	"godis/logger"
	"godis/types"
	"sync"
	"time"
)

var log = logger.NewModuleLogger("DATASTORE")

// Item 带有过期时间的数据项
type Item struct {
	Type       types.DataType
	Value      interface{} // *types.StringValue / *types.HashValue / *types.ListValue / *types.SetValue / *types.ZSetValue
	ExpiresAt  time.Time
	IsNeverDie bool
}

// GodisDB 数据库结构
type GodisDB struct {
	mu   sync.RWMutex
	data map[string]Item
}

// NewGodisDB 初始化并返回数据库实例
func NewGodisDB() *GodisDB {
	db := &GodisDB{
		data: make(map[string]Item),
	}
	return db
}

// Set 存入字符串数据
func (db *GodisDB) Set(key, val string, ttlSeconds int) {
	item := Item{
		Type:       types.TypeString,
		Value:      types.NewStringValue(val),
		IsNeverDie: true,
	}

	if ttlSeconds > 0 {
		item.IsNeverDie = false
		item.ExpiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}

	db.mu.Lock()
	db.data[key] = item
	db.mu.Unlock()
}

// Get 读取字符串数据（自带惰性删除逻辑）
func (db *GodisDB) Get(key string) (string, bool) {
	db.mu.RLock()
	item, exists := db.data[key]
	db.mu.RUnlock()

	if exists && !item.IsNeverDie && time.Now().After(item.ExpiresAt) {
		// 触发惰性删除
		db.mu.Lock()
		delete(db.data, key)
		db.mu.Unlock()
		return "", false
	}

	if !exists {
		return "", false
	}
	sv, ok := item.Value.(*types.StringValue)
	if !ok {
		return "", false
	}
	return sv.Value, true
}

// GetItem 获取原始 Item（供命令层判断类型）
func (db *GodisDB) GetItem(key string) (*Item, bool) {
	db.mu.RLock()
	item, exists := db.data[key]
	db.mu.RUnlock()

	if exists && !item.IsNeverDie && time.Now().After(item.ExpiresAt) {
		db.mu.Lock()
		delete(db.data, key)
		db.mu.Unlock()
		return nil, false
	}

	if !exists {
		return nil, false
	}
	return &item, true
}

// putItem 内部通用写入（供各类型命令层调用）
func (db *GodisDB) putItem(key string, item Item) {
	db.mu.Lock()
	db.data[key] = item
	db.mu.Unlock()
}

// Del 删除一个或多个 key，返回实际删除的数量
func (db *GodisDB) Del(keys ...string) int {
	db.mu.Lock()
	deleted := 0
	for _, key := range keys {
		if _, exists := db.data[key]; exists {
			delete(db.data, key)
			deleted++
		}
	}
	db.mu.Unlock()
	return deleted
}

// TypeOf 返回 key 对应的数据类型，不存在返回 -1
func (db *GodisDB) TypeOf(key string) types.DataType {
	db.mu.RLock()
	item, exists := db.data[key]
	db.mu.RUnlock()

	if !exists {
		return -1
	}
	return item.Type
}

// DBStats 数据库统计信息
type DBStats struct {
	Keys    int
	Expires int // 设有过期时间的 key 数量
}

// Stats 返回当前数据库的统计信息
func (db *GodisDB) Stats() DBStats {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var s DBStats
	s.Keys = len(db.data)
	for _, item := range db.data {
		if !item.IsNeverDie {
			s.Expires++
		}
	}
	return s
}

// Keys 返回所有 key（用于 KEYS 命令）
func (db *GodisDB) Keys() []string {
	db.mu.RLock()
	keys := make([]string, 0, len(db.data))
	for k := range db.data {
		keys = append(keys, k)
	}
	db.mu.RUnlock()
	return keys
}

// StartGcWorker 全局 GC，负责在后台定期清理所有数据库中过期的 Key
func StartGcWorker(dbs []*GodisDB) {
	ticker := time.NewTicker(1 * time.Second)
	log.Debug("GC coroutine started successfully")
	go func() {
		for range ticker.C {
			now := time.Now()
			for i, db := range dbs {
				db.mu.Lock()
				for key, item := range db.data {
					if !item.IsNeverDie && now.After(item.ExpiresAt) {
						log.Info("clear expired key [%s] from db %d", key, i)
						delete(db.data, key)
					}
				}
				db.mu.Unlock()
			}
		}
	}()
}
