package commands

import (
	"net"
	"strings"

	"godis/datastore"
	"godis/logger"
	"godis/protocol"
)

var cmdLog = logger.NewModuleLogger("COMMANDS")

// Flag 命令标志
type Flag string

const (
	FlagWrite    Flag = "write"    // 修改数据
	FlagReadonly Flag = "readonly" // 只读
	FlagFast     Flag = "fast"     // O(1) / O(log N)，不阻塞
	FlagAdmin    Flag = "admin"    // 管理类命令
)

// CommandContext 命令上下文
type CommandContext struct {
	Args         []string
	DB           *datastore.GodisDB   // 当前正在操作的数据库实例
	AllDBs       []*datastore.GodisDB // 所有数据库列表
	CurrentDBID  *int                 // 当前绑定的库ID
	Aof          *datastore.AofLogger // AOF 持久化实例
	Conn         net.Conn             // 客户端连接（供 Pub/Sub 推送消息）
	PubSubClient any                  // *pubsub.Client（循环引用避免，server 注入）
}

type HandlerFunc func(ctx *CommandContext) string

// Command 命令定义，包含元数据和处理函数
type Command struct {
	Name     string
	Arity    int  // 参数个数，负数表示最少参数（如 -2 表示至少 2 个）
	Flags    Flag // 命令标志
	FirstKey int  // 第一个 key 参数的位置（1-based，0 表示无 key）
	LastKey  int  // 最后一个 key 参数的位置（负值表示到末尾，按 KeyStep 步进）
	KeyStep  int  // key 参数的步长
	Handler  HandlerFunc
}

// Register 注册命令，自动转为大写
//
//	name     - 命令名称，如 "SET"、"GET"
//	arity    - 参数个数（含命令名本身），负数表示最少参数，如 -2 表示至少 2 个
//	flags    - 命令标志：FlagWrite / FlagReadonly / FlagFast / FlagAdmin
//	firstKey - 第一个 key 参数的位置（1-based），0 表示无 key 参数
//	lastKey  - 最后一个 key 参数的位置，负值表示到参数末尾，按 keyStep 步进
//	keyStep  - key 参数之间的步长
//	handler  - 命令处理函数
func Register(name string, arity int, flags Flag, firstKey, lastKey, keyStep int, handler HandlerFunc) {
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
	if reason := validateArity(cmd, len(ctx.Args)); reason != "" {
		return reason, &cmd, true
	}
	cmdLog.Debug("executing command: %s, db=%d, args=%v", cmdName, *ctx.CurrentDBID, ctx.Args)
	reply := cmd.Handler(ctx)
	return reply, &cmd, true
}

// validateArity 校验参数个数，返回错误响应；通过则返回空字符串
func validateArity(cmd Command, argCount int) string {
	a := cmd.Arity
	if a > 0 && argCount != a {
		return protocol.WrongArgsErr(cmd.Name)
	}
	if a < 0 && argCount < -a {
		return protocol.WrongArgsErr(cmd.Name)
	}
	return ""
}
