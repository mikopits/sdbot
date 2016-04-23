package sdbot

import (
	"fmt"
)

const (
	Administrator = `~`
	TheImmortal   = `>`
	Leader        = `&`
	RoomOwner     = `#`
	Driver        = `%`
	Battler       = `★`
	Voiced        = `+`
	Unvoiced      = ` `
	Muted         = `?`
	Locked        = `‽`
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

// Responds to a user in private message.
func (u *User) Reply(res string, msg *Message, bot *Bot) {
	bot.Connection.QueueMessage(fmt.Sprintf("|/pm %s,(%s) %s", u.Name, msg.User.Name, res))
}

// Responds to a user in a room.
func (r *Room) Reply(res string, msg *Message, bot *Bot) {
	bot.Connection.QueueMessage(fmt.Sprintf("%s|(%s) %s", r.Name, msg.User.Name, res))
}

// Responds to a user in private message without prepending their username.
func (u *User) RawReply(res string, msg *Message, bot *Bot) {
	bot.Connection.QueueMessage(fmt.Sprintf("|/pm %s,%s", u.Name, res))
}

// Responds to a user in a room without prepending their username.
func (r *Room) RawReply(res string, msg *Message, bot *Bot) {
	bot.Connection.QueueMessage(fmt.Sprintf("%s|%s", r.Name, res))
}

// Add a room authority to the user.
func (u *User) AddAuth(room string, auth string) {
	u.Auths[Sanitize(room)] = auth
}

// Adds a user to the room.
func (r *Room) AddUser(name string) {
	sn := Sanitize(name)
	for _, n := range r.Users {
		if n == sn {
			return
		}
	}

	r.Users = append(r.Users, sn)
}

// Removes a user from the room.
func (r *Room) RemoveUser(name string) {
	sn := Sanitize(name)
	for i, n := range r.Users {
		if n == sn {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			return
		}
	}
}

// Finds a user if it exists, creates the user if it doesn't.
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

// Finds a room if it exists, creates the room if it doesn't.
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

// Renames a user.
func Rename(old string, s string, bot *Bot) {
	so := Sanitize(old)
	sn := Sanitize(s)

	var rename = func() interface{} {
		if bot.UserList[so] != nil {
			u := bot.UserList[so]
			delete(bot.UserList, so)
			u.Name = s
			bot.UserList[sn] = u
		}
		return nil
	}

	bot.Synchronize("user", &rename)
}

// A Target can be either a Room or a User. It represents where the bot will
// send its message in response.
type Target interface {
	Reply(string, *Message, *Bot)
}
