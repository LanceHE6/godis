package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"godis/protocol"
	"godis/types"
)

func init() {
	// 删除一个或多个 key
	Register("DEL", -2, FlagWrite, 1, -1, 1, handleDel)
	// 检查 key 是否存在
	Register("EXISTS", -2, FlagReadonly, 1, -1, 1, handleExists)
	// 设置 key 的过期时间（秒）
	Register("EXPIRE", 3, FlagWrite, 1, 1, 1, handleExpire)
	// 将 key 移动到另一个数据库
	Register("MOVE", 3, FlagWrite, 1, 1, 1, handleMove)
	// 移除 key 的过期时间，使其永久有效
	Register("PERSIST", 2, FlagWrite, 1, 1, 1, handlePersist)
	// 设置 key 的过期时间（毫秒）
	Register("PEXPIRE", 3, FlagWrite, 1, 1, 1, handlePExpire)
	// 获取 key 的剩余过期时间（毫秒）
	Register("PTTL", 2, FlagReadonly, 1, 1, 1, handlePTTL)
	// 对列表、集合或有序集合的元素排序
	Register("SORT", -2, FlagWrite, 1, 1, 1, handleSort)
	// 更新 key 的最后访问时间
	Register("TOUCH", -2, FlagReadonly, 1, -1, 1, handleTouch)
	// 获取 key 的剩余过期时间（秒）
	Register("TTL", 2, FlagReadonly, 1, 1, 1, handleTTL)
	// 获取 key 存储的数据类型
	Register("TYPE", 2, FlagReadonly, 1, 1, 1, handleType)
	// 异步删除一个或多个 key（功能同 DEL，无异步实现）
	Register("UNLINK", -2, FlagWrite, 1, -1, 1, handleDel)
	// 返回当前数据库 key 的数量
	Register("DBSIZE", 1, FlagFast, 0, 0, 0, handleDBSize)
	// 清空当前数据库的所有 key
	Register("FLUSHDB", 1, FlagWrite, 0, 0, 0, handleFlushDB)
	// 清空所有数据库的所有 key
	Register("FLUSHALL", 1, FlagWrite, 0, 0, 0, handleFlushAll)
	// 增量迭代当前数据库中的 key
	Register("SCAN", -2, FlagReadonly, 0, 0, 0, handleScan)
}

func handleDel(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("del")
	}
	keys := ctx.Args[1:]
	deleted := ctx.DB.Del(keys...)
	return protocol.MakeInt(deleted)
}

func handleExists(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("exists")
	}
	keys := ctx.Args[1:]
	count := ctx.DB.Exists(keys...)
	return protocol.MakeInt(count)
}

func handleExpire(ctx *CommandContext) string {
	if len(ctx.Args) < 3 {
		return protocol.WrongArgsErr("expire")
	}
	key := ctx.Args[1]
	seconds, err := strconv.Atoi(ctx.Args[2])
	if err != nil || seconds <= 0 {
		return protocol.MakeError("ERR invalid expire time")
	}
	if ctx.DB.Expire(key, seconds) {
		return protocol.MakeInt(1)
	}
	return protocol.MakeInt(0)
}

func handleMove(ctx *CommandContext) string {
	if len(ctx.Args) < 3 {
		return protocol.WrongArgsErr("move")
	}
	key := ctx.Args[1]
	dbIdx, err := strconv.Atoi(ctx.Args[2])
	if err != nil || dbIdx < 0 || dbIdx >= len(ctx.AllDBs) {
		return protocol.MakeError("ERR invalid DB index")
	}
	if dbIdx == *ctx.CurrentDBID {
		return protocol.MakeError("ERR source and destination DB are the same")
	}
	if ctx.DB.Move(key, ctx.AllDBs[dbIdx]) {
		return protocol.MakeInt(1)
	}
	return protocol.MakeInt(0)
}

func handlePersist(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("persist")
	}
	key := ctx.Args[1]
	if ctx.DB.Persist(key) {
		return protocol.MakeInt(1)
	}
	return protocol.MakeInt(0)
}

func handlePExpire(ctx *CommandContext) string {
	if len(ctx.Args) < 3 {
		return protocol.WrongArgsErr("pexpire")
	}
	key := ctx.Args[1]
	ms, err := strconv.ParseInt(ctx.Args[2], 10, 64)
	if err != nil || ms <= 0 {
		return protocol.MakeError("ERR invalid expire time")
	}
	if ctx.DB.PExpire(key, ms) {
		return protocol.MakeInt(1)
	}
	return protocol.MakeInt(0)
}

func handlePTTL(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("pttl")
	}
	key := ctx.Args[1]
	return protocol.MakeInt(int(ctx.DB.PTTL(key)))
}

func handleTTL(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("ttl")
	}
	key := ctx.Args[1]
	return protocol.MakeInt(ctx.DB.TTL(key))
}

func handleType(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("type")
	}
	key := ctx.Args[1]
	dt := ctx.DB.TypeOf(key)
	switch dt {
	case types.TypeString:
		return protocol.MakeSimpleString("string")
	case types.TypeHash:
		return protocol.MakeSimpleString("hash")
	case types.TypeList:
		return protocol.MakeSimpleString("list")
	case types.TypeSet:
		return protocol.MakeSimpleString("set")
	case types.TypeZSet:
		return protocol.MakeSimpleString("zset")
	default:
		return protocol.MakeSimpleString("none")
	}
}

func handleTouch(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("touch")
	}
	keys := ctx.Args[1:]
	count := ctx.DB.Touch(keys...)
	return protocol.MakeInt(count)
}

