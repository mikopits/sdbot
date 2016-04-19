package psbot

import (
	"log"
)

type Bot struct {
	handlers *Handlers
	loggers  *Loggers
	rooms    []Room
}

type ResponseHandler func(target, message string, sender *User)

type Handlers struct {
	Response ResponseHandler
}
