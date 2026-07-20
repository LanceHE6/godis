package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"godis/datastore"
	"godis/protocol"
	"godis/types"
)

func init() {
	// 删除 hash 中一个或多个 field
	Register("HDEL", -3, FlagWrite, 1, 1, 1, handleHDel)
	// 检查 hash 中指定 field 是否存在
	Register("HEXISTS", 3, FlagReadonly, 1, 1, 1, handleHExists)
	// 获取 hash 中指定 field 的值
	Register("HGET", 3, FlagReadonly, 1, 1, 1, handleHGet)
	// 返回 hash 中所有的 field-value 对
	Register("HGETALL", 2, FlagReadonly, 1, 1, 1, handleHGetAll)
	// 返回 hash 中所有的 field 名称
	Register("HKEYS", 2, FlagReadonly, 1, 1, 1, handleHKeys)
	// 返回 hash 中 field 的数量
	Register("HLEN", 2, FlagReadonly, 1, 1, 1, handleHLen)
	// 获取 hash 中多个 field 的值
	Register("HMGET", -3, FlagReadonly, 1, 1, 1, handleHMGet)
	// 批量设置 hash 中多个 field 的值
	Register("HMSET", -4, FlagWrite, 1, 1, 1, handleHMSet)
	// 增量迭代 hash 中的 field
	Register("HSCAN", -3, FlagReadonly, 1, 1, 1, handleHScan)
	// 设置 hash 中一个或多个 field 的值，返回新增 field 数量
	Register("HSET", -4, FlagWrite, 1, 1, 1, handleHSet)
	// 返回 hash 中指定 field 的值的长度
	Register("HSTRLEN", 3, FlagReadonly, 1, 1, 1, handleHStrLen)
	// 返回 hash 中所有的 value
	Register("HVALS", 2, FlagReadonly, 1, 1, 1, handleHVals)
}

// getHashItem 获取并校验 HashValue 类型，不存在则创建空 hash（仅用于写入命令）
func getHashItem(ctx *CommandContext, key string, createOnMissing bool) (*types.HashValue, bool, string) {
	item, exists := ctx.DB.GetItem(key)
	if !exists {
		if createOnMissing {
			// 写入命令时，key 不存在则创建一个空 hash
			hv := types.NewHashValue()
			ctx.DB.SetItem(key, datastore.Item{
				Type:       types.TypeHash,
				Value:      hv,
				IsNeverDie: true,
			})
			return hv, true, ""
		}
		return nil, false, ""
	}
	if item.Type != types.TypeHash {
		return nil, false, protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	hv, ok := item.Value.(*types.HashValue)
	if !ok {
		return nil, false, protocol.MakeError("ERR hash value corruption")
	}
	return hv, true, ""
}

// getHashItemRead 获取并校验 HashValue 类型，不存在返回 nil
func getHashItemRead(ctx *CommandContext, key string) (*types.HashValue, string) {
	item, exists := ctx.DB.GetItem(key)
	if !exists {
		return nil, ""
	}
	if item.Type != types.TypeHash {
		return nil, protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	hv, ok := item.Value.(*types.HashValue)
	if !ok {
		return nil, protocol.MakeError("ERR hash value corruption")
	}
	return hv, ""
}

// handleHSet 设置 hash 中一个或多个 field 的值
// 命令: HSET key field value [field value ...]
// 返回: 新增的 field 数量
func handleHSet(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	pairs := args[1:]

	if len(pairs)%2 != 0 {
		return protocol.MakeError("ERR wrong number of arguments for HSET")
	}

	hv, _, errStr := getHashItem(ctx, key, true)
	if errStr != "" {
		return errStr
	}

	newCount := 0
	for i := 0; i < len(pairs); i += 2 {
		if !hv.Exists(pairs[i]) {
			newCount++
		}
		hv.Set(pairs[i], pairs[i+1])
	}
	return protocol.MakeInt(newCount)
}

// handleHMSet 批量设置多个 field-value 对（与 HSET 同）
// 命令: HMSET key field value [field value ...]
func handleHMSet(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	pairs := args[1:]

	if len(pairs)%2 != 0 {
		return protocol.MakeError("ERR wrong number of arguments for HMSET")
	}

	hv, _, errStr := getHashItem(ctx, key, true)
	if errStr != "" {
		return errStr
	}

	for i := 0; i < len(pairs); i += 2 {
		hv.Set(pairs[i], pairs[i+1])
	}
	return protocol.MakeSimpleString("OK")
}

// handleHGet 获取 hash 中指定 field 的值
// 命令: HGET key field
func handleHGet(ctx *CommandContext) string {
	key, field := ctx.Args[1], ctx.Args[2]
	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeNull()
	}
	val, ok := hv.Get(field)
	if !ok {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(val)
}

// handleHGetAll 返回 hash 中所有的 field-value 对
// 命令: HGETALL key
// 返回: 扁平数组 [field1, val1, field2, val2, ...]
func handleHGetAll(ctx *CommandContext) string {
	key := ctx.Args[1]
	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeArray([]string{})
	}
	all := hv.GetAll()
	elements := make([]string, 0, len(all)*2)
	// 为了保证测试确定性，按 field 排序
	fields := make([]string, 0, len(all))
	for f := range all {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	for _, f := range fields {
		elements = append(elements, protocol.MakeBulkString(f))
		elements = append(elements, protocol.MakeBulkString(all[f]))
	}
	return protocol.MakeArray(elements)
}

// handleHKeys 返回 hash 中所有的 field 名称
// 命令: HKEYS key
func handleHKeys(ctx *CommandContext) string {
	key := ctx.Args[1]
	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeArray([]string{})
	}
	keys := hv.Keys()
	sort.Strings(keys)
	elements := make([]string, len(keys))
	for i, k := range keys {
		elements[i] = protocol.MakeBulkString(k)
	}
	return protocol.MakeArray(elements)
}

// handleHVals 返回 hash 中所有的 value
// 命令: HVALS key
func handleHVals(ctx *CommandContext) string {
	key := ctx.Args[1]
	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeArray([]string{})
	}
	all := hv.GetAll()
	// 按 field 排序保证确定性
	fields := make([]string, 0, len(all))
	for f := range all {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	elements := make([]string, len(fields))
	for i, f := range fields {
		elements[i] = protocol.MakeBulkString(all[f])
	}
	return protocol.MakeArray(elements)
}

// handleHLen 返回 hash 中 field 的数量
// 命令: HLEN key
func handleHLen(ctx *CommandContext) string {
	key := ctx.Args[1]
	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeInt(0)
	}
	return protocol.MakeInt(hv.Len())
}

