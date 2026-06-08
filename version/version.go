package version

// 编译时通过 -ldflags 注入，默认值用于本地开发
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)
