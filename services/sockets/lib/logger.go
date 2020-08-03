// Copyright sasha.los.0148@gmail.com
// All Rights have been taken by Mafia :)

// Common output stream interface
package lib

import (
	"fmt"
	"time"
)

// Function to print message
type LogFunction func(string, string)

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

func getTime() string {
	return time.Now().Format(time.RFC1123)
}

var logsMap = map[LogType]LogFunction{
	Info: func(message string, key string) {
		fmt.Printf(infoColor, "/***LOG START***/\n"+getTime()+"\nKEY: "+key+"\n\n"+message+"\n/---LOG END---/\n") // Colorize message with Info color = "\033[1;34m%s\033[0m"
	},
	Notice: func(message string, key string) {
		fmt.Printf(noticeColor, "/***LOG START***/\n"+getTime()+"\nKEY: "+key+"\n\n"+message+"\n/---LOG END---/\n") // Colorize message with Notice color = "\033[1;36m%s\033[0m"
	},
	Warn: func(message string, key string) {
		fmt.Printf(warningColor, "/***LOG START***/\n"+getTime()+"\nKEY: "+key+"\n\n"+message+"\n/---LOG END---/\n") // Colorize message with Warn color = "\033[1;33m%s\033[0m"
	},
	Error: func(message string, key string) {
		fmt.Printf(errorColor, "/***LOG START***/\n"+getTime()+"\nKEY: "+key+"\n\n"+message+"\n/---LOG END---/\n") // Colorize message with Error color = "\033[1;31m%s\033[0m"
	},
	Debug: func(message string, key string) {
		fmt.Printf(debugColor, "/***LOG START***/\n"+getTime()+"\nKEY: "+key+"\n\n"+message+"\n/---LOG END---/\n") // Colorize message with Debug color = "\033[0;36m%s\033[0m"
	},
}

// Root logger interface
type Logger interface {
	Log(logType LogType, message string)
}

// Implementation for current project
type MafiaLogger struct {
	IsEnabled bool
	LogKey    string
}

// Print message to the current output stream
func (l *MafiaLogger) Log(logType LogType, message string) {
	if l.IsEnabled {
		if logFunc, ok := logsMap[logType]; ok {
			logFunc(message, l.LogKey)
		}
	}
}
