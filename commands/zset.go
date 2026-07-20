package commands

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"godis/datastore"
	"godis/protocol"
	"godis/types"
)

func init() {
	// 向有序集合添加一个或多个成员及分数
	Register("ZADD", -4, FlagWrite, 1, 1, 1, handleZAdd)
	// 返回有序集合的成员数
	Register("ZCARD", 2, FlagReadonly, 1, 1, 1, handleZCard)
	// 统计分数范围内的成员数
	Register("ZCOUNT", 4, FlagReadonly, 1, 1, 1, handleZCount)
	// 增加成员的分数
	Register("ZINCRBY", 4, FlagWrite, 1, 1, 1, handleZIncrBy)
	// 弹出最高分的一个或多个成员
	Register("ZPOPMAX", -2, FlagWrite, 1, 1, 1, handleZPopMax)
	// 弹出最低分的一个或多个成员
	Register("ZPOPMIN", -2, FlagWrite, 1, 1, 1, handleZPopMin)
	// 随机返回成员
	Register("ZRANDMEMBER", -2, FlagReadonly, 1, 1, 1, handleZRandMember)
	// 返回成员的排名（正序，从0开始）
	Register("ZRANK", 3, FlagReadonly, 1, 1, 1, handleZRank)
	// 移除一个或多个成员
	Register("ZREM", -3, FlagWrite, 1, 1, 1, handleZRem)
	// 按分数范围移除成员
	Register("ZREMRANGEBYSCORE", 4, FlagWrite, 1, 1, 1, handleZRemRangeByScore)
	// 按索引范围移除成员
	Register("ZREMRANGEBYRANK", 4, FlagWrite, 1, 1, 1, handleZRemRangeByRank)
	// 按字典序范围移除成员
	Register("ZREMRANGEBYLEX", 4, FlagWrite, 1, 1, 1, handleZRemRangeByLex)
	// 返回成员的倒序排名（从高到低，0最高）
	Register("ZREVRANK", 3, FlagReadonly, 1, 1, 1, handleZRevRank)
	// 返回成员的分数
	Register("ZSCORE", 3, FlagReadonly, 1, 1, 1, handleZScore)
	// 返回多个成员的分数
	Register("ZMSCORE", -3, FlagReadonly, 1, 1, 1, handleZMScore)
	// 增量迭代有序集合中的成员
	Register("ZSCAN", -3, FlagReadonly, 1, 1, 1, handleZScan)
	// 按字典序返回成员
	Register("ZRANGEBYLEX", -4, FlagReadonly, 1, 1, 1, handleZRangeByLex)
	// 按分数范围返回成员
	Register("ZRANGEBYSCORE", -4, FlagReadonly, 1, 1, 1, handleZRangeByScore)
	// 按索引范围返回成员（支持 WITHSCORES/REV/BYLEX/BYSCORE/LIMIT）
	Register("ZRANGE", -4, FlagReadonly, 1, 1, 1, handleZRange)
	// ZREVRANGE 是 ZRANGE + REV 的别名
	Register("ZREVRANGE", -4, FlagReadonly, 1, 1, 1, handleZRevRange)
	// 保存 ZRANGE 结果到目标集合
	Register("ZRANGESTORE", -5, FlagWrite, 1, 1, 1, handleZRangeStore)
	// 多个有序集合的并集
	Register("ZUNION", -3, FlagReadonly, 1, 1, 1, handleZUnion)
	// 并集保存到目标集合
	Register("ZUNIONSTORE", -4, FlagWrite, 1, 1, 1, handleZUnionStore)
	// 多个有序集合的交集
	Register("ZINTER", -3, FlagReadonly, 1, 1, 1, handleZInter)
	// 交集保存到目标集合
	Register("ZINTERSTORE", -4, FlagWrite, 1, 1, 1, handleZInterStore)
	// 返回交集的成员数
	Register("ZINTERCARD", -3, FlagReadonly, 1, 1, 1, handleZInterCard)
	// 多个有序集合的差集
	Register("ZDIFF", -3, FlagReadonly, 1, 1, 1, handleZDiff)
	// 差集保存到目标集合
	Register("ZDIFFSTORE", -4, FlagWrite, 1, 1, 1, handleZDiffStore)
	// 统计字典序范围内的成员数
	Register("ZLEXCOUNT", 4, FlagReadonly, 1, 1, 1, handleZLexCount)
	// 弹出最高/最低分成员（批量多集合）
	Register("ZMPOP", -4, FlagWrite, 1, -1, 1, handleZMPop)
}

