package datastore

import (
	"fmt"

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

// Append 在字符串值末尾追加内容，key 不存在时创建，返回新长度
func (db *GodisDB) Append(key, suffix string) (int, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	now := time.Now()
	item, exists := db.data[key]

	if exists {
		// 惰性删除
		if !item.IsNeverDie && now.After(item.ExpiresAt) {
			delete(db.data, key)
			exists = false
		}
	}

	if exists {
		sv, ok := item.Value.(*types.StringValue)
		if !ok {
			return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
		}
		sv.Value += suffix
		item.Value = sv
		db.data[key] = item
		return len(sv.Value), nil
	}

	// key 不存在，创建
	db.data[key] = Item{
		Type:       types.TypeString,
		Value:      types.NewStringValue(suffix),
		IsNeverDie: true,
	}
	return len(suffix), nil
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

// Flush 清空当前数据库的所有 key
func (db *GodisDB) Flush() {
	db.mu.Lock()
	db.data = make(map[string]Item)
	db.mu.Unlock()
}

// Exists 返回存在的 key 数量
func (db *GodisDB) Exists(keys ...string) int {
	db.mu.RLock()
	count := 0
	now := time.Now()
	for _, key := range keys {
		if item, ok := db.data[key]; ok {
			if item.IsNeverDie || now.Before(item.ExpiresAt) {
				count++
			}
		}
	}
	db.mu.RUnlock()
	return count
}

// Expire 设置 key 的过期时间（秒）
func (db *GodisDB) Expire(key string, seconds int) bool {
	db.mu.Lock()
	item, exists := db.data[key]
	if !exists || (!item.IsNeverDie && time.Now().After(item.ExpiresAt)) {
		db.mu.Unlock()
		return false
	}
	item.IsNeverDie = false
	item.ExpiresAt = time.Now().Add(time.Duration(seconds) * time.Second)
	db.data[key] = item
	db.mu.Unlock()
	return true
}

// PExpire 设置 key 的过期时间（毫秒）
func (db *GodisDB) PExpire(key string, milliseconds int64) bool {
	db.mu.Lock()
	item, exists := db.data[key]
	if !exists || (!item.IsNeverDie && time.Now().After(item.ExpiresAt)) {
		db.mu.Unlock()
		return false
	}
	item.IsNeverDie = false
	item.ExpiresAt = time.Now().Add(time.Duration(milliseconds) * time.Millisecond)
	db.data[key] = item
	db.mu.Unlock()
	return true
}

// Persist 移除 key 的过期时间，使其永久有效
func (db *GodisDB) Persist(key string) bool {
	db.mu.Lock()
	item, exists := db.data[key]
	if !exists || item.IsNeverDie {
		db.mu.Unlock()
		return false
	}
	if !item.IsNeverDie && time.Now().After(item.ExpiresAt) {
		delete(db.data, key)
		db.mu.Unlock()
		return false
	}
	item.IsNeverDie = true
	item.ExpiresAt = time.Time{}
	db.data[key] = item
	db.mu.Unlock()
	return true
}

// TTL 返回 key 的剩余生存时间（秒），-1 表示永久，-2 表示不存在
func (db *GodisDB) TTL(key string) int {
	db.mu.RLock()
	item, exists := db.data[key]
	db.mu.RUnlock()

	if !exists {
		return -2
	}
	if item.IsNeverDie {
		return -1
	}
	if time.Now().After(item.ExpiresAt) {
		db.mu.Lock()
		delete(db.data, key)
		db.mu.Unlock()
		return -2
	}
	return int(time.Until(item.ExpiresAt).Seconds())
}

// PTTL 返回 key 的剩余生存时间（毫秒），-1 表示永久，-2 表示不存在
func (db *GodisDB) PTTL(key string) int64 {
	db.mu.RLock()
	item, exists := db.data[key]
	db.mu.RUnlock()

	if !exists {
		return -2
	}
	if item.IsNeverDie {
		return -1
	}
	if time.Now().After(item.ExpiresAt) {
		db.mu.Lock()
		delete(db.data, key)
		db.mu.Unlock()
		return -2
	}
	return time.Until(item.ExpiresAt).Milliseconds()
}

// Move 将 key 移动到目标数据库，成功返回 true
func (db *GodisDB) Move(key string, target *GodisDB) bool {
	db.mu.Lock()
	item, exists := db.data[key]
	if !exists || (!item.IsNeverDie && time.Now().After(item.ExpiresAt)) {
		db.mu.Unlock()
		return false
	}
	delete(db.data, key)
	db.mu.Unlock()

	target.mu.Lock()
	if _, exists := target.data[key]; exists {
		target.mu.Unlock()
		// 目标已存在，回滚
		db.mu.Lock()
		db.data[key] = item
		db.mu.Unlock()
		return false
	}
	target.data[key] = item
	target.mu.Unlock()
	return true
}

// Touch 更新 key 的最后访问时间，返回存在的 key 数量
func (db *GodisDB) Touch(keys ...string) int {
	db.mu.RLock()
	count := 0
	now := time.Now()
	for _, key := range keys {
		if item, ok := db.data[key]; ok {
			if item.IsNeverDie || now.Before(item.ExpiresAt) {
				count++
			}
		}
	}
	db.mu.RUnlock()
	return count
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
