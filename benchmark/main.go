package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

type benchResult struct {
	Name    string
	Ops     int64
	Elapsed time.Duration
	Errors  int64
}

func main() {
	host := flag.String("host", "127.0.0.1", "redis host")
	port := flag.Int("port", 6389, "redis port")
	clients := flag.Int("c", 50, "concurrent clients")
	requests := flag.Int("n", 100000, "total requests")
	payload := flag.Int("d", 3, "data size in bytes")
	_ = flag.String("P", "1", "pipeline size (reserved for future use)")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)
	val := strings.Repeat("x", *payload)

	tests := []struct {
		name string
		fn   func(client *redis.Client) error
	}{
		{"PING", func(c *redis.Client) error { return c.Ping(context.Background()).Err() }},
		{"SET", func(c *redis.Client) error { return c.Set(context.Background(), "bench:set", val, 0).Err() }},
		{"GET", func(c *redis.Client) error {
			_, err := c.Get(context.Background(), "bench:set").Result()
			return err
		}},
		{"LPUSH", func(c *redis.Client) error { return c.LPush(context.Background(), "bench:list", val).Err() }},
		{"RPOP", func(c *redis.Client) error {
			_, err := c.RPop(context.Background(), "bench:list").Result()
			if err == redis.Nil {
				return nil
			}
			return err
		}},
		{"SADD", func(c *redis.Client) error { return c.SAdd(context.Background(), "bench:sset", val).Err() }},
		{"SPOP", func(c *redis.Client) error {
			_, err := c.SPop(context.Background(), "bench:sset").Result()
			if err == redis.Nil {
				return nil
			}
			return err
		}},
		{"HSET", func(c *redis.Client) error { return c.HSet(context.Background(), "bench:hash", "f", val).Err() }},
		{"HGET", func(c *redis.Client) error { return c.HGet(context.Background(), "bench:hash", "f").Err() }},
		{"ZADD", func(c *redis.Client) error {
			return c.ZAdd(context.Background(), "bench:zset", redis.Z{Score: float64(time.Now().UnixNano()), Member: val}).Err()
		}},
		{"ZRANGE", func(c *redis.Client) error {
			_, err := c.ZRange(context.Background(), "bench:zset", 0, 9).Result()
			return err
		}},
		{"PUBLISH", func(c *redis.Client) error { return c.Publish(context.Background(), "bench:chan", val).Err() }},
	}

	rdb := redis.NewClient(&redis.Options{Addr: addr, PoolSize: *clients})
	defer rdb.Close()

	// 预热 SET（确保 GET 有数据）
	rdb.Set(context.Background(), "bench:set", val, 0)
	rdb.Del(context.Background(), "bench:list", "bench:set", "bench:hash", "bench:zset")

	var results []benchResult

	for _, t := range tests {
		reqPerClient := *requests / *clients
		var ops, errs int64
		start := time.Now()

		var wg sync.WaitGroup
		for i := 0; i < *clients; i++ {
			wg.Add(1)
			go func(fn func(c *redis.Client) error) {
				defer wg.Done()
				c := redis.NewClient(&redis.Options{Addr: addr})
				defer c.Close()

				for j := 0; j < reqPerClient; j++ {
					if err := fn(c); err != nil {
						atomic.AddInt64(&errs, 1)
					} else {
						atomic.AddInt64(&ops, 1)
					}
				}
			}(t.fn)
		}
		wg.Wait()

		r := benchResult{
			Name:    t.name,
			Ops:     ops,
			Elapsed: time.Since(start),
			Errors:  errs,
		}
		results = append(results, r)
		fmt.Printf("%-12s %10d ops  %8s  %6.1f ops/sec  errors: %d\n",
			r.Name, r.Ops, r.Elapsed.Round(time.Millisecond),
			float64(r.Ops)/r.Elapsed.Seconds(), r.Errors)
	}

	// summary
	fmt.Println("\n--- Pipeline + Concurrent Summary ---")
	for _, r := range results {
		fmt.Printf("%-12s %10d ops/sec\n", r.Name, int64(float64(r.Ops)/r.Elapsed.Seconds()))
	}
}

// Helper to run if any flag errors
func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go run benchmark/main.go [flags]\n\n")
		flag.PrintDefaults()
	}
}
