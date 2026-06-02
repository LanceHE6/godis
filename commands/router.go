package commands

import (
	"godis/datastore"
)

// CommandContext 传递给每个命令的上下文参数
type CommandContext struct {
	Args []string           // 客户端发来的参数，例如 ["SET", "key", "val"]
	DB   *datastore.GodisDB // 统一操作的数据库实例
}

// HandlerFunc 每个具体命令要实现的函数签名
// 返回值是给客户端的 RESP 回复字符串
type HandlerFunc func(ctx *CommandContext) string

// 全局命令注册表
var CommandRegistry = make(map[string]HandlerFunc)

func init() {
	// 注册通用命令
	CommandRegistry["COMMAND"] = func(ctx *CommandContext) string {
		return "*0\r\n"
	}
}
