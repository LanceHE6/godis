package datastore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AofLogger struct {
	mu              sync.Mutex
	filename        string
	file            *os.File
	lastRewriteSize int64 // 用于记录上次重写后的文件大小
}

func NewAofLogger(filename string) (*AofLogger, error) {
	dir := filepath.Dir(filename)
	// 自动递归创建文件夹
	// os.ModePerm 代表 0777 最高读写权限
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to auto-create aof directory [%s]: %v", dir, err)
	}

	// 以读写、创建、追加模式打开文件
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &AofLogger{
		filename: filename,
		file:     file,
	}, nil
}

// WriteCmd 向文件末尾追加 RESP 文本命令
// dbID 为当前操作的数据库编号，非 0 时自动前插 SELECT 命令
func (aof *AofLogger) WriteCmd(args []string, dbID int) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	var resp string

	// 非 db0 时，先写一条 SELECT 命令标记切换库
	if dbID != 0 {
		resp += fmt.Sprintf("*2\r\n$6\r\nSELECT\r\n$%d\r\n%d\r\n", len(fmt.Sprintf("%d", dbID)), dbID)
	}

	resp += fmt.Sprintf("*%d\r\n", len(args))
	for _, arg := range args {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}

	_, err := aof.file.Write([]byte(resp))
	return err
}

// RewriteToHybrid 触发混合持久化重写
// 它会把所有内存 db 拍成二进制快照写在文件开头，并保持文件句柄可用
func (aof *AofLogger) RewriteToHybrid(dbs []*GodisDB) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// 先把原有的 AOF 文件句柄关闭，因为要重写整个文件
	aof.file.Close()

	// 创建一个临时的内存缓冲区，先把二进制 GDB 写进去
	var buf bytes.Buffer

	// 用来在启动时区分这个文件是"纯文本AOF"还是"混合AOF"
	buf.Write([]byte("GODIS-HYBRID\n"))

	// GDB 序列化方法，保存所有数据库
	if err := SaveAllToBinary(&buf, dbs); err != nil {
		return fmt.Errorf("failed to serialize GDB: %v", err)
	}

	// 以截断清空（O_TRUNC）的模式重新打开 AOF 文件，写入二进制数据
	file, err := os.OpenFile(aof.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	_, err = file.Write(buf.Bytes())
	if err != nil {
		file.Close()
		return err
	}

	// 将持久化句柄重新切换回原来的文件，后续的客户端命令可以继续在这个文件末尾追加了
	aof.file = file

	fileInfo, _ := file.Stat()
	aof.lastRewriteSize = fileInfo.Size()

	log.Debug("GDB snapshot created successfully.")
	return nil
}

// GetLastRewriteSize 查询上一次的大小
func (aof *AofLogger) GetLastRewriteSize() int64 {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	return aof.lastRewriteSize
}

func (aof *AofLogger) Close() {
	if aof.file != nil {
		aof.file.Close()
	}
}

// StartAutoRewriteWorker 启动后台自动重写监控协程，适配多数据库
// filename: 需要监控的 AOF 文件名
// aofLogger: 持久化组件实例
// dbs: 所有数据库实例
func StartAutoRewriteWorker(filename string, aofLogger *AofLogger, dbs []*GodisDB) {
	ticker := time.NewTicker(10 * time.Second)
	log.Debug("AOF coroutine started successfully")

	// 绝对增量上限：只要新写满 64MB 的文本命令，不管比例到没到，必须重写
	const maxAbsoluteGrowthBytes int64 = 64 * 1024 * 1024

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
				if currentSize >= 2*1024 {
					_ = aofLogger.RewriteToHybrid(dbs)
				}
				continue
			}

			// 计算当前的增长量
			growthBytes := currentSize - lastSize

			// 只有当新追加的文本命令体积，达到了硬上限或达到了上一次总大小的 50% 时，才触发重写
			if growthBytes >= maxAbsoluteGrowthBytes || (float64(growthBytes)/float64(lastSize) >= 0.5) {
				log.Debug("AOF has been triggered")
				_ = aofLogger.RewriteToHybrid(dbs)
			}
		}
	}()
}
