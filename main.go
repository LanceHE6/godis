package main

import (
	"fmt"
	"godis/commands"
	"godis/datastore"
	"godis/logger"
	"godis/server"
)

const aofFilename = "./data/godis.aof"
const logFilename = "./logs/godis.log"
const dbCount = 16

func main() {

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

	// 尝试从 AOF 文件中恢复历史数据
	// TODO需支持所有数据库恢复数据
	commands.ReloadHistoryData(aofFilename, dbs[0])

	// 初始化 AOF 记录器
	aof, err := datastore.NewAofLogger(aofFilename)
	if err != nil {
		panic(fmt.Sprintf("failed to create AOF file: %v", err))
	}
	defer aof.Close()

	// 将 aof 实例也注册到命令层的上下文，方便后续提供“手动重写”命令
	commands.GlobalAof = aof

	// 启动 GBD 监控协程
	dbs[0].StartAutoRewriteWorker(aofFilename, aof)

	// 创建并启动网络服务器
	srv := server.NewServer(dbs, aof)
	srv.Start("0.0.0.0:6379")
}
