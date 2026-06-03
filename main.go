package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"godis/commands"
	"godis/datastore"
	"godis/protocol"
	"godis/server"
)

const aofFilename = "godis.aof"

func main() {
	// 初始化存储引擎
	db := datastore.NewGodisDB()

	// 尝试从 AOF 文件中恢复历史数据
	reloadHistoryData(db)

	// 初始化 AOF 记录器
	aof, err := datastore.NewAofLogger(aofFilename)
	if err != nil {
		panic(fmt.Sprintf("创建 AOF 失败: %v", err))
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

func reloadHistoryData(db *datastore.GodisDB) {
	file, err := os.Open(aofFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return // 不存在任何历史数据，跳过
		}
		fmt.Printf("读取 AOF 恢复文件失败: %v\n", err)
		return
	}
	defer file.Close()

	fmt.Println("【💾 混合恢复】检测到历史数据，正在分析文件格式...")
	reader := bufio.NewReader(file)

	// 1. 判断是否是混合持久化格式
	firstLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	if strings.TrimSpace(firstLine) == "GODIS-HYBRID" {
		fmt.Println("【💾 混合恢复】识别为混合格式，优先载入 GDB 二进制快照头部...")
		// 调用二进制加载方法，它会正好把二进制数据块全部读完
		if err := db.LoadFromBinary(reader); err != nil {
			fmt.Printf("加载 GDB 二进制快照失败: %v\n", err)
			return
		}
		fmt.Println("【💾 混合恢复】GDB 头部载入成功！开始追加读取后续增量 AOF 文本...")
	} else {
		// 如果不是魔数，说明是纯文本老 AOF，我们需要把文件指针重置回开头
		fmt.Println("【💾 混合恢复】识别为传统纯文本 AOF 格式，从头重放指令...")
		_, _ = file.Seek(0, 0)
		reader = bufio.NewReader(file)
	}

	// 2. 处理剩下的（或者全量的）RESP 文本指令追加
	count := 0
	for {
		args, err := protocol.ParseRESP(reader)
		if err != nil {
			if err == io.EOF {
				break // 读到文件末尾，安全退出
			}
			fmt.Printf("解析增量 AOF 出错: %v\n", err)
			break
		}

		if len(args) == 0 {
			continue
		}

		cmdName := strings.ToUpper(args[0])
		if handler, exists := commands.CommandRegistry[cmdName]; exists {
			ctx := &commands.CommandContext{
				Args: args,
				DB:   db,
			}
			handler(ctx)
			count++
		}
	}
	fmt.Printf("【💾 混合恢复】数据完全复活！共重放了 %d 条增量文本命令\n", count)
}
