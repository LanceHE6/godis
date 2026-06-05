package commands

import "godis/protocol"

func init() {
	// PING
	CommandRegistry["PING"] = func(ctx *CommandContext) string {
		if len(ctx.Args) > 1 {
			return protocol.MakeBulkString(string(rune(len(ctx.Args[1]))))
		}
		return protocol.MakeSimpleString("PONG")
	}

}
