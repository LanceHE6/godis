package webadmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"godis/commands"
	"godis/config"
	"godis/logger"
	"godis/types"
	"godis/version"
)

// ---- Auth ----

func handleAuth(w http.ResponseWriter, r *http.Request) {
	logAPI(r)

	if r.Method == "GET" {
		sendJSON(w, map[string]any{"requirepass": config.Global.RequirePass != ""})
		return
	}
	var body struct{ Password string `json:"password"` }
	json.NewDecoder(r.Body).Decode(&body)
	ok := body.Password == config.Global.RequirePass
	sendJSON(w, map[string]any{"ok": ok})
}

// ---- Server Info ----

func handleServerInfo(w http.ResponseWriter, r *http.Request) {
	logAPI(r)

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	totalKeys := 0
	for _, db := range dbs {
		totalKeys += len(db.Keys())
	}
	sendJSON(w, map[string]any{
		"version":   version.Version,
		"uptime":    time.Since(startTime).Round(time.Second).String(),
		"keys":      totalKeys,
		"memory":    fmt.Sprintf("%.1f MB", float64(mem.Alloc)/1024/1024),
		"clients":   "1",
		"port":      config.Global.Port,
		"databases": config.Global.Databases,
	})
}

func handleServerStats(w http.ResponseWriter, r *http.Request) {
	logAPI(r)

	keysPerDB := make([]int, len(dbs))
	for i, db := range dbs {
		keysPerDB[i] = len(db.Keys())
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	cpuPct := readCPUUsage()
	appendStatsHistory(cpuPct, float64(mem.Alloc)/1024/1024)

	sendJSON(w, map[string]any{
		"keys_per_db": keysPerDB,
		"cpu_pct":     cpuPct,
		"memory_mb":   fmt.Sprintf("%.1f", float64(mem.Alloc)/1024/1024),
		"history":     getStatsHistory(),
	})
}

// ---- Keys ----

func handleKeys(w http.ResponseWriter, r *http.Request) {
	logAPI(r)

	pattern := r.URL.Query().Get("pattern")
	if pattern == "" { pattern = "*" }
	db := getDB(r)
	allKeys := db.Keys()

	var matched []string
	for _, k := range allKeys {
		if pattern == "*" || matchGlob(k, pattern) {
			matched = append(matched, k)
		}
	}
	sort.Strings(matched)

	total := len(matched)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 15 }

	start := (page - 1) * pageSize
	if start > len(matched) { start = len(matched) }
	end := start + pageSize
	if end > len(matched) { end = len(matched) }

	result := make([]map[string]any, 0, end-start)
	for _, k := range matched[start:end] {
		result = append(result, map[string]any{
			"key":  k,
			"type": typeName(db.TypeOf(k)),
			"ttl":  db.TTL(k),
			"size": 1,
		})
	}
	sendJSON(w, map[string]any{
		"keys":      result,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func handleKeysDelete(w http.ResponseWriter, r *http.Request) {
	logAPI(r)

	var body struct{ Keys []string `json:"keys"` }
	json.NewDecoder(r.Body).Decode(&body)
	db := getDB(r)
	deleted := db.Del(body.Keys...)
	logAOF(0, append([]string{"DEL"}, body.Keys...)...)
	sendJSON(w, map[string]any{"deleted": deleted})
}

// ---- Key Detail / Edit ----

func handleKeyDetail(w http.ResponseWriter, r *http.Request) {
	logAPI(r)

	key := r.URL.Query().Get("key")
	db := getDB(r)
	item, exists := db.GetItem(key)
	if !exists {
		http.Error(w, "key not found", 404)
		return
	}

	result := map[string]any{
		"key":  key,
		"type": typeName(item.Type),
		"ttl":  db.TTL(key),
	}

	switch item.Type {
	case types.TypeString:
		if sv, ok := item.Value.(*types.StringValue); ok { result["value"] = sv.Value }
	case types.TypeHash:
		if hv, ok := item.Value.(*types.HashValue); ok { result["fields"] = hv.GetAll() }
	case types.TypeList:
		if lv, ok := item.Value.(*types.ListValue); ok { result["values"] = lv.Data() }
	case types.TypeSet:
		if sv, ok := item.Value.(*types.SetValue); ok { result["members"] = sv.MembersList() }
	case types.TypeZSet:
		if zv, ok := item.Value.(*types.ZSetValue); ok { result["members"] = zv.Data() }
	}

	sendJSON(w, result)
}

func handleKeyEdit(w http.ResponseWriter, r *http.Request) {
	logAPI(r)

	var body struct {
		Key    string `json:"key"`
		Action string `json:"action"`
		Value  string `json:"value"`
		Field  string `json:"field"`
		DB     int    `json:"db"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	db := dbs[0]
	if body.DB >= 0 && body.DB < len(dbs) {
		db = dbs[body.DB]
	}

	switch body.Action {
	case "set_value":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeString {
			sendJSON(w, map[string]string{"error": "key not found or not a string"})
			return
		}
		db.Set(body.Key, body.Value, 0)
		logAOF(0, "SET", body.Key, body.Value)
		sendJSON(w, map[string]string{"status": "ok"})

	case "hset":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeHash {
			sendJSON(w, map[string]string{"error": "key not found or not a hash"})
			return
		}
		item.Value.(*types.HashValue).Set(body.Field, body.Value)
		logAOF(0, "HSET", body.Key, body.Field, body.Value)
		db.Touch(body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "hdel":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeHash {
			sendJSON(w, map[string]string{"error": "key not found or not a hash"})
			return
		}
		item.Value.(*types.HashValue).Del(body.Field)
		logAOF(0, "HDEL", body.Key, body.Field)
		db.Touch(body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "rpush":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeList {
			sendJSON(w, map[string]string{"error": "key not found or not a list"})
			return
		}
		item.Value.(*types.ListValue).PushRight(body.Value)
		logAOF(0, "RPUSH", body.Key, body.Value)
		db.Touch(body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "lrem":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeList {
			sendJSON(w, map[string]string{"error": "key not found or not a list"})
			return
		}
		item.Value.(*types.ListValue).Remove(body.Value, 1)
		logAOF(0, "LREM", body.Key, body.Value)
		db.Touch(body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "lset":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeList {
			sendJSON(w, map[string]string{"error": "key not found or not a list"})
			return
		}
		lv := item.Value.(*types.ListValue)
		idx, err := strconv.Atoi(body.Field)
		if err != nil || !lv.Set(idx, body.Value) {
			logAOF(0, "LSET", body.Key, body.Field, body.Value)
			sendJSON(w, map[string]string{"error": "invalid index or set failed"})
			return
		}
		db.Touch(body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "sadd":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeSet {
			sendJSON(w, map[string]string{"error": "key not found or not a set"})
			return
		}
		item.Value.(*types.SetValue).Add(body.Value)
		logAOF(0, "SADD", body.Key, body.Value)
		db.Touch(body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "srem":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeSet {
			sendJSON(w, map[string]string{"error": "key not found or not a set"})
			return
		}
		item.Value.(*types.SetValue).Remove(body.Value)
		logAOF(0, "SREM", body.Key, body.Value)
		db.Touch(body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "zset_score":
		item, exists := db.GetItem(body.Key)
		if !exists || item.Type != types.TypeZSet {
			sendJSON(w, map[string]string{"error": "key not found or not a zset"})
			return
		}
		zv := item.Value.(*types.ZSetValue)
		score, err := strconv.ParseFloat(body.Value, 64)
		if err != nil {
			sendJSON(w, map[string]string{"error": "invalid score"})
			return
		}
		zv.Add(score, body.Field)
		logAOF(0, "ZADD", body.Key, fmt.Sprintf("%f", score), body.Field)
		db.Touch(body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "set_ttl":
		sec, err := strconv.Atoi(body.Value)
		if err != nil || sec <= 0 {
			sendJSON(w, map[string]string{"error": "invalid TTL"})
			return
		}
		db.Expire(body.Key, sec)
		logAOF(0, "EXPIRE", body.Key, body.Value)
		sendJSON(w, map[string]string{"status": "ok"})

	case "persist":
		db.Persist(body.Key)
		logAOF(0, "PERSIST", body.Key)
		sendJSON(w, map[string]string{"status": "ok"})

	case "rename":
		newName := body.Value
		if newName == "" || newName == body.Key {
			sendJSON(w, map[string]string{"error": "invalid new name"})
			return
		}
		item, exists := db.GetItem(body.Key)
		if !exists {
			sendJSON(w, map[string]string{"error": "key not found"})
			return
		}
		db.SetItem(newName, *item)
		db.Del(body.Key)
		logAOF(0, "RENAME", body.Key, newName)
		sendJSON(w, map[string]string{"status": "ok"})
	}
}

// ---- Exec / Commands / Logs ----

func handleExec(w http.ResponseWriter, r *http.Request) {
	logAPI(r)

	var body struct{ Command string `json:"command"`; Db int `json:"db"` }
	json.NewDecoder(r.Body).Decode(&body)

	args := parseArgs(body.Command)
	if len(args) == 0 {
		http.Error(w, "empty command", 400)
		return
	}

	cmdName := strings.ToUpper(args[0])
	dbid := body.Db
	if dbid < 0 || dbid >= len(dbs) { dbid = 0 }
	db := dbs[dbid]

	ctx := &commands.CommandContext{
		Args:        args,
		DB:          db,
		AllDBs:      dbs,
		CurrentDBID: &dbid,
	}

	reply, cmd, ok := commands.Execute(cmdName, ctx)
	if !ok {
		reply = fmt.Sprintf("-ERR unknown command '%s'", cmdName)
	} else if cmd != nil && cmd.Flags == commands.FlagWrite && !strings.HasPrefix(reply, "-ERR") {
		logAOF(dbid, args...)
	}
	reply = strings.TrimSuffix(reply, "\r\n")
	sendJSON(w, map[string]string{"reply": reply})
}

func handleCommands(w http.ResponseWriter, r *http.Request) {
	names := make([]string, 0, len(commands.CommandRegistry))
	for name := range commands.CommandRegistry {
		names = append(names, name)
	}
	sendJSON(w, map[string]any{"commands": names})
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, map[string]any{"logs": logger.GetLogRing()})
}
