package sdbot

// LoggerList represents a list of Loggers with methods that allow you to
// log to every one of them as per each loggers' individual logging behaviour.
type LoggerList struct {
	Loggers []Logger
}

// NewLoggerList creates a new LoggerList from a slice of Loggers.
func NewLoggerList(loggers ...Logger) *LoggerList {
	return &LoggerList{Loggers: loggers}
}

// loggers is the list of loggers that the bot will log to. Access to these
// loggers is provided in the helpers in helpers.go exports.
var loggers *LoggerList

// AddLogger adds a logger to the LoggerList Loggers.
func AddLogger(lo Logger) {
	loggers.Loggers = append(loggers.Loggers, lo)
}

// RemoveLogger removes a logger from the LoggerLit Loggers.
// Returns true if the logger was successfully removed.
func RemoveLogger(lo Logger) bool {
	for i, logger := range loggers.Loggers {
		if logger == lo {
			loggers.Loggers = append(loggers.Loggers[:i], loggers.Loggers[i+1:]...)
			return true
		}
	}
	return false
}

// Log debug messages to all loggers in the LoggerList.
func logDebugAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		logDebug(&lo, s)
	}
}

// Log informatic messages to all loggers in the LoggerList.
func logInfoAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		logInfo(&lo, s)
	}
}

// Log warning messages to all loggers in the LoggerList.
func logWarnAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		logWarn(&lo, s)
	}
}

// Log errors to all loggers in the LoggerList.
func logErrorAll(lol *LoggerList, err error) {
	for _, lo := range lol.Loggers {
		logError(&lo, err)
	}
}

// Log fatal messages to all loggers in the LoggerList.
func logFatalAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		logFatal(&lo, s)
	}
}

// Log messages from the websocket to all loggers in the LoggerList.
func logIncomingAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		logIncoming(&lo, s)
	}
}

// Log messages being sent to the websocket to all loggers in the LoggerList.
func logOutgoingAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		logOutgoing(&lo, s)
	}
}

// Log debug messages with arguments to all loggers in the LoggerList.
func logDebugAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		logDebugf(&lo, format, a...)
	}
}

// Log informatic messages with arguments to all loggers in the LoggerList.
func logInfoAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		logInfof(&lo, format, a...)
	}
}

// Log warning messages with arguments to all loggers in the LoggerList.
func logWarnAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		logWarnf(&lo, format, a...)
	}
}

// Log errors with arguments to all loggers in the LoggerList.
func logErrorAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		logErrorf(&lo, format, a...)
	}
}

// Log fatal messages with arguments to all loggers in the LoggerList.
func logFatalAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		logFatalf(&lo, format, a...)
	}
}

// Log messages from the websocket with arguments to all loggers in the
// LoggerList. TODO Probably don't need this.
func logIncomingAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		logIncomingf(&lo, format, a...)
	}
}

// Log messages being sent to the websocket with arguments to all loggers in
// the LoggerList. TODO Probably don't need this.
func logOutgoingAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		logOutgoingf(&lo, format, a...)
	}
}
