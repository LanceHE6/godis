package datastore

import (
	"godis/logger"
	"sync"
	"time"
)

var log = logger.NewModuleLogger("DATASTORE")

// Item 带有过期时间的数据项
type Item struct {
	Value      string
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

// Set 存入数据
func (db *GodisDB) Set(key, val string, ttlSeconds int) {
	item := Item{
		Value:      val,
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

// Get 读取数据（自带惰性删除逻辑）
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

	return item.Value, exists
}

// StartGcWorker 全局 GC，负责在后台定期清理所有数据库中过期的 Key
func StartGcWorker(dbs []*GodisDB) {
	ticker := time.NewTicker(1 * time.Second)
	log.Info("GC coroutine started successfully")
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
