package commands

import (
	"strconv"
	"strings"

	"godis/protocol"
)

func init() {
	// 设置 key 的值，支持 EX 参数设置过期时间（秒）
	Register("SET", -3, FlagWrite, 1, 1, 1, handleSet)
	// 获取 key 对应的值
	Register("GET", 2, FlagReadonly, 1, 1, 1, handleGet)
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

	ctx.DB.Set(key, val, ttl)
	return protocol.MakeSimpleString("OK")
}

func handleGet(ctx *CommandContext) string {
	args := ctx.Args
	if len(args) < 2 {
		return protocol.MakeError("ERR wrong number of arguments for 'get' command")
	}

	key := args[1]
	val, exists := ctx.DB.Get(key)
	if !exists {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(val)
}
