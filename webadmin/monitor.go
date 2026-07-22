package webadmin

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var startTime time.Time

// CPU 监控
var (
	prevCPUTicks uint64
	prevCPUTime  time.Time
	statsHistory []map[string]any
	statsMu      sync.Mutex
)

func readCPUUsage() float64 {
	data, err := os.ReadFile("/proc/self/stat")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 15 {
		return 0
	}
	utime, _ := strconv.ParseUint(fields[13], 10, 64)
	stime, _ := strconv.ParseUint(fields[14], 10, 64)
	totalTicks := utime + stime
	now := time.Now()

	var pct float64
	if prevCPUTicks > 0 {
		elapsed := now.Sub(prevCPUTime).Seconds()
		deltaTicks := totalTicks - prevCPUTicks
		pct = (float64(deltaTicks) / elapsed / float64(runtime.NumCPU())) * 100
	}
	prevCPUTicks = totalTicks
	prevCPUTime = now
	if pct > 100 { pct = 100 }
	if pct < 0 { pct = 0 }
	return pct
}

func appendStatsHistory(cpuPct, memMb float64) {
	statsMu.Lock()
	statsHistory = append(statsHistory, map[string]any{
		"time": time.Now().Format("15:04:05"),
		"cpu":  cpuPct,
		"mem":  memMb,
	})
	if len(statsHistory) > 60 {
		statsHistory = statsHistory[len(statsHistory)-60:]
	}
	statsMu.Unlock()
}

func getStatsHistory() []map[string]any {
	statsMu.Lock()
	defer statsMu.Unlock()
	result := make([]map[string]any, len(statsHistory))
	copy(result, statsHistory)
	return result
}
