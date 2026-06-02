package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"godis/commands" // 引入命令层
	"godis/datastore"
	"godis/protocol"
)

type Server struct {
	db *datastore.GodisDB
}

func NewServer(db *datastore.GodisDB) *Server {
	return &Server{db: db}
}

func (s *Server) Start(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("启动监听失败: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Godis 服务器已启动，正在监听 %s ...\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("接收客户端连接失败: %v\n", err)
			continue
		}
		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		args, err := protocol.ParseRESP(reader)
		if err != nil {
			if err == io.EOF {
				return
			}
			conn.Write([]byte("-ERR Protocol error\r\n"))
			return
		}

		if len(args) == 0 {
			continue
		}

		cmdName := strings.ToUpper(args[0])

		// 命令注册表里查找命令
		var reply string
		if handler, exists := commands.CommandRegistry[cmdName]; exists {
			// 组装上下文，丢给命令层去跑
			ctx := &commands.CommandContext{
				Args: args,
				DB:   s.db,
			}
			reply = handler(ctx)
		} else {
			reply = fmt.Sprintf("-ERR unknown command '%s'\r\n", cmdName)
		}

		conn.Write([]byte(reply))
	}
}
