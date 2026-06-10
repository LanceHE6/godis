package commands

// key 操作相关命令

func init() {
	// 删除一个或多个 key
	Register("DEL", -2, FlagWrite, 1, -1, 1, UnimplementedHandlerFunc)
	// 检查 key 是否存在
	Register("EXISTS", -2, FlagReadonly, 1, -1, 1, UnimplementedHandlerFunc)
	// 设置 key 的过期时间（秒）
	Register("EXPIRE", 3, FlagWrite, 1, 1, 1, UnimplementedHandlerFunc)
	// 将 key 移动到另一个数据库
	Register("MOVE", 3, FlagWrite, 1, 1, 1, UnimplementedHandlerFunc)
	// 移除 key 的过期时间，使其永久有效
	Register("PERSIST", 2, FlagWrite, 1, 1, 1, UnimplementedHandlerFunc)
	// 设置 key 的过期时间（毫秒）
	Register("PEXPIRE", 3, FlagWrite, 1, 1, 1, UnimplementedHandlerFunc)
	// 获取 key 的剩余过期时间（毫秒）
	Register("PTTL", 2, FlagReadonly, 1, 1, 1, UnimplementedHandlerFunc)
	// 对列表、集合或有序集合的元素排序
	Register("SORT", -2, FlagWrite, 1, 1, 1, UnimplementedHandlerFunc)
	// 更新 key 的最后访问时间
	Register("TOUCH", -2, FlagReadonly, 1, -1, 1, UnimplementedHandlerFunc)
	// 获取 key 的剩余过期时间（秒）
	Register("TTL", 2, FlagReadonly, 1, 1, 1, UnimplementedHandlerFunc)
	// 获取 key 存储的数据类型
	Register("TYPE", 2, FlagReadonly, 1, 1, 1, UnimplementedHandlerFunc)
	// 异步删除一个或多个 key
	Register("UNLINK", -2, FlagWrite, 1, -1, 1, UnimplementedHandlerFunc)
}
