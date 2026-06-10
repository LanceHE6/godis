package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"godis/commands"
	"godis/datastore"
	"godis/logger"
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
	defer func() {
		log.Info("client disconnected: %s", conn.RemoteAddr())
		conn.Close()
	}()
	log.Info("new client connected: %s", conn.RemoteAddr())
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

		ctx := &commands.CommandContext{
			Args:        args,
			DB:          s.dbs[currentDBID],
			AllDBs:      s.dbs,
			CurrentDBID: &currentDBID,
			Aof:         s.aof,
		}

		reply, cmd, ok := commands.Execute(cmdName, ctx)
		if ok {
			// 写类型命令且执行成功时持久化 AOF（非 db0 自动携带 SELECT）
			if strings.Contains(cmd.Flags, "write") && strings.HasPrefix(reply, "+OK") {
				_ = s.aof.WriteCmd(args, currentDBID)
			}
		} else {
			reply = fmt.Sprintf("-ERR unknown command '%s'\r\n", cmdName)
		}

		conn.Write([]byte(reply))
	}
}
