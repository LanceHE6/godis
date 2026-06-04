package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// 全局变量
var (
	GlobalFile  *os.File
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
	// 打开或创建日志文件（追加模式）
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	GlobalFile = file
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

	if GlobalFile != nil {
		fileLogger := log.New(GlobalFile, "", log.LstdFlags)
		_ = fileLogger.Output(3, fullMsg)
	}

	if lvl >= globalLevel {
		_ = ml.outConsole.Output(3, fullMsg)
	}
}

func (ml *ModuleLogger) Debug(format string, v ...interface{}) { ml.Log(LevelDebug, format, v...) }
func (ml *ModuleLogger) Info(format string, v ...interface{})  { ml.Log(LevelInfo, format, v...) }
func (ml *ModuleLogger) Warn(format string, v ...interface{})  { ml.Log(LevelWarn, format, v...) }
func (ml *ModuleLogger) Error(format string, v ...interface{}) { ml.Log(LevelError, format, v...) }
