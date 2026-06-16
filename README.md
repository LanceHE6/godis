# Godis

> 使用 Go 语言编写的 Redis 兼容内存键值数据库。

Godis 是一款轻量级、基于 AI 开发的、以学习为目的的键值数据库，支持 [RESP 协议](https://redis.io/docs/latest/develop/reference/protocol-spec/)，与标准 Redis 客户端（redis-cli、go-redis 等）完全兼容。

## 特性

- **Redis 兼容命令**：覆盖字符串、键操作、数据库、服务器、内存等类别
- **RESP 协议**：同时支持标准 `*数组` 格式和内联命令
- **16 个逻辑数据库**：通过 SELECT 实现多数据库隔离
- **AOF 持久化**：Append-Only File，混合二进制快照 + 文本命令增量日志
- **自动重写**：当 AOF 文件超过阈值时后台自动压缩
- **键过期**：支持 TTL，访问时惰性删除 + 定期 GC 清理
- **Redis 客户端兼容**：支持 redis-cli、go-redis 等所有 RESP 客户端

## 快速开始

### 构建

```bash
go build -o godis .
```

### 运行

```bash
./godis --config ./etc/godis.yaml
```

配置文件不存在时会自动生成默认配置，`--config` 参数可省略。

### 连接

```bash
redis-cli -p 6389
```

```
127.0.0.1:6389> PING
PONG
127.0.0.1:6389> SET foo bar
OK
127.0.0.1:6389> GET foo
"bar"
```

## 命令列表

### 字符串

| 命令     | 语法                              | 说明                         |
|----------|-----------------------------------|------------------------------|
| `SET`    | `SET key value [EX seconds]`      | 设置字符串值，可选过期时间    |
| `GET`    | `GET key`                         | 获取字符串值                  |
| `APPEND` | `APPEND key value`                | 追加到已有字符串或创建新键    |

### 键操作

| 命令      | 语法                                            | 说明                            |
|-----------|-------------------------------------------------|---------------------------------|
| `DEL`     | `DEL key [key ...]`                             | 删除一个或多个键                 |
| `EXISTS`  | `EXISTS key [key ...]`                          | 统计存在的键数量                 |
| `EXPIRE`  | `EXPIRE key seconds`                            | 设置过期时间（秒）               |
| `PEXPIRE` | `PEXPIRE key milliseconds`                      | 设置过期时间（毫秒）             |
| `TTL`     | `TTL key`                                       | 查看剩余过期时间（秒）            |
| `PTTL`    | `PTTL key`                                      | 查看剩余过期时间（毫秒）          |
| `PERSIST` | `PERSIST key`                                   | 移除过期时间，永久保存            |
| `MOVE`    | `MOVE key db`                                   | 将键移动到另一个数据库            |
| `TYPE`    | `TYPE key`                                      | 查看键的数据类型                  |
| `TOUCH`   | `TOUCH key [key ...]`                           | 更新最后访问时间                  |
| `UNLINK`  | `UNLINK key [key ...]`                          | 异步删除键                        |
| `SORT`    | `SORT key [ASC\|DESC] [ALPHA] [LIMIT off count]`| 对列表/集合/有序集合排序          |
| `SCAN`    | `SCAN cursor [MATCH pattern] [COUNT count] [TYPE type]` | 增量迭代键空间            |
| `DBSIZE`  | `DBSIZE`                                        | 返回当前数据库键数量              |
| `FLUSHDB` | `FLUSHDB`                                       | 清空当前数据库                    |
| `FLUSHALL`| `FLUSHALL`                                      | 清空所有数据库                    |

### 数据库

| 命令     | 语法         | 说明                    |
|----------|-------------|------------------------|
| `SELECT` | `SELECT db` | 切换数据库（0-15）      |

### 服务器

| 命令          | 语法                                        | 说明                  |
|---------------|---------------------------------------------|-----------------------|
| `COMMAND`     | `COMMAND [COUNT\|INFO [cmd ...]\|GETKEYS cmd args]` | 命令信息查询    |
| `CONFIG`      | `CONFIG GET\|SET\|REWRITE\|RESETSTAT`        | 配置管理              |
| `INFO`        | `INFO [section]`                             | 服务器信息和统计       |
| `BGREWRITEAOF`| `BGREWRITEAOF`                               | 触发 AOF 混合重写     |

### 连接

| 命令       | 语法           | 说明                               |
|------------|----------------|------------------------------------|
| `PING`     | `PING [msg]`   | 测试连接，返回 PONG 或指定消息      |
| `TIME`     | `TIME`         | 返回服务器时间（秒 + 微秒）         |
| `SHUTDOWN` | `SHUTDOWN`     | 关闭服务器                         |

### 内存

| 命令     | 语法                   | 说明                  |
|----------|------------------------|-----------------------|
| `MEMORY` | `MEMORY STATS\|USAGE`  | 内存统计和键占用估算   |

## 项目结构

```
godis/
├── main.go                  # 入口：配置加载、日志初始化、数据库创建、AOF、启动服务
├── commands/                # 命令处理器（按类别分文件）
│   ├── router.go            #   命令注册、Execute 分发、Flag 常量定义
│   ├── strings.go           #   SET、GET、APPEND
│   ├── keys.go              #   DEL、EXISTS、EXPIRE、SORT、SCAN、FLUSHDB 等
│   ├── connection.go        #   SHUTDOWN、TIME
│   ├── handshake.go         #   PING
│   ├── database.go          #   SELECT
│   ├── server.go            #   INFO、COMMAND、CONFIG、BGREWRITEAOF
│   └── memory.go            #   MEMORY STATS/USAGE
├── config/                  # YAML 配置（加载、保存、默认值）
├── datastore/               # 核心存储引擎
│   ├── db.go                #   GodisDB：键值 CRUD、过期处理、GC 协程
│   ├── aof.go               #   AOF 日志：追加写入、混合重写、自动重写协程
│   └── gdb.go               #   Gob 二进制快照（序列化/反序列化）
├── recovery/                # 启动时 AOF 数据恢复
├── server/                  # TCP 服务器：监听、接受连接、处理客户端
├── protocol/                # RESP 协议解析器 + 响应构造
├── types/                   # 数据类型实现（String、Hash、List、Set、ZSet）
├── logger/                  # 模块化日志（lumberjack 滚动切割）
├── version/                 # 编译时注入的版本信息
├── integration/             # 集成测试（独立 Go module，依赖 go-redis）
│   ├── main_test.go         #   TestMain：编译二进制、启动服务器、清理
│   ├── connection_test.go   #   PING、TIME
│   ├── strings_test.go      #   SET、GET、APPEND
│   ├── database_test.go     #   SELECT
│   ├── keys_test.go         #   所有键操作命令
│   ├── server_test.go       #   INFO、COMMAND、CONFIG
│   └── memory_test.go       #   MEMORY STATS/USAGE
├── etc/godis.yaml           # 配置文件
├── go.mod                   # 主模块依赖
├── go.work                  # Go workspace（主模块 + 集成测试）
└── all_test.go              # 单元测试编排
```

## 架构设计

### 数据流

```
redis-cli ──► TCP(6389) ──► ParseRESP ──► Execute(命令) ──► GodisDB ──► AOF 写入
                                                │
                                                ▼
                                          响应构造 ──► redis-cli
```

### AOF 持久化

- **仅追加写入**：所有写命令（`FlagWrite` + 执行成功）以 RESP 文本格式追加到 AOF 文件，非 0 号数据库自动添加 `SELECT` 前缀
- **混合重写**：定期将 AOF 压缩为完整的 Gob 二进制快照（`GODIS-HYBRID` 头部），后续增量命令以文本格式追加。兼顾快速恢复和低磁盘 I/O
- **自动重写触发条件**（每 10 秒检查）：
  - 首次写入超过 2 KB
  - 文件增长超过 64 MB
  - 文件相对上次重写增长超过 50%

### 键过期

- **惰性删除**：访问时检查并删除已过期的键
- **主动 GC**：后台协程每秒遍历所有数据库，批量清理过期键
- TTL 语义：`-1` = 永久有效，`-2` = 键不存在

### 多数据库

每个连接维护独立的 `currentDBID`（默认 0），`SELECT n` 切换后续命令操作的数据库。

## 配置说明

```yaml
# Godis configuration file

bind: 0.0.0.0
port: 6389
databases: 16
aof_file: ./data/godis.aof
log_file: ./logs/godis.log
log_level: info
```

| 配置项      | 默认值             | 说明                                         |
|-------------|-------------------|----------------------------------------------|
| `bind`      | `0.0.0.0`         | 监听地址                                     |
| `port`      | `6379`            | TCP 端口                                     |
| `databases` | `16`              | 逻辑数据库数量（0-15）                       |
| `aof_file`  | `./data/godis.aof`| AOF 持久化文件路径                           |
| `log_file`  | `./logs/godis.log`| 日志文件，单文件最大 20MB，保留 10 个备份     |
| `log_level` | `info`            | 日志级别：`debug` / `info` / `warn` / `error` |

支持运行时通过 `CONFIG SET log_level` 和 `CONFIG REWRITE` 修改配置并持久化。

## 测试

### 单元测试

```bash
go test ./...
```

### 集成测试

需要 `go-redis`（自动下载），会编译 godis 二进制并在真实 TCP 服务器上测试所有命令：

```bash
go test -v ./integration/ -count=1
```

输出包含每个测试的操作日志，以及最终统计汇总：

```
==================================================
  Test Results
  Total: 34 | Passed: 34 | Failed: 0
  Status: ✅ ALL PASSED
==================================================
```

## 构建时注入版本信息

```bash
go build -ldflags "\
  -X godis/version.Version=1.0.0 \
  -X godis/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%S) \
  -X godis/version.GitCommit=$(git rev-parse --short HEAD)" \
  -o godis .
```

## 依赖

| 包                                | 用途              |
|-----------------------------------|-------------------|
| `gopkg.in/yaml.v3`                | 配置文件解析       |
| `gopkg.in/natefinch/lumberjack.v2`| 日志文件滚动       |
| `github.com/redis/go-redis/v9`    | 集成测试（独立模块）|

## 待实现

- [ ] Hash 命令：HSET、HGET、HDEL、HGETALL、HEXISTS、HLEN
- [ ] List 命令：LPUSH、RPUSH、LPOP、RPOP、LLEN、LRANGE
- [ ] Set 命令：SADD、SREM、SMEMBERS、SISMEMBER、SCARD
- [ ] ZSet 命令：ZADD、ZREM、ZRANGE、ZSCORE、ZCARD、ZRANK
- [ ] 发布订阅：SUBSCRIBE、PUBLISH
- [ ] 事务：MULTI、EXEC、DISCARD、WATCH
- [ ] Lua 脚本：EVAL、EVALSHA
- [ ] 主从复制
- [ ] 集群模式
- [ ] RESP3 协议支持
- [ ] TLS 支持
- [ ] 性能压测套件
- [ ] 命令行管理工具
