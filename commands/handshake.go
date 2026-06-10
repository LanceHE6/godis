package commands

import "godis/protocol"

func init() {
	// 测试连接活性，可选携带消息参数
	Register("PING", -1, "fast", 0, 0, 0, func(ctx *CommandContext) string {
		if len(ctx.Args) > 1 {
			return protocol.MakeBulkString(ctx.Args[1])
		}
		return protocol.MakeSimpleString("PONG")
	})
}
