package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"time"
	"web-api/conf/consts"

	"github.com/sirupsen/logrus"
)

const (
	LevelError = iota
	LevelWarning
	LevelInformational
	LevelDebug
)

var logger *Logger

type Logger struct {
	level int
}

func (ll *Logger) Println(msg string) {
	fmt.Printf("%s %s\n", time.Now().Format("2006-01-02 15:04:05 -0700"), msg)
}

func (ll *Logger) Panic(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	msg := fmt.Sprintf("[Panic] "+format+"\n", v...)
	ll.Println(msg)
	os.Exit(0)
}

func (ll *Logger) Error(format string, v ...interface{}) {
	if LevelError > ll.level {
		return
	}
	msg := fmt.Sprintf("[E] "+format+"\n", v...)
	ll.Println(msg)
}

func (ll *Logger) Warning(format string, v ...interface{}) {
	if LevelWarning > ll.level {
		return
	}
	msg := fmt.Sprintf("[W] "+format+"\n", v...)
	ll.Println(msg)
}

func (ll *Logger) Info(format string, v ...interface{}) {
	if LevelInformational > ll.level {
		return
	}
	msg := fmt.Sprintf("[I] "+format+"\n", v...)
	ll.Println(msg)
}

func (ll *Logger) Debug(format string, v ...interface{}) {
	if LevelDebug > ll.level {
		return
	}
	msg := fmt.Sprintf("[D] "+format+"\n", v...)
	ll.Println(msg)
}

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
	stk := string(debug.Stack())
	logger := c.Value(consts.LogKey)
	if logger == nil {
		logLevel := logrus.DebugLevel
		if l, e := logrus.ParseLevel(os.Getenv("LOG_LEVEL")); e == nil {
			logLevel = l
		}
		logrus.SetLevel(logLevel)
		return logrus.WithField(consts.CorrelationKey, c.Value(consts.CorrelationKey)).WithField(consts.StackTraceKey, stk)
	}
	return logger.(*logrus.Entry).WithField(consts.StackTraceKey, stk)
}

func GetCorrelationID(c context.Context) string {
	return c.Value(consts.CorrelationKey).(string)
}