// handleHDel 删除 hash 中一个或多个 field
// 命令: HDEL key field [field ...]
// 返回: 实际删除的 field 数量
func handleHDel(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	fields := args[1:]

	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeInt(0)
	}

	deleted := 0
	for _, f := range fields {
		if hv.Del(f) {
			deleted++
		}
	}
	return protocol.MakeInt(deleted)
}

// handleHExists 检查 hash 中指定 field 是否存在
// 命令: HEXISTS key field
func handleHExists(ctx *CommandContext) string {
	key, field := ctx.Args[1], ctx.Args[2]
	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeInt(0)
	}
	if hv.Exists(field) {
		return protocol.MakeInt(1)
	}
	return protocol.MakeInt(0)
}

// handleHMGet 返回 hash 中指定多个 field 的值
// 命令: HMGET key field [field ...]
func handleHMGet(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	fields := args[1:]

	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}

	elements := make([]string, len(fields))
	for i, f := range fields {
		if hv == nil {
			elements[i] = protocol.MakeNull()
		} else {
			val, ok := hv.Get(f)
			if !ok {
				elements[i] = protocol.MakeNull()
			} else {
				elements[i] = protocol.MakeBulkString(val)
			}
		}
	}
	return protocol.MakeArray(elements)
}

// handleHStrLen 返回 hash 中指定 field 的值的长度
// 命令: HSTRLEN key field
func handleHStrLen(ctx *CommandContext) string {
	key, field := ctx.Args[1], ctx.Args[2]
	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeInt(0)
	}
	val, ok := hv.Get(field)
	if !ok {
		return protocol.MakeInt(0)
	}
	return protocol.MakeInt(len(val))
}

// handleHScan 增量迭代 hash 中的 field
// 命令: HSCAN key cursor [MATCH pattern] [COUNT count]
func handleHScan(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	cursor := args[1]

	hv, errStr := getHashItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if hv == nil {
		return protocol.MakeArray([]string{
			protocol.MakeBulkString("0"),
			protocol.MakeArray([]string{}),
		})
	}

	// 解析 MATCH pattern
	matchPattern := ""
	// 解析 COUNT（默认 10）
	count := 10

	for i := 2; i < len(args); i++ {
		upper := strings.ToUpper(args[i])
		if upper == "MATCH" && i+1 < len(args) {
			matchPattern = args[i+1]
			i++
		} else if upper == "COUNT" && i+1 < len(args) {
			n, err := strconv.Atoi(args[i+1])
			if err == nil && n > 0 {
				count = n
			}
			i++
		} else {
			return protocol.MakeError(fmt.Sprintf("ERR syntax error at '%s'", args[i]))
		}
	}

	allFields := hv.Keys()
	sort.Strings(allFields)

	cur, _ := strconv.Atoi(cursor)
	if cur < 0 {
		cur = 0
	}
	if uint64(cur) >= uint64(len(allFields)) {
		return protocol.MakeArray([]string{
			protocol.MakeBulkString("0"),
			protocol.MakeArray([]string{}),
		})
	}

	end := cur + count
	if uint64(end) > uint64(len(allFields)) {
		end = len(allFields)
	}

	result := []string{}
	for _, f := range allFields[cur:end] {
		if matchPattern != "" && !matchGlob(f, matchPattern) {
			continue
		}
		result = append(result, protocol.MakeBulkString(f))
	}

	nextCursor := "0"
	if end < len(allFields) {
		nextCursor = fmt.Sprintf("%d", end)
	}

	return protocol.MakeArray([]string{
		protocol.MakeBulkString(nextCursor),
		protocol.MakeArray(result),
	})
}
