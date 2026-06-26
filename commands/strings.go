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
	// 在字符串值末尾追加内容，若 key 不存在则创建
	Register("APPEND", 3, FlagWrite, 1, 1, 1, handleAppend)
	// 统计字符串值中比特位为1的数量
	Register("BITCOUNT", -2, FlagReadonly, 1, 1, 1, handleBitCount)
	// 将 key 的整数值减 1
	Register("DECR", 2, FlagWrite, 1, 1, 1, handleDecr)
	// 将 key 的整数值减指定值
	Register("DECRBY", 3, FlagWrite, 1, 1, 1, handleDecrBy)
	// 将 key 的整数值加 1
	Register("INCR", 2, FlagWrite, 1, 1, 1, handleIncr)
	// 将 key 的整数值加指定值
	Register("INCRBY", 3, FlagWrite, 1, 1, 1, handleIncrBy)
	// 返回字符串值的子串
	Register("GETRANGE", 4, FlagReadonly, 1, 1, 1, handleGetRange)
	// 设置新值并返回旧值
	Register("GETSET", 3, FlagWrite, 1, 1, 1, handleGetSet)
}

func handleSet(ctx *CommandContext) string {
	args := ctx.Args
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
	key := ctx.Args[1]
	val, exists := ctx.DB.Get(key)
	if !exists {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(val)
}

func handleAppend(ctx *CommandContext) string {
	key, suffix := ctx.Args[1], ctx.Args[2]
	newLen, err := ctx.DB.Append(key, suffix)
	if err != nil {
		return protocol.MakeError("ERR " + err.Error())
	}
	return protocol.MakeInt(newLen)
}

func handleBitCount(ctx *CommandContext) string {
	key := ctx.Args[1]

	val, exists := ctx.DB.Get(key)
	if !exists {
		return protocol.MakeInt(0)
	}

	bytes := []byte(val)
	start, end := 0, len(bytes)-1

	// 解析可选的 start end 参数（字节范围，支持负值）
	if len(ctx.Args) >= 4 {
		var err error
		start, err = strconv.Atoi(ctx.Args[2])
		if err != nil {
			return protocol.MakeError("ERR value is not an integer or out of range")
		}
		end, err = strconv.Atoi(ctx.Args[3])
		if err != nil {
			return protocol.MakeError("ERR value is not an integer or out of range")
		}
	}

	// 处理负索引
	if start < 0 {
		start = len(bytes) + start
	}
	if end < 0 {
		end = len(bytes) + end
	}

	// 范围裁剪
	if start < 0 {
		start = 0
	}
	if end >= len(bytes) {
		end = len(bytes) - 1
	}

	// 范围无效时返回 0
	if start > end || start >= len(bytes) {
		return protocol.MakeInt(0)
	}

	count := 0
	for i := start; i <= end; i++ {
		count += popcountTable[bytes[i]]
	}
	return protocol.MakeInt(count)
}

func handleIncrBy(ctx *CommandContext) string {
	key := ctx.Args[1]
	delta, err := strconv.ParseInt(ctx.Args[2], 10, 64)
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}
	n, err := ctx.DB.IncrBy(key, delta)
	if err != nil {
		return protocol.MakeError("ERR " + err.Error())
	}
	return protocol.MakeInt(int(n))
}

func handleIncr(ctx *CommandContext) string {
	n, err := ctx.DB.IncrBy(ctx.Args[1], 1)
	if err != nil {
		return protocol.MakeError("ERR " + err.Error())
	}
	return protocol.MakeInt(int(n))
}

func handleDecrBy(ctx *CommandContext) string {
	key := ctx.Args[1]
	delta, err := strconv.ParseInt(ctx.Args[2], 10, 64)
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}
	n, err := ctx.DB.IncrBy(key, -delta)
	if err != nil {
		return protocol.MakeError("ERR " + err.Error())
	}
	return protocol.MakeInt(int(n))
}

func handleDecr(ctx *CommandContext) string {
	n, err := ctx.DB.IncrBy(ctx.Args[1], -1)
	if err != nil {
		return protocol.MakeError("ERR " + err.Error())
	}
	return protocol.MakeInt(int(n))
}

func handleGetRange(ctx *CommandContext) string {
	key := ctx.Args[1]
	val, exists := ctx.DB.Get(key)
	if !exists {
		return protocol.MakeBulkString("")
	}

	start, err := strconv.Atoi(ctx.Args[2])
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}
	end, err := strconv.Atoi(ctx.Args[3])
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}

	bytes := []byte(val)
	n := len(bytes)

	// 处理负索引
	if start < 0 {
		start = n + start
	}
	if end < 0 {
		end = n + end
	}

	// 范围裁剪
	if start < 0 {
		start = 0
	}
	if end >= n {
		end = n - 1
	}

	if start > end || start >= n {
		return protocol.MakeBulkString("")
	}

	return protocol.MakeBulkString(string(bytes[start : end+1]))
}

func handleGetSet(ctx *CommandContext) string {
	key, newVal := ctx.Args[1], ctx.Args[2]
	oldVal, exists := ctx.DB.GetSet(key, newVal)
	if !exists {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(oldVal)
}