// ---- helpers ----

func getZSetRead(ctx *CommandContext, key string) (*types.ZSetValue, string) {
	item, exists := ctx.DB.GetItem(key)
	if !exists {
		return nil, ""
	}
	if item.Type != types.TypeZSet {
		return nil, protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	zv, ok := item.Value.(*types.ZSetValue)
	if !ok {
		return nil, protocol.MakeError("ERR zset value corruption")
	}
	return zv, ""
}

func getZSetWrite(ctx *CommandContext, key string) (*types.ZSetValue, string) {
	item, exists := ctx.DB.GetItem(key)
	if !exists {
		zv := types.NewZSetValue()
		ctx.DB.SetItem(key, datastore.Item{
			Type:       types.TypeZSet,
			Value:      zv,
			IsNeverDie: true,
		})
		return zv, ""
	}
	if item.Type != types.TypeZSet {
		return nil, protocol.MakeError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}
	zv, ok := item.Value.(*types.ZSetValue)
	if !ok {
		return nil, protocol.MakeError("ERR zset value corruption")
	}
	return zv, ""
}

func membersToArray(members []types.ZSetMember) []string {
	result := make([]string, 0, len(members)*2)
	for _, m := range members {
		result = append(result, protocol.MakeBulkString(m.Member))
		result = append(result, protocol.MakeBulkString(formatFloat(m.Score)))
	}
	return result
}

func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// ---- ZADD ----

func handleZAdd(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]

	// 默认不支持 NX/XX/GT/LT/CH 等高级选项
	args = args[1:] // 去掉 key

	zv, errStr := getZSetWrite(ctx, key)
	if errStr != "" {
		return errStr
	}

	added := 0
	for i := 0; i+1 < len(args); i += 2 {
		score, err := strconv.ParseFloat(args[i], 64)
		if err != nil {
			return protocol.MakeError("ERR value is not a valid float")
		}
		added += zv.Add(score, args[i+1])
	}
	return protocol.MakeInt(added)
}

// ---- ZCARD ----

func handleZCard(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeInt(0)
	}
	return protocol.MakeInt(zv.Len())
}

// ---- ZCOUNT ----

func handleZCount(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeInt(0)
	}
	minB, err := types.ParseScoreBound(ctx.Args[2])
	if err != nil {
		return protocol.MakeError("ERR min or max is not a float")
	}
	maxB, err := types.ParseScoreBound(ctx.Args[3])
	if err != nil {
		return protocol.MakeError("ERR min or max is not a float")
	}
	minV := minB.Value
	if minB.Exclusive {
		minV = math.Nextafter(minV, math.Inf(1))
	}
	maxV := maxB.Value
	if maxB.Exclusive {
		maxV = math.Nextafter(maxV, math.Inf(-1))
	}
	return protocol.MakeInt(zv.CountByScore(minV, maxV))
}

// ---- ZINCRBY ----

func handleZIncrBy(ctx *CommandContext) string {
	incr, err := strconv.ParseFloat(ctx.Args[2], 64)
	if err != nil {
		return protocol.MakeError("ERR value is not a valid float")
	}
	zv, errStr := getZSetWrite(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	newScore := zv.IncrBy(incr, ctx.Args[3])
	return protocol.MakeBulkString(formatFloat(newScore))
}

// ---- ZPOPMAX / ZPOPMIN ----

func handleZPopMax(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeArray([]string{})
	}
	count := 1
	if len(ctx.Args) >= 3 {
		count, _ = strconv.Atoi(ctx.Args[2])
	}
	result := zv.PopMax(count)
	if count == 1 && len(ctx.Args) < 3 {
		return protocol.MakeArray(membersToArray(result))
	}
	return protocol.MakeArray(membersToArray(result))
}

