package datastore

import (
	"godis/logger"
	"os"
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

// StartAutoRewriteWorker 启动后台自动重写监控协程
// filename: 需要监控的 AOF 文件名
// aofLogger: 持久化组件实例
func (db *GodisDB) StartAutoRewriteWorker(filename string, aofLogger *AofLogger) {
	ticker := time.NewTicker(10 * time.Second)
	log.Info("AOF coroutine started successfully")

	// 绝对增量上限：只要新写满 64MB 的文本命令，不管比例到没到，必须重写
	const maxAbsoluteGrowthBytes int64 = 64 * 1024 * 1024 // 1MB

	go func() {
		for range ticker.C {
			fileInfo, err := os.Stat(filename)
			if err != nil {
				continue
			}

			currentSize := fileInfo.Size()
			lastSize := aofLogger.GetLastRewriteSize()

			// 上一次重写大小为 0，给予2k初始线
			if lastSize == 0 {
				if currentSize >= 2*1024 { // 超过 2KB 先触发第一次
					_ = aofLogger.RewriteToHybrid(db)
				}
				continue
			}

			// 计算当前的增长量
			growthBytes := currentSize - lastSize

			// 只有当新追加的文本命令体积，达到了硬上限或达到了上一次总大小的 50% 时，才触发重写
			// 比如上次重写后是 3KB，必须等文件涨到 4.5KB 以上，才会触发下一次瘦身
			if growthBytes >= maxAbsoluteGrowthBytes || (float64(growthBytes)/float64(lastSize) >= 0.5) {
				log.Info("AOF has been triggered")
				_ = aofLogger.RewriteToHybrid(db)
			}
		}
	}()
}
