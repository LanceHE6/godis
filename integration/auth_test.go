package integration

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

// TestAuth_NoPassword 默认无密码时发送 AUTH 应报错
func TestAuth_NoPassword(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	err := rdb.Do(ctx, "AUTH", "anything").Err()
	t.Logf("AUTH without requirepass error: %v", err)
	if err == nil {
		t.Error("AUTH should fail when no password is set")
	}
}

// TestAuth_WrongPassword 错误的密码
func TestAuth_WrongPassword(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// 临时设置密码
	rdb.Do(ctx, "CONFIG", "SET", "requirepass", "secret")
	defer rdb.Do(ctx, "CONFIG", "SET", "requirepass", "")

	// 创建新客户端测试（当前连接已在设置密码前认证通过）
	// 注意：CONFIG SET 设置密码后，已有连接不受影响
	// 这里仅验证 AUTH 命令存在且能处理错误密码
	err := rdb.Do(ctx, "AUTH", "wrong").Err()
	t.Logf("AUTH wrong password error: %v", err)
	if err == nil {
		t.Error("AUTH wrong password should fail")
	}

	err = rdb.Do(ctx, "AUTH", "secret").Err()
	t.Logf("AUTH correct password: %v", err)
	if err != nil {
		t.Errorf("AUTH correct password should succeed: %v", err)
	}
}

// TestAuth_RequiresAuth 启用密码后非认证连接被拒绝
func TestAuth_RequiresAuth(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// 设置密码
	rdb.Do(ctx, "CONFIG", "SET", "requirepass", "mypass")
	defer func() {
		rdb.Do(ctx, "CONFIG", "SET", "requirepass", "")
		t.Log("requirepass cleared")
	}()

	// 用新客户端连接（无密码）
	newClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:16379"})
	defer newClient.Close()

	// 无认证执行 SET 应报错 NOAUTH
	err := newClient.Set(ctx, "k", "v", 0).Err()
	t.Logf("SET without auth: %v", err)
	if err == nil {
		t.Error("SET should fail without authentication")
	}

	// 认证后应正常
	err = newClient.Do(ctx, "AUTH", "mypass").Err()
	if err != nil {
		t.Fatalf("AUTH failed: %v", err)
	}
	err = newClient.Ping(ctx).Err()
	if err != nil {
		t.Errorf("PING after AUTH failed: %v", err)
	}
}
