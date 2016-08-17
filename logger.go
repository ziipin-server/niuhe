package niuhe

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	LOG_DEBUG = iota
	LOG_INFO
	LOG_WARN
	LOG_ERROR
	LOG_FATAL
	end_log_level
	DEFAULT_LOG_FLAGS = log.Ldate | log.Ltime | log.Lshortfile
)

type LoggerT struct {
	minLevel int
	loggers  []*log.Logger
}

var defaultLogger *LoggerT

func MustParseLogLevelName(name string) int {
	if name == "DEBUG" {
		return LOG_DEBUG
	}
	if name == "INFO" {
		return LOG_INFO
	}
	if name == "WARN" {
		return LOG_WARN
	}
	if name == "ERROR" {
		return LOG_ERROR
	}
	if name == "FATAL" {
		return LOG_FATAL
	}
	panic("Unknown level name " + name)
}

func NewLogger(out io.Writer, prefix string) *LoggerT {
	return &LoggerT{
		minLevel: LOG_INFO,
		loggers: []*log.Logger{
			log.New(out, prefix+"[DBG]", DEFAULT_LOG_FLAGS),
			log.New(out, prefix+"[INF]", DEFAULT_LOG_FLAGS),
			log.New(out, prefix+"[WRN]", DEFAULT_LOG_FLAGS),
			log.New(out, prefix+"[ERR]", DEFAULT_LOG_FLAGS),
			log.New(out, prefix+"[FAT]", DEFAULT_LOG_FLAGS),
			log.New(out, prefix+"[PAN]", DEFAULT_LOG_FLAGS),
		},
	}
}

func (l *LoggerT) SetLogLevel(logLevel int) {
	l.minLevel = logLevel
}

func (l *LoggerT) log(level int, calldepth int, format string, args ...interface{}) bool {
	if level < l.minLevel || level >= end_log_level {
		return false
	}
	logger := l.loggers[level]
	if len(args) < 1 {
		logger.Output(calldepth, format)
	} else {
		logger.Output(calldepth, fmt.Sprintf(format, args...))
	}
	return true
}

func (l *LoggerT) Debug(format string, args ...interface{}) bool {
	return l.log(LOG_DEBUG, 3, format, args...)
}

func (l *LoggerT) Info(format string, args ...interface{}) bool {
	return l.log(LOG_INFO, 3, format, args...)
}

func (l *LoggerT) Warn(format string, args ...interface{}) bool {
	return l.log(LOG_WARN, 3, format, args...)
}

func (l *LoggerT) Error(format string, args ...interface{}) bool {
	return l.log(LOG_ERROR, 3, format, args...)
}

func (l *LoggerT) Fatal(format string, args ...interface{}) bool {
	return l.log(LOG_FATAL, 3, format, args...)
}

func SetLogLevel(logLevel int) {
	defaultLogger.SetLogLevel(logLevel)
}

func LogDebug(format string, args ...interface{}) bool {
	return defaultLogger.log(LOG_DEBUG, 3, format, args...)
}

func LogInfo(format string, args ...interface{}) bool {
	return defaultLogger.log(LOG_INFO, 3, format, args...)
}

func LogWarn(format string, args ...interface{}) bool {
	return defaultLogger.log(LOG_WARN, 3, format, args...)
}

func LogError(format string, args ...interface{}) bool {
	return defaultLogger.log(LOG_ERROR, 3, format, args...)
}

func LogFatal(format string, args ...interface{}) bool {
	return defaultLogger.log(LOG_FATAL, 3, format, args...)
}

func init() {
	defaultLogger = NewLogger(os.Stderr, "")
}
