package sdbot

import (
	"fmt"
)

const (
	Administrator = "~"
	TheImmortal   = ">"
	Leader        = "&"
	RoomOwner     = "#"
	Driver        = "%"
	Battler       = "★"
	Voiced        = "+"
	Unvoiced      = " "
	Muted         = "?"
	Locked        = "‽"
)

type User struct {
	Name  string
	Auths map[string]string // Map of room names to auth strings
}

type Room struct {
	Name  string
	Users []string // List of unique sanitized names of users in the room
}

func NewUser(name string) *User {
	return &User{
		Name:  name,
		Auths: make(map[string]string),
	}
}

// TODO Add an option or another method to reply without the user name.
func (u *User) Reply(res string, msg *Message, bot *Bot) {
	bot.Connection.QueueMessage(fmt.Sprintf("|/pm %s,(%s) %s", u.Name, msg.User.Name, res))
}

// TODO Add an option or another method to reply without the user name.
func (r *Room) Reply(res string, msg *Message, bot *Bot) {
	bot.Connection.QueueMessage(fmt.Sprintf("%s|(%s) %s", r.Name, msg.User.Name, res))
}

func (u *User) AddAuth(room string, auth string) {
	u.Auths[Sanitize(room)] = auth
}

func (r *Room) AddUser(name string) {
	sn := Sanitize(name)
	for _, n := range r.Users {
		if n == sn {
			return
		}
	}

	r.Users = append(r.Users, sn)
}

func (r *Room) RemoveUser(name string) {
	sn := Sanitize(name)
	for i, n := range r.Users {
		if n == sn {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			return
		}
	}
}

func FindUserEnsured(name string, bot *Bot) *User {
	sn := Sanitize(name)

	var updateUsers = func() interface{} {
		if bot.UserList[sn] != nil {
			return bot.UserList[sn]
		}

		user := NewUser(name)
		bot.UserList[sn] = user

		return user
	}

	return bot.Synchronize("room", &updateUsers).(*User)
}

func FindRoomEnsured(name string, bot *Bot) *Room {
	sn := Sanitize(name)

	var updateRooms = func() interface{} {
		if bot.RoomList[sn] != nil {
			return bot.RoomList[sn]
		}

		room := &Room{Name: sn}
		bot.RoomList[name] = room

		return room
	}

	return bot.Synchronize("user", &updateRooms).(*Room)
}

func Rename(old string, s string, bot *Bot) {
	so := Sanitize(old)
	sn := Sanitize(s)

	if bot.UserList[so] != nil {
		u := bot.UserList[so]
		delete(bot.UserList, so)
		u.Name = s
		bot.UserList[sn] = u
	}
}

type Target interface {
	Reply(string, *Message, *Bot)
}
