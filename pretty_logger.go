package sdbot

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	Reset   string = `\e[0m`
	Bold    string = `\e[1m`
	Red     string = `\e[31m`
	Green   string = `\e[32m`
	Yellow  string = `\e[33m`
	Blue    string = `\e[34m`
	Black   string = `\e[30m`
	BgWhite string = `\e[47m`
)

type PrettyLogger struct {
	Output io.Writer
	mutex  sync.Mutex
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
	return fmt.Sprintf("%s %s %s", timestamp(), colourize("!!", Yellow), s)
}

func formatInfo(s string) string {
	return fmt.Sprintf("%s %s %s", timestamp(), "II", s)
}

func formatError(s string) string {
	return fmt.Sprintf("%s %s %s", timestamp(), colourize("red"), s)
}

func formatIncoming(s string) string {
	split := strings.Split(s, `\n`)
	if len(split) < 2 {
		return formatGeneral(s)
	}
	room := split[0]
	rest := split[1]

	parts := strings.Split(rest, "|")
	prefix := colourize(">>", Green)

	if room == "" {
		// Private messages
		return fmt.Sprintf("%s %s %s|%s", timestamp(), prefix, colourize(parts[0], Blue), strings.Join(parts[1:], "|"))
	}

	room = colourize(room[1:], Bold)

	if len(parts) == 0 {
		return fmt.Sprintf("%s %s %s", timestamp(), prefix, room)
	}

	if len(parts) == 1 {
		// Raw server messages
		return fmt.Sprintf("%s %s %s|%s", timestamp(), prefix, room, colourize(parts[0], Red))
	}

	cmd := colourize(parts[1], Blue)
	params := strings.Join(parts[2:], "|")

	return fmt.Sprintf("%s %s %s|%s|%s", timestamps(), prefix, room, cmd, params)
}

func formatOutgoing(s string) string {
	split := strings.Split(s, "|")
	room := colourize(split[0], Bold)
	rest := split[1]
	prefix := colourize("<<", Red)

	return fmt.Sprintf("%s %s %s", timestamp(), colourize("!!", Red), s)
}

func timestamp() string {
	now := time.Now()
	buffer := bytes.Buffer
	buffer.WriteString("[")
	buffer.WriteString(now.Date())
	buffer.WriteString(" ")
	buffer.WriteString(now.Clock())
	buffer.WriteString("]")
	return buffer.String()
}

func colourize(s string, codes ...string) string {
	reg, err := regexp.Compile(regexp.QuoteMeta(Reset))
	if err != nil {
		log.Fatal(formatError(s))
	}

	buffer := bytes.Buffer
	for _, code := range codes {
		buffer.WriteString(code)
	}
	codeStr := buffer.String()

	text := reg.ReplaceAllString(s, strings.Join([]string{Reset, codeStr}, ""))

	return strings.Join([]string{codeStr, text, Reset}, "")
}
