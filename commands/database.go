package commands

import (
	"fmt"
	"strconv"

	"godis/protocol"
)

func init() {
	// 切换当前连接的数据库
	Register("SELECT", 2, FlagFast, 0, 0, 0, handleSelect)
}

func handleSelect(ctx *CommandContext) string {
	if len(ctx.Args) < 2 {
		return protocol.MakeError("ERR wrong number of arguments for 'select' command")
	}

	// 解析输入的库序号
	dbIdx, err := strconv.Atoi(ctx.Args[1])
	if err != nil || dbIdx < 0 || dbIdx >= len(ctx.AllDBs) {
		return protocol.MakeError(fmt.Sprintf("ERR DB index is out of range (0-%d)", len(ctx.AllDBs)-1))
	}

	// 直接修改上下文中的指针值，从而改变这个客户端连接的全局状态
	*ctx.CurrentDBID = dbIdx

	return protocol.MakeSimpleString("OK")
}
