package config

import (
	"log"
	"os"
)

// 定义日志级别常量
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// 初始化不同的日志输出器，可根据需要设置不同的输出目标和格式
var (
	// 调试日志：输出到标准输出
	debugLog = log.New(os.Stdout, LevelDebug+" ", log.Ldate|log.Ltime|log.Lshortfile)
	// 信息日志：输出到标准输出
	infoLog = log.New(os.Stdout, LevelInfo+" ", log.Ldate|log.Ltime|log.Lshortfile)
	// 警告日志：输出到标准输出
	warnLog = log.New(os.Stdout, LevelWarn+" ", log.Ldate|log.Ltime|log.Lshortfile)
	// 错误日志：输出到标准错误
	errorLog = log.New(os.Stderr, LevelError+" ", log.Ldate|log.Ltime|log.Lshortfile)
)

func Debug(v ...interface{}) {
	debugLog.Println(v...)
}

func Info(v ...interface{}) {
	infoLog.Println(v...)
}

func Warn(v ...interface{}) {
	warnLog.Println(v...)
}

func Error(v ...interface{}) {
	errorLog.Println(v...)
}
