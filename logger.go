package sdbot

import (
	"io"
	"os"
	"sync"
)

const (
	LevelDebug     = 0
	LevelLog       = 1
	LevelIncoming  = 10
	LevelOutgoing  = 11
	LevelInfo      = 2
	LevelWarn      = 3
	LevelError     = 4
	LevelException = 40
	LevelFatal     = 5
)

var DefaultWriter = os.Stderr

type DefaultLogger struct {
	Output io.Writer
	mutex  sync.Mutex // Safely access the logger concurrently
}

// Log debug messages at different levels of severity.
func Debug(lo *Logger, s string) {
	go lo.log(s, LevelDebug)
}

func Info(lo *Logger, s string) {
	go lo.log(s, LevelInfo)
}

func Warn(lo *Logger, s string) {
	go lo.log(s, LevelWarn)
}

func Error(lo *Logger, s string) {
	go lo.log(s, LevelError)
}

func Fatal(lo *Logger, s string) {
	go lo.log(s, LevelFatal)
}

// Log messages sent to and from the websocket.
func Incoming(lo *Logger, s string) {
	go lo.log(s, LevelIncoming)
}

func Outgoing(lo *Logger, s string) {
	go lo.log(s, LevelOutgoing)
}

func (lo *Logger) log(s string, i int) {
	lo.mutex.Lock()
	defer lo.mutex.Unlock()
	message = lo.formatMessage(s, i)
	lo.Output.Write([]byte(message))
}

// Default message formatting for DefaultLogger
func (lo *DefaultLogger) formatMessage(s string, i int) string {
	return s
}

// Allow for creation your own loggers with custom formatting.
// See PrettyLogger for how this is done.
type Logger interface {
	log(s string, i int)
	formatMessage(s string, i int) string
}
