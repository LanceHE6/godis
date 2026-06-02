package commands

import (
	"fmt"
	"strconv"
	"strings"
)

func init() {
	// 注册当前文件里的命令
	CommandRegistry["PING"] = handlePing
	CommandRegistry["SET"] = handleSet
	CommandRegistry["GET"] = handleGet
}

func handlePing(ctx *CommandContext) string {
	return "+PONG\r\n"
}

func handleSet(ctx *CommandContext) string {
	args := ctx.Args
	if len(args) < 3 {
		return "-ERR wrong number of arguments for 'set' command\r\n"
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
	return "+OK\r\n"
}

func handleGet(ctx *CommandContext) string {
	args := ctx.Args
	if len(args) < 2 {
		return "-ERR wrong number of arguments for 'get' command\r\n"
	}

	key := args[1]
	val, exists := ctx.DB.Get(key)
	if !exists {
		return "$-1\r\n"
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
}
