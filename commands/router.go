package commands

import (
	"fmt"
	"godis/datastore"
	"godis/logger"
)

var cmdLog = logger.NewModuleLogger("COMMANDS")

type CommandContext struct {
	Args []string
	DB   *datastore.GodisDB
}

type HandlerFunc func(ctx *CommandContext) string

var CommandRegistry = make(map[string]HandlerFunc)

// GlobalAof 留出指针，方便命令层调用持久化组件
var GlobalAof *datastore.AofLogger

func init() {
	CommandRegistry["COMMAND"] = func(ctx *CommandContext) string {
		cmdLog.Debug("COMMAND received")
		return "*0\r\n"
	}

	// 手动触发混合重写的命令
	CommandRegistry["BGREWRITEAOF"] = func(ctx *CommandContext) string {
		cmdLog.Debug("BGREWRITEAOF received")
		if GlobalAof == nil {
			return "-ERR AOF logger not initialized\r\n"
		}

		// 触发数据二进制写入
		err := GlobalAof.RewriteToHybrid(ctx.DB)
		if err != nil {
			return fmt.Sprintf("-ERR Rewrite failed: %v\r\n", err)
		}
		return "+Background append only file rewriting started (Godis Hybrid Mode)\r\n"
	}
}
