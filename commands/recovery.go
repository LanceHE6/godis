package commands

import (
	"bufio"
	"godis/datastore"
	"godis/logger"
	"io"
	"os"
	"strings"

	"godis/protocol"
)

var log = logger.NewModuleLogger("RECOVERY")

// ReloadHistoryData 从指定的 AOF 文件中分析格式并恢复历史数据，适配多数据库
func ReloadHistoryData(filename string, dbs []*datastore.GodisDB) {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return // 无历史数据，直接跳过
		}
		log.Error("failed to read AOF recovery file: %v", err)
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// 识别是"纯文本"还是"GDB 混合二进制"
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	if strings.TrimSpace(firstLine) == "GODIS-HYBRID" {
		if err := datastore.LoadAllFromBinary(reader, dbs); err != nil {
			log.Error("failed to load GDB binary snapshot: %v", err)
			return
		}
		log.Debug("GDB header loaded successfully. start reading subsequent incremental AOF commands...")
	} else {
		log.Debug("detected plain text AOF format, replaying all commands...")
		// 如果是纯文本，刚才指针被读走了一行，需要通过 Seek 将文件指针重置回开头位置
		_, _ = file.Seek(0, 0)
		reader = bufio.NewReader(file)
	}

	// 重放后续（或全量）的 RESP 文本追加指令
	// 用 currentDBID 跟踪当前正在操作的数据库，SELECT 命令会修改它
	currentDBID := 0
	count := 0
	for {
		args, err := protocol.ParseRESP(reader)
		if err != nil {
			if err == io.EOF {
				break // 读到文件末尾，说明安全恢复完毕
			}
			log.Error("error parsing incremental AOF text: %v", err)
			break
		}

		if len(args) == 0 {
			continue
		}

		cmdName := strings.ToUpper(args[0])
		// 去命令层的全局注册表里寻找对应的 Handler
		if cmd, exists := CommandRegistry[cmdName]; exists {
			ctx := &CommandContext{
				Args:        args,
				DB:          dbs[currentDBID],
				AllDBs:      dbs,
				CurrentDBID: &currentDBID,
			}
			cmd.Handler(ctx)
			count++
		}
	}
	log.Debug("data recovery completed, replayed %d commands across all databases", count)
}
