# Godis Benchmark Suite

## 快速开始

```bash
# 确保 godis 正在运行
./godis --config etc/godis.yaml &

# 一键运行全量压测（默认 127.0.0.1:6389）
bash benchmark/run.sh

# 自定义端口
bash benchmark/run.sh 6379
```

## 工具

### 1. 自定义压测工具

Go 编写的并发压测器，支持 Pipeline 和多客户端模拟：

```bash
go run benchmark/ -host 127.0.0.1 -port 6389 -n 100000 -c 50
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-host` | `127.0.0.1` | godis 地址 |
| `-port` | `6389` | godis 端口 |
| `-n` | `100000` | 总请求数 |
| `-c` | `50` | 并发客户端数 |
| `-d` | `3` | payload 大小（字节） |

### 2. Go 微基准

```bash
# 单命令基准
go test -bench . -benchtime=3s ./benchmark/

# 指定命令
go test -bench "SET|GET" -benchtime=3s ./benchmark/

# 多次取样
go test -bench . -benchtime=3s -count=5 ./benchmark/
```

### 3. redis-benchmark（需安装）

```bash
make benchmark

# 或直接调用
redis-benchmark -h 127.0.0.1 -p 6389 -t set,get -n 100000 -q
redis-benchmark -h 127.0.0.1 -p 6389 -t set,get -n 500000 -P 16 -q
redis-benchmark -h 127.0.0.1 -p 6389 -t lpush,rpop,sadd,spop,hset,hget -n 100000 -q
```

## 结果

所有压测结果自动保存到 `benchmark/results/` 目录，文件名包含时间戳：

- `custom_*.log` — 自定义工具输出
- `micro_*.log` — Go benchmark 输出
- `redis_*.log` — redis-benchmark 输出

## 对比 Redis

在同一台机器上运行可对比 godis vs Redis 性能：

```bash
# godis
bash benchmark/run.sh 6389

# redis（需要先安装 redis-server）
bash benchmark/run.sh 6379
```

## 典型结果（本地 MacBook + godis @ 6389, c=50, n=100000）

```
PING         100000 ops     524ms  190839 ops/sec
SET          100000 ops     728ms  137362 ops/sec
GET          100000 ops     631ms  158478 ops/sec
LPUSH         99999 ops    1024ms   97655 ops/sec
SADD         100000 ops     687ms  145560 ops/sec
HSET         100000 ops     712ms  140449 ops/sec
HGET         100000 ops     589ms  169779 ops/sec
ZADD         100000 ops     891ms  112233 ops/sec
ZRANGE       100000 ops     677ms  147710 ops/sec
PUBLISH      100000 ops     521ms  191938 ops/sec
```
