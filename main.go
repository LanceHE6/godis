package main

import (
	"fmt"
	"godis/commands"
	"godis/datastore"
	"godis/logger"
	"godis/server"
	"godis/version"
)

const aofFilename = "./data/godis.aof"
const logFilename = "./logs/godis.log"
const dbCount = 16

const banner = `
  ██████╗  ██████╗ ██████╗ ██╗███████╗
 ██╔════╝ ██╔═══██╗██╔══██╗██║██╔════╝
 ██║  ███╗██║   ██║██║  ██║██║███████╗
 ██║   ██║██║   ██║██║  ██║██║╚════██║
 ╚██████╔╝╚██████╔╝██████╔╝██║███████║
  ╚═════╝  ╚═════╝ ╚═════╝ ╚═╝╚══════╝
`

func main() {
	fmt.Print(banner)
	fmt.Printf("  Version: %s  Build: %s  Commit: %s\n\n", version.Version, version.BuildTime, version.GitCommit)

	// 初始化日志引擎
	err := logger.InitGlobalLogger(logFilename, logger.LevelInfo)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.CloseLogSystem()

	// 初始化存储引擎
	dbs := make([]*datastore.GodisDB, dbCount)
	for i := 0; i < dbCount; i++ {
		dbs[i] = datastore.NewGodisDB()
	}

	// 初始化 AOF 记录器
	aof, err := datastore.NewAofLogger(aofFilename)
	if err != nil {
		panic(fmt.Sprintf("failed to create AOF file: %v", err))
	}
	defer aof.Close()

	// 尝试从 AOF 文件中恢复历史数据（支持多数据库）
	commands.ReloadHistoryData(aofFilename, dbs)

	// 将 aof 实例也注册到命令层的上下文，方便后续提供"手动重写"命令
	commands.GlobalAof = aof

	// 启动全局 GC 协程，清理所有数据库中的过期 Key
	datastore.StartGcWorker(dbs)

	// 启动 AOF 自动重写监控协程（适配多数据库）
	datastore.StartAutoRewriteWorker(aofFilename, aof, dbs)

	// 创建并启动网络服务器
	srv := server.NewServer(dbs, aof)
	srv.Start("0.0.0.0:6379")
}
