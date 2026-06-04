package server

import (
	"bufio"
	"fmt"
	"godis/logger"
	"io"
	"net"
	"strings"

	"godis/commands"
	"godis/datastore"
	"godis/protocol"
)

var log = logger.NewModuleLogger("SERVER")

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
		log.Error("failed to start listener: %v\n", err)
		return
	}
	defer listener.Close()

	log.Info("godis server started, listening on %s ...\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("failed to accept client connection: %v\n", err)
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
