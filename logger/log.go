package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Level 定义日志级别数字权重
type Level int

const (
	LevelDebug Level = iota // 0
	LevelInfo               // 1
	LevelWarn               // 2
	LevelError              // 3
)

var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
}

var levelMap = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

// ParseLevel 将字符串转为日志级别，无法识别时返回 LevelInfo
func ParseLevel(s string) Level {
	if lvl, ok := levelMap[strings.ToLower(s)]; ok {
		return lvl
	}
	return LevelInfo
}

// 全局变量
var (
	fileRolling *lumberjack.Logger
	globalLevel = LevelInfo // 仅控制控制台的过滤级别
)

// ModuleLogger 每个模块专属的日志记录器
type ModuleLogger struct {
	moduleName string
	outConsole *log.Logger // 控制台的句柄
}

// InitGlobalLogger 初始化全局日志环境
func InitGlobalLogger(filename string, consoleLvl Level) error {
	globalLevel = consoleLvl // 设置控制台的全局过滤级别

	dir := filepath.Dir(filename)
	// 自动递归创建文件夹
	// os.ModePerm 代表 0777 最高读写权限
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to auto-create log directory [%s]: %v", dir, err)
	}
	// 初始化 lumberjack 智能滚动写入器
	fileRolling = &lumberjack.Logger{
		Filename:   filename, // 日志文件路径
		MaxSize:    20,       // 【单位：MB】单文件最大 20MB，超过就切片
		MaxBackups: 10,       // 最多保留 10 个历史备份文件
		MaxAge:     7,        // 【单位：天】保留最近 7 天的历史日志
		Compress:   true,     // 核心大招：历史日志自动用 Gzip 压缩（变成 .log.gz）
		LocalTime:  true,     // 使用本地时间命名备份文件
	}

	return nil
}

// NewModuleLogger 模块专属 Logger
func NewModuleLogger(moduleName string) *ModuleLogger {
	name := strings.ToUpper(moduleName)
	return &ModuleLogger{
		moduleName: name,
		outConsole: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Log 分流输出
func (ml *ModuleLogger) Log(lvl Level, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fullMsg := fmt.Sprintf("[%s] [%s] %s", levelNames[lvl], ml.moduleName, msg)

	if fileRolling != nil {
		fileLogger := log.New(fileRolling, "", log.LstdFlags)
		_ = fileLogger.Output(3, fullMsg)
	}

	if lvl >= globalLevel {
		_ = ml.outConsole.Output(3, fullMsg)
	}
}

// CloseLogSystem 用于程序安全退出时，刷新缓冲区并关闭滚动器
func CloseLogSystem() {
	if fileRolling != nil {
		_ = fileRolling.Close()
	}
}

func (ml *ModuleLogger) Debug(format string, v ...interface{}) { ml.Log(LevelDebug, format, v...) }
func (ml *ModuleLogger) Info(format string, v ...interface{})  { ml.Log(LevelInfo, format, v...) }
func (ml *ModuleLogger) Warn(format string, v ...interface{})  { ml.Log(LevelWarn, format, v...) }
func (ml *ModuleLogger) Error(format string, v ...interface{}) { ml.Log(LevelError, format, v...) }
