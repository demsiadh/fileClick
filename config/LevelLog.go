package config

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

// 定义日志级别常量
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

// 初始化不同的日志输出器，输出到文件
var (
	// 调试日志：输出到文件
	debugLog *log.Logger
	// 信息日志：输出到文件
	infoLog *log.Logger
	// 警告日志：输出到文件
	warnLog *log.Logger
	// 错误日志：输出到文件
	errorLog *log.Logger
)

func init() {
	// 创建或追加写入日志文件
	file, err := os.OpenFile(LogPath+"app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}

	// 初始化不同的日志记录器，启用调用者信息
	debugLog = log.New(file, LevelDebug+" ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLog = log.New(file, LevelInfo+" ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLog = log.New(file, LevelWarn+" ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(file, LevelError+" ", log.Ldate|log.Ltime|log.Lshortfile)
}

// getCallerInfo 获取调用者信息
func getCallerInfo() string {
	// 获取调用栈信息，跳过2层（getCallerInfo本身和调用函数）
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return ""
	}

	// 提取文件名
	short := file
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			short = file[i+1:]
			break
		}
	}

	return fmt.Sprintf("[%s:%d] ", short, line)
}

func Debug(v ...interface{}) {
	// 在日志内容前添加调用者信息
	args := append([]interface{}{getCallerInfo()}, v...)
	debugLog.Println(args...)
}

func Info(v ...interface{}) {
	args := append([]interface{}{getCallerInfo()}, v...)
	infoLog.Println(args...)
}

func Warn(v ...interface{}) {
	args := append([]interface{}{getCallerInfo()}, v...)
	warnLog.Println(args...)
}

func Error(v ...interface{}) {
	args := append([]interface{}{getCallerInfo()}, v...)
	errorLog.Println(args...)
}
