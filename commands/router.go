package commands

import (
	"fmt"
	"strings"

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
	CurrentDBID *int                 // 当前绑定的库ID
	Aof         *datastore.AofLogger // AOF 持久化实例
}

type HandlerFunc func(ctx *CommandContext) string

// UnimplementedHandlerFunc 未实现的命令handler
var UnimplementedHandlerFunc HandlerFunc = func(ctx *CommandContext) string {
	return protocol.MakeError(fmt.Sprintf("%s is not supported", ctx.Args[0]))
}

// Command 命令定义，包含元数据和处理函数
type Command struct {
	Name     string
	Arity    int    // 参数个数，负数表示最少参数（如 -2 表示至少 2 个）
	Flags    string // write / readonly / fast / admin
	FirstKey int    // 第一个 key 参数的位置（1-based，0 表示无 key）
	LastKey  int    // 最后一个 key 参数的位置（负值表示到末尾，按 KeyStep 步进）
	KeyStep  int    // key 参数的步长
	Handler  HandlerFunc
}

// Register 注册命令，自动转为大写
//
//	name     - 命令名称，如 "SET"、"GET"
//	arity    - 参数个数（含命令名本身），负数表示最少参数，如 -2 表示至少 2 个
//	flags    - 命令标志：write / readonly / fast / admin
//	firstKey - 第一个 key 参数的位置（1-based），0 表示无 key 参数
//	lastKey  - 最后一个 key 参数的位置，负值表示到参数末尾，按 keyStep 步进
//	keyStep  - key 参数之间的步长
//	handler  - 命令处理函数
func Register(name string, arity int, flags string, firstKey, lastKey, keyStep int, handler HandlerFunc) {
	CommandRegistry[strings.ToUpper(name)] = Command{
		Name:     strings.ToUpper(name),
		Arity:    arity,
		Flags:    flags,
		FirstKey: firstKey,
		LastKey:  lastKey,
		KeyStep:  keyStep,
		Handler:  handler,
	}
}

// CommandRegistry 全局命令注册
var CommandRegistry = make(map[string]Command)

// Execute 查找并执行命令，自动打印调用日志，返回响应和命令元数据
func Execute(cmdName string, ctx *CommandContext) (string, *Command, bool) {
	cmd, exists := CommandRegistry[cmdName]
	if !exists {
		return "", nil, false
	}
	cmdLog.Debug("executing command: %s, db=%d, args=%v", cmdName, *ctx.CurrentDBID, ctx.Args)
	reply := cmd.Handler(ctx)
	return reply, &cmd, true
}
