package sdbot

import (
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Bot       *Bot
	Time      time.Time
	Command   string
	Params    []string
	Timestamp int
	Room      *Room
	User      *User
	Auth      string
	Target    Target
	Message   string
}

func NewMessage(rawMessage string, bot *Bot) *Message {
	m := &Message{
		Bot:  bot,
		Time: time.Now(),
	}
	m.Command, m.Params, m.Timestamp, m.Room, m.User, m.Auth, m.Target, m.Message = parseMessage(rawMessage, bot)
	return m
}

// Parse a raw message and return data in the following order:
// Command, Params, Timestamp, Room, User, Auth, Target, Message
func parseMessage(s string, b *Bot) (string, []string, int, *Room, *User, string, Target, string) {
	newlineDelimited := strings.Split(s, "\n")
	vertbarDelimited := strings.Split(s, "|")
	var command string
	var params []string
	var timestamp int
	var room *Room
	var user *User
	var auth string
	var message string

	// The command is always after the first vertical bar.
	if len(vertbarDelimited) < 2 {
		command = "none"
	} else {
		command = string(vertbarDelimited[1])
	}

	// Parse the parameters following a command.
	if command == "" {
		params = []string{}
	} else {
		params = vertbarDelimited[2:]
	}

	// Parse the timestamp of a chat event.
	var err error
	if strings.Contains(command, ":") {
		timestamp, err = strconv.Atoi(params[0])
		if err != nil {
			Error(&Log, err)
		}
	} else {
		timestamp = 0
	}

	// If the message starts with a ">" then it comes from a room.
	if string(newlineDelimited[0][0]) == ">" {
		room = &Room{Name: string(newlineDelimited[0][1:])}
		b.RoomList[room] = true
	} else {
		room = &Room{}
	}

	// Parse the user sending a command, and their auth level.
	switch strings.ToLower(command) {
	case "c:":
		auth = string(vertbarDelimited[3][0])
		user = &User{Name: string(vertbarDelimited[3][1:])}
		b.UserList[user] = true
	case "c":
		fallthrough
	case "j":
		fallthrough
	case "l":
		fallthrough
	case "n":
		fallthrough
	case "pm":
		auth = string(vertbarDelimited[2][0])
		user = &User{Name: string(vertbarDelimited[2][1:])}
		b.UserList[user] = true
	}

	// Parse the message
	if command == "" {
		message = ""
	} else {
		switch strings.ToLower(command) {
		case "c:", "pm":
			message = strings.Join(vertbarDelimited[4:], "|")
		case "none":
			message = newlineDelimited[len(newlineDelimited)-1]
		}
	}

	// Decide the target
	if strings.ToLower(command) == "pm" {
		return command, params, timestamp, room, user, auth, user, message
	}
	return command, params, timestamp, room, user, auth, room, message
}

// Reply to a message
func (m *Message) Reply(res string) {
	m.Target.Reply(res, m, m.Bot)
}
