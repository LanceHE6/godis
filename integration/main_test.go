package integration

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

const testPort = 16379

var rdb *redis.Client
var cmd *exec.Cmd

func TestMain(m *testing.M) {
	fmt.Println("[SETUP] building godis binary...")
	bin := filepath.Join(os.TempDir(), "godis-test")
	rootDir := filepath.Join(".", "..")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = rootDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[SETUP] build failed: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(bin)
	fmt.Printf("[SETUP] binary built: %s\n", bin)

	tmpDir := os.TempDir()
	aofFile := filepath.Join(tmpDir, "godis-integration.aof")
	logFile := filepath.Join(tmpDir, "godis-integration.log")
	cfgFile := filepath.Join(tmpDir, "godis-integration.yaml")

	cfgContent := fmt.Sprintf(`bind: 127.0.0.1
port: %d
databases: 16
aof_file: %s
log_file: %s
log_level: error
`, testPort, aofFile, logFile)
	os.WriteFile(cfgFile, []byte(cfgContent), 0644)
	os.Remove(aofFile)
	os.Remove(logFile)

	fmt.Printf("[SETUP] starting godis server on port %d...\n", testPort)
	cmd = exec.Command(bin, "--config", cfgFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "[SETUP] start failed: %v\n", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", testPort)
	if err := waitForPort(addr, 5*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "[SETUP] server not ready: %v\n", err)
		cmd.Process.Kill()
		os.Exit(1)
	}
	fmt.Printf("[SETUP] server ready at %s (pid=%d)\n", addr, cmd.Process.Pid)

	rdb = redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()

	code := m.Run()

	fmt.Println("[TEARDOWN] cleaning up...")
	rdb.Do(context.Background(), "FLUSHALL")
	rdb.Close()
	cmd.Process.Kill()
	cmd.Wait()
	os.Remove(cfgFile)
	os.Remove(aofFile)
	os.Remove(logFile)
	fmt.Println("[TEARDOWN] done")
	os.Exit(code)
}

func waitForPort(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("port %s not reachable after %s", addr, timeout)
}

func cleanDB(t *testing.T) {
	t.Helper()
	rdb.Do(context.Background(), "FLUSHALL")
}
