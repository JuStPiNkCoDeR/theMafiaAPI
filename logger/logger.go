// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Common output stream interface
package logger

import (
	"fmt"
	"log"
)

// Function to print message
type LogFunction func(string)

// Type of log case
type LogType string

const (
	infoColor            = "\033[1;34m%s\033[0m\n"
	noticeColor          = "\033[1;36m%s\033[0m\n"
	warningColor         = "\033[1;33m%s\033[0m\n"
	errorColor           = "\033[1;31m%s\033[0m\n"
	debugColor           = "\033[0;36m%s\033[0m\n"
	Info         LogType = "info"
	Notice       LogType = "notice"
	Warn         LogType = "warn"
	Error        LogType = "error"
	Debug        LogType = "debug"
)

var logsMap = map[LogType]LogFunction{
	Info: func(message string) {
		fmt.Printf(infoColor, message) // Colorize message with Info color = "\033[1;34m%s\033[0m"
	},
	Notice: func(message string) {
		fmt.Printf(noticeColor, message) // Colorize message with Notice color = "\033[1;36m%s\033[0m"
	},
	Warn: func(message string) {
		fmt.Printf(warningColor, message) // Colorize message with Warn color = "\033[1;33m%s\033[0m"
	},
	Error: func(message string) {
		fmt.Printf(errorColor, message) // Colorize message with Error color = "\033[1;31m%s\033[0m"
	},
	Debug: func(message string) {
		fmt.Printf(debugColor, message) // Colorize message with Debug color = "\033[0;36m%s\033[0m"
	},
}

// Root logger interface
type Logger interface {
	Log(logType LogType, message string)
}

// Implementation for current project
type MafiaLogger struct {
	IsEnabled bool
}

// Print message to the current output stream
func (l *MafiaLogger) Log(logType LogType, message string) {
	if l.IsEnabled == true {
		if logFunc, ok := logsMap[logType]; ok {
			logFunc(message)
		} else {
			log.Fatal("Unknown log function")
		}
	}
}
