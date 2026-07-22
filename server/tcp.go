package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"godis/commands"
	"godis/config"
	"godis/datastore"
	"godis/logger"
	"godis/protocol"
	"godis/pubsub"
)

var log = logger.NewModuleLogger("SERVER")

type Server struct {
	dbs []*datastore.GodisDB
	aof *datastore.AofLogger
}

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

	currentDBID := 0
	pubsubClient := pubsub.NewClient(conn)
	authenticated := false
	requirepass := config.Global.RequirePass

	defer pubsub.GlobalHub.Disconnect(pubsubClient)

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
			Args:         args,
			DB:           s.dbs[currentDBID],
			AllDBs:       s.dbs,
			CurrentDBID:  &currentDBID,
			Aof:          s.aof,
			Conn:         conn,
			PubSubClient: pubsubClient,
		}

		// 认证检查：未认证时仅允许 AUTH / PING
		if requirepass != "" && !authenticated {
			if cmdName == "AUTH" {
				reply, _, ok := commands.Execute(cmdName, ctx)
				if ok && strings.HasPrefix(reply, "+OK") {
					authenticated = true
				}
				conn.Write([]byte(reply))
				continue
			}
			if cmdName == "PING" {
				reply, _, _ := commands.Execute(cmdName, ctx)
				conn.Write([]byte(reply))
				continue
			}
			conn.Write([]byte("-NOAUTH Authentication required.\r\n"))
			continue
		}

		// 订阅模式下只允许订阅相关命令 + PING
		if pubsubClient.IsSubscribed() {
			if !commands.IsPubSubCmd(cmdName) {
				conn.Write([]byte("-ERR only (P)SUBSCRIBE/(P)UNSUBSCRIBE/PING allowed in subscribed mode\r\n"))
				continue
			}
			reply, _, ok := commands.Execute(cmdName, ctx)
			if ok && !commands.SubCmdWritesDirect(cmdName) && reply != "" {
				conn.Write([]byte(reply))
			}
			continue
		}

		reply, cmd, ok := commands.Execute(cmdName, ctx)
		if ok {
			// 订阅命令需直接写连接
			if commands.SubCmdWritesDirect(cmdName) {
				// 确认消息已在 handler 中写入
				continue
			}
			if cmd.Flags == commands.FlagWrite && !strings.HasPrefix(reply, "-ERR") {
				_ = s.aof.WriteCmd(args, currentDBID)
			}
		} else {
			reply = fmt.Sprintf("-ERR unknown command '%s'\r\n", cmdName)
		}

		conn.Write([]byte(reply))
	}
}
