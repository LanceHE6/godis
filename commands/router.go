package commands

import (
	"fmt"
	"godis/datastore"
	"godis/logger"
	"godis/protocol"
)

var cmdLog = logger.NewModuleLogger("COMMANDS")

// CommandContext 命令上下文
type CommandContext struct {
	Args        []string
	DB          *datastore.GodisDB   // 当前正在操作的数据库实例
	AllDBs      []*datastore.GodisDB // 所有数据库列表
	CurrentDBID *int                 //当前绑定的库ID
	Aof         *datastore.AofLogger // AOF 持久化实例
}

type HandlerFunc func(ctx *CommandContext) string

// UnimplementedHandlerFunc 未实现的命令handler
var UnimplementedHandlerFunc HandlerFunc = func(ctx *CommandContext) string {
	return protocol.MakeError(fmt.Sprintf("%s is not supported", ctx.Args[0]))
}

// CommandRegistry 全局命令注册
var CommandRegistry = make(map[string]HandlerFunc)