func handleZPopMin(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeArray([]string{})
	}
	count := 1
	if len(ctx.Args) >= 3 {
		count, _ = strconv.Atoi(ctx.Args[2])
	}
	result := zv.PopMin(count)
	return protocol.MakeArray(membersToArray(result))
}

// ---- ZRANDMEMBER ----

func handleZRandMember(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeNull()
	}
	count := 1
	withScores := false
	args := ctx.Args[2:]
	for i := 0; i < len(args); i++ {
		if strings.ToUpper(args[i]) == "WITHSCORES" {
			withScores = true
		} else {
			count, _ = strconv.Atoi(args[i])
		}
	}
	result := zv.RandomMembers(count)
	if result == nil {
		return protocol.MakeNull()
	}
	if len(ctx.Args) < 3 {
		return protocol.MakeBulkString(result[0].Member)
	}
	if withScores {
		return protocol.MakeArray(membersToArray(result))
	}
	elements := make([]string, len(result))
	for i, m := range result {
		elements[i] = protocol.MakeBulkString(m.Member)
	}
	return protocol.MakeArray(elements)
}

// ---- ZRANK / ZREVRANK ----

func handleZRank(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeNull()
	}
	r := zv.Rank(ctx.Args[2])
	if r < 0 {
		return protocol.MakeNull()
	}
	return protocol.MakeInt(r)
}

func handleZRevRank(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeNull()
	}
	r := zv.RevRank(ctx.Args[2])
	if r < 0 {
		return protocol.MakeNull()
	}
	return protocol.MakeInt(r)
}

// ---- ZREM ----

func handleZRem(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeInt(0)
	}
	removed := 0
	for _, m := range ctx.Args[2:] {
		if zv.Remove(m) {
			removed++
		}
	}
	return protocol.MakeInt(removed)
}

// ---- ZREMRANGEBYSCORE ----

func handleZRemRangeByScore(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeInt(0)
	}
	minB, _ := types.ParseScoreBound(ctx.Args[2])
	maxB, _ := types.ParseScoreBound(ctx.Args[3])
	minV := minB.Value
	if minB.Exclusive {
		minV = math.Nextafter(minV, math.Inf(1))
	}
	maxV := maxB.Value
	if maxB.Exclusive {
		maxV = math.Nextafter(maxV, math.Inf(-1))
	}
	return protocol.MakeInt(zv.RemoveByScore(minV, maxV))
}

// ---- ZREMRANGEBYRANK ----

func handleZRemRangeByRank(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeInt(0)
	}
	start, _ := strconv.Atoi(ctx.Args[2])
	stop, _ := strconv.Atoi(ctx.Args[3])
	return protocol.MakeInt(zv.RemoveByRank(start, stop))
}

// ---- ZREMRANGEBYLEX ----

func handleZRemRangeByLex(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeInt(0)
	}
	minStr, minEx := parseLexBound(ctx.Args[2])
	maxStr, maxEx := parseLexBound(ctx.Args[3])
	return protocol.MakeInt(zv.RemoveByLex(minStr, maxStr, minEx, maxEx))
}

// parseLexBound 解析 "-" / "+" / "[abc" / "(abc" 等
func parseLexBound(s string) (string, bool) {
	if s == "-" {
		return "\x00", false
	}
	if s == "+" {
		return "\xff", false
	}
	if strings.HasPrefix(s, "[") {
		return s[1:], false
	}
	if strings.HasPrefix(s, "(") {
		return s[1:], true
	}
	return s, false
}

// ---- ZSCORE / ZMSCORE ----

func handleZScore(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeNull()
	}
	s, ok := zv.Score(ctx.Args[2])
	if !ok {
		return protocol.MakeNull()
	}
	return protocol.MakeBulkString(formatFloat(s))
}

func handleZMScore(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	result := make([]string, len(ctx.Args)-2)
	for i := 2; i < len(ctx.Args); i++ {
		if zv == nil {
			result[i-2] = protocol.MakeNull()
			continue
		}
		s, ok := zv.Score(ctx.Args[i])
		if ok {
			result[i-2] = protocol.MakeBulkString(formatFloat(s))
		} else {
			result[i-2] = protocol.MakeNull()
		}
	}
	return protocol.MakeArray(result)
}

