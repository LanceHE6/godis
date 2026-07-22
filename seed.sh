#!/bin/bash
# godis-seed.sh — 向 godis 写入测试数据
# 用法: bash seed.sh [port]

PORT="${1:-6389}"
CLI="redis-cli -p $PORT"

echo "==> 写入字符串 (5 keys)"
$CLI SET user:1:name "Alice"
$CLI SET user:2:name "Bob"
$CLI SET user:3:name "Charlie"
$CLI SET config:app "godis"
$CLI SET counter:visits 0

echo "==> 写入 Hash (2 keys)"
$CLI HSET user:1:profile age 25 city "Beijing" role admin
$CLI HSET session:abc token "xyz123" ip "10.0.0.1" ttl 3600

echo "==> 写入 List (2 keys)"
$CLI LPUSH queue:tasks "task3" "task2" "task1"
$CLI RPUSH log:recent "2024-01-01" "2024-01-02" "2024-01-03"

echo "==> 写入 Set (2 keys)"
$CLI SADD tags:golang "高性能" "并发" "简洁"
$CLI SADD tags:redis "缓存" "持久化" "高性能"

echo "==> 写入 ZSet (2 keys)"
$CLI ZADD leaderboard 100 "player1" 200 "player2" 300 "player3"
$CLI ZADD scores:math 95 "Alice" 87 "Bob" 92 "Charlie"

echo "==> 设置 TTL (3 keys)"
$CLI EXPIRE session:abc 60
$CLI EXPIRE user:3:name 10
$CLI EXPIRE counter:visits 300

echo "==> 切换 DB1 写入隔离数据"
$CLI SELECT 1
$CLI SET db1:key "from-db1"
$CLI SELECT 0

echo ""
echo "数据写入完成。验证："
echo "  $CLI KEYS '*'                # 列出所有 key"
echo "  $CLI TTL session:abc         # 查看 TTL"
echo "  $CLI TTL user:3:name         # 应剩余 ~10s"
echo "  $CLI DBSIZE                  # key 数量"
echo ""
echo "删除 AOF 后重启，TTL 不应重置："
echo "  rm -f ./data/godis.aof && ./godis --config etc/godis.yaml"
echo "  $CLI TTL counter:visits       # 应 < 300，而非正好 300"
