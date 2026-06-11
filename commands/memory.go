package commands

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"godis/datastore"
	"godis/protocol"
	"godis/types"
)

func init() {
	// 内存相关命令，包含 STATS/USAGE 子命令
	Register("MEMORY", -2, FlagAdmin, 0, 0, 0, handleMemory)
}

func handleMemory(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.MakeError("ERR wrong number of arguments for 'memory' command")
	}
	sub := strings.ToUpper(ctx.Args[1])
	switch sub {
	case "STATS":
		return memoryStats(ctx.AllDBs)
	case "USAGE":
		if len(ctx.Args) < 3 {
			return protocol.MakeError("ERR wrong number of arguments for 'memory usage' command")
		}
		return memoryUsage(ctx.DB, ctx.Args[2])
	default:
		return protocol.MakeError(fmt.Sprintf("ERR unknown subcommand '%s'", ctx.Args[1]))
	}
}

func memoryStats(dbs []*datastore.GodisDB) string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	totalKeys := 0
	for _, db := range dbs {
		totalKeys += len(db.Keys())
	}

	stats := []string{
		protocol.MakeBulkString("used_memory"),
		protocol.MakeInt(int(m.Alloc)),
		protocol.MakeBulkString("used_memory_rss"),
		protocol.MakeInt(int(m.Sys)),
		protocol.MakeBulkString("used_memory_peak"),
		protocol.MakeInt(int(m.TotalAlloc)),
		protocol.MakeBulkString("databases"),
		protocol.MakeInt(len(dbs)),
		protocol.MakeBulkString("total_keys"),
		protocol.MakeInt(totalKeys),
	}
	return protocol.MakeArray(stats)
}

// estimateValueSize 估算 value 的内存占用（字节）
func estimateValueSize(v interface{}) int {
	switch val := v.(type) {
	case *types.StringValue:
		return int(unsafe.Sizeof(*val)) + len(val.Value)
	case *types.HashValue:
		size := int(unsafe.Sizeof(*val))
		val.Mu.RLock()
		for k, v := range val.Fields {
			size += len(k) + len(v)
		}
		val.Mu.RUnlock()
		return size
	case *types.ListValue:
		size := int(unsafe.Sizeof(*val))
		for _, s := range val.Data() {
			size += len(s)
		}
		return size
	case *types.SetValue:
		size := int(unsafe.Sizeof(*val))
		val.Mu.RLock()
		for m := range val.Members {
			size += len(m)
		}
		val.Mu.RUnlock()
		return size
	case *types.ZSetValue:
		size := int(unsafe.Sizeof(*val))
		val.Mu.RLock()
		for m, s := range val.Scores {
			size += len(m) + int(unsafe.Sizeof(s))
		}
		val.Mu.RUnlock()
		return size
	default:
		return 64
	}
}

func memoryUsage(db *datastore.GodisDB, key string) string {
	item, ok := db.GetItem(key)
	if !ok {
		return protocol.MakeNull()
	}

	// key 占用 + Item 结构体 + value 实际数据
	usage := len(key) + int(unsafe.Sizeof(*item)) + estimateValueSize(item.Value)
	return protocol.MakeInt(usage)
}
