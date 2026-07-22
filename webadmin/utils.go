package webadmin

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"godis/datastore"
	"godis/types"
)

func sendJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func logAOF(dbid int, args ...string) {
	if aof != nil {
		aof.WriteCmd(args, dbid)
	}
}

func logAPI(r *http.Request) {
	if r.Method != "GET" {
		log.Info("%s %s", r.Method, r.URL.String())
	}
}

func getDB(r *http.Request) *datastore.GodisDB {
	dbid, _ := strconv.Atoi(r.URL.Query().Get("db"))
	if dbid < 0 || dbid >= len(dbs) {
		dbid = 0
	}
	return dbs[dbid]
}

func getDBID(r *http.Request) int {
	dbid, _ := strconv.Atoi(r.URL.Query().Get("db"))
	if dbid < 0 || dbid >= len(dbs) {
		return 0
	}
	return dbid
}

func parseArgs(cmd string) []string {
	return strings.Fields(cmd)
}

func matchGlob(s, pattern string) bool {
	si, pi := 0, 0
	starSi, starPi := -1, -1
	for si < len(s) {
		if pi < len(pattern) && (pattern[pi] == '?' || pattern[pi] == s[si]) {
			si++
			pi++
		} else if pi < len(pattern) && pattern[pi] == '*' {
			starPi = pi
			starSi = si
			pi++
		} else if starPi != -1 {
			pi = starPi + 1
			starSi++
			si = starSi
		} else {
			return false
		}
	}
	for pi < len(pattern) && pattern[pi] == '*' {
		pi++
	}
	return pi == len(pattern)
}

func typeName(dt types.DataType) string {
	switch dt {
	case types.TypeString:
		return "string"
	case types.TypeHash:
		return "hash"
	case types.TypeList:
		return "list"
	case types.TypeSet:
		return "set"
	case types.TypeZSet:
		return "zset"
	default:
		return "unknown"
	}
}
