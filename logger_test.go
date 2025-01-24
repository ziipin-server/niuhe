package niuhe

import (
	"fmt"
	"os"
	"testing"
)

func TestSimpleLog(t *testing.T) {
	LogDebug("1")
	LogInfo("2")
	LogWarn("3")
	LogError("4")
	LogFatal("5")
}

func TestSetLogLevel(t *testing.T) {
	SetLogLevel(LOG_WARN)
	LogDebug("1")
	LogInfo("2")
	LogWarn("3")
	LogError("4")
	LogFatal("5")
}

func TestNewLogger(t *testing.T) {
	AddLogCallback(LOG_ERROR, func(msg string) {
		fmt.Println("LOG_ERROR", msg)
	})
	SetLogPreHook(func(level int, f string, args []any) (int, string, []any) {
		if level == LOG_INFO && f == "2" {
			fmt.Println(">>>>>>>>> hooking info 2, change level into ERR")
			return LOG_ERROR, "!!%s!!" + f, append([]any{"lalala"}, args...)
		}
		return level, f, args
	})
	SetLogLevel(LOG_WARN)
	LogDebug("1")
	LogInfo("2")
	LogWarn("%d", 3)
	LogError("4")
	LogFatal("5")
	logger := NewLogger(os.Stdout, "[logger]-")
	logger.SetLogLevel(LOG_DEBUG)
	logger.Debug("x %d", 1)
	logger.Info("x 2")
	logger.Warn("x 3")
	logger.Error("x 4")
	logger.Fatal("x 5")
}
