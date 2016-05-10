package sdbot

import (
	"errors"
	"fmt"
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
	queue     chan string
	rk        Killable
}

// TODO Automatic reconnection to the socket.
// Connects to the server websocket and initialize reading and writing threads.
func (c *Connection) connect() {
	host := c.Bot.Config.Server + ":" + c.Bot.Config.Port
	u := url.URL{Scheme: "ws", Host: host, Path: "/showdown/websocket"}
	Info(&Log, fmt.Sprintf("Connecting to %s...", u.String()))

	var res *http.Response
	var err error
	dialer := websocket.DefaultDialer
	c.conn, res, err = dialer.Dial(u.String(), http.Header{
		"Origin": []string{"https://play.pokemonshowdown.com"},
	})
	CheckErr(err)

	c.Connected = true

	defer res.Body.Close()
	defer c.conn.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	c.startReading()
	c.startSending()
	defer wg.Done()
	wg.Wait()
}

// ErrUnexpectedMessageType is returned when we receive a message from the
// websocket that isn't a websocket.TextMessage or a normal closure.
var ErrUnexpectedMessageType = errors.New("sdbot: unexpected message type from the websocket")

// Listens for messages from the websocket.
// TODO gracefully break out of this loop on interrupt.
func (c *Connection) startReading() {
	go func() {
		for {
			select {
			case <-c.rk.Dying():
				// Break out of the read loop when we kill its Killable.
				return
			default:
				msgType, msg, err := c.conn.ReadMessage()
				CheckErr(err)

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
	}()
}

// Initiates the message sending goroutine.
func (c *Connection) startSending() {
	go func() {
		interrupt = make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		done := make(chan struct{})

		for {
			select {
			case msg := <-c.queue:
				Send(c, msg)
				ms := 1000.0 / c.Bot.Config.MessagesPerSecond
				time.Sleep(time.Duration(ms) * time.Millisecond)
			case <-interrupt:
				Warn(&Log, "Process was interrupted. Closing connection...")

				// Send a close frame and wait for the server to close the connection.
				err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					Error(&Log, err)
					return
				}

				// Kill the reading goroutine.
				c.rk.Kill()
				c.rk.Wait()

				// Close all plugin goroutines gracefully.
				var wg sync.WaitGroup
				wg.Add(1)
				c.Bot.StopTimedPlugins()
				c.Bot.UnregisterPlugins()
				defer wg.Done()
				wg.Wait()

				select {
				case <-done:
					os.Exit(0)
				case <-time.After(time.Second):
					os.Exit(0)
				}
				c.conn.Close()
				c.Connected = false
				return
			}
		}
	}()
}

// Adds a message to the outgoing queue. Prefer to use this over Send.
func (c *Connection) QueueMessage(msg string) {
	c.queue <- msg
}

// Sends a message upstream to the websocket.
func Send(c *Connection, s string) {
	Outgoing(&Log, s)

	err := c.conn.WriteMessage(websocket.TextMessage, []byte(s))
	CheckErr(err)
}

// Parses the message and difers it to a relevant handler.
func (c *Connection) parse(s string) {
	msg := NewMessage(s, c.Bot)

	// Log the incoming messages to every logger.
	IncomingAll(ActiveLoggers, s)

	cmd := strings.ToLower(msg.Command)

	switch cmd {
	case ":":
		LoginTime = msg.Timestamp
	}

	callHandler(handlers, cmd, msg)
}
