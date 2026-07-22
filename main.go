package main

import (
	"flag"
	"fmt"

	"godis/commands"
	"godis/config"
	"godis/datastore"
	"godis/logger"
	"godis/recovery"
	"godis/server"
	"godis/version"
	"godis/webadmin"
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
	configPath := flag.String("config", "./etc/godis.yaml", "config file path")
	flag.Parse()

	fmt.Print(banner)
	fmt.Printf("  Version: %s  Build: %s  Commit: %s\n\n", version.Version, version.BuildTime, version.GitCommit)

	// 加载配置
	if err := config.Init(*configPath); err != nil {
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

	// 从 AOF 文件中恢复历史数据
	recovery.ReloadHistoryData(cfg.AofFile, dbs, commands.CommandRegistry)

	// 启动全局 GC 协程
	datastore.StartGcWorker(dbs)

	// 启动 AOF 自动重写监控协程
	datastore.StartAutoRewriteWorker(cfg.AofFile, aof, dbs)

	// 启动 Web 管理后台
	webadmin.Start(dbs, aof, webAssets())

	// 启动 TCP 网络服务器
	srv := server.NewServer(dbs, aof)
	srv.Start(fmt.Sprintf("%s:%d", cfg.Bind, cfg.Port))
}