// ---- ZSCAN ----

func handleZScan(ctx *CommandContext) string {
	args := ctx.Args[1:]
	key := args[0]
	cursor := args[1]

	zv, errStr := getZSetRead(ctx, key)
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeArray([]string{protocol.MakeBulkString("0"), protocol.MakeArray([]string{})})
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
		}
	}

	members := zv.Members()
	cur, _ := strconv.Atoi(cursor)
	if cur < 0 {
		cur = 0
	}
	if cur >= len(members) {
		return protocol.MakeArray([]string{protocol.MakeBulkString("0"), protocol.MakeArray([]string{})})
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
	return protocol.MakeArray([]string{protocol.MakeBulkString(nextCursor), protocol.MakeArray(result)})
}

// ---- ZRANGEBYLEX ----

func handleZRangeByLex(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeArray([]string{})
	}
	minStr, minEx := parseLexBound(ctx.Args[2])
	maxStr, maxEx := parseLexBound(ctx.Args[3])
	offset, count := 0, 0
	args := ctx.Args[4:]
	for i := 0; i < len(args); i++ {
		if strings.ToUpper(args[i]) == "LIMIT" && i+2 < len(args) {
			offset, _ = strconv.Atoi(args[i+1])
			count, _ = strconv.Atoi(args[i+2])
			break
		}
	}
	members := zv.RangeByLex(minStr, maxStr, minEx, maxEx, offset, count)
	result := make([]string, len(members))
	for i, m := range members {
		result[i] = protocol.MakeBulkString(m)
	}
	return protocol.MakeArray(result)
}

// ---- ZRANGEBYSCORE ----

func handleZRangeByScore(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeArray([]string{})
	}
	minB, _ := types.ParseScoreBound(ctx.Args[2])
	maxB, _ := types.ParseScoreBound(ctx.Args[3])
	minV := minB.Value
	if minB.Exclusive {
		minV = math.Nextafter(minV, math.Inf(1))
	}
	maxV := maxB.Value
	if maxB.Exclusive {
		maxV = math.Nextafter(maxV, math.Inf(-1))
	}

	withScores := false
	offset, limit := 0, 0
	for i := 4; i < len(ctx.Args); i++ {
		upper := strings.ToUpper(ctx.Args[i])
		if upper == "WITHSCORES" {
			withScores = true
		} else if upper == "LIMIT" && i+2 < len(ctx.Args) {
			offset, _ = strconv.Atoi(ctx.Args[i+1])
			limit, _ = strconv.Atoi(ctx.Args[i+2])
			i += 2
		}
	}

	result := zv.RangeByScore(minV, maxV, offset, limit)
	if withScores {
		return protocol.MakeArray(membersToArray(result))
	}
	elements := make([]string, len(result))
	for i, m := range result {
		elements[i] = protocol.MakeBulkString(m.Member)
	}
	return protocol.MakeArray(elements)
}

// ---- ZRANGE (unified Redis 6.2+) ----

