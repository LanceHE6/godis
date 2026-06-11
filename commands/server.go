package commands

import (
	"fmt"
	"sort"
	"strings"

	"godis/config"
	"godis/protocol"
	"godis/version"
)

func init() {
	// 获取服务器信息和统计数据
	Register("INFO", -1, FlagFast, 0, 0, 0, handleInfo)
	// 手动触发 AOF 混合持久化重写
	Register("BGREWRITEAOF", 1, FlagAdmin, 0, 0, 0, handleBGRewriteAOF)
	// 获取所有命令的详细信息，支持 COUNT/INFO/GETKEYS 子命令
	Register("COMMAND", -1, FlagFast, 0, 0, 0, handleCommand)
	// 配置相关命令，包含 GET/RESETSTAT/REWRITE/SET 子命令
	Register("CONFIG", -1, FlagAdmin, 0, 0, 0, handleConfig)
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
		sb.WriteString(fmt.Sprintf("tcp_port:%d\r\n", config.Global.Port))
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
		return protocol.WrongArgsErr("command getkeys")
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
			protocol.MakeSimpleString(string(cmd.Flags)),
		}),
		protocol.MakeInt(cmd.FirstKey),
		protocol.MakeInt(cmd.LastKey),
		protocol.MakeInt(cmd.KeyStep),
	})
}

func handleConfig(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.WrongArgsErr("config")
	}

	sub := strings.ToUpper(ctx.Args[1])
	switch sub {
	case "GET":
		return configGet(ctx.Args[2:])
	case "SET":
		return configSet(ctx.Args[2:])
	case "RESETSTAT":
		return protocol.MakeSimpleString("OK")
	case "REWRITE":
		return configRewrite()
	default:
		return protocol.MakeError(fmt.Sprintf("ERR unknown subcommand '%s'", ctx.Args[1]))
	}
}

// configFields 定义配置字段名到值的映射（有序）
func configFields() map[string]string {
	cfg := config.Global
	return map[string]string{
		"bind":      cfg.Bind,
		"port":      fmt.Sprintf("%d", cfg.Port),
		"databases": fmt.Sprintf("%d", cfg.Databases),
		"aof_file":  cfg.AofFile,
		"log_file":  cfg.LogFile,
		"log_level": cfg.LogLevel,
	}
}

func configGet(args []string) string {
	if len(args) < 1 {
		return protocol.WrongArgsErr("config get")
	}

	pattern := args[0]
	fields := configFields()

	// 收集匹配的 key，保持输出有序
	var keys []string
	for k := range fields {
		if matchConfigPattern(pattern, k) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	elements := make([]string, 0, len(keys)*2)
	for _, k := range keys {
		elements = append(elements, protocol.MakeBulkString(k), protocol.MakeBulkString(fields[k]))
	}
	return protocol.MakeArray(elements)
}

func matchConfigPattern(pattern, name string) bool {
	if pattern == "*" {
		return true
	}
	// 简单的前缀/后缀通配符匹配
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		return strings.Contains(name, pattern[1:len(pattern)-1])
	}
	if strings.HasPrefix(pattern, "*") {
		return strings.HasSuffix(name, pattern[1:])
	}
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(name, pattern[:len(pattern)-1])
	}
	return name == pattern
}

// configSet 配置设置，目前仅支持log_level
func configSet(args []string) string {
	if len(args) < 2 {
		return protocol.WrongArgsErr("config set")
	}

	key := strings.ToLower(args[0])
	value := args[1]

	switch key {
	case "log_level":
		config.Global.LogLevel = value
	default:
		return protocol.MakeError(fmt.Sprintf("ERR CONFIG SET '%s' is not supported", key))
	}

	return protocol.MakeSimpleString("OK")
}

func configRewrite() string {
	if err := config.Save(); err != nil {
		return protocol.MakeError(fmt.Sprintf("ERR rewrite failed: %v", err))
	}
	return protocol.MakeSimpleString("OK")
}
