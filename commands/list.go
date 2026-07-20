package commands

import (
	"fmt"
	"strconv"
	"strings"

	"godis/datastore"
	"godis/protocol"
	"godis/types"
)

func init() {
	// 向列表头部添加一个或多个元素
	Register("LPUSH", -3, FlagWrite, 1, 1, 1, handleLPush)
	// 仅当 key 存在时向列表头部添加元素
	Register("LPUSHX", -3, FlagWrite, 1, 1, 1, handleLPushX)
	// 向列表尾部添加一个或多个元素
	Register("RPUSH", -3, FlagWrite, 1, 1, 1, handleRPush)
	// 仅当 key 存在时向列表尾部添加元素
	Register("RPUSHX", -3, FlagWrite, 1, 1, 1, handleRPushX)
	// 移除并返回列表头部的元素
	Register("LPOP", -2, FlagWrite, 1, 1, 1, handleLPop)
	// 移除并返回列表尾部的元素
	Register("RPOP", -2, FlagWrite, 1, 1, 1, handleRPop)
	// 返回列表的元素数量
	Register("LLEN", 2, FlagReadonly, 1, 1, 1, handleLLen)
	// 获取列表中指定位置的元素
	Register("LINDEX", 3, FlagReadonly, 1, 1, 1, handleLIndex)
	// 在列表中指定元素的前/后插入新元素
	Register("LINSERT", 5, FlagWrite, 1, 1, 1, handleLInsert)
	// 获取列表指定范围内的元素
	Register("LRANGE", 4, FlagReadonly, 1, 1, 1, handleLRange)
	// 移除列表中与给定值匹配的元素
	Register("LREM", 4, FlagWrite, 1, 1, 1, handleLRem)
	// 设置列表中指定位置的元素值
	Register("LSET", 4, FlagWrite, 1, 1, 1, handleLSet)
	// 返回元素在列表中的位置
	Register("LPOS", -3, FlagReadonly, 1, 1, 1, handleLPos)
	// 将元素从源列表移动到目标列表（原子 pop + push）
	Register("LMOVE", 5, FlagWrite, 1, 2, 1, handleLMove)
	// 阻塞版 LMOVE（非阻塞简化实现）
	Register("BLMOVE", 6, FlagWrite, 1, 2, 1, handleBLMove)
	// 阻塞式弹出列表头部元素（非阻塞简化实现）
	Register("BLPOP", -3, FlagWrite, 1, -2, 1, handleBLPop)
	// 阻塞式弹出列表尾部元素（非阻塞简化实现）
	Register("BRPOP", -3, FlagWrite, 1, -2, 1, handleBRPop)
}

