package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"godis/commands"
	"godis/datastore"
	"godis/protocol"
)

type Server struct {
	db  *datastore.GodisDB
	aof *datastore.AofLogger
}

// NewServer 接收 aof 参数
func NewServer(db *datastore.GodisDB, aof *datastore.AofLogger) *Server {
	return &Server{db: db, aof: aof}
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

		var reply string
		if handler, exists := commands.CommandRegistry[cmdName]; exists {
			ctx := &commands.CommandContext{
				Args: args,
				DB:   s.db,
			}
			reply = handler(ctx)

			// 如果命令执行成功（返回 +OK 状态），且是 SET 命令，将其落盘！
			if cmdName == "SET" && strings.HasPrefix(reply, "+OK") {
				_ = s.aof.WriteCmd(args)
			}

		} else {
			reply = fmt.Sprintf("-ERR unknown command '%s'\r\n", cmdName)
		}

		conn.Write([]byte(reply))
	}
}