func handleSort(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("sort")
	}
	key := ctx.Args[1]

	// 解析可选参数
	desc := false
	byAlpha := false
	limitOffset := -1
	limitCount := -1

	args := ctx.Args[2:]
	for i := 0; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "DESC":
			desc = true
		case "ASC":
			// 默认
		case "ALPHA":
			byAlpha = true
		case "LIMIT":
			if i+2 < len(args) {
				limitOffset, _ = strconv.Atoi(args[i+1])
				limitCount, _ = strconv.Atoi(args[i+2])
				i += 2
			}
		}
	}

	item, ok := ctx.DB.GetItem(key)
	if !ok {
		return protocol.MakeArray([]string{})
	}

	var elements []string
	switch item.Type {
	case types.TypeList:
		elements = item.Value.(*types.ListValue).Data()
	case types.TypeSet:
		elements = item.Value.(*types.SetValue).MembersList()
	case types.TypeZSet:
		zv := item.Value.(*types.ZSetValue)
		data := zv.Data()
		elements = make([]string, len(data))
		for i, m := range data {
			if byAlpha {
				elements[i] = m.Member
			} else {
				elements[i] = fmt.Sprintf("%f", m.Score)
			}
		}
	default:
		return protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// 排序
	if byAlpha {
		sortStrings(elements, desc)
	} else {
		sortNumeric(elements, desc)
	}

	// LIMIT
	if limitOffset >= 0 && limitCount > 0 {
		if limitOffset >= len(elements) {
			return protocol.MakeArray([]string{})
		}
		end := limitOffset + limitCount
		if end > len(elements) {
			end = len(elements)
		}
		elements = elements[limitOffset:end]
	}

	result := make([]string, len(elements))
	for i, v := range elements {
		result[i] = protocol.MakeBulkString(v)
	}
	return protocol.MakeArray(result)
}

func sortStrings(data []string, desc bool) {
	for i := 1; i < len(data); i++ {
		key := data[i]
		j := i - 1
		for j >= 0 {
			if desc {
				if data[j] < key {
					data[j+1] = data[j]
					j--
				} else {
					break
				}
			} else {
				if data[j] > key {
					data[j+1] = data[j]
					j--
				} else {
					break
				}
			}
		}
		data[j+1] = key
	}
}

func sortNumeric(data []string, desc bool) {
	for i := 1; i < len(data); i++ {
		key := data[i]
		keyVal := parseFloat(key)
		j := i - 1
		for j >= 0 {
			jVal := parseFloat(data[j])
			if desc {
				if jVal < keyVal {
					data[j+1] = data[j]
					j--
				} else {
					break
				}
			} else {
				if jVal > keyVal {
					data[j+1] = data[j]
					j--
				} else {
					break
				}
			}
		}
		data[j+1] = key
	}
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func handleDBSize(ctx *CommandContext) string {
	return protocol.MakeInt(len(ctx.DB.Keys()))
}

func handleFlushDB(ctx *CommandContext) string {
	ctx.DB.Flush()
	return protocol.MakeSimpleString("OK")
}

func handleFlushAll(ctx *CommandContext) string {
	for _, db := range ctx.AllDBs {
		db.Flush()
	}
	return protocol.MakeSimpleString("OK")
}

func handleScan(ctx *CommandContext) string {
	cursor, err := strconv.ParseUint(ctx.Args[1], 10, 64)
	if err != nil {
		return protocol.MakeError("ERR invalid cursor")
	}

	count := 10
	var matchPattern string
	var typeFilter string

	args := ctx.Args[2:]
	for i := 0; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "COUNT":
			if i+1 < len(args) {
				c, err := strconv.Atoi(args[i+1])
				if err == nil && c > 0 {
					count = c
				}
				i++
			}
		case "MATCH":
			if i+1 < len(args) {
				matchPattern = args[i+1]
				i++
			}
		case "TYPE":
			if i+1 < len(args) {
				typeFilter = strings.ToLower(args[i+1])
				i++
			}
		}
	}

	keys := ctx.DB.Keys()
	sort.Strings(keys)

	if cursor >= uint64(len(keys)) {
		return protocol.MakeArray([]string{
			protocol.MakeBulkString("0"),
			protocol.MakeArray([]string{}),
		})
	}

	end := cursor + uint64(count)
	if end > uint64(len(keys)) {
		end = uint64(len(keys))
	}

	result := []string{}
	for _, key := range keys[cursor:end] {
		if matchPattern != "" && !matchGlob(key, matchPattern) {
			continue
		}
		if typeFilter != "" {
			dt := ctx.DB.TypeOf(key)
			typeName := ""
			switch dt {
			case types.TypeString:
				typeName = "string"
			case types.TypeHash:
				typeName = "hash"
			case types.TypeList:
				typeName = "list"
			case types.TypeSet:
				typeName = "set"
			case types.TypeZSet:
				typeName = "zset"
			}
			if typeName != typeFilter {
				continue
			}
		}
		result = append(result, protocol.MakeBulkString(key))
	}

	nextCursor := "0"
	if end < uint64(len(keys)) {
		nextCursor = fmt.Sprintf("%d", end)
	}

	return protocol.MakeArray([]string{
		protocol.MakeBulkString(nextCursor),
		protocol.MakeArray(result),
	})
}

// matchGlob 简单 glob 模式匹配，支持 * 和 ?
func matchGlob(s, pattern string) bool {
	si, pi := 0, 0
	starPi, starSi := -1, -1

	for si < len(s) {
		if pi < len(pattern) && (pattern[pi] == '?' || pattern[pi] == s[si]) {
			si++
			pi++
		} else if pi < len(pattern) && pattern[pi] == '*' {
			starPi = pi
			starSi = si
			pi++
		} else if starPi != -1 {
			pi = starPi + 1
			starSi++
			si = starSi
		} else {
			return false
		}
	}

	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}
	return pi == len(pattern)
}
