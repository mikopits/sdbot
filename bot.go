package sdbot

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	LoginURL = "https://play.pokemonshowdown.com/action.php"
)

type Bot struct {
	Config     *Config
	Connection *Connection    // The websocket connection
	Loggers    *Logger        // The logger used to debug TODO: list of loggers
	UserList   *map[User]bool // List of all users the bot knows about
	RoomList   *map[Room]bool // List of all rooms the bot knows about
	Rooms      *map[Room]bool // List of all the rooms the bot is in
	Nick       string         // The bot's username
}

// Creates a new bot instance.
func NewBot() *Bot {
	b := &Bot{
		Config:   ReadConfig(),
		UserList: &make(map[User]bool),
		RoomList: &make(map[Room]bool),
	}
	b.Nick = b.Config.Nick
	b.Connection = &make(Connection{
		Bot:       b,
		Connected: false,
		inQueue:   make(chan string, 128),
		outQueue:  make(chan string, 128),
	})
	return b
}

// Connects to the Pokemon Showdown server.
func (b *Bot) Login(msg *Message) {
	var res *http.Response
	var err error

	if b.Config.Password == "" {
		res, err = http.Get(strings.Join([]string{
			LoginURL,
			"?act=getassertion&userid=",
			Sanitize(b.Config.Nick),
			"&challengekeyid=",
			msg.Params[0],
			"&challenge=",
			msg.Params[1],
		}, ""))
	} else {
		res, err = http.PostForm(LoginURL, url.Values{
			"act":            {"login"},
			"name":           {b.Config.Nick},
			"pass":           {b.Config.Password},
			"challengekeyid": {msg.Params[0]},
			"challenge":      {msg.Params[1]},
		})
	}
	if err != nil {
		log.Println("login:", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("login ioutil read:", err)
	}

	if b.Config.Password == "" {
		b.Connection.QueueMessage(strings.Join([]string{
			"/trn ",
			b.Config.Nick,
			",0,",
			string(body),
		}, ""))
	} else {
		type LoginDetails struct {
			Assertion string
		}
		data := LoginDetails{}
		err = json.Unmarshal(body[1:], &data)
		if err != nil {
			log.Println("login json unmarshal:", err)
		}

		b.Connection.QueueMessage(strings.Join([]string{
			"|/trn ",
			b.Config.Nick,
			",0,",
			data.Assertion,
		}, ""))
	}
}

// Joins a room.
func (b *Bot) JoinRoom(room *Room) {
	b.Rooms[*room] = true
	b.RoomList[*room] = true
	b.Connection.QueueMessage("|/join " + room.Name)
}

// Leaves a room.
func (b *Bot) LeaveRoom(room *Room) {
	delete(b.Rooms, *room)
	b.RoomList[*room] = true
	b.Connection.QueueMessage("|/leave " + room.Name)
}
