package integration

import (
	"context"
	"testing"
	"time"
)

func TestPublishSubscribe(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sub := rdb.Subscribe(ctx, "mychannel")
	defer sub.Close()

	// 订阅确认由 go-redis 内部处理，不阻塞等待

	// 发布消息
	n, err := rdb.Publish(ctx, "mychannel", "hello").Result()
	if err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}
	t.Logf("PUBLISH returned %d", n)
	if n != 1 {
		t.Errorf("PUBLISH = %d, want 1", n)
	}

	// 接收消息
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("receive message failed: %v", err)
	}
	t.Logf("received: channel=%s payload=%s", msg.Channel, msg.Payload)
	if msg.Channel != "mychannel" || msg.Payload != "hello" {
		t.Errorf("msg = %v, want channel=mychannel payload=hello", msg)
	}
}

func TestPSubscribe(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sub := rdb.PSubscribe(ctx, "news.*")
	defer sub.Close()

	n, err := rdb.Publish(ctx, "news.sports", "sports news").Result()
	if err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}
	t.Logf("PUBLISH = %d", n)
	if n != 1 {
		t.Errorf("PUBLISH = %d, want 1", n)
	}

	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		t.Fatalf("receive message failed: %v", err)
	}
	t.Logf("received psubscribe: pattern=%s channel=%s payload=%s", msg.Pattern, msg.Channel, msg.Payload)
	if msg.Pattern != "news.*" || msg.Channel != "news.sports" || msg.Payload != "sports news" {
		t.Errorf("msg = %v, want pattern=news.* channel=news.sports payload='sports news'", msg)
	}
}

func TestUnsubscribe(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sub := rdb.Subscribe(ctx, "ch1", "ch2")
	defer sub.Close()

	// 退订 ch1
	err := sub.Unsubscribe(ctx, "ch1")
	if err != nil {
		t.Fatalf("UNSUBSCRIBE failed: %v", err)
	}
	t.Log("unsubscribed ch1")

	// PUBSUB NUMSUB 验证
	result := rdb.Do(ctx, "PUBSUB", "NUMSUB", "ch1", "ch2").Val()
	t.Logf("PUBSUB NUMSUB = %v", result)
}

func TestPubSubChannels(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sub := rdb.Subscribe(ctx, "news", "sports")
	defer sub.Close()

	result, err := rdb.PubSubChannels(ctx, "*").Result()
	if err != nil {
		t.Fatalf("PUBSUB CHANNELS failed: %v", err)
	}
	t.Logf("PUBSUB CHANNELS * = %v", result)
	if len(result) < 2 {
		t.Errorf("PUBSUB CHANNELS = %v, want at least 2 channels", result)
	}
}

func TestPublishNoSubscribers(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	n, err := rdb.Publish(ctx, "noone", "msg").Result()
	if err != nil {
		t.Fatalf("PUBLISH failed: %v", err)
	}
	if n != 0 {
		t.Errorf("PUBLISH no subscribers = %d, want 0", n)
	}
}

func TestPubSubNumSub(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sub := rdb.Subscribe(ctx, "ch1")
	defer sub.Close()

	// 短暂等待确保注册完成
	time.Sleep(50 * time.Millisecond)

	result, err := rdb.PubSubNumSub(ctx, "ch1", "ch2").Result()
	if err != nil {
		t.Fatalf("PUBSUB NUMSUB failed: %v", err)
	}
	t.Logf("PUBSUB NUMSUB = %v", result)
	if len(result) != 2 {
		t.Fatalf("NUMSUB len = %d, want 2 (ch1 + ch2)", len(result))
	}
	if result["ch1"] != 1 {
		t.Errorf("NUMSUB ch1 = %d, want 1", result["ch1"])
	}
}

func TestPubSubNumPat(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sub := rdb.PSubscribe(ctx, "n*")
	defer sub.Close()

	time.Sleep(50 * time.Millisecond)

	n, err := rdb.PubSubNumPat(ctx).Result()
	if err != nil {
		t.Fatalf("PUBSUB NUMPAT failed: %v", err)
	}
	t.Logf("PUBSUB NUMPAT = %d", n)
	if n < 1 {
		t.Errorf("NUMPAT = %d, want >= 1", n)
	}
}
