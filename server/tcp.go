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
	dbs []*datastore.GodisDB // 数据库切片，支持多数据库
	aof *datastore.AofLogger // aof命令日志记录
}

// NewServer 接收 aof 参数
func NewServer(dbs []*datastore.GodisDB, aof *datastore.AofLogger) *Server {
	return &Server{dbs: dbs, aof: aof}
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

	currentDBID := 0 // 默认数据库0
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
			activeDB := s.dbs[currentDBID]
			ctx := &commands.CommandContext{
				Args:        args,
				DB:          activeDB,
				AllDBs:      s.dbs,
				CurrentDBID: &currentDBID,
				Aof:         s.aof,
			}
			reply = handler(ctx)

			// 持久化 AOF（所有数据库均写入，非 db0 自动携带 SELECT）
			if cmdName == "SET" && strings.HasPrefix(reply, "+OK") {
				_ = s.aof.WriteCmd(args, currentDBID)
			}

		} else {
			reply = fmt.Sprintf("-ERR unknown command '%s'\r\n", cmdName)
		}

		conn.Write([]byte(reply))
	}
}
