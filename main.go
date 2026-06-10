package main

import (
	"fmt"
	"godis/commands"
	"godis/config"
	"godis/datastore"
	"godis/logger"
	"godis/recovery"
	"godis/server"
	"godis/version"
)

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

	// 加载配置（不存在则自动生成 godis.conf）
	if err := config.Init("./etc/godis.yaml"); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	cfg := config.Global

	// 初始化日志引擎
	err := logger.InitGlobalLogger(cfg.LogFile, logger.ParseLevel(cfg.LogLevel))
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.CloseLogSystem()

	// 初始化存储引擎
	dbs := make([]*datastore.GodisDB, cfg.Databases)
	for i := 0; i < cfg.Databases; i++ {
		dbs[i] = datastore.NewGodisDB()
	}

	// 初始化 AOF 记录器
	aof, err := datastore.NewAofLogger(cfg.AofFile)
	if err != nil {
		panic(fmt.Sprintf("failed to create AOF file: %v", err))
	}
	defer aof.Close()

	// 从 AOF 文件中恢复历史数据（支持多数据库）
	recovery.ReloadHistoryData(cfg.AofFile, dbs, commands.CommandRegistry)

	// 启动全局 GC 协程，清理所有数据库中的过期 Key
	datastore.StartGcWorker(dbs)

	// 启动 AOF 自动重写监控协程（适配多数据库）
	datastore.StartAutoRewriteWorker(cfg.AofFile, aof, dbs)

	// 创建并启动网络服务器
	addr := fmt.Sprintf("%s:%d", cfg.Bind, cfg.Port)
	srv := server.NewServer(dbs, aof)
	srv.Start(addr)
}
