package webadmin

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"godis/config"
	"godis/datastore"
	"godis/logger"
)

var log *logger.ModuleLogger
var aof *datastore.AofLogger
var dbs []*datastore.GodisDB

func Start(dbList []*datastore.GodisDB, aofLogger *datastore.AofLogger, assets fs.FS) {
	dbs = dbList
	aof = aofLogger
	log = logger.NewModuleLogger("WEB")
	startTime = time.Now()

	if !config.Global.WebAdmin {
		log.Info("web admin disabled by config")
		return
	}

	addr := fmt.Sprintf("%s:%d", config.Global.WebBind, config.Global.WebPort)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth", handleAuth)
	mux.HandleFunc("/api/server/info", handleServerInfo)
	mux.HandleFunc("/api/server/stats", handleServerStats)
	mux.HandleFunc("/api/keys", handleKeys)
	mux.HandleFunc("/api/keys/delete", handleKeysDelete)
	mux.HandleFunc("/api/key", handleKeyDetail)
	mux.HandleFunc("/api/key/edit", handleKeyEdit)
	mux.HandleFunc("/api/exec", handleExec)
	mux.HandleFunc("/api/commands", handleCommands)
	mux.HandleFunc("/api/logs", handleLogs)

	mux.Handle("/", http.FileServer(http.FS(assets)))

	go func() {
		fmt.Printf("[web] admin dashboard at http://%s\n", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			fmt.Fprintf(os.Stderr, "[web] server error: %v\n", err)
		}
	}()
}
