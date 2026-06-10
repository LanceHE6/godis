package commands

import (
	"fmt"
	"strings"

	"godis/protocol"
	"godis/version"
)

func init() {
	// 获取服务器信息和统计数据
	Register("INFO", -1, "fast", 0, 0, 0, handleInfo)
	// 手动触发 AOF 混合持久化重写
	Register("BGREWRITEAOF", 1, "admin", 0, 0, 0, handleBGRewriteAOF)
	// 获取所有命令的详细信息，支持 COUNT/INFO/GETKEYS 子命令
	Register("COMMAND", -1, "fast", 0, 0, 0, handleCommand)
}

func handleBGRewriteAOF(ctx *CommandContext) string {
	cmdLog.Debug("BGREWRITEAOF received")
	if ctx.Aof == nil {
		return protocol.MakeError("ERR AOF logger not initialized")
	}

	err := ctx.Aof.RewriteToHybrid(ctx.AllDBs)
	if err != nil {
		return protocol.MakeError(fmt.Sprintf("ERR Rewrite failed: %v", err))
	}
	return protocol.MakeSimpleString("Background append only file rewriting started (Godis Hybrid Mode)")
}

func handleInfo(ctx *CommandContext) string {
	var section string
	if len(ctx.Args) > 1 {
		section = strings.ToLower(ctx.Args[1])
	}

	var sb strings.Builder

	if section == "" || section == "server" {
		sb.WriteString("# Server\r\n")
		sb.WriteString(fmt.Sprintf("godis_version:%s\r\n", version.Version))
		sb.WriteString("godis_mode:standalone\r\n")
		sb.WriteString("tcp_port:6379\r\n")
		sb.WriteString(fmt.Sprintf("build_time:%s\r\n", version.BuildTime))
		sb.WriteString(fmt.Sprintf("git_commit:%s\r\n", version.GitCommit))
	}

	if section == "" || section == "clients" {
		sb.WriteString("# Clients\r\n")
		sb.WriteString("connected_clients:1\r\n")
	}

	if section == "" || section == "keyspace" {
		sb.WriteString("# Keyspace\r\n")
		for i, db := range ctx.AllDBs {
			stats := db.Stats()
			if stats.Keys > 0 {
				sb.WriteString(fmt.Sprintf("db%d:keys=%d,expires=%d,avg_ttl=0\r\n", i, stats.Keys, stats.Expires))
			}
		}
	}

	return protocol.MakeBulkString(sb.String())
}

func handleCommand(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return commandInfoAll()
	}

	sub := strings.ToUpper(ctx.Args[1])
	switch sub {
	case "COUNT":
		return protocol.MakeInt(len(CommandRegistry))
	case "INFO":
		return commandInfo(ctx.Args[2:])
	case "GETKEYS":
		return commandGetKeys(ctx.Args[2:])
	default:
		return protocol.MakeError(fmt.Sprintf("ERR unknown subcommand '%s'", ctx.Args[1]))
	}
}

func commandInfoAll() string {
	elements := make([]string, 0, len(CommandRegistry))
	for _, cmd := range CommandRegistry {
		elements = append(elements, formatCommandInfo(cmd))
	}
	return protocol.MakeArray(elements)
}

func commandInfo(names []string) string {
	elements := make([]string, 0, len(names))
	for _, name := range names {
		if cmd, ok := CommandRegistry[strings.ToUpper(name)]; ok {
			elements = append(elements, formatCommandInfo(cmd))
		} else {
			elements = append(elements, protocol.MakeNull())
		}
	}
	return protocol.MakeArray(elements)
}

func commandGetKeys(args []string) string {
	if len(args) < 1 {
		return protocol.MakeError("ERR wrong number of arguments for 'command getkeys' command")
	}

	cmd, ok := CommandRegistry[strings.ToUpper(args[0])]
	if !ok {
		return protocol.MakeError(fmt.Sprintf("ERR unknown command '%s'", args[0]))
	}

	if cmd.FirstKey == 0 {
		return protocol.MakeArray([]string{})
	}

	keys := []string{}
	if cmd.LastKey > 0 {
		for i := cmd.FirstKey - 1; i < cmd.LastKey && i < len(args)-1; i++ {
			keys = append(keys, protocol.MakeBulkString(args[i+1]))
		}
	} else if cmd.LastKey < 0 {
		for i := cmd.FirstKey - 1; i < len(args)-1; i += cmd.KeyStep {
			keys = append(keys, protocol.MakeBulkString(args[i+1]))
		}
	}
	return protocol.MakeArray(keys)
}

func formatCommandInfo(cmd Command) string {
	return protocol.MakeArray([]string{
		protocol.MakeBulkString(cmd.Name),
		protocol.MakeInt(cmd.Arity),
		protocol.MakeArray([]string{
			protocol.MakeSimpleString(cmd.Flags),
		}),
		protocol.MakeInt(cmd.FirstKey),
		protocol.MakeInt(cmd.LastKey),
		protocol.MakeInt(cmd.KeyStep),
	})
}
