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
	// 向集合添加一个或多个成员
	Register("SADD", -3, FlagWrite, 1, 1, 1, handleSAdd)
	// 返回集合的成员数
	Register("SCARD", 2, FlagReadonly, 1, 1, 1, handleSCard)
	// 返回多个集合的差集
	Register("SDIFF", -2, FlagReadonly, 1, -1, 1, handleSDiff)
	// 将差集保存到目标集合
	Register("SDIFFSTORE", -3, FlagWrite, 1, 1, 1, handleSDiffStore)
	// 返回多个集合的交集
	Register("SINTER", -2, FlagReadonly, 1, -1, 1, handleSInter)
	// 返回交集的成员数
	Register("SINTERCARD", -3, FlagReadonly, 1, 1, 1, handleSInterCard)
	// 将交集保存到目标集合
	Register("SINTERSTORE", -3, FlagWrite, 1, 1, 1, handleSInterStore)
	// 判断成员是否在集合中
	Register("SISMEMBER", 3, FlagReadonly, 1, 1, 1, handleSIsMember)
	// 返回集合所有成员
	Register("SMEMBERS", 2, FlagReadonly, 1, 1, 1, handleSMembers)
	// 批量判断多个成员是否在集合中
	Register("SMISMEMBER", -3, FlagReadonly, 1, 1, 1, handleSMIsMember)
	// 将成员从源集合移动到目标集合
	Register("SMOVE", 4, FlagWrite, 1, 2, 1, handleSMove)
	// 随机移除并返回集合中的一个或多个成员
	Register("SPOP", -2, FlagWrite, 1, 1, 1, handleSPop)
	// 随机返回集合中的一个或多个成员（不移除）
	Register("SRANDMEMBER", -2, FlagReadonly, 1, 1, 1, handleSRandMember)
	// 移除集合中一个或多个成员
	Register("SREM", -3, FlagWrite, 1, 1, 1, handleSRem)
	// 增量迭代集合中的成员
	Register("SSCAN", -3, FlagReadonly, 1, 1, 1, handleSScan)
	// 返回多个集合的并集
	Register("SUNION", -2, FlagReadonly, 1, -1, 1, handleSUnion)
	// 将并集保存到目标集合
	Register("SUNIONSTORE", -3, FlagWrite, 1, 1, 1, handleSUnionStore)
}

