package datastore

import (
	"bytes"
	"fmt"
	"os"
	"sync"
)

type AofLogger struct {
	mu              sync.Mutex
	filename        string
	file            *os.File
	lastRewriteSize int64 // 用于记录上次重写后的文件大小
}

func NewAofLogger(filename string) (*AofLogger, error) {
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
func (aof *AofLogger) WriteCmd(args []string) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	resp := fmt.Sprintf("*%d\r\n", len(args))
	for _, arg := range args {
		resp += fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)
	}

	_, err := aof.file.Write([]byte(resp))
	return err
}

// RewriteToHybrid 触发混合持久化重写
// 它会把当前的内存 db 拍成二进制快照写在文件开头，并保持文件句柄可用
func (aof *AofLogger) RewriteToHybrid(db *GodisDB) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	// 1. 先把原有的 AOF 文件句柄关闭，因为要重写整个文件
	aof.file.Close()

	// 2. 创建一个临时的内存缓冲区，先把二进制 GDB 写进去
	var buf bytes.Buffer

	// 在二进制头部写入一个我们自定义的魔数（Magic Number）标记
	// 用来在启动时区分这个文件是“纯文本AOF”还是“混合AOF”
	buf.Write([]byte("GODIS-HYBRID\n"))

	// GDB 序列化方法
	if err := db.SaveToBinary(&buf); err != nil {
		return fmt.Errorf("failed to serialize GDB: %v", err)
	}

	// 3. 以截断清空（O_TRUNC）的模式重新打开 AOF 文件，写入二进制数据
	file, err := os.OpenFile(aof.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	_, err = file.Write(buf.Bytes())
	if err != nil {
		file.Close()
		return err
	}

	// 4. 将持久化句柄重新切换回原来的文件，后续的客户端命令可以继续在这个文件末尾追加了
	aof.file = file

	fileInfo, _ := file.Stat()
	aof.lastRewriteSize = fileInfo.Size()

	log.Info("GDB snapshot created successfully.")
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
