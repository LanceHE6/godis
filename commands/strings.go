package commands

import (
	"strconv"
	"strings"

	"godis/protocol"
)

func init() {
	// 注册当前文件里的命令
	CommandRegistry["SET"] = handleSet
	CommandRegistry["GET"] = handleGet
}

func handleSet(ctx *CommandContext) string {
	args := ctx.Args
	if len(args) < 3 {
		return protocol.MakeError("ERR wrong number of arguments for 'set' command")
	}

	key := args[1]
	val := args[2]
	ttl := 0

	if len(args) >= 5 && strings.ToUpper(args[3]) == "EX" {
		seconds, err := strconv.Atoi(args[4])
		if err == nil && seconds > 0 {
			ttl = seconds
		}
	}

	cmdLog.Debug("SET key=%s val=%s ttl=%d", key, val, ttl)
	ctx.DB.Set(key, val, ttl)
	return protocol.MakeSimpleString("OK")
}

func handleGet(ctx *CommandContext) string {
	args := ctx.Args
	if len(args) < 2 {
		return protocol.MakeError("ERR wrong number of arguments for 'get' command")
	}

	key := args[1]
	cmdLog.Debug("GET key=%s", key)
	val, exists := ctx.DB.Get(key)
	if !exists {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(val)
}
