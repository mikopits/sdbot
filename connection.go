package sdbot

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var LoginTime int

var interrupt chan os.Signal

type Connection struct {
	Bot       *Bot
	Connected bool
	conn      *websocket.Conn
	outQueue  chan string
}

// Connect to the server websocket and initialize reading and writing threads.
func (c *Connection) Connect() {
	host := c.Bot.Config.Server + ":" + c.Bot.Config.Port
	u := url.URL{Scheme: "ws", Host: host, Path: "/showdown/websocket"}
	Info(&Log, fmt.Sprintf("Connecting to %s...", u.String()))

	var res *http.Response
	var err error
	dialer := websocket.DefaultDialer
	c.conn, res, err = dialer.Dial(u.String(), http.Header{
		"Origin": []string{"https://play.pokemonshowdown.com"},
	})
	if err != nil {
		Error(&Log, err)
	}

	c.Connected = true

	defer res.Body.Close()
	defer c.conn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go c.startReading()
	go c.startSending()
	defer wg.Done()
	wg.Wait()
}

// ErrUnexpectedMessageType is returned when we receive a message from the
// websocket that isn't a websocket.TextMessage or websocket.CloseNormalClosure.
var ErrUnexpectedMessageType = errors.New("sdbot: unexpected message type from the websocket")

// Listens for messages from the websocket.
// FIXME This routine will panic with a runtime error: index out of range if
// the bot is killed.
func (c *Connection) startReading() {
	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			Error(&Log, err)
		}

		if msgType != websocket.TextMessage && msgType != -1 {
			Error(&Log, ErrUnexpectedMessageType)
		}

		var room string
		messages := strings.Split(string(msg), "\n")
		if string(messages[0][0]) == ">" {
			room, messages = messages[0], messages[1:]
		}

		for _, rawmessage := range messages {
			c.parse(fmt.Sprintf("%s\n%s", room, rawmessage))
		}
	}

}

func (c *Connection) startSending() {
	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	for {
		select {
		case msg := <-c.outQueue:
			Send(c, msg)
			ms := math.Floor(1000.0 / c.Bot.Config.MessagesPerSecond)
			time.Sleep(time.Duration(ms) * time.Millisecond)
		case <-interrupt:
			Warn(&Log, "Process was interrupted. Closing connection...")

			// Send a close frame and wait for the server to close the connection.
			err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				Error(&Log, err)
				return
			}
			select {
			case <-done:
				os.Exit(15)
			case <-time.After(time.Second):
				os.Exit(15)
			}
			c.conn.Close()
			c.Connected = false
			return
		}
	}
}

// Adds a message to the outgoing queue.
func (c *Connection) QueueMessage(msg string) {
	c.outQueue <- msg
}

// Sends a message upstream to the websocket.
func Send(c *Connection, s string) {
	Outgoing(&Log, s)

	err := c.conn.WriteMessage(websocket.TextMessage, []byte(s))
	if err != nil {
		Error(&Log, err)
	}
}

func (c *Connection) parse(s string) {
	msg := NewMessage(s, c.Bot)
	events := make(chan string, 16)

	Incoming(&Log, s)

	cmd := strings.ToLower(msg.Command)

	switch cmd {
	case ":":
		LoginTime = msg.Timestamp
	case "c":
		if msg.Params[0] != "~" {
			events <- "message"
		}
	}

	events <- cmd

	CallHandler(Handlers, cmd, msg, events)
}
