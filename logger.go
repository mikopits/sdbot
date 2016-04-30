package sdbot

import (
	"fmt"
	"io"
	"sync"
)

const (
	LevelDebug = iota
	LevelLog
	LevelIncoming
	LevelOutgoing
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

type AnyLogger struct {
	Output io.Writer
	mutex  sync.Mutex // Safely access the logger concurrently
}

func (lo *AnyLogger) GetLogger() *AnyLogger {
	return lo
}

type DefaultLogger struct {
	AnyLogger
}

// Log debug messages at different levels of severity.
func Debug(lo *Logger, s string) {
	log(lo, s, LevelDebug)
}

func Info(lo *Logger, s string) {
	log(lo, s, LevelInfo)
}

func Warn(lo *Logger, s string) {
	log(lo, s, LevelWarn)
}

func Error(lo *Logger, e error) {
	log(lo, e.Error(), LevelError)
}

func Fatal(lo *Logger, s string) {
	log(lo, s, LevelFatal)
}

// Log messages sent to and from the websocket.
func Incoming(lo *Logger, s string) {
	log(lo, s, LevelIncoming)
}

func Outgoing(lo *Logger, s string) {
	log(lo, s, LevelOutgoing)
}

// fmt.Sprintf shortcuts for convenience and so that the fmt package need not
// be imported where not needed.
func Debugf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), LevelDebug)
}

func Infof(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), LevelInfo)
}

func Warnf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), LevelWarn)
}

func Errorf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), LevelError)
}

func Fatalf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), LevelFatal)
}

func Incomingf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), LevelIncoming)
}

func Outgoingf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), LevelOutgoing)
}

func log(lo *Logger, s string, i int) {
	m := GetMutex(*lo)
	m.Lock()
	message := (*lo).formatMessage(s, i)
	GetOutput(*lo).Write([]byte(message))
	m.Unlock()
}

// Default message formatting for DefaultLogger
func (lo *DefaultLogger) formatMessage(s string, i int) string {
	return s
}

func GetMutex(lp LoggerProvider) sync.Mutex {
	return lp.GetLogger().mutex
}

func GetOutput(lp LoggerProvider) io.Writer {
	return lp.GetLogger().Output
}

// Allow for creation your own loggers with custom formatting.
// See PrettyLogger for how this is done.
type Logger interface {
	formatMessage(s string, i int) string
	GetLogger() *AnyLogger
}

// This interface is used to allow access to the struct fields.
type LoggerProvider interface {
	GetLogger() *AnyLogger
}
