package sdbot

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

const (
	Reset   string = "\x1b[0m"
	Bold    string = "\x1b[1m"
	Red     string = "\x1b[31m"
	Green   string = "\x1b[32m"
	Yellow  string = "\x1b[33m"
	Blue    string = "\x1b[34m"
	Black   string = "\x1b[30m"
	BgWhite string = "\x1b[47m"
)

type PrettyLogger struct {
	AnyLogger
}

func (lo *PrettyLogger) formatMessage(s string, level int) string {
	switch level {
	case LevelDebug:
		fallthrough
	case LevelWarn:
		return formatDebug(s)
	case LevelInfo:
		return formatInfo(s)
	case LevelError:
		return formatError(s)
	case LevelIncoming:
		return formatIncoming(s)
	case LevelOutgoing:
		return formatOutgoing(s)
	default:
		return formatGeneral(s)
	}
}

func formatDebug(s string) string {
	return fmt.Sprintf("%s %s %s\n", timestamp(), colourize("!!", Yellow), s)
}

func formatInfo(s string) string {
	return fmt.Sprintf("%s %s %s\n", timestamp(), "II", s)
}

func formatError(s string) string {
	// TODO: Make the stack trace output readable
	debug.PrintStack()
	return fmt.Sprintf("%s %s %s\n", timestamp(), colourize("!!", Red), s)
}

func formatGeneral(s string) string {
	return fmt.Sprintf("%s %s\n", timestamp(), s)
}

func formatIncoming(s string) string {
	split := strings.Split(s, "\n")
	if len(split) < 2 {
		return formatGeneral(s)
	}
	room := split[0]
	rest := split[1]

	parts := strings.Split(rest, "|")
	prefix := colourize(">>", Green)

	if room == "" {
		// Private messages
		return fmt.Sprintf("%s %s %s|%s\n", timestamp(), prefix, colourize(parts[0], Blue), strings.Join(parts[1:], "|"))
	}

	room = colourize(room[1:], Bold)

	if len(parts) == 0 {
		return fmt.Sprintf("%s %s %s\n", timestamp(), prefix, room)
	}

	if len(parts) == 1 {
		// Raw server messages
		return fmt.Sprintf("%s %s %s|%s\n", timestamp(), prefix, room, colourize(parts[0], Red))
	}

	cmd := colourize(parts[1], Blue)
	params := strings.Join(parts[2:], "|")

	return fmt.Sprintf("%s %s %s|%s|%s\n", timestamp(), prefix, room, cmd, params)
}

func formatOutgoing(s string) string {
	split := strings.Split(s, "|")
	room := colourize(split[0], Bold)
	rest := split[1]
	prefix := colourize("<<", Red)

	return fmt.Sprintf("%s %s %s|%s\n", timestamp(), prefix, room, rest)
}

func timestamp() string {
	return strings.Join([]string{"[", time.Now().Format(time.RFC1123), "]"}, "")
}

func colourize(s string, codes ...string) string {
	var buffer bytes.Buffer
	for _, code := range codes {
		buffer.WriteString(code)
	}
	codeStr := buffer.String()

	return strings.Join([]string{codeStr, s, Reset}, "")
}
