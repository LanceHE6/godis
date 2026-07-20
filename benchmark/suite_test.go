package main

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

// 微基准测试：go test -bench . -benchtime=3s -count=3 ./benchmark/

var benchClient = redis.NewClient(&redis.Options{Addr: "127.0.0.1:6389"})
var benchCtx = context.Background()
var benchData = "benchvalue"

func init() {
	benchClient.Set(benchCtx, "bench:get", benchData, 0)
	benchClient.Del(benchCtx, "bench:list", "bench:sadd", "bench:hash", "bench:zset")
}

// ---- single ops ----

func BenchmarkPING(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchClient.Ping(benchCtx)
	}
}

func BenchmarkSET(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchClient.Set(benchCtx, "bench:set", benchData, 0)
	}
}

func BenchmarkGET(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchClient.Get(benchCtx, "bench:get")
	}
}

func BenchmarkLPUSH(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchClient.LPush(benchCtx, "bench:list", benchData)
	}
}

func BenchmarkLPOP(b *testing.B) {
	// preload
	for i := 0; i < b.N; i++ {
		benchClient.LPush(benchCtx, "bench:list", benchData)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchClient.LPop(benchCtx, "bench:list")
	}
}

func BenchmarkSADD(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchClient.SAdd(benchCtx, "bench:sadd", benchData)
	}
}

func BenchmarkHSET(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchClient.HSet(benchCtx, "bench:hash", "f", benchData)
	}
}

func BenchmarkHGET(b *testing.B) {
	benchClient.HSet(benchCtx, "bench:hash", "f", benchData)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchClient.HGet(benchCtx, "bench:hash", "f")
	}
}

func BenchmarkZADD(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchClient.ZAdd(benchCtx, "bench:zset", redis.Z{Score: float64(i), Member: benchData})
	}
}

// ---- pipeline ops ----

func BenchmarkSET_Pipeline_16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pipe := benchClient.Pipeline()
		for j := 0; j < 16; j++ {
			pipe.Set(benchCtx, "bench:pset", benchData, 0)
		}
		pipe.Exec(benchCtx)
	}
}

func BenchmarkGET_Pipeline_16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pipe := benchClient.Pipeline()
		for j := 0; j < 16; j++ {
			pipe.Get(benchCtx, "bench:get")
		}
		pipe.Exec(benchCtx)
	}
}
