package sdbot

import (
	"bytes"
	"errors"
	"log"
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

// Define function handlers to call depending on the command we get.
var handlers = map[string]interface{}{
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

// The time at which the bot logs in.
var LoginTime int

// Timer used to periodically ping the server.
var PingTicker *time.Ticker

// ErrUnexpectedMessageType is returned when we receive a message from the
// websocket that isn't a websocket.TextMessage.
var ErrUnexpectedMessageType = errors.New("sdbot: unexpected message type from the websocket")

type Connection struct {
	Bot       *Bot
	Connected bool
	ws        *websocket.Conn
	inQueue   chan string
	outQueue  chan string
}

//func NewConnection(bot *Bot) *Connection {
//	return &Connection{
//		Bot:       bot,
//		Connected: false,
//		inQueue:   make(chan string, 128),
//		outQueue:  make(chan string, 128),
//	}
//}

// Connect to the server websocket and initialize reading and writing threads.
func (c *Connection) Connect() {
	host := c.Bot.Config.Server + ":" + c.Bot.Config.Port
	u := url.URL{Scheme: "ws", Host: host, Path: "/showdown/websocket"}
	log.Printf("connecting to %s", u.String())

	var res *http.Response
	var err error
	c.ws, res, err = websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Origin": []string{"https://play.pokemonshowdown.com"},
	})
	if err != nil {
		log.Fatal("dial:", err)
	}

	c.Connected = true

	PingTicker = time.NewTicker(time.Minute)
	c.ws.SetPongHandler(func(s string) error {
		log.Println("received pong:", s)
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
	go func() error {
		for {
			msgType, msg, err := c.ws.ReadMessage()
			if err != nil {
				log.Println("readthread:", err)
			}

			if msgType != websocket.TextMessage {
				return ErrUnexpectedMessageType
			}

			log.Printf("\nReceived: %s.\n", msg)
			c.inQueue <- string(msg)
		}
		return nil
	}()

	go func() {
		for {
			select {
			case event := <-c.inQueue:
				var room string
				messages := strings.Split(event, "\n")
				if string([]rune(messages[0])[0]) == ">" {
					room, messages = messages[0], messages[1:]
				}

				var buffer bytes.Buffer
				for _, rawMessage := range messages {
					buffer.WriteString(room)
					buffer.WriteString("\n")
					buffer.WriteString(rawMessage)
					parse(buffer.String(), c.Bot)
				}
			}
		}
	}()
}

func (c *Connection) startSendingThread() {
	interrupt := make(chan os.Signal, 1)
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
				log.Println("interrupt")
				// Send a close frame and wait for the server to close the connection.
				err := c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Println("write close:", err)
					return
				}
				select {
				case <-done:
				case <-time.After(time.Second):
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
	log.Printf("\nSent message: %s\n", s)
	err := c.ws.WriteMessage(websocket.TextMessage, []byte(s))
	if err != nil {
		log.Println("send:", err)
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
	case "c":
		if msg.Params[0] != "~" {
			events <- "message"
		}
	}

	events <- cmd
	Call(handlers, cmd, msg, events)
}

// TODO Handlers and dispatching them on events
