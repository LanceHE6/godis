package commands

import (
	"strings"

	"godis/protocol"
	"godis/pubsub"
)

func init() {
	// 发布消息到频道
	Register("PUBLISH", 3, FlagWrite, 0, 0, 0, handlePublish)
	// 订阅频道
	Register("SUBSCRIBE", -2, FlagReadonly, 0, 0, 0, handleSubscribe)
	// 退订频道
	Register("UNSUBSCRIBE", -1, FlagReadonly, 0, 0, 0, handleUnsubscribe)
	// 按模式订阅频道
	Register("PSUBSCRIBE", -2, FlagReadonly, 0, 0, 0, handlePSubscribe)
	// 按模式退订频道
	Register("PUNSUBSCRIBE", -1, FlagReadonly, 0, 0, 0, handlePUnsubscribe)
	// 查看订阅状态
	Register("PUBSUB", -2, FlagReadonly, 0, 0, 0, handlePubSub)
}

func getPubSubClient(ctx *CommandContext) *pubsub.Client {
	if ctx.PubSubClient != nil {
		return ctx.PubSubClient.(*pubsub.Client)
	}
	return pubsub.NewClient(ctx.Conn)
}

func handlePublish(ctx *CommandContext) string {
	channel := ctx.Args[1]
	message := ctx.Args[2]
	count := pubsub.GlobalHub.Publish(channel, message)
	return protocol.MakeInt(count)
}

func handleSubscribe(ctx *CommandContext) string {
	if ctx.Conn == nil {
		return protocol.MakeError("ERR SUBSCRIBE requires a connection")
	}
	client := getPubSubClient(ctx)
	channels := ctx.Args[1:]
	confirms := pubsub.GlobalHub.Subscribe(client, channels...)
	for _, c := range confirms {
		ctx.Conn.Write([]byte(c))
	}
	return ""
}

func handlePSubscribe(ctx *CommandContext) string {
	if ctx.Conn == nil {
		return protocol.MakeError("ERR PSUBSCRIBE requires a connection")
	}
	client := getPubSubClient(ctx)
	patterns := ctx.Args[1:]
	confirms := pubsub.GlobalHub.PSubscribe(client, patterns...)
	for _, c := range confirms {
		ctx.Conn.Write([]byte(c))
	}
	return ""
}

func handleUnsubscribe(ctx *CommandContext) string {
	if ctx.Conn == nil {
		return protocol.MakeError("ERR UNSUBSCRIBE requires a connection")
	}
	client := getPubSubClient(ctx)
	channels := ctx.Args[1:]
	confirms := pubsub.GlobalHub.Unsubscribe(client, channels...)
	for _, c := range confirms {
		ctx.Conn.Write([]byte(c))
	}
	return ""
}

func handlePUnsubscribe(ctx *CommandContext) string {
	if ctx.Conn == nil {
		return protocol.MakeError("ERR PUNSUBSCRIBE requires a connection")
	}
	client := getPubSubClient(ctx)
	patterns := ctx.Args[1:]
	confirms := pubsub.GlobalHub.PUnsubscribe(client, patterns...)
	for _, c := range confirms {
		ctx.Conn.Write([]byte(c))
	}
	return ""
}

func handlePubSub(ctx *CommandContext) string {
	args := ctx.Args[1:]
	if len(args) == 0 {
		return protocol.MakeError("ERR wrong number of arguments for 'PUBSUB' command")
	}
	subCmd := strings.ToUpper(args[0])

	switch subCmd {
	case "CHANNELS":
		pattern := ""
		if len(args) >= 2 {
			pattern = args[1]
		}
		chans := pubsub.GlobalHub.Channels(pattern)
		if chans == nil {
			return protocol.MakeArray([]string{})
		}
		elements := make([]string, len(chans))
		for i, ch := range chans {
			elements[i] = protocol.MakeBulkString(ch)
		}
		return protocol.MakeArray(elements)

	case "NUMSUB":
		var chans []string
		if len(args) > 1 {
			chans = args[1:]
		}
		elements := pubsub.GlobalHub.NumSub(chans...)
		return protocol.MakeArray(elements)

	case "NUMPAT":
		count := pubsub.GlobalHub.NumPat()
		return protocol.MakeInt(count)

	default:
		return protocol.MakeError("ERR Unknown PUBSUB subcommand")
	}
}

// IsPubSubCmd 判断命令是否是订阅模式专属命令
func IsPubSubCmd(cmd string) bool {
	upper := strings.ToUpper(cmd)
	switch upper {
	case "SUBSCRIBE", "PSUBSCRIBE", "UNSUBSCRIBE", "PUNSUBSCRIBE", "PING":
		return true
	}
	return false
}

// SubCmdWritesDirect 判断命令是否需要服务器写入确认消息（返回空 reply）
func SubCmdWritesDirect(cmd string) bool {
	upper := strings.ToUpper(cmd)
	switch upper {
	case "SUBSCRIBE", "PSUBSCRIBE", "UNSUBSCRIBE", "PUNSUBSCRIBE":
		return true
	}
	return false
}
