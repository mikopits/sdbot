package sdbot

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// The colour codes for tty to be used with colourize.
const (
	reset   string = "\x1b[0m"
	bold    string = "\x1b[1m"
	red     string = "\x1b[31m"
	green   string = "\x1b[32m"
	yellow  string = "\x1b[33m"
	blue    string = "\x1b[34m"
	black   string = "\x1b[30m"
	bgWhite string = "\x1b[47m"
)

// PrettyLogger logs everything with pretty colours. Looks good with the
// Solarized terminal theme.
// Don't like how it looks? Make your own!
type PrettyLogger struct {
	AnyLogger
}

func (lo *PrettyLogger) formatMessage(s string, level int) string {
	switch level {
	case levelDebug:
		fallthrough
	case levelWarn:
		return formatDebug(s)
	case levelInfo:
		return formatInfo(s)
	case levelError:
		return formatError(s)
	case levelIncoming:
		return formatIncoming(s)
	case levelOutgoing:
		return formatOutgoing(s)
	default:
		return formatGeneral(s)
	}
}

func formatDebug(s string) string {
	return fmt.Sprintf("%s %s %s\n", timestamp(), colourize("!!", yellow), s)
}

func formatInfo(s string) string {
	return fmt.Sprintf("%s %s %s\n", timestamp(), "II", s)
}

func formatError(s string) string {
	return fmt.Sprintf("%s %s %s\n", timestamp(), colourize("!!", red), s)
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
	prefix := colourize(">>", green)

	if room == "" {
		// Private messages
		if len(parts) > 1 {
			parts = parts[1:]
			return fmt.Sprintf("%s %s %s|%s\n", timestamp(), prefix, colourize(parts[0], blue), strings.Join(parts[1:], "|"))
		} else {
			return fmt.Sprintf("%s %s %s\n", timestamp(), prefix, parts[0])
		}
	}

	room = colourize(room[1:], bold)

	if len(parts) == 0 {
		return fmt.Sprintf("%s %s %s\n", timestamp(), prefix, room)
	}

	if len(parts) == 1 {
		// Raw server messages/ban messages
		return fmt.Sprintf("%s %s %s|%s\n", timestamp(), prefix, room, colourize(parts[0], bold))
	}

	cmd := colourize(parts[1], blue)
	params := strings.Join(parts[2:], "|")

	return fmt.Sprintf("%s %s %s|%s|%s\n", timestamp(), prefix, room, cmd, params)
}

func formatOutgoing(s string) string {
	split := strings.Split(s, "|")
	room := colourize(split[0], bold)
	rest := split[1]
	prefix := colourize("<<", red)

	return fmt.Sprintf("%s %s %s|%s\n", timestamp(), prefix, room, rest)
}

func timestamp() string {
	return strings.Join([]string{"[", time.Now().Format(time.RFC3339), "]"}, "")
}

func colourize(s string, codes ...string) string {
	var buffer bytes.Buffer
	for _, code := range codes {
		buffer.WriteString(code)
	}

	return strings.Join([]string{buffer.String(), s, reset}, "")
}
