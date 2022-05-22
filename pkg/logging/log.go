package logging

import (
	"fmt"
	goLogger "log"
)

type logLevel uint

const (
	DEBUG logLevel = iota
	INFO  logLevel = iota
	WARN  logLevel = iota
	FATAL logLevel = iota
)

type Log struct {
	level logLevel
}

func NewLog(level logLevel) Log {
	return Log{level: level}
}
func (l logLevel) String() string {
	switch l {
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case WARN:
		return "WARN"
	case FATAL:
		return "FATAL"
	default:
		return ""
	}
}
func (l Log) Log(level logLevel, msg ...any) {
	if level < l.level {
		return
	}
	goLogger.Printf("[%s] %s\n", l.level, fmt.Sprintln(msg...))
}
