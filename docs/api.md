# Godis Web Admin API

> Base URL: `http://127.0.0.1:6390/api`

所有返回均为 JSON。写操作（非 GET）会记录到日志并写入 AOF。

## 认证

### GET /auth

检查是否启用密码认证。

**Response**
```json
{ "requirepass": true }
```

### POST /auth

验证登录密码。

**Request**
```json
{ "password": "mypassword" }
```

**Response**
```json
{ "ok": true }
```

## 服务器

### GET /server/info

获取服务器运行状态。

**Response**
```json
{
  "version": "dev",
  "uptime": "1h23m45s",
  "keys": 42,
  "memory": "3.2 MB",
  "clients": "1",
  "port": 6389,
  "databases": 16
}
```

### GET /server/stats

获取实时性能统计（含历史趋势）。

**Response**
```json
{
  "keys_per_db": [5, 3, 0, 7],
  "cpu_pct": 1.23,
  "memory_mb": "3.2",
  "history": [
    { "time": "15:04:05", "cpu": 1.20, "mem": 3.1 },
    { "time": "15:04:08", "cpu": 1.25, "mem": 3.2 }
  ]
}
```

| 字段 | 说明 |
|------|------|
| `keys_per_db` | 每个数据库的键数量数组 |
| `cpu_pct` | 当前 CPU 占用百分比（Linux `/proc/self/stat`） |
| `history` | 最近 60 个采样点（时间 / CPU % / 内存 MB） |

## 键管理

### GET /keys

分页获取键列表。

**Query**
| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `db` | int | 0 | 数据库编号 |
| `pattern` | string | `*` | glob 匹配模式 |
| `page` | int | 1 | 页码 |
| `page_size` | int | 15 | 每页条数 |

**Response**
```json
{
  "keys": [
    { "key": "user:1", "type": "string", "ttl": 3600, "size": 1 }
  ],
  "total": 42,
  "page": 1,
  "page_size": 15
}
```

### POST /keys/delete

批量删除键。

**Request**
```json
{ "keys": ["key1", "key2"] }
```

**Response**
```json
{ "deleted": 2 }
```

### GET /key

获取单个键详情。

**Query**
| 参数 | 类型 | 说明 |
|------|------|------|
| `db` | int | 数据库编号 |
| `key` | string | 键名 |

**Response（string）**
```json
{ "key": "foo", "type": "string", "ttl": -1, "value": "hello" }
```

**Response（hash）**
```json
{ "key": "h", "type": "hash", "ttl": 3600, "fields": { "f1": "v1", "f2": "v2" } }
```

**Response（list）**
```json
{ "key": "l", "type": "list", "ttl": -1, "values": ["a", "b", "c"] }
```

**Response（set）**
```json
{ "key": "s", "type": "set", "ttl": -1, "members": ["x", "y"] }
```

**Response（zset）**
```json
{ "key": "z", "type": "zset", "ttl": 100, "members": [{ "Member": "p1", "Score": 100 }] }
```

### POST /key/edit

修改键（值、TTL、重命名、子元素操作）。

**Request**
```json
{
  "key": "foo",
  "action": "set_value",
  "value": "new_value",
  "field": "field_name",
  "db": 0
}
```

**Actions**

| action | 说明 | 必填字段 |
|--------|------|----------|
| `set_value` | 设置 string 值 | `key`, `value` |
| `set_ttl` | 设置过期秒数 | `key`, `value`(秒) |
| `persist` | 移除过期 | `key` |
| `rename` | 重命名 | `key`, `value`(新名) |
| `hset` | hash 设置 field | `key`, `field`, `value` |
| `hdel` | hash 删除 field | `key`, `field` |
| `rpush` | list 尾部追加 | `key`, `value` |
| `lrem` | list 移除元素 (count=1) | `key`, `value` |
| `lset` | list 按索引设值 | `key`, `field`(索引), `value` |
| `sadd` | set 添加成员 | `key`, `value` |
| `srem` | set 移除成员 | `key`, `value` |
| `zset_score` | zset 添加/更新成员 | `key`, `field`(member), `value`(score) |

**Response**
```json
{ "status": "ok" }
```
错误时返回 `{ "error": "..." }`。

## 命令执行

### POST /exec

执行任意 Redis 命令。

**Request**
```json
{ "command": "SET foo bar", "db": 0 }
```

**Response**
```json
{ "reply": "OK" }
```

> 写命令会自动写入 AOF 持久化，非 DB0 的操作会带上 `SELECT` 前缀。

## 命令列表

### GET /commands

返回所有已注册的命令名。

**Response**
```json
{ "commands": ["SET", "GET", "HSET", ...] }
```

## 日志

### GET /logs

返回最近的日志行（环形缓冲区，最多 200 条）。

**Response**
```json
{
  "logs": [
    "2026/01/15 10:30:00 [INFO] [DATASTORE] clear expired key [x]",
    "2026/01/15 10:30:01 [WARN] [WEB] POST /api/exec"
  ]
}
```
