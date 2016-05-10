package sdbot

type LoggerList struct {
	Loggers []Logger
}

func NewLoggerList(loggers ...Logger) *LoggerList {
	return &LoggerList{Loggers: loggers}
}

func DebugAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		Debug(&lo, s)
	}
}

func InfoAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		Info(&lo, s)
	}
}

func WarnAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		Warn(&lo, s)
	}
}

func ErrorAll(lol *LoggerList, err error) {
	for _, lo := range lol.Loggers {
		Error(&lo, err)
	}
}

func FatalAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		Fatal(&lo, s)
	}
}

func IncomingAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		Incoming(&lo, s)
	}
}

func OutgoingAll(lol *LoggerList, s string) {
	for _, lo := range lol.Loggers {
		Outgoing(&lo, s)
	}
}

func DebugAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		Debugf(&lo, format, a...)
	}
}

func InfoAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		Infof(&lo, format, a...)
	}
}

func WarnAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		Warnf(&lo, format, a...)
	}
}

func ErrorAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		Errorf(&lo, format, a...)
	}
}

func FatalAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		Fatalf(&lo, format, a...)
	}
}

func IncomingAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		Incomingf(&lo, format, a...)
	}
}

func OutgoingAllf(lol *LoggerList, format string, a ...interface{}) {
	for _, lo := range lol.Loggers {
		Outgoingf(&lo, format, a...)
	}
}
