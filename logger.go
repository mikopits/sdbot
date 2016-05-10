package sdbot

import (
	"fmt"
	"io"
	"sync"
)

// Constants representing the different debug levels for loggers.
const (
	levelDebug = iota
	levelLog
	levelIncoming
	levelOutgoing
	levelInfo
	levelWarn
	levelError
	levelFatal
)

// AnyLogger represents a container for a logger. It knows WHERE to log
// the messages and uses a mutex so as to not log out of order. What it does
// not know is HOW to format its messages. See DefaultLogger and
// PrettyLogger to see how this works.
type AnyLogger struct {
	Output io.Writer
	mutex  sync.Mutex
}

func (lo *AnyLogger) getLogger() *AnyLogger {
	return lo
}

// DefaultLogger is the default logger type. Formats messages by doing nothing to them.
type DefaultLogger struct {
	AnyLogger
}

// Default message formatting for DefaultLogger.
func (lo *DefaultLogger) formatMessage(s string, i int) string {
	return s
}

// Log debug messages.
func logDebug(lo *Logger, s string) {
	log(lo, s, levelDebug)
}

// Log informatic messages.
func logInfo(lo *Logger, s string) {
	log(lo, s, levelInfo)
}

// Log warnings.
func logWarn(lo *Logger, s string) {
	log(lo, s, levelWarn)
}

// Log errors.
func logError(lo *Logger, err error) {
	log(lo, err.Error(), levelError)
}

// Log otherwise fatal messages when something that shouldn't happen happens.
func logFatal(lo *Logger, s string) {
	log(lo, s, levelFatal)
}

// Log messages sent from the websocket.
func logIncoming(lo *Logger, s string) {
	log(lo, s, levelIncoming)
}

// Log messages sent to the websocket.
func logOutgoing(lo *Logger, s string) {
	log(lo, s, levelOutgoing)
}

// fmt.Sprintf shortcuts for convenience and so that the fmt package need not
// be imported where not needed.

// Log debug messages with arguments.
func logDebugf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), levelDebug)
}

// Log informatic messages with arguments.
func logInfof(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), levelInfo)
}

// Log warning messages with arguments.
func logWarnf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), levelWarn)
}

// Log errors with arguments.
func logErrorf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), levelError)
}

// Log fatal messages with arguments.
func logFatalf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), levelFatal)
}

// Log incoming messages with arguments. TODO Probably don't need this.
func logIncomingf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), levelIncoming)
}

// Log outgoing messages with arguments. TODO Probably don't need this.
func logOutgoingf(lo *Logger, format string, a ...interface{}) {
	log(lo, fmt.Sprintf(format, a...), levelOutgoing)
}

func log(lo *Logger, s string, i int) {
	m := getMutex(*lo)
	m.Lock()
	message := (*lo).formatMessage(s, i)
	getOutput(*lo).Write([]byte(message))
	m.Unlock()
}

func getMutex(lp LoggerProvider) *sync.Mutex {
	return &lp.getLogger().mutex
}

func getOutput(lp LoggerProvider) io.Writer {
	return lp.getLogger().Output
}

// Logger is an interface that allows for creation your own loggers with
// custom formatting. See PrettyLogger for how this is done.
type Logger interface {
	formatMessage(s string, i int) string
	getLogger() *AnyLogger
}

// LoggerProvider is an interface used to allow access to the struct fields
// of AnyLoggers. Somewhat of an golang hack.
type LoggerProvider interface {
	getLogger() *AnyLogger
}
