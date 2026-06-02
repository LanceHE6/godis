package main

import (
	"godis/datastore"
	"godis/server"
)

func main() {
	// 初始化存储引擎
	db := datastore.NewGodisDB()

	// 创建并启动网络服务器
	srv := server.NewServer(db)
	srv.Start("0.0.0.0:6379")
}
