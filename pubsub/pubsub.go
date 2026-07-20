package pubsub

import (
	"fmt"
	"net"
	"sync"

	"godis/protocol"
)

// Client 订阅客户端
type Client struct {
	conn     net.Conn
	mu       sync.Mutex
	channels map[string]struct{} // 已订阅频道
	patterns map[string]struct{} // 已订阅模式
}

// NewClient 创建客户端
func NewClient(conn net.Conn) *Client {
	return &Client{
		conn:     conn,
		channels: make(map[string]struct{}),
		patterns: make(map[string]struct{}),
	}
}

// Write 线程安全写 RESP 数据到连接
func (c *Client) Write(data string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Write([]byte(data))
	}
}

// SubscribeCount 返回已订阅频道数
func (c *Client) SubscribeCount() int {
	return len(c.channels)
}

// PatternCount 返回已订阅模式数
func (c *Client) PatternCount() int {
	return len(c.patterns)
}

// IsSubscribed 是否处于订阅模式（有频道或模式）
func (c *Client) IsSubscribed() bool {
	return len(c.channels) > 0 || len(c.patterns) > 0
}

// UnsubscribeAll 退订所有频道和模式
func (c *Client) UnsubscribeAll() {
	c.channels = make(map[string]struct{})
	c.patterns = make(map[string]struct{})
}

// Hub 全局发布订阅管理器
type Hub struct {
	mu       sync.RWMutex
	channels map[string]map[*Client]struct{} // channel → clients
	patterns map[string]map[*Client]struct{} // pattern → clients
}

var GlobalHub = &Hub{
	channels: make(map[string]map[*Client]struct{}),
	patterns: make(map[string]map[*Client]struct{}),
}

// Subscribe 订阅频道，返回确认消息列表
func (h *Hub) Subscribe(c *Client, channels ...string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var confirms []string
	for _, ch := range channels {
		if _, ok := c.channels[ch]; ok {
			continue // 已订阅
		}
		c.channels[ch] = struct{}{}
		if h.channels[ch] == nil {
			h.channels[ch] = make(map[*Client]struct{})
		}
		h.channels[ch][c] = struct{}{}
		confirms = append(confirms, protocol.MakeArray([]string{
			protocol.MakeBulkString("subscribe"),
			protocol.MakeBulkString(ch),
			protocol.MakeInt(len(c.channels)),
		}))
	}
	return confirms
}

// PSubscribe 模式订阅
func (h *Hub) PSubscribe(c *Client, patterns ...string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var confirms []string
	for _, p := range patterns {
		if _, ok := c.patterns[p]; ok {
			continue
		}
		c.patterns[p] = struct{}{}
		if h.patterns[p] == nil {
			h.patterns[p] = make(map[*Client]struct{})
		}
		h.patterns[p][c] = struct{}{}
		confirms = append(confirms, protocol.MakeArray([]string{
			protocol.MakeBulkString("psubscribe"),
			protocol.MakeBulkString(p),
			protocol.MakeInt(len(c.patterns)),
		}))
	}
	return confirms
}

// Unsubscribe 退订频道，无参数则退订全部
func (h *Hub) Unsubscribe(c *Client, channels ...string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var confirms []string
	if len(channels) == 0 {
		// 退订全部
		for ch := range c.channels {
			delete(h.channels[ch], c)
			if len(h.channels[ch]) == 0 {
				delete(h.channels, ch)
			}
			confirms = append(confirms, protocol.MakeArray([]string{
				protocol.MakeBulkString("unsubscribe"),
				protocol.MakeBulkString(ch),
				protocol.MakeInt(0),
			}))
		}
		c.channels = make(map[string]struct{})
	} else {
		for _, ch := range channels {
			if _, ok := c.channels[ch]; ok {
				delete(c.channels, ch)
				delete(h.channels[ch], c)
				if len(h.channels[ch]) == 0 {
					delete(h.channels, ch)
				}
			}
			confirms = append(confirms, protocol.MakeArray([]string{
				protocol.MakeBulkString("unsubscribe"),
				protocol.MakeBulkString(ch),
				protocol.MakeInt(len(c.channels)),
			}))
		}
	}
	return confirms
}

// PUnsubscribe 退订模式
func (h *Hub) PUnsubscribe(c *Client, patterns ...string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	var confirms []string
	if len(patterns) == 0 {
		for p := range c.patterns {
			delete(h.patterns[p], c)
			if len(h.patterns[p]) == 0 {
				delete(h.patterns, p)
			}
			confirms = append(confirms, protocol.MakeArray([]string{
				protocol.MakeBulkString("punsubscribe"),
				protocol.MakeBulkString(p),
				protocol.MakeInt(0),
			}))
		}
		c.patterns = make(map[string]struct{})
	} else {
		for _, p := range patterns {
			delete(c.patterns, p)
			delete(h.patterns[p], c)
			if len(h.patterns[p]) == 0 {
				delete(h.patterns, p)
			}
			confirms = append(confirms, protocol.MakeArray([]string{
				protocol.MakeBulkString("punsubscribe"),
				protocol.MakeBulkString(p),
				protocol.MakeInt(len(c.patterns)),
			}))
		}
	}
	return confirms
}

// Publish 向频道发布消息，返回接收者数量
func (h *Hub) Publish(channel, message string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msgResp := protocol.MakeArray([]string{
		protocol.MakeBulkString("message"),
		protocol.MakeBulkString(channel),
		protocol.MakeBulkString(message),
	})

	count := 0

	// 精确匹配的订阅者
	if subs, ok := h.channels[channel]; ok {
		for c := range subs {
			c.Write(msgResp)
			count++
		}
	}

	// 模式匹配
	for pattern, subs := range h.patterns {
		if matchGlob(channel, pattern) {
			pmResp := protocol.MakeArray([]string{
				protocol.MakeBulkString("pmessage"),
				protocol.MakeBulkString(pattern),
				protocol.MakeBulkString(channel),
				protocol.MakeBulkString(message),
			})
			for c := range subs {
				c.Write(pmResp)
				count++
			}
		}
	}
	return count
}

// Channels 返回活跃频道列表（支持模式过滤）
func (h *Hub) Channels(pattern string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, 0)
	for ch := range h.channels {
		if pattern == "" || matchGlob(ch, pattern) {
			result = append(result, ch)
		}
	}
	return result
}

// NumSub 返回指定频道的订阅数
func (h *Hub) NumSub(channels ...string) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]string, 0, len(channels)*2)
	for _, ch := range channels {
		result = append(result, protocol.MakeBulkString(ch))
		if subs, ok := h.channels[ch]; ok {
			result = append(result, protocol.MakeInt(len(subs)))
		} else {
			result = append(result, protocol.MakeInt(0))
		}
	}
	return result
}

// NumPat 返回模式订阅总数
func (h *Hub) NumPat() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	count := 0
	for _, subs := range h.patterns {
		count += len(subs)
	}
	return count
}

// matchGlob 简单 glob 匹配，支持 * 和 ?
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

// Disconnect 客户端断开时清理订阅
func (h *Hub) Disconnect(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range c.channels {
		delete(h.channels[ch], c)
		if len(h.channels[ch]) == 0 {
			delete(h.channels, ch)
		}
	}
	for p := range c.patterns {
		delete(h.patterns[p], c)
		if len(h.patterns[p]) == 0 {
			delete(h.patterns, p)
		}
	}
	c.channels = nil
	c.patterns = nil
}

// 辅助方法返回错误（不在 pubsub 包中引用 protocol）
var MakeErrorFn = func(format string, args ...interface{}) string {
	return fmt.Sprintf("-ERR "+format+"\r\n", args...)
}
