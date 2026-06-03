package commands

import (
	"fmt"
	"godis/datastore"
)

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
		return "*0\r\n"
	}

	// 【新增】：手动触发混合重写的命令
	CommandRegistry["BGREWRITEAOF"] = func(ctx *CommandContext) string {
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
