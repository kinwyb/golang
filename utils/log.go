package utils

import (
	"github.com/astaxie/beego/logs"
)

//LoggerLevel 日志等级
type LoggerLevel int

const (
	LogLevelTrace LoggerLevel = iota
	LogLevelDebug
	LogLevelWarn
	LogLevelInfo
	LogLevelError
)

//Logger 日志接口
type Logger interface {
	//追踪
	Trace(format string, args ...interface{})
	//调试
	Debug(format string, args ...interface{})
	//输出
	Info(format string, args ...interface{})
	//警告
	Warning(format string, args ...interface{})
	//错误
	Error(format string, args ...interface{})
}

//GetDefaultLogger 获取一个默认的Logger
func GetDefaultLogger() Logger {
	return logs.NewLogger()
}

//WriteLog 写入日志
func WriteLog(log Logger, level LoggerLevel, format string, args ...interface{}) {
	if log == nil {
		return
	}
	switch level {
	case LogLevelDebug:
		log.Debug(format, args...)
	case LogLevelWarn:
		log.Warning(format, args...)
	case LogLevelInfo:
		log.Info(format, args...)
	case LogLevelError:
		log.Error(format, args...)
	default:
		log.Trace(format, args...)
	}
}