func handleZRange(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeArray([]string{})
	}

	start, _ := strconv.Atoi(ctx.Args[2])
	stop, _ := strconv.Atoi(ctx.Args[3])

	byScore, byLex, rev, withScores := false, false, false, false
	offset, limit := 0, 0

	for i := 4; i < len(ctx.Args); i++ {
		upper := strings.ToUpper(ctx.Args[i])
		switch {
		case upper == "BYSCORE":
			byScore = true
		case upper == "BYLEX":
			byLex = true
		case upper == "REV":
			rev = true
		case upper == "WITHSCORES":
			withScores = true
		case upper == "LIMIT" && i+2 < len(ctx.Args):
			offset, _ = strconv.Atoi(ctx.Args[i+1])
			limit, _ = strconv.Atoi(ctx.Args[i+2])
			i += 2
		}
	}

	if byScore || byLex {
		// BYSCORE/BYLEX: start/stop 是字符串
		if byScore {
			minB, _ := types.ParseScoreBound(ctx.Args[2])
			maxB, _ := types.ParseScoreBound(ctx.Args[3])
			minV := minB.Value
			if minB.Exclusive {
				minV = math.Nextafter(minV, math.Inf(1))
			}
			maxV := maxB.Value
			if maxB.Exclusive {
				maxV = math.Nextafter(maxV, math.Inf(-1))
			}
			result := zv.RangeByScore(minV, maxV, offset, limit)
			if rev {
				reverseMembers(result)
			}
			if withScores {
				return protocol.MakeArray(membersToArray(result))
			}
			return protocol.MakeArray(bulkMembers(result))
		}
		// BYLEX
		minStr, minEx := parseLexBound(ctx.Args[2])
		maxStr, maxEx := parseLexBound(ctx.Args[3])
		members := zv.RangeByLex(minStr, maxStr, minEx, maxEx, offset, limit)
		if rev {
			reverseStrings(members)
		}
		result := make([]string, len(members))
		for i, m := range members {
			result[i] = protocol.MakeBulkString(m)
		}
		return protocol.MakeArray(result)
	}

	// Default: rank range
	var result []types.ZSetMember
	if rev {
		result = zv.RangeRev(start, stop)
	} else {
		result = zv.Range(start, stop)
	}

	if withScores {
		return protocol.MakeArray(membersToArray(result))
	}
	return protocol.MakeArray(bulkMembers(result))
}

func bulkMembers(members []types.ZSetMember) []string {
	result := make([]string, len(members))
	for i, m := range members {
		result[i] = protocol.MakeBulkString(m.Member)
	}
	return result
}

func reverseMembers(a []types.ZSetMember) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func reverseStrings(a []string) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func handleZRevRange(ctx *CommandContext) string {
	// 插入 REV 标志：ZREVRANGE key start stop 等同于 ZRANGE key start stop REV
	args := make([]string, 0, len(ctx.Args)+1)
	args = append(args, ctx.Args[0], ctx.Args[1], ctx.Args[2], ctx.Args[3], "REV")
	args = append(args, ctx.Args[4:]...)
	newCtx := &CommandContext{
		Args:        args,
		DB:          ctx.DB,
		AllDBs:      ctx.AllDBs,
		CurrentDBID: ctx.CurrentDBID,
		Aof:         ctx.Aof,
	}
	return handleZRange(newCtx)
}

// ---- ZRANGESTORE ----

func handleZRangeStore(ctx *CommandContext) string {
	dstKey := ctx.Args[1]
	// 直接调用 ZRANGE 获取结果
	zv, errStr := getZSetRead(ctx, ctx.Args[2])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		dst, _ := getZSetWrite(ctx, dstKey)
		dst.Mu.Lock()
		dst.Scores = make(map[string]float64)
		dst.Dirty = true
		dst.Mu.Unlock()
		return protocol.MakeInt(0)
	}

	start, _ := strconv.Atoi(ctx.Args[3])
	stop, _ := strconv.Atoi(ctx.Args[4])

	var result []types.ZSetMember
	rev := false
	for i := 5; i < len(ctx.Args); i++ {
		if strings.ToUpper(ctx.Args[i]) == "REV" {
			rev = true
		}
	}
	if rev {
		result = zv.RangeRev(start, stop)
	} else {
		result = zv.Range(start, stop)
	}

	dst, _ := getZSetWrite(ctx, dstKey)
	dst.Mu.Lock()
	dst.Scores = make(map[string]float64)
	for _, m := range result {
		dst.Scores[m.Member] = m.Score
	}
	dst.Dirty = true
	dst.Mu.Unlock()
	return protocol.MakeInt(len(result))
}

// ---- ZUNION / ZUNIONSTORE / ZINTER / ZINTERSTORE / ZINTERCARD / ZDIFF / ZDIFFSTORE ----

type zAggOpts struct {
	keys       []string
	weights    []float64
	aggregate  string // SUM, MIN, MAX
	withScores bool
}