// getListItemRead 获取 ListValue 只读
func getListItemRead(ctx *CommandContext, key string) (*types.ListValue, string) {
	item, exists := ctx.DB.GetItem(key)
	if !exists {
		return nil, ""
	}
	if item.Type != types.TypeList {
		return nil, protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	lv, ok := item.Value.(*types.ListValue)
	if !ok {
		return nil, protocol.MakeError("ERR list value corruption")
	}
	return lv, ""
}

// getListItemWrite 获取 ListValue，不存在时创建
func getListItemWrite(ctx *CommandContext, key string) (*types.ListValue, string) {
	item, exists := ctx.DB.GetItem(key)
	if !exists {
		lv := types.NewListValue()
		ctx.DB.SetItem(key, datastore.Item{
			Type:       types.TypeList,
			Value:      lv,
			IsNeverDie: true,
		})
		return lv, ""
	}
	if item.Type != types.TypeList {
		return nil, protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	lv, ok := item.Value.(*types.ListValue)
	if !ok {
		return nil, protocol.MakeError("ERR list value corruption")
	}
	return lv, ""
}

// ---- LPUSH / LPUSHX / RPUSH / RPUSHX ----

func handleLPush(ctx *CommandContext) string {
	lv, errStr := getListItemWrite(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	n := lv.PushLeft(ctx.Args[2:]...)
	return protocol.MakeInt(n)
}

func handleLPushX(ctx *CommandContext) string {
	lv, errStr := getListItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeInt(0)
	}
	n := lv.PushLeft(ctx.Args[2:]...)
	return protocol.MakeInt(n)
}

func handleRPush(ctx *CommandContext) string {
	lv, errStr := getListItemWrite(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	n := lv.PushRight(ctx.Args[2:]...)
	return protocol.MakeInt(n)
}

func handleRPushX(ctx *CommandContext) string {
	lv, errStr := getListItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeInt(0)
	}
	n := lv.PushRight(ctx.Args[2:]...)
	return protocol.MakeInt(n)
}

// ---- LPOP / RPOP ----

func handleLPop(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	lv, errStr := getListItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if lv == nil || lv.Len() == 0 {
		return protocol.MakeNull()
	}

	// 解析 count，Redis 6.2+ 支持
	if len(args) >= 2 {
		count, err := strconv.Atoi(args[1])
		if err != nil || count <= 0 {
			return protocol.MakeError("ERR count must be a positive integer")
		}
		elements := make([]string, count)
		for i := 0; i < count; i++ {
			val, ok := lv.PopLeft()
			if !ok {
				elements = elements[:i]
				break
			}
			elements[i] = protocol.MakeBulkString(val)
		}
		return protocol.MakeArray(elements)
	}

	// 默认 count=1
	val, _ := lv.PopLeft()
	return protocol.MakeBulkString(val)
}

func handleRPop(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	lv, errStr := getListItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if lv == nil || lv.Len() == 0 {
		return protocol.MakeNull()
	}

	if len(args) >= 2 {
		count, err := strconv.Atoi(args[1])
		if err != nil || count <= 0 {
			return protocol.MakeError("ERR count must be a positive integer")
		}
		elements := make([]string, count)
		for i := 0; i < count; i++ {
			val, ok := lv.PopRight()
			if !ok {
				elements = elements[:i]
				break
			}
			elements[i] = protocol.MakeBulkString(val)
		}
		return protocol.MakeArray(elements)
	}

	val, _ := lv.PopRight()
	return protocol.MakeBulkString(val)
}

// ---- LLEN ----

func handleLLen(ctx *CommandContext) string {
	lv, errStr := getListItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeInt(0)
	}
	return protocol.MakeInt(lv.Len())
}

// ---- LINDEX ----

func handleLIndex(ctx *CommandContext) string {
	key := ctx.Args[1]
	idx, err := strconv.Atoi(ctx.Args[2])
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}
	lv, errStr := getListItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeNull()
	}
	val, ok := lv.Index(idx)
	if !ok {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(val)
}

// ---- LINSERT ----

func handleLInsert(ctx *CommandContext) string {
	key := ctx.Args[1]
	where := strings.ToUpper(ctx.Args[2])
	pivot := ctx.Args[3]
	element := ctx.Args[4]

	lv, errStr := getListItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeInt(0)
	}

	var n int
	if where == "BEFORE" {
		n = lv.InsertBefore(pivot, element)
	} else if where == "AFTER" {
		n = lv.InsertAfter(pivot, element)
	} else {
		return protocol.MakeError("ERR syntax error")
	}

	if n < 0 {
		return protocol.MakeInt(-1)
	}
	return protocol.MakeInt(n)
}

// ---- LRANGE ----

func handleLRange(ctx *CommandContext) string {
	start, err := strconv.Atoi(ctx.Args[2])
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}
	stop, err := strconv.Atoi(ctx.Args[3])
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}
	lv, errStr := getListItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeArray([]string{})
	}
	values := lv.Range(start, stop)
	elements := make([]string, len(values))
	for i, v := range values {
		elements[i] = protocol.MakeBulkString(v)
	}
	return protocol.MakeArray(elements)
}

// ---- LREM ----

func handleLRem(ctx *CommandContext) string {
	count, err := strconv.Atoi(ctx.Args[2])
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}
	element := ctx.Args[3]

	lv, errStr := getListItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeInt(0)
	}
	return protocol.MakeInt(lv.Remove(element, count))
}

// ---- LSET ----

func handleLSet(ctx *CommandContext) string {
	idx, err := strconv.Atoi(ctx.Args[2])
	if err != nil {
		return protocol.MakeError("ERR value is not an integer or out of range")
	}
	lv, errStr := getListItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeError("ERR no such key")
	}
	if !lv.Set(idx, ctx.Args[3]) {
		return protocol.MakeError("ERR index out of range")
	}
	return protocol.MakeSimpleString("OK")
}

// ---- LPOS ----

