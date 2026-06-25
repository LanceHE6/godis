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
}

func handleSet(ctx *CommandContext) string {
	args := ctx.Args
	if len(args) < 3 {
		return protocol.WrongArgsErr("set")
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
		return protocol.WrongArgsErr("get")
	}

	key := args[1]
	val, exists := ctx.DB.Get(key)
	if !exists {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(val)
}

func handleAppend(ctx *CommandContext) string {
	if len(ctx.Args) < 3 {
		return protocol.WrongArgsErr("append")
	}
	key, suffix := ctx.Args[1], ctx.Args[2]
	newLen, err := ctx.DB.Append(key, suffix)
	if err != nil {
		return protocol.MakeError("ERR " + err.Error())
	}
	return protocol.MakeInt(newLen)
}

// popcountTable 预计算的 256 个值的 popcount
var popcountTable = [256]int{
	0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
	1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
	2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
	3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7,
	4, 5, 5, 6, 5, 6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8,
}

func handleBitCount(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("bitcount")
	}
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