func parseZOpts(args []string, numKeys int) (*zAggOpts, string) {
	opts := &zAggOpts{
		keys:      args[1 : 1+numKeys],
		aggregate: "SUM",
	}
	opts.weights = make([]float64, numKeys)
	for i := range opts.weights {
		opts.weights[i] = 1.0
	}

	rest := args[1+numKeys:]
	for i := 0; i < len(rest); i++ {
		upper := strings.ToUpper(rest[i])
		switch {
		case upper == "WEIGHTS":
			for j := 0; j < numKeys && i+1+j < len(rest); j++ {
				w, err := strconv.ParseFloat(rest[i+1+j], 64)
				if err == nil {
					opts.weights[j] = w
				}
			}
			i += numKeys
		case upper == "AGGREGATE" && i+1 < len(rest):
			opts.aggregate = strings.ToUpper(rest[i+1])
			i++
		case upper == "WITHSCORES":
			opts.withScores = true
		default:
			return nil, protocol.MakeError(fmt.Sprintf("ERR syntax error at '%s'", rest[i]))
		}
	}
	return opts, ""
}

func computeAgg(zvs []*types.ZSetValue, opts *zAggOpts) []types.ZSetMember {
	allMembers := make(map[string]struct{})
	for _, z := range zvs {
		if z == nil {
			continue
		}
		z.Mu.RLock()
		for m := range z.Scores {
			allMembers[m] = struct{}{}
		}
		z.Mu.RUnlock()
	}

	result := make([]types.ZSetMember, 0, len(allMembers))
	for m := range allMembers {
		score := computeScore(m, zvs, opts)
		result = append(result, types.ZSetMember{Member: m, Score: score})
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Member < result[j].Member
		}
		return result[i].Score < result[j].Score
	})
	return result
}

func computeScore(member string, zvs []*types.ZSetValue, opts *zAggOpts) float64 {
	switch opts.aggregate {
	case "MIN":
		result := math.Inf(1)
		hasVal := false
		for i, z := range zvs {
			if z == nil {
				continue
			}
			s, ok := z.Score(member)
			if ok {
				hasVal = true
				ws := s * opts.weights[i]
				if ws < result {
					result = ws
				}
			}
		}
		if !hasVal {
			return 0
		}
		return result
	case "MAX":
		result := math.Inf(-1)
		hasVal := false
		for i, z := range zvs {
			if z == nil {
				continue
			}
			s, ok := z.Score(member)
			if ok {
				hasVal = true
				ws := s * opts.weights[i]
				if ws > result {
					result = ws
				}
			}
		}
		if !hasVal {
			return 0
		}
		return result
	default: // SUM
		sum := 0.0
		for i, z := range zvs {
			if z == nil {
				continue
			}
			s, ok := z.Score(member)
			if ok {
				sum += s * opts.weights[i]
			}
		}
		return sum
	}
}

func handleZUnion(ctx *CommandContext) string {
	numKeys, err := strconv.Atoi(ctx.Args[1])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}
	opts, errStr := parseZOpts(ctx.Args[1:], numKeys)
	if errStr != "" {
		return errStr
	}
	zvs := make([]*types.ZSetValue, len(opts.keys))
	for i, k := range opts.keys {
		zvs[i], _ = getZSetRead(ctx, k)
	}
	result := computeAgg(zvs, opts)
	if opts.withScores {
		return protocol.MakeArray(membersToArray(result))
	}
	return protocol.MakeArray(bulkMembers(result))
}

func handleZUnionStore(ctx *CommandContext) string {
	dstKey := ctx.Args[1]
	numKeys, err := strconv.Atoi(ctx.Args[2])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}
	opts, errStr := parseZOpts(ctx.Args[2:], numKeys)
	if errStr != "" {
		return errStr
	}
	zvs := make([]*types.ZSetValue, len(opts.keys))
	for i, k := range opts.keys {
		zvs[i], _ = getZSetRead(ctx, k)
	}
	result := computeAgg(zvs, opts)
	dst, _ := getZSetWrite(ctx, dstKey)
	dst.Mu.Lock()
	dst.Scores = make(map[string]float64)
	for _, m := range result {
		dst.Scores[m.Member] = m.Score
	}
	dst.Dirty = true
	dst.Mu.Unlock()
	return protocol.MakeInt(len(result))
}

