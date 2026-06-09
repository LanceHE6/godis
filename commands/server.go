package commands

import (
	"fmt"
	"strings"

	"godis/protocol"
	"godis/version"
)

func init() {
	// 获取服务器信息
	CommandRegistry["INFO"] = handleInfo
	// 手动异步重写 AOF 持久化文件
	CommandRegistry["BGREWRITEAOF"] = handleBGRewriteAOF
	// 获取所有 Godis 命令的详细信息
	CommandRegistry["COMMAND"] = handleCommand
}

// 手动触发混合重写的命令
func handleBGRewriteAOF(ctx *CommandContext) string {
	cmdLog.Debug("BGREWRITEAOF received")
	if ctx.Aof == nil {
		return protocol.MakeError("ERR AOF logger not initialized")
	}

	// 触发所有数据库的二进制写入
	err := ctx.Aof.RewriteToHybrid(ctx.AllDBs)
	if err != nil {
		return protocol.MakeError(fmt.Sprintf("ERR Rewrite failed: %v", err))
	}
	return protocol.MakeSimpleString("Background append only file rewriting started (Godis Hybrid Mode)")
}

func handleInfo(ctx *CommandContext) string {
	cmdLog.Debug("INFO received")

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
	cmdLog.Debug("COMMAND received")
	return protocol.MakeArray([]string{})
}
