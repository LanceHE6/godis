package pubsub

import (
	"testing"
)

func TestHubSubscribe(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	confirms := h.Subscribe(c, "news", "sports")
	if len(confirms) != 2 {
		t.Fatalf("Subscribe confirms len = %d, want 2", len(confirms))
	}
	if !c.IsSubscribed() {
		t.Fatal("client should be subscribed")
	}
	if c.SubscribeCount() != 2 {
		t.Fatalf("SubscribeCount = %d, want 2", c.SubscribeCount())
	}
}

func TestHubUnsubscribe(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	h.Subscribe(c, "news", "sports")
	h.Unsubscribe(c, "news")

	if c.IsSubscribed() {
		if c.SubscribeCount() != 1 {
			t.Fatalf("after unsub news, count = %d, want 1", c.SubscribeCount())
		}
	}
}

func TestHubUnsubscribeAll(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	h.Subscribe(c, "news", "sports")
	h.Unsubscribe(c) // 无参数 = 退订全部

	if c.IsSubscribed() {
		t.Fatal("client should not be subscribed after UnsubscribeAll")
	}
}

func TestHubPublish(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	h.Subscribe(c, "news")
	count := h.Publish("news", "hello world")
	if count != 1 {
		t.Fatalf("Publish count = %d, want 1", count)
	}
}

func TestHubPSubscribe(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	confirms := h.PSubscribe(c, "news.*")
	if len(confirms) != 1 {
		t.Fatalf("PSubscribe confirms = %d, want 1", len(confirms))
	}
	count := h.Publish("news.sports", "hello")
	if count != 1 {
		t.Fatalf("PSubscribe Publish count = %d, want 1", count)
	}
}

func TestHubPUnsubscribe(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	h.PSubscribe(c, "news.*")
	h.PUnsubscribe(c, "news.*")
	if c.IsSubscribed() {
		t.Fatal("client should not be subscribed after PUnsubscribe")
	}
}

func TestHubPubSubChannels(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	h.Subscribe(c, "news", "sports", "music")
	chans := h.Channels("")
	if len(chans) != 3 {
		t.Fatalf("Channels len = %d, want 3", len(chans))
	}
}

func TestHubPubSubNumSub(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	h.Subscribe(c, "news")
	result := h.NumSub("news", "sports")
	// result should be [news, 1, sports, 0]
	if len(result) != 4 {
		t.Fatalf("NumSub len = %d, want 4", len(result))
	}
}

func TestHubNumPat(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	h.PSubscribe(c, "n*")
	if h.NumPat() != 1 {
		t.Fatalf("NumPat = %d, want 1", h.NumPat())
	}
}

func TestHubDisconnect(t *testing.T) {
	h := &Hub{channels: make(map[string]map[*Client]struct{}), patterns: make(map[string]map[*Client]struct{})}
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}

	h.Subscribe(c, "news")
	h.Disconnect(c)

	if c.channels != nil || c.patterns != nil {
		t.Fatal("Disconnect should nil out channels/patterns")
	}
	if len(h.channels) != 0 {
		t.Fatal("Hub should have no channels after Disconnect")
	}
}

func TestClientIsSubscribed(t *testing.T) {
	c := &Client{channels: make(map[string]struct{}), patterns: make(map[string]struct{})}
	if c.IsSubscribed() {
		t.Fatal("new client should not be subscribed")
	}
	c.channels["x"] = struct{}{}
	if !c.IsSubscribed() {
		t.Fatal("client with channel should be subscribed")
	}
}
