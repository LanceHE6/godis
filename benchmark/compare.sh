#!/bin/bash
set -e

REQUESTS=100000
CLIENTS=50
BENCH_DIR="$(cd "$(dirname "$0")" && pwd)"
RESULTS_DIR="$BENCH_DIR/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

REDIS_PORT="${1:-6379}"
GODIS_PORT="${2:-6389}"

echo "========== Godis vs Redis Benchmark =========="
echo "Requests: $REQUESTS  Clients: $CLIENTS"
echo "Redis port: $REDIS_PORT    Godis port: $GODIS_PORT"
echo ""

# 检查 redis 是否在对应端口运行
check_redis() {
    local port=$1
    redis-cli -p "$port" PING > /dev/null 2>&1 && return 0 || return 1
}

# ---- Redis ----
if check_redis "$REDIS_PORT"; then
    echo "[Redis :$REDIS_PORT] Starting benchmarks..."
    echo ""
    go run "$BENCH_DIR" -host 127.0.0.1 -port "$REDIS_PORT" -n "$REQUESTS" -c "$CLIENTS" 2>&1 | tee "$RESULTS_DIR/redis_${TIMESTAMP}.log"
else
    echo "[Redis :$REDIS_PORT] Not running, skipping."
    echo "  Start one: redis-server --port $REDIS_PORT --daemonize yes"
fi

echo ""

# ---- Godis ----
if check_redis "$GODIS_PORT"; then
    echo "[Godis :$GODIS_PORT] Starting benchmarks..."
    echo ""
    go run "$BENCH_DIR" -host 127.0.0.1 -port "$GODIS_PORT" -n "$REQUESTS" -c "$CLIENTS" 2>&1 | tee "$RESULTS_DIR/godis_${TIMESTAMP}.log"
else
    echo "[Godis :$GODIS_PORT] Not running, skipping."
    echo "  Start godis: ./godis --config etc/godis.yaml"
fi

echo ""
echo "========== Comparison =========="
echo ""
printf "%-12s %15s %15s %10s\n" "Command" "Godis (ops/s)" "Redis (ops/s)" "Ratio"
printf "%s\n" "----------------------------------------------------------"

# 如果两边都存在，做对比
GODIS_LOG="$RESULTS_DIR/godis_${TIMESTAMP}.log"
REDIS_LOG="$RESULTS_DIR/redis_${TIMESTAMP}.log"

if [ -f "$GODIS_LOG" ] && [ -f "$REDIS_LOG" ]; then
    for cmd in PING SET GET LPUSH RPOP SADD SPOP HSET HGET ZADD ZRANGE PUBLISH; do
        godis_val=$(grep "^$cmd " "$GODIS_LOG" | awk '{print $5}')
        redis_val=$(grep "^$cmd " "$REDIS_LOG" | awk '{print $5}')
        if [ -n "$godis_val" ] && [ -n "$redis_val" ] && [ "$redis_val" != "0.0" ]; then
            ratio=$(awk "BEGIN {printf \"%.2f\", $godis_val / $redis_val}")
            printf "%-12s %15s %15s %10sx\n" "$cmd" "$godis_val" "$redis_val" "$ratio"
        fi
    done
else
    echo "(Run both godis and redis to see comparison)"
fi

echo ""
echo "Full logs:"
echo "  Godis: ${GODIS_LOG:-"N/A"}"
echo "  Redis: ${REDIS_LOG:-"N/A"}"
