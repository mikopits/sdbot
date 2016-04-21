package sdbot

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// Define function handlers to call depending on the command we get.
var Handlers = map[string]interface{}{
	"challstr":   onChallstr,
	"updateuser": onUpdateuser,
	"l":          onLeave,
	"j":          onJoin,
	"n":          onNick,
	"init":       onInit,
	"deinit":     onDeinit,
	"users":      onUsers,
	"popup":      onPopup,
	"c:":         onChat,
	"pm":         onPrivateMessage,
	"tournament": onTournament,
}

func CallHandler(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	if !f.IsValid() {
		return
	}

	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

func onChallstr(msg *Message, events chan string) {
	Info(&Log, "Attempting to log in...")
	msg.Bot.Login(msg)
}

var AvatarSet bool = false

func onUpdateuser(msg *Message, events chan string) {
	switch msg.Params[1] {
	case "0":
		msg.Bot.Connection.QueueMessage("|/avatar " + msg.Bot.Config.Avatar)
		AvatarSet = true
	case "1":
		for _, r := range msg.Bot.Config.Rooms {
			room := &Room{Name: strings.ToLower(r)}
			msg.Bot.RoomList[room] = true
			msg.Bot.JoinRoom(room)
		}
	}
}

func onLeave(msg *Message, events chan string) {
	// TODO Add Room logic
	if msg.User.Name == msg.Bot.Nick {
		delete(msg.Bot.Rooms, msg.Room)
	}
}

func onJoin(msg *Message, events chan string) {
	// TODO Add Room logic
	if msg.User.Name == msg.Bot.Nick {
		onInit(msg, events)
	}
}

func onNick(msg *Message, events chan string) {
	// TODO Add Room logic
	oldNick := Sanitize(msg.Params[0])
	if oldNick == Sanitize(msg.Bot.Nick) {
		msg.Bot.Nick = msg.User.Name
	} else {
		delete(msg.Bot.UserList, &User{Name: oldNick})
		msg.Bot.UserList[&User{Name: Sanitize(msg.User.Name)}] = true
	}
}

func onInit(msg *Message, events chan string) {
	// This may occur if the bot is redirected. Leave the room if it
	// is not a room it should be in, and try to rejoin any rooms that
	// it should be in.
	//
	// Note that a successful /join will trigger another init event.
	// So be careful to not cause an infinite loop.
	if !includes(msg.Bot.Config.Rooms, msg.Room.Name) {
		msg.Bot.LeaveRoom(msg.Room)
		// Try to join each of the config rooms, as you may have been redirected.
		// It is safe to call "/join [room]" if you are already in it, so there is
		// not really a need to check.
		for _, room := range msg.Bot.Config.Rooms {
			msg.Bot.JoinRoom(&Room{Name: room})
		}
	}
}

func includes(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}
	return false
}

func onDeinit(msg *Message, events chan string) {
	// TODO Add room logic
	msg.Bot.Rooms[msg.Room] = false
}

func onUsers(msg *Message, events chan string) {
	// TODO Populate the room with its users and their auth levels
}

func onPopup(msg *Message, events chan string) {
	// Handle bans
	if strings.Contains(msg.Params[2], "has banned you from the room") {
		reg, err := regexp.Compile("(?P<user>[^ ]+) has banned you from the room (?P<room>[^ ]*).</p><p>To appeal")
		if err != nil {
			Error(&Log, err)
		}
		match := reg.FindStringSubmatch(msg.Params[2])
		result := make(map[string]string)
		for i, name := range reg.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		user, room := result["user"], result["room"]
		Warn(&Log, "You have been banned from the room "+room+" by the user "+user)
		// TODO: This is a lazy solution, but the server no longer notifies you when
		// you are unbanned. Try to rejoin after a potential kick.
		time.Sleep(time.Second)
		msg.Bot.JoinRoom(&Room{Name: room})
	}
}

func onChat(msg *Message, events chan string) {
	Debug(&Log, "in onChat")
	if LoginTime == 0 {
		return
	}
	if msg.Message != "" && msg.Timestamp >= LoginTime {
		events <- "message"
		ChatEvents <- msg
	}
}

func onPrivateMessage(msg *Message, events chan string) {
	events <- "private"
	PrivateEvents <- msg
}

func onTournament(msg *Message, events chan string) {
	param := msg.Params[0]
	// Tournaments are very noisy and send a lot of information we don't need,
	// only want to listen to tournament create, update, and start events. You
	// can still use the "tournament" event for other purposes.
	if param == "create" || param == "update" || param == "start" {
		events <- "tour"
	}
}
