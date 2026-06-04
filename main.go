package main

import (
	"fmt"
	"godis/commands"
	"godis/datastore"
	"godis/logger"
	"godis/server"
)

const aofFilename = "godis.aof"
const logFilename = "./logs/godis.log"

func main() {

	// 初始化日志引擎
	err := logger.InitGlobalLogger(logFilename, logger.LevelInfo)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	defer logger.CloseLogSystem()

	// 初始化存储引擎
	db := datastore.NewGodisDB()

	// 尝试从 AOF 文件中恢复历史数据
	commands.ReloadHistoryData(aofFilename, db)

	// 初始化 AOF 记录器
	aof, err := datastore.NewAofLogger(aofFilename)
	if err != nil {
		panic(fmt.Sprintf("failed to create AOF file: %v", err))
	}
	defer aof.Close()

	// 将 aof 实例也注册到命令层的上下文，方便后续提供“手动重写”命令
	commands.GlobalAof = aof

	// 启动 GBD 监控协程
	db.StartAutoRewriteWorker(aofFilename, aof)

	// 创建并启动网络服务器
	srv := server.NewServer(db, aof)
	srv.Start("0.0.0.0:6379")
}
