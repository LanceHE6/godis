package commands

// key操作相关命令

func init() {
	// 删除一个或多个 key
	CommandRegistry["DEL"] = UnimplementedHandlerFunc
	// 检查 key 是否存在
	CommandRegistry["EXISTS"] = UnimplementedHandlerFunc
	// 设置 key 的过期时间（秒）
	CommandRegistry["EXPIRE"] = UnimplementedHandlerFunc
	// 将 key 移动到另一个数据库
	CommandRegistry["MOVE"] = UnimplementedHandlerFunc
	// 移除 key 的过期时间，使其永久有效
	CommandRegistry["PERSIST"] = UnimplementedHandlerFunc
	// 设置 key 的过期时间（毫秒）
	CommandRegistry["PEXPIRE"] = UnimplementedHandlerFunc
	// 获取 key 的剩余过期时间（毫秒）
	CommandRegistry["PTTL"] = UnimplementedHandlerFunc
	// 对列表、集合或有序集合的元素排序
	CommandRegistry["SORT"] = UnimplementedHandlerFunc
	// 更新 key 的最后访问时间
	CommandRegistry["TOUCH"] = UnimplementedHandlerFunc
	// 获取 key 的剩余过期时间（秒）
	CommandRegistry["TTL"] = UnimplementedHandlerFunc
	// 获取 key 存储的数据类型
	CommandRegistry["TYPE"] = UnimplementedHandlerFunc
	// 异步删除一个或多个 key
	CommandRegistry["UNLINK"] = UnimplementedHandlerFunc
}
