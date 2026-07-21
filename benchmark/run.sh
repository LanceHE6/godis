#!/bin/bash
set -e

PORT="${1:-6389}"
HOST="127.0.0.1"
REQUESTS=100000
CLIENTS=50
BENCH_DIR="$(cd "$(dirname "$0")" && pwd)"
RESULTS_DIR="$BENCH_DIR/results"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

REDIS_BENCH=$(which redis-benchmark 2>/dev/null || echo "")

echo "========== Godis Benchmark Suite =========="
echo "Host: $HOST  Port: $PORT  Requests: $REQUESTS  Clients: $CLIENTS"
echo ""

# ---- custom Go benchmark ----
echo "[1/3] Custom benchmark tool..."
echo ""
go run "$BENCH_DIR" -host "$HOST" -port "$PORT" -n "$REQUESTS" -c "$CLIENTS" 2>&1 | tee "$RESULTS_DIR/custom_${TIMESTAMP}.log"

echo ""
echo "[2/3] Go micro-benchmarks..."
go test -bench . -benchtime=1s -count=1 "$BENCH_DIR" 2>&1 | tee "$RESULTS_DIR/micro_${TIMESTAMP}.log"

# ---- redis-benchmark (if available) ----
if [ -n "$REDIS_BENCH" ]; then
    echo ""
    echo "[3/3] redis-benchmark (official tool)..."
    echo ""

    echo "--- SET/GET ---"
    $REDIS_BENCH -h "$HOST" -p "$PORT" -t set,get -n "$REQUESTS" -c "$CLIENTS" -q 2>&1 | tee -a "$RESULTS_DIR/redis_${TIMESTAMP}.log"

    echo ""
    echo "--- Pipeline (P=16) ---"
    $REDIS_BENCH -h "$HOST" -p "$PORT" -t set,get -n "$((REQUESTS * 5))" -c "$CLIENTS" -P 16 -q 2>&1 | tee -a "$RESULTS_DIR/redis_${TIMESTAMP}.log"

    echo ""
    echo "--- List / Set / Hash / ZSet ---"
    $REDIS_BENCH -h "$HOST" -p "$PORT" -t lpush,rpop,sadd,spop,hset,hget,zadd,zrange -n "$REQUESTS" -c "$CLIENTS" -q 2>&1 | tee -a "$RESULTS_DIR/redis_${TIMESTAMP}.log"

    echo ""
    echo "--- Large payload (128 bytes) ---"
    $REDIS_BENCH -h "$HOST" -p "$PORT" -t set,get -n "$REQUESTS" -c "$CLIENTS" -d 128 -q 2>&1 | tee -a "$RESULTS_DIR/redis_${TIMESTAMP}.log"
else
    echo ""
    echo "[3/3] redis-benchmark not found, skipping."
    echo "      Install: apt install redis-tools / brew install redis"
fi

echo ""
echo "========== Done =========="
echo "Results saved to $RESULTS_DIR/"
echo ""
echo "Latest summary:"
echo "  Custom:  $(ls -t "$RESULTS_DIR"/custom_*.log | head -1)"
echo "  Micro:   $(ls -t "$RESULTS_DIR"/micro_*.log | head -1)"
echo "  Redis:   $(ls -t "$RESULTS_DIR"/redis_*.log 2>/dev/null | head -1)"
