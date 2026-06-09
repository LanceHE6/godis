package commands

import "godis/protocol"

func init() {
	Register("PING", -1, "fast", 0, 0, 0, func(ctx *CommandContext) string {
		if len(ctx.Args) > 1 {
			return protocol.MakeBulkString(string(rune(len(ctx.Args[1]))))
		}
		return protocol.MakeSimpleString("PONG")
	})
}
