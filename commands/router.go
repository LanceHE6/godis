package commands

import (
	"fmt"
	"godis/datastore"
	"godis/logger"
	"godis/protocol"
)

var cmdLog = logger.NewModuleLogger("COMMANDS")

type CommandContext struct {
	Args        []string
	DB          *datastore.GodisDB   // 当前正在操作的数据库实例
	AllDBs      []*datastore.GodisDB // 所有数据库列表
	CurrentDBID *int                 //当前绑定的库ID
}

type HandlerFunc func(ctx *CommandContext) string

var CommandRegistry = make(map[string]HandlerFunc)

// GlobalAof 留出指针，方便命令层调用持久化组件
var GlobalAof *datastore.AofLogger

func init() {
	CommandRegistry["COMMAND"] = func(ctx *CommandContext) string {
		cmdLog.Debug("COMMAND received")
		return protocol.MakeArray([]string{})
	}

	// 手动触发混合重写的命令
	CommandRegistry["BGREWRITEAOF"] = func(ctx *CommandContext) string {
		cmdLog.Debug("BGREWRITEAOF received")
		if GlobalAof == nil {
			return protocol.MakeError("ERR AOF logger not initialized")
		}

		// 触发数据二进制写入
		err := GlobalAof.RewriteToHybrid(ctx.DB)
		if err != nil {
			return protocol.MakeError(fmt.Sprintf("ERR Rewrite failed: %v", err))
		}
		return protocol.MakeSimpleString("Background append only file rewriting started (Godis Hybrid Mode)")
	}
}
