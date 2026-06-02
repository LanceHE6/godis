package datastore

import (
	"fmt"
	"sync"
	"time"
)

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
	// GC 协程
	go db.startGcWorker()
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

// startGcWorker GC，负责在后台定期清理过期的 Key
func (db *GodisDB) startGcWorker() {
	ticker := time.NewTicker(1 * time.Second)
	fmt.Println("【🔥 GC 清理】常驻协程启动成功")
	for range ticker.C {
		now := time.Now()
		db.mu.Lock()
		for key, item := range db.data {
			if !item.IsNeverDie && now.After(item.ExpiresAt) {
				fmt.Printf("【🔥 GC 清理】检测到 Key [%s] 已过期，执行内存释放\n", key)
				delete(db.data, key)
			}
		}
		db.mu.Unlock()
	}
}
