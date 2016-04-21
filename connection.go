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

var PingTicker *time.Ticker

var interrupt chan os.Signal

// ErrUnexpectedMessageType is returned when we receive a message from the
// websocket that isn't a websocket.TextMessage or websocket.CloseNormalClosure.
var ErrUnexpectedMessageType = errors.New("sdbot: unexpected message type from the websocket")

type Connection struct {
	Bot       *Bot
	Connected bool
	ws        *websocket.Conn
	inQueue   chan string
	outQueue  chan string
}

// Connect to the server websocket and initialize reading and writing threads.
func (c *Connection) Connect() {
	host := c.Bot.Config.Server + ":" + c.Bot.Config.Port
	u := url.URL{Scheme: "ws", Host: host, Path: "/showdown/websocket"}
	Info(&Log, fmt.Sprintf("Connecting to %s...", u.String()))

	var res *http.Response
	var err error
	c.ws, res, err = websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Origin": []string{"https://play.pokemonshowdown.com"},
	})
	if err != nil {
		Error(&Log, err)
	}

	c.Connected = true

	PingTicker = time.NewTicker(time.Minute)
	c.ws.SetPongHandler(func(s string) error {
		Info(&Log, fmt.Sprintf("Received pong: %s", s))
		return nil
	})

	defer res.Body.Close()
	defer c.ws.Close()
	defer PingTicker.Stop()

	var wg sync.WaitGroup
	wg.Add(2)

	c.startReadingThread()
	defer wg.Done()
	c.startSendingThread()
	defer wg.Done()
	wg.Wait()
}

func (c *Connection) startReadingThread() {
	// Listen for messages from the websocket
	go func() error {
		for {
			msgType, msg, err := c.ws.ReadMessage()
			if err != nil {
				Error(&Log, err)
			}

			if msgType != websocket.TextMessage || msgType != websocket.CloseNormalClosure {
				err = ErrUnexpectedMessageType
				Error(&Log, err)
				return err
			}

			Incoming(&Log, string(msg))
			c.inQueue <- string(msg)
		}
		return nil
	}()

	// Parse the messages and do stuff with them
	go func() {
		for {
			select {
			case event := <-c.inQueue:
				var room string
				messages := strings.Split(event, `\n`)

				if string([]rune(messages[0])[0]) == ">" {
					room, messages = messages[0], messages[1:]
				}

				for _, rawMessage := range messages {
					parse(strings.Join([]string{room, `\n`, rawMessage}, ""), c.Bot)
				}
			case <-interrupt:
				return
			}
		}
	}()
}

func (c *Connection) startSendingThread() {
	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	go func() {
		for {
			select {
			case msg := <-c.outQueue:
				Send(c, msg)
				ms := math.Floor(1000.0 / c.Bot.Config.MessagesPerSecond)
				time.Sleep(time.Duration(ms) * time.Millisecond)
			case <-interrupt:
				Warn(&Log, "Process was interrupted. Closing connection...")

				// Send a close frame and wait for the server to close the connection.
				err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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
				c.ws.Close()
				c.Connected = false
				PingTicker.Stop()
				return
			}
		}
	}()
}

// Adds a message to the outgoing queue.
func (c *Connection) QueueMessage(msg string) {
	c.outQueue <- msg
}

// Sends a message upstream to the websocket.
func Send(c *Connection, s string) {
	Outgoing(&Log, s)

	err := c.ws.WriteMessage(websocket.TextMessage, []byte(s))
	if err != nil {
		Error(&Log, err)
	}
}

func parse(s string, b *Bot) {
	if strings.TrimRight(s, "\n") == "" {
		return
	}

	msg := NewMessage(s, b)
	cmd := strings.ToLower(msg.Command)
	events := make(chan string, 64)

	switch cmd {
	case ":":
		LoginTime = msg.Timestamp
		Debug(&Log, fmt.Sprintf("LoginTime: %d", LoginTime))
	case "c":
		if msg.Params[0] != "~" {
			events <- "message"
		}
	}

	events <- cmd
	CallHandler(Handlers, cmd, msg, events)
}
