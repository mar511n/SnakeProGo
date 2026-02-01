package main

import (
	"fmt"
	"os"
	"time"
)

// LogLevel represents the importance of a log message.
type LogLevel int

const (
	// LogLevelError is for critical errors that might prevent the game from working.
	LogLevelError LogLevel = iota
	// LogLevelWarning is for issues that are handled but might be unexpected.
	LogLevelWarning
	// LogLevelInfo is for general informational messages.
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
	// If the message importance is higher (numerically) than the configured level, ignore it.
	// E.g. Config=1 (Warning), Message=2 (Info) -> 2 > 1 -> Ignore
	// Config=2 (Info), Message=0 (Error) -> 0 <= 2 -> Print
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

// LogError is a helper for logging errors
func LogError(format string, args ...interface{}) {
	Log(LogLevelError, format, args...)
}

// LogWarning is a helper for logging warnings
func LogWarning(format string, args ...interface{}) {
	Log(LogLevelWarning, format, args...)
}

// LogInfo is a helper for logging info
func LogInfo(format string, args ...interface{}) {
	Log(LogLevelInfo, format, args...)
}

// FatalError logs the error and exits the program
func FatalError(format string, args ...interface{}) {
	Log(LogLevelError, format, args...)
	os.Exit(1)
}