func handleZInter(ctx *CommandContext) string {
	numKeys, err := strconv.Atoi(ctx.Args[1])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}
	opts, errStr := parseZOpts(ctx.Args[1:], numKeys)
	if errStr != "" {
		return errStr
	}
	zvs := make([]*types.ZSetValue, len(opts.keys))
	for i, k := range opts.keys {
		zvs[i], _ = getZSetRead(ctx, k)
		if zvs[i] == nil {
			return protocol.MakeArray([]string{})
		}
	}
	result := computeAggWithIntersection(zvs, opts)
	if opts.withScores {
		return protocol.MakeArray(membersToArray(result))
	}
	return protocol.MakeArray(bulkMembers(result))
}

func computeAggWithIntersection(zvs []*types.ZSetValue, opts *zAggOpts) []types.ZSetMember {
	first := zvs[0]
	first.Mu.RLock()
	members := make([]string, 0, len(first.Scores))
	for m := range first.Scores {
		// check all others contain m
		all := true
		for _, z := range zvs[1:] {
			if z == nil || !z.Exists(m) {
				all = false
				break
			}
		}
		if all {
			members = append(members, m)
		}
	}
	first.Mu.RUnlock()

	result := make([]types.ZSetMember, len(members))
	for i, m := range members {
		result[i] = types.ZSetMember{Member: m, Score: computeScore(m, zvs, opts)}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Member < result[j].Member
		}
		return result[i].Score < result[j].Score
	})
	return result
}

func handleZInterStore(ctx *CommandContext) string {
	dstKey := ctx.Args[1]
	numKeys, err := strconv.Atoi(ctx.Args[2])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}
	opts, errStr := parseZOpts(ctx.Args[2:], numKeys)
	if errStr != "" {
		return errStr
	}
	zvs := make([]*types.ZSetValue, len(opts.keys))
	allExist := true
	for i, k := range opts.keys {
		zvs[i], _ = getZSetRead(ctx, k)
		if zvs[i] == nil {
			allExist = false
		}
	}
	var result []types.ZSetMember
	if allExist {
		result = computeAggWithIntersection(zvs, opts)
	}
	dst, _ := getZSetWrite(ctx, dstKey)
	dst.Mu.Lock()
	dst.Scores = make(map[string]float64)
	for _, m := range result {
		dst.Scores[m.Member] = m.Score
	}
	dst.Dirty = true
	dst.Mu.Unlock()
	return protocol.MakeInt(len(result))
}

func handleZInterCard(ctx *CommandContext) string {
	numKeys, err := strconv.Atoi(ctx.Args[1])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}
	// 格式: ZINTERCARD numkeys key [key ...] [LIMIT limit]
	keys := ctx.Args[2:]
	if len(keys) < numKeys {
		return protocol.MakeError("ERR wrong number of arguments")
	}
	count := 0
	limit := 0

	// Parse optional LIMIT
	rem := keys[numKeys:]
	for i := 0; i < len(rem); i++ {
		if strings.ToUpper(rem[i]) == "LIMIT" && i+1 < len(rem) {
			limit, _ = strconv.Atoi(rem[i+1])
			break
		}
	}

	zvs := make([]*types.ZSetValue, numKeys)
	for i, k := range keys[:numKeys] {
		zvs[i], _ = getZSetRead(ctx, k)
		if zvs[i] == nil {
			return protocol.MakeInt(0)
		}
	}

	first := zvs[0]
	first.Mu.RLock()
	for m := range first.Scores {
		all := true
		for _, z := range zvs[1:] {
			if !z.Exists(m) {
				all = false
				break
			}
		}
		if all {
			count++
			if limit > 0 && count >= limit {
				break
			}
		}
	}
	first.Mu.RUnlock()
	return protocol.MakeInt(count)
}

