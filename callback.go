package sdbot

// Used to encaspulate handlers and prevent them from overwriting instance
// variables in a Bot due to unsafe thread access.
type Callback struct {
	Bot *Bot
}

// See Bot.Synchronize
func (c *Callback) Synchronize(name string, lambda *func() interface{}) interface{} {
	return c.Bot.Synchronize(name, lambda)
}
