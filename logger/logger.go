package logger

import (
	"fmt"
	"log"
)

const (
	infoColor    = "\033[1;34m%s\033[0m\n"
	noticeColor  = "\033[1;36m%s\033[0m\n"
	warningColor = "\033[1;33m%s\033[0m\n"
	errorColor   = "\033[1;31m%s\033[0m\n"
	debugColor   = "\033[0;36m%s\033[0m\n"
)

type logFunction func(string)

var logsMap = map[string]logFunction{
	"info": func(message string) {
		fmt.Printf(infoColor, message) // Colorize message with Info color = "\033[1;34m%s\033[0m"
	},
	"notice": func(message string) {
		fmt.Printf(noticeColor, message) // Colorize message with Notice color = "\033[1;36m%s\033[0m"
	},
	"warn": func(message string) {
		fmt.Printf(warningColor, message) // Colorize message with Warn color = "\033[1;33m%s\033[0m"
	},
	"error": func(message string) {
		fmt.Printf(errorColor, message) // Colorize message with Error color = "\033[1;31m%s\033[0m"
	},
	"debug": func(message string) {
		fmt.Printf(debugColor, message) // Colorize message with Debug color = "\033[0;36m%s\033[0m"
	},
}

type Logger interface {
	Log(logType string, message string)
}

type MafiaLogger struct {
	IsEnabled bool
}

// Print message to the current output stream
func (l *MafiaLogger) Log(logType string, message string) {
	if l.IsEnabled == true {
		if logFunc, ok := logsMap[logType]; ok {
			logFunc(message)
		} else {
			log.Fatal("Unknown log function")
		}
	}
}
