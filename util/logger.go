package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"time"
	"web-api/conf/consts"

	"github.com/sirupsen/logrus"
)

const (
	// LevelError 错误
	LevelError = iota
	// LevelWarning 警告
	LevelWarning
	// LevelInformational 提示
	LevelInformational
	// LevelDebug 除错
	LevelDebug
)

var logger *Logger

// Logger 日志
type Logger struct {
	level int
}

// Println 打印
func (ll *Logger) Println(msg string) {
	fmt.Printf("%s %s\n", time.Now().Format("2006-01-02 15:04:05 -0700"), msg)
}

// Panic 极端错误
func (ll *Logger) Panic(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	msg := fmt.Sprintf("[Panic] "+format+"\n", v...)
	ll.Println(msg)
	os.Exit(0)
}

// Error 错误
func (ll *Logger) Error(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	msg := fmt.Sprintf("[E] "+format+"\n", v...)
	ll.Println(msg)
}

// Warning 警告
func (ll *Logger) Warning(format string, v ...interface{}) {
	if LevelWarning > ll.level {
		return
	}
	msg := fmt.Sprintf("[W] "+format+"\n", v...)
	ll.Println(msg)
}

// Info 信息
func (ll *Logger) Info(format string, v ...interface{}) {
	if LevelInformational > ll.level {
		return
	}
	msg := fmt.Sprintf("[I] "+format+"\n", v...)
	ll.Println(msg)
}

// Debug 校验
func (ll *Logger) Debug(format string, v ...interface{}) {
	if LevelDebug > ll.level {
		return
	}
	msg := fmt.Sprintf("[D] "+format+"\n", v...)
	ll.Println(msg)
}

// BuildLogger 构建logger
func BuildLogger(level string) {
	intLevel := LevelError
	switch level {
	case "error":
		intLevel = LevelError
	case "warning":
		intLevel = LevelWarning
	case "info":
		intLevel = LevelInformational
	case "debug":
		intLevel = LevelDebug
	}
	l := Logger{
		level: intLevel,
	}
	logger = &l
}

// Log 返回日志对象
func Log() *Logger {
	if logger == nil {
		l := Logger{
			level: LevelDebug,
		}
		logger = &l
	}
	return logger
}

func MarshalService(service any) string {
	if service == nil {
		return ""
	}
	value := reflect.ValueOf(service)
	var field reflect.Value
	if value.Kind() == reflect.Ptr {
		field = value.Elem().FieldByName("Password")
	} else {
		field = value.FieldByName("Password")
	}
	if field.IsValid() && !field.IsZero() {
		field.SetString("***")
	}
	str, _ := json.Marshal(service)
	return string(str)
}

// Logger creates a new Ginrus logger with a UUID included
func GetLoggerEntry(c context.Context) *logrus.Entry {
	logger := c.Value(consts.LogKey)
	if logger == nil {
		logrus.SetLevel(logrus.DebugLevel)
		return logrus.WithField(consts.CorrelationKey, c.Value(consts.CorrelationKey))
	}
	return logger.(*logrus.Entry)
}

func GetCorrelationID(c context.Context) string {
	return c.Value(consts.CorrelationKey).(string)
}
