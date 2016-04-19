package sdbot

import (
	"bytes"
)

type User struct {
	Name string
}

type Room struct {
	Name string
}

func (u *User) Reply(res string, msg *Message, bot *Bot) {
	var buffer bytes.Buffer
	buffer.WriteString("/pm ")
	buffer.WriteString(msg.User.Name)
	buffer.WriteString("|(")
	buffer.WriteString(u.Name)
	buffer.WriteString(") ")
	buffer.WriteString(res)
	bot.Connection.QueueMessage(buffer.String())
}

func (r *Room) Reply(res string, msg *Message, bot *Bot) {
	var buffer bytes.Buffer
	buffer.WriteString(r.Name)
	buffer.WriteString("|(")
	buffer.WriteString(msg.User.Name)
	buffer.WriteString(") ")
	buffer.WriteString(res)
	bot.Connection.QueueMessage(buffer.String())
}

type Target interface {
	Reply(string, *Message, *Bot)
}