func handleZDiff(ctx *CommandContext) string {
	numKeys, err := strconv.Atoi(ctx.Args[1])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}
	rest := ctx.Args[2:]
	if len(rest) < numKeys {
		return protocol.MakeArray([]string{})
	}
	keys := rest[:numKeys]
	withScores := false
	if len(rest) > numKeys && strings.ToUpper(rest[numKeys]) == "WITHSCORES" {
		withScores = true
	}

	zvs := make([]*types.ZSetValue, len(keys))
	for i, k := range keys {
		zvs[i], _ = getZSetRead(ctx, k)
		if i > 0 && zvs[i] == nil {
			zvs[i] = types.NewZSetValue()
		}
	}
	if zvs[0] == nil {
		return protocol.MakeArray([]string{})
	}

	result := computeDiff(zvs, 1.0)
	if withScores {
		return protocol.MakeArray(membersToArray(result))
	}
	return protocol.MakeArray(bulkMembers(result))
}

func computeDiff(zvs []*types.ZSetValue, weight float64) []types.ZSetMember {
	first := zvs[0]
	first.Mu.RLock()
	result := make([]types.ZSetMember, 0)
	for m := range first.Scores {
		inOthers := false
		for _, z := range zvs[1:] {
			if z.Exists(m) {
				inOthers = true
				break
			}
		}
		if !inOthers {
			result = append(result, types.ZSetMember{Member: m, Score: first.Scores[m] * weight})
		}
	}
	first.Mu.RUnlock()
	sort.Slice(result, func(i, j int) bool {
		if result[i].Score == result[j].Score {
			return result[i].Member < result[j].Member
		}
		return result[i].Score < result[j].Score
	})
	return result
}

func handleZDiffStore(ctx *CommandContext) string {
	dstKey := ctx.Args[1]
	numKeys, err := strconv.Atoi(ctx.Args[2])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}
	keys := ctx.Args[3:]
	if len(keys) < numKeys {
		return protocol.MakeError("ERR wrong number of arguments")
	}

	zvs := make([]*types.ZSetValue, numKeys)
	for i, k := range keys[:numKeys] {
		zvs[i], _ = getZSetRead(ctx, k)
		if i > 0 && zvs[i] == nil {
			zvs[i] = types.NewZSetValue()
		}
	}
	if zvs[0] == nil {
		zvs[0] = types.NewZSetValue()
	}

	result := computeDiff(zvs, 1.0)
	dst, _ := getZSetWrite(ctx, dstKey)
	dst.Mu.Lock()
	dst.Scores = make(map[string]float64)
	for _, m := range result {
		dst.Scores[m.Member] = m.Score
	}
	dst.Dirty = true
	dst.Mu.Unlock()
	return protocol.MakeInt(len(result))
}

// ---- ZLEXCOUNT ----

func handleZLexCount(ctx *CommandContext) string {
	zv, errStr := getZSetRead(ctx, ctx.Args[1])
	if errStr != "" {
		return errStr
	}
	if zv == nil {
		return protocol.MakeInt(0)
	}
	minStr, minEx := parseLexBound(ctx.Args[2])
	maxStr, maxEx := parseLexBound(ctx.Args[3])
	return protocol.MakeInt(zv.CountByLex(minStr, maxStr, minEx, maxEx))
}

// ---- ZMPOP ----

func handleZMPop(ctx *CommandContext) string {
	args := ctx.Args[1:]
	numKeys, err := strconv.Atoi(args[0])
	if err != nil || numKeys < 1 {
		return protocol.MakeError("ERR numkey must be a positive integer")
	}

	keys := args[1 : 1+numKeys]
	direction := strings.ToUpper(args[1+numKeys]) // MIN or MAX

	count := 1
	if len(args) > 1+numKeys+1 {
		// 可能有 COUNT count
		rest := args[1+numKeys+1:]
		for i := 0; i < len(rest); i++ {
			if strings.ToUpper(rest[i]) == "COUNT" && i+1 < len(rest) {
				count, _ = strconv.Atoi(rest[i+1])
				break
			}
		}
	}

	for _, key := range keys {
		zv, _ := getZSetRead(ctx, key)
		if zv == nil || zv.Len() == 0 {
			continue
		}
		var popped []types.ZSetMember
		if direction == "MAX" {
			popped = zv.PopMax(count)
		} else {
			popped = zv.PopMin(count)
		}
		arr := []string{
			protocol.MakeBulkString(key),
			protocol.MakeArray(membersToArray(popped)),
		}
		return protocol.MakeArray(arr)
	}
	return protocol.MakeNull()
}
