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
	LOG_ASSERT
	LOG_FATAL
	end_log_level
	DEFAULT_LOG_FLAGS = log.Ldate | log.Ltime | log.Lshortfile
)

type LoggerT struct {
	minLevel  int
	loggers   []*log.Logger
	callbacks struct {
		debug  []func(string)
		info   []func(string)
		warn   []func(string)
		error  []func(string)
		assert []func(string)
		fatal  []func(string)
	}
	plugin func(string) string // replacement plugin
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
	if name == "ASSERT" {
		return LOG_ASSERT
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
			log.New(out, prefix+"[AST]", DEFAULT_LOG_FLAGS),
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
	var content = format
	if len(args) > 0 {
		content = fmt.Sprintf(format, args...)
	}

	if nil != l.plugin {
		content = l.plugin(content)
	}

	logger.Output(calldepth, content)
	switch level {
	case LOG_DEBUG:
		for _, cb := range l.callbacks.debug {
			cb(content)
		}
	case LOG_INFO:
		for _, cb := range l.callbacks.info {
			cb(content)
		}
	case LOG_WARN:
		for _, cb := range l.callbacks.warn {
			cb(content)
		}
	case LOG_ERROR:
		for _, cb := range l.callbacks.error {
			cb(content)
		}
	case LOG_ASSERT:
		for _, cb := range l.callbacks.assert {
			cb(content)
		}
	case LOG_FATAL:
		for _, cb := range l.callbacks.fatal {
			cb(content)
		}
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

func (l *LoggerT) Assert(format string, args ...interface{}) bool {
	return l.log(LOG_ASSERT, 3, format, args...)
}

func (l *LoggerT) Fatal(format string, args ...interface{}) bool {
	return l.log(LOG_FATAL, 3, format, args...)
}

// DebugUp calldepth +up
func (l *LoggerT) DebugUp(up int, format string, args ...interface{}) bool {
	return l.log(LOG_DEBUG, 3+up, format, args...)
}

func (l *LoggerT) InfoUp(up int, format string, args ...interface{}) bool {
	return l.log(LOG_INFO, 3+up, format, args...)
}

func (l *LoggerT) WarnUp(up int, format string, args ...interface{}) bool {
	return l.log(LOG_WARN, 3+up, format, args...)
}

func (l *LoggerT) ErrorUp(up int, format string, args ...interface{}) bool {
	return l.log(LOG_ERROR, 3+up, format, args...)
}

func (l *LoggerT) AssertUp(up int, format string, args ...interface{}) bool {
	return l.log(LOG_ASSERT, 3+up, format, args...)
}

func (l *LoggerT) FatalUp(up int, format string, args ...interface{}) bool {
	return l.log(LOG_FATAL, 3+up, format, args...)
}

func (l *LoggerT) logLevel() int {
	return l.minLevel
}

// Logger export default looger
func Logger() *LoggerT {
	return defaultLogger
}

func (l *LoggerT) AddCallback(level int, callbackFunc func(string)) {
	switch level {
	case LOG_DEBUG:
		l.callbacks.debug = append(l.callbacks.debug, callbackFunc)
	case LOG_INFO:
		l.callbacks.info = append(l.callbacks.info, callbackFunc)
	case LOG_WARN:
		l.callbacks.warn = append(l.callbacks.warn, callbackFunc)
	case LOG_ERROR:
		l.callbacks.error = append(l.callbacks.error, callbackFunc)
	case LOG_ASSERT:
		l.callbacks.assert = append(l.callbacks.assert, callbackFunc)
	case LOG_FATAL:
		l.callbacks.fatal = append(l.callbacks.fatal, callbackFunc)
	}
}

func (l *LoggerT) SetPlugin(plugin func(string) string) {
	l.plugin = plugin
}

func LogLevel() int {
	return defaultLogger.minLevel
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

func LogAssert(format string, args ...interface{}) bool {
	return defaultLogger.log(LOG_ASSERT, 3, format, args...)
}

func LogFatal(format string, args ...interface{}) bool {
	return defaultLogger.log(LOG_FATAL, 3, format, args...)
}

func AddLogCallback(level int, callback func(string)) {
	defaultLogger.AddCallback(level, callback)
}

func SetLogPlugin(plugin func(string) string) {
	defaultLogger.SetPlugin(plugin)
}

func init() {
	defaultLogger = NewLogger(os.Stderr, "")
}
