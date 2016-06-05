package sdbot

import (
	"fmt"
)

// String values of all the auth levels.
const (
	Administrator = `~`
	Leader        = `&`
	RoomOwner     = `#`
	Moderator     = `@`
	Driver        = `%`
	TheImmortal   = `>`
	Battler       = `★`
	Voiced        = `+`
	Unvoiced      = ` `
	Muted         = `?`
	Locked        = `‽`
)

var authLevels = map[string]int{
	Locked:        0,
	Muted:         1,
	Unvoiced:      2,
	Voiced:        3,
	Battler:       4,
	TheImmortal:   5,
	Driver:        6,
	Moderator:     7,
	RoomOwner:     8,
	Leader:        9,
	Administrator: 10,
}

// User represents a user with their username and the auth levels in rooms
// that the bot knows about.
type User struct {
	Name  string
	Auths map[string]string
}

// Room represents a room with its name and the users currently in it.
type Room struct {
	Name  string
	Users []string // List of unique SANITIZED names of users in the room
}

// NewUser creates a new User, initializing the Auths map.
func NewUser(name string) *User {
	return &User{
		Name:  name,
		Auths: make(map[string]string),
	}
}

// Reply responds to a user in private message and prepends the user's name to
// the response.
func (u *User) Reply(m *Message, res string) {
	s := fmt.Sprintf("(%s) %s", m.User.Name, res)
	if len(s) > 300 {
		s = s[:300]
	}
	m.Bot.Connection.QueueMessage(fmt.Sprintf("|/w %s,%s", u.Name, s))
}

// Reply responds to a user in a chat message and prepends the user's name to
// the response. The message is sent to the Room of the method's receiver.
func (r *Room) Reply(m *Message, res string) {
	s := fmt.Sprintf("(%s) %s", m.User.Name, res)
	if len(s) > 300 {
		s = s[:300]
	}
	m.Bot.Connection.QueueMessage(fmt.Sprintf("%s|%s", r.Name, s))
}

// RawReply responds to a user in private message without prepending their
// username.
func (u *User) RawReply(m *Message, res string) {
	if len(res) > 300 {
		res = res[:300]
	}
	m.Bot.Connection.QueueMessage(fmt.Sprintf("|/w %s,%s", u.Name, res))
}

// RawReply responds to a user in a room without prepending their username.
func (r *Room) RawReply(m *Message, res string) {
	if len(res) > 300 {
		res = res[:300]
	}
	m.Bot.Connection.QueueMessage(fmt.Sprintf("%s|%s", r.Name, res))
}

// AddAuth adds a room authority level to a user.
func (u *User) AddAuth(room string, auth string) {
	u.Auths[Sanitize(room)] = auth
}

// AddUser adds a user to the room.
func (r *Room) AddUser(name string) {
	sn := Sanitize(name)
	for _, n := range r.Users {
		if n == sn {
			return
		}
	}

	r.Users = append(r.Users, sn)
}

// RemoveUser removes a user from the room.
func (r *Room) RemoveUser(name string) {
	sn := Sanitize(name)
	for i, n := range r.Users {
		if n == sn {
			r.Users = append(r.Users[:i], r.Users[i+1:]...)
			return
		}
	}
}

// FindUserEnsured finds a user if it exists, creates the user if it doesn't.
func FindUserEnsured(name string, b *Bot) *User {
	sn := Sanitize(name)

	var updateUsers = func() interface{} {
		if b.UserList[sn] != nil {
			return b.UserList[sn]
		}

		user := NewUser(name)
		b.UserList[sn] = user

		return user
	}

	return b.Synchronize("room", &updateUsers).(*User)
}

// FindRoomEnsured finds a room if it exists, creates the room if it doesn't.
func FindRoomEnsured(name string, b *Bot) *Room {
	sn := Sanitize(name)

	var updateRooms = func() interface{} {
		if b.RoomList[sn] != nil {
			return b.RoomList[sn]
		}

		room := &Room{Name: name}
		b.RoomList[sn] = room

		return room
	}

	return b.Synchronize("user", &updateRooms).(*Room)
}

// Rename renames a user and updates their record in the UserList.
func Rename(old string, s string, r *Room, b *Bot, auths ...string) {
	so := Sanitize(old)
	sn := Sanitize(s)

	var rename = func() interface{} {
		if so == sn {
			b.UserList[so].Name = s
		} else {
			r.RemoveUser(so)
			r.AddUser(sn)
			FindUserEnsured(s, b)
			u := b.UserList[sn]
			for _, a := range auths {
				u.AddAuth(r.Name, a)
			}
		}
		return nil
	}

	b.Synchronize("user", &rename)
}

// HasAuth checks if a user has AT LEAST a given authorization level in a given room.
func (u *User) HasAuth(roomname string, level string) bool {
	return authLevels[u.Auths[roomname]] >= authLevels[level]
}

// Target represents either a Room or a User. The distinction is in where the bot will
// send its message in response.
type Target interface {
	Reply(*Message, string)
	RawReply(*Message, string)
}