// getSetItemRead 获取 SetValue 只读
func getSetItemRead(ctx *CommandContext, key string) (*types.SetValue, string) {
	item, exists := ctx.DB.GetItem(key)
	if !exists {
		return nil, ""
	}
	if item.Type != types.TypeSet {
		return nil, protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	sv, ok := item.Value.(*types.SetValue)
	if !ok {
		return nil, protocol.MakeError("ERR set value corruption")
	}
	return sv, ""
}

// getSetItemWrite 获取 SetValue，不存在时创建
func getSetItemWrite(ctx *CommandContext, key string) (*types.SetValue, string) {
	item, exists := ctx.DB.GetItem(key)
	if !exists {
		sv := types.NewSetValue()
		ctx.DB.SetItem(key, datastore.Item{
			Type:       types.TypeSet,
			Value:      sv,
			IsNeverDie: true,
		})
		return sv, ""
	}
	if item.Type != types.TypeSet {
		return nil, protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	sv, ok := item.Value.(*types.SetValue)
	if !ok {
		return nil, protocol.MakeError("ERR set value corruption")
	}
	return sv, ""
}

// getSetKeys 读取多个集合，返回 (集合列表, 错误响应, 成功数)
func getSetKeys(ctx *CommandContext, keys []string) ([]*types.SetValue, string) {
	sets := make([]*types.SetValue, len(keys))
	for i, key := range keys {
		sv, errStr := getSetItemRead(ctx, key)
		if errStr != "" {
			return nil, errStr
		}
		sets[i] = sv
	}
	return sets, ""
}

// ---- SADD ----

func handleSAdd(ctx *CommandContext) string {
	sv, errStr := getSetItemWrite(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	return protocol.MakeInt(sv.Add(ctx.Args[2:]...))
}

// ---- SCARD ----

func handleSCard(ctx *CommandContext) string {
	sv, errStr := getSetItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if sv == nil {
		return protocol.MakeInt(0)
	}
	return protocol.MakeInt(sv.Len())
}

// ---- SDIFF ----

func handleSDiff(ctx *CommandContext) string {
	keys := ctx.Args[1:]
	if len(keys) == 0 {
		return protocol.MakeArray([]string{})
	}

	sv, errStr := getSetItemRead(ctx, keys[0])
	if errStr != "" {
		return errStr
	}
	if sv == nil {
		return protocol.MakeArray([]string{})
	}

	// 读取其他集合
	others := make([]*types.SetValue, 0, len(keys)-1)
	for _, k := range keys[1:] {
		o, _ := getSetItemRead(ctx, k)
		others = append(others, o)
	}

	result := sv.Diff(others...)
	sort.Strings(result)
	elements := bulkStrings(result)
	return protocol.MakeArray(elements)
}

// ---- SDIFFSTORE ----

func handleSDiffStore(ctx *CommandContext) string {
	keys := ctx.Args[1:]
	dst := keys[0]
	srcKeys := keys[1:]

	if len(srcKeys) == 0 {
		// 空源集，创建空目标并返回0
		dstSv, _ := getSetItemWrite(ctx, dst)
		dstSv.Mu.Lock()
		dstSv.Members = make(map[string]struct{})
		dstSv.Mu.Unlock()
		return protocol.MakeInt(0)
	}

	sv, _ := getSetItemRead(ctx, srcKeys[0])
	others := make([]*types.SetValue, 0, len(srcKeys)-1)
	for _, k := range srcKeys[1:] {
		o, _ := getSetItemRead(ctx, k)
		others = append(others, o)
	}

	var result []string
	if sv != nil {
		result = sv.Diff(others...)
	}

	dstSv, _ := getSetItemWrite(ctx, dst)
	dstSv.Mu.Lock()
	dstSv.Members = make(map[string]struct{})
	for _, m := range result {
		dstSv.Members[m] = struct{}{}
	}
	dstSv.Mu.Unlock()
	return protocol.MakeInt(len(result))
}

// ---- SINTER ----

func handleSInter(ctx *CommandContext) string {
	keys := ctx.Args[1:]
	if len(keys) == 0 {
		return protocol.MakeArray([]string{})
	}

	sets, errStr := getSetKeys(ctx, keys)
	if errStr != "" {
		return errStr
	}
	for _, s := range sets {
		if s == nil {
			return protocol.MakeArray([]string{})
		}
	}

	others := sets[1:]
	result := sets[0].Intersect(others...)
	sort.Strings(result)
	elements := bulkStrings(result)
	return protocol.MakeArray(elements)
}

// ---- SINTERCARD ----

func handleSInterCard(ctx *CommandContext) string {
	args := ctx.Args[1:]
	if len(args) == 0 {
		return protocol.MakeInt(0)
	}

	// 解析 numkeys，格式: SINTERCARD numkeys key [key ...] [LIMIT limit]
	numKeys, err := strconv.Atoi(args[0])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}

	keys := args[1:]
	if len(keys) < numKeys {
		return protocol.MakeError("ERR wrong number of arguments")
	}
	srcKeys := keys[:numKeys]

	sets, errStr := getSetKeys(ctx, srcKeys)
	if errStr != "" {
		return errStr
	}
	for _, s := range sets {
		if s == nil {
			return protocol.MakeInt(0)
		}
	}

	count := sets[0].IntersectLen(sets[1:]...)
	return protocol.MakeInt(count)
}

// ---- SINTERSTORE ----

func handleSInterStore(ctx *CommandContext) string {
	keys := ctx.Args[1:]
	if len(keys) < 2 {
		return protocol.MakeError("ERR wrong number of arguments")
	}
	dst := keys[0]
	srcKeys := keys[1:]

	sets, errStr := getSetKeys(ctx, srcKeys)
	if errStr != "" {
		return errStr
	}

	var result []string
	allExist := true
	for _, s := range sets {
		if s == nil {
			allExist = false
			break
		}
	}

	if allExist {
		result = sets[0].Intersect(sets[1:]...)
	}

	dstSv, _ := getSetItemWrite(ctx, dst)
	dstSv.Mu.Lock()
	dstSv.Members = make(map[string]struct{})
	for _, m := range result {
		dstSv.Members[m] = struct{}{}
	}
	dstSv.Mu.Unlock()
	return protocol.MakeInt(len(result))
}

// ---- SISMEMBER ----

func handleSIsMember(ctx *CommandContext) string {
	sv, errStr := getSetItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if sv == nil {
		return protocol.MakeInt(0)
	}
	if sv.Exists(ctx.Args[2]) {
		return protocol.MakeInt(1)
	}
	return protocol.MakeInt(0)
}

// ---- SMEMBERS ----

func handleSMembers(ctx *CommandContext) string {
	sv, errStr := getSetItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if sv == nil {
		return protocol.MakeArray([]string{})
	}
	members := sv.MembersList()
	sort.Strings(members)
	elements := bulkStrings(members)
	return protocol.MakeArray(elements)
}

// ---- SMISMEMBER ----

func handleSMIsMember(ctx *CommandContext) string {
	sv, errStr := getSetItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	members := ctx.Args[2:]
	result := make([]string, len(members))
	for i, m := range members {
		if sv != nil && sv.Exists(m) {
			result[i] = protocol.MakeInt(1)
		} else {
			result[i] = protocol.MakeInt(0)
		}
	}
	return protocol.MakeArray(result)
}

// ---- SMOVE ----

func handleSMove(ctx *CommandContext) string {
	srcKey, dstKey, member := ctx.Args[1], ctx.Args[2], ctx.Args[3]

	src, errStr := getSetItemRead(ctx, srcKey)
	if errStr != "" {
		return errStr
	}
	if src == nil || !src.Exists(member) {
		return protocol.MakeInt(0)
	}

	dst, _ := getSetItemWrite(ctx, dstKey)
	src.Remove(member)
	dst.Add(member)
	return protocol.MakeInt(1)
}

// ---- SPOP ----

func handleSPop(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	sv, errStr := getSetItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if sv == nil {
		return protocol.MakeNull()
	}

	count := 1
	if len(args) >= 2 {
		c, err := strconv.Atoi(args[1])
		if err != nil || c < 0 {
			return protocol.MakeError("ERR count must be a non-negative integer")
		}
		if c > 0 {
			count = c
		}
	}

	result := sv.Pop(count)
	if result == nil {
		return protocol.MakeNull()
	}

	if count == 1 && len(args) < 2 {
		return protocol.MakeBulkString(result[0])
	}
	return protocol.MakeArray(bulkStrings(result))
}

// ---- SRANDMEMBER ----

func handleSRandMember(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	sv, errStr := getSetItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if sv == nil {
		return protocol.MakeNull()
	}

	count := 1
	if len(args) >= 2 {
		c, err := strconv.Atoi(args[1])
		if err != nil {
			return protocol.MakeError("ERR value is not an integer or out of range")
		}
		count = c
		if count == 0 {
			return protocol.MakeArray([]string{})
		}
	}

	if count < 0 {
		// 负数时允许重复
		result := sv.RandomMembers(count)
		return protocol.MakeArray(bulkStrings(result))
	}

	// 默认不重复，且 count > len 时不报错，返回所有
	result := sv.RandomMembers(count)
	if len(args) < 2 {
		return protocol.MakeBulkString(result[0])
	}
	return protocol.MakeArray(bulkStrings(result))
}

// ---- SREM ----

func handleSRem(ctx *CommandContext) string {
	sv, errStr := getSetItemRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if sv == nil {
		return protocol.MakeInt(0)
	}
	return protocol.MakeInt(sv.Remove(ctx.Args[2:]...))
}

// ---- SSCAN ----

func handleSScan(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	cursor := args[1]

	sv, errStr := getSetItemRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if sv == nil {
		return protocol.MakeArray([]string{
			protocol.MakeBulkString("0"),
			protocol.MakeArray([]string{}),
		})
	}

	matchPattern := ""
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

	members := sv.MembersList()
	sort.Strings(members)

	cur, _ := strconv.Atoi(cursor)
	if cur < 0 {
		cur = 0
	}
	if cur >= len(members) {
		return protocol.MakeArray([]string{
			protocol.MakeBulkString("0"),
			protocol.MakeArray([]string{}),
		})
	}

	end := cur + count
	if end > len(members) {
		end = len(members)
	}

	result := []string{}
	for _, m := range members[cur:end] {
		if matchPattern != "" && !matchGlob(m, matchPattern) {
			continue
		}
		result = append(result, protocol.MakeBulkString(m))
	}

	nextCursor := "0"
	if end < len(members) {
		nextCursor = fmt.Sprintf("%d", end)
	}

	return protocol.MakeArray([]string{
		protocol.MakeBulkString(nextCursor),
		protocol.MakeArray(result),
	})
}

// ---- SUNION ----

func handleSUnion(ctx *CommandContext) string {
	keys := ctx.Args[1:]
	if len(keys) == 0 {
		return protocol.MakeArray([]string{})
	}

	sv, _ := getSetItemRead(ctx, keys[0])
	others := make([]*types.SetValue, 0, len(keys)-1)
	for _, k := range keys[1:] {
		o, _ := getSetItemRead(ctx, k)
		others = append(others, o)
	}

	var result []string
	if sv != nil {
		result = sv.Union(others...)
	} else {
		// 第一个 key 不存在，取后续集合的并集
		for _, o := range others {
			if o != nil {
				result = o.Union()
				break
			}
		}
	}

	sort.Strings(result)
	return protocol.MakeArray(bulkStrings(result))
}

// ---- SUNIONSTORE ----

func handleSUnionStore(ctx *CommandContext) string {
	keys := ctx.Args[1:]
	dst := keys[0]
	srcKeys := keys[1:]

	sv, _ := getSetItemRead(ctx, srcKeys[0])
	others := make([]*types.SetValue, 0, len(srcKeys)-1)
	for _, k := range srcKeys[1:] {
		o, _ := getSetItemRead(ctx, k)
		others = append(others, o)
	}

	var result []string
	if sv != nil {
		result = sv.Union(others...)
	}

	dstSv, _ := getSetItemWrite(ctx, dst)
	dstSv.Mu.Lock()
	dstSv.Members = make(map[string]struct{})
	for _, m := range result {
		dstSv.Members[m] = struct{}{}
	}
	dstSv.Mu.Unlock()
	return protocol.MakeInt(len(result))
}

// ---- helpers ----

func bulkStrings(vals []string) []string {
	result := make([]string, len(vals))
	for i, v := range vals {
		result[i] = protocol.MakeBulkString(v)
	}
	return result
}
