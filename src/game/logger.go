package game

import (
	"fmt"
	"os"
	"time"
)

type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarning
	LogLevelInfo
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// Log prints a message to the console if the importance level is within the configured DebugLevel.
// level: The importance of the message (Error, Warning, Info)
// format: Sprintf style format string
// args: Arguments for the format string
func Log(level LogLevel, format string, args ...interface{}) {
	if ConfigLoaded && int(level) > GConfig.DebugLevel {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	var prefix string
	var color string

	switch level {
	case LogLevelError:
		prefix = "[ERROR]"
		color = colorRed
	case LogLevelWarning:
		prefix = "[WARN ]"
		color = colorYellow
	case LogLevelInfo:
		prefix = "[INFO ]"
		color = colorCyan
	default:
		prefix = "[UNK  ]"
		color = colorWhite
	}

	msg := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s %s %s%s\n", color, timestamp, prefix, msg, colorReset)
}

func LogError(format string, args ...interface{}) {
	Log(LogLevelError, format, args...)
}

func LogWarning(format string, args ...interface{}) {
	Log(LogLevelWarning, format, args...)
}

func LogInfo(format string, args ...interface{}) {
	Log(LogLevelInfo, format, args...)
}

func FatalError(format string, args ...interface{}) {
	Log(LogLevelError, format, args...)
	os.Exit(1)
}