func handleLPos(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	element := args[1]

	lv, errStr := getListItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if lv == nil {
		return protocol.MakeNull()
	}

	rank := 1   // 第几次出现的匹配
	count := 1  // 返回几个结果
	maxlen := 0 // 最大搜索长度

	for i := 2; i < len(args); i++ {
		upper := strings.ToUpper(args[i])
		switch {
		case upper == "RANK" && i+1 < len(args):
			r, err := strconv.Atoi(args[i+1])
			if err == nil {
				rank = r
			}
			i++
		case upper == "COUNT" && i+1 < len(args):
			c, err := strconv.Atoi(args[i+1])
			if err == nil && c > 0 {
				count = c
			}
			i++
		case upper == "MAXLEN" && i+1 < len(args):
			ml, err := strconv.Atoi(args[i+1])
			if err == nil && ml > 0 {
				maxlen = ml
			}
			i++
		default:
			return protocol.MakeError(fmt.Sprintf("ERR syntax error at '%s'", args[i]))
		}
	}

	skip := 0
	if rank > 0 {
		skip = rank - 1
	} else {
		// 负 rank：从尾开始，暂时不支持
		return protocol.MakeError("ERR negative RANK not supported")
	}

	result := []string{}
	pos := 0
	for c := 0; c < count; c++ {
		idx := lv.Find(element, pos, skip)
		if idx < 0 {
			break
		}
		if maxlen > 0 && idx-pos >= maxlen {
			break
		}
		result = append(result, protocol.MakeInt(idx))
		pos = idx + 1
		skip = 0 // 第二次及以后不再跳过
	}

	if len(result) == 0 {
		return protocol.MakeNull()
	}
	if count == 1 && !strings.Contains(strings.ToUpper(strings.Join(args, " ")), "COUNT") {
		// 没有显式 COUNT 时返回单个元素
		return result[0]
	}
	return protocol.MakeArray(result)
}

// ---- LMOVE / BLMOVE ----

func moveListElement(ctx *CommandContext, srcKey, dstKey, srcDir, dstDir string) (string, string) {
	src, errStr := getListItemRead(ctx, srcKey)
	if errStr != "" {
		return "", errStr
	}
	if src == nil || src.Len() == 0 {
		return "", protocol.MakeNull()
	}

	var val string
	var ok bool
	if strings.ToUpper(srcDir) == "LEFT" {
		val, ok = src.PopLeft()
	} else if strings.ToUpper(srcDir) == "RIGHT" {
		val, ok = src.PopRight()
	} else {
		return "", protocol.MakeError("ERR syntax error")
	}
	if !ok {
		return "", protocol.MakeNull()
	}

	dst, _ := getListItemWrite(ctx, dstKey)
	if strings.ToUpper(dstDir) == "LEFT" {
		dst.PushLeft(val)
	} else {
		dst.PushRight(val)
	}

	return val, ""
}

func handleLMove(ctx *CommandContext) string {
	val, errStr := moveListElement(ctx, ctx.Args[1], ctx.Args[2], ctx.Args[3], ctx.Args[4])
	if errStr != "" {
		return errStr
	}
	if val == "" {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(val)
}

func handleBLMove(ctx *CommandContext) string {
	// 简化实现：忽略 timeout，直接 LMOVE
	val, errStr := moveListElement(ctx, ctx.Args[1], ctx.Args[2], ctx.Args[3], ctx.Args[4])
	if errStr != "" {
		return errStr
	}
	if val == "" {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(val)
}

// ---- BLPOP / BRPOP ----

func blockingPop(ctx *CommandContext, isLeft bool) string {
	// 遍历 keys（排除最后一个 timeout 参数）
	keys := ctx.Args[1 : len(ctx.Args)-1]
	if len(keys) == 0 {
		return protocol.MakeNull()
	}

	for _, key := range keys {
		lv, errStr := getListItemRead(ctx, key)
		if errStr != "" {
			return errStr
		}
		if lv != nil && lv.Len() > 0 {
			var val string
			var ok bool
			if isLeft {
				val, ok = lv.PopLeft()
			} else {
				val, ok = lv.PopRight()
			}
			if ok {
				return protocol.MakeArray([]string{
					protocol.MakeBulkString(key),
					protocol.MakeBulkString(val),
				})
			}
		}
	}

	// 简化：忽略 timeout，直接返回 null
	return protocol.MakeNull()
}

func handleBLPop(ctx *CommandContext) string {
	return blockingPop(ctx, true)
}

func handleBRPop(ctx *CommandContext) string {
	return blockingPop(ctx, false)
}
