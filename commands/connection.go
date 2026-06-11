package commands

import (
	"fmt"
	"os"
	"time"

	"godis/protocol"
)

func init() {
	// 关闭服务器
	Register("SHUTDOWN", -1, FlagAdmin, 0, 0, 0, handleShutdown)
	// 返回服务器当前时间
	Register("TIME", 1, FlagFast, 0, 0, 0, handleTime)
}

func handleShutdown(ctx *CommandContext) string {
	cmdLog.Info("SHUTDOWN received, stopping server...")
	if ctx.Aof != nil {
		ctx.Aof.Close()
	}
	os.Exit(0)
	return "" // unreachable
}

func handleTime(ctx *CommandContext) string {
	now := time.Now()
	sec := now.Unix()
	usec := now.Nanosecond() / 1000
	return protocol.MakeArray([]string{
		protocol.MakeBulkString(fmt.Sprintf("%d", sec)),
		protocol.MakeBulkString(fmt.Sprintf("%d", usec)),
	})
}
