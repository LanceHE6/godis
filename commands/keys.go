package commands

// key 操作相关命令

func init() {
	Register("DEL", -2, "write", 1, -1, 1, UnimplementedHandlerFunc)
	Register("EXISTS", -2, "readonly", 1, -1, 1, UnimplementedHandlerFunc)
	Register("EXPIRE", 3, "write", 1, 1, 1, UnimplementedHandlerFunc)
	Register("MOVE", 3, "write", 1, 1, 1, UnimplementedHandlerFunc)
	Register("PERSIST", 2, "write", 1, 1, 1, UnimplementedHandlerFunc)
	Register("PEXPIRE", 3, "write", 1, 1, 1, UnimplementedHandlerFunc)
	Register("PTTL", 2, "readonly", 1, 1, 1, UnimplementedHandlerFunc)
	Register("SORT", -2, "write", 1, 1, 1, UnimplementedHandlerFunc)
	Register("TOUCH", -2, "readonly", 1, -1, 1, UnimplementedHandlerFunc)
	Register("TTL", 2, "readonly", 1, 1, 1, UnimplementedHandlerFunc)
	Register("TYPE", 2, "readonly", 1, 1, 1, UnimplementedHandlerFunc)
	Register("UNLINK", -2, "write", 1, -1, 1, UnimplementedHandlerFunc)
}
