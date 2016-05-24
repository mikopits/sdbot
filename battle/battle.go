package battle

import (
	"fmt"
	"regexp"
	"sdbot"
)

// Constants that represent the enumeration of the types of battles.
const (
	Singles = iota
	Doubles
	Triples
)

// Battle represents a pokemon battle.
type Battle struct {
	Bot    *sdbot.Bot
	Name   string
	Format string
	Id     string
	Logic  BattleLogic
	Type   int
	Gen    int
	Me     *Player
	Opp    *Player
	Turn   int
	Rqid   int
	Room   *sdbot.Room
}

// NewBattle creates a new Battle with a given name. The name must be of the
// format "battle-[battleformat]-[battleid]" as assigned by the server. The
// gen must be the generation of the battle, an int in [1, 6].
func NewBattle(name string, gen int, bot *sdbot.Bot) *Battle {
	b := &Battle{Name: name, Gen: gen, Bot: bot}
	fmtReg := regexp.MustCompile("battle-(.+)-([0-9]+)")
	sm := fmtReg.FindStringSubmatch(name)
	b.Format = sm[1]
	b.Id = sm[2]
	b.Room = &sdbot.Room{Name: name}
	return b
}

// Reply calls sdbot.Room.Reply
func (b *Battle) Reply(m *sdbot.Message, s string) {
	b.Room.Reply(m, s)
}

// RawReply calls sdbot.Room.RawReply
func (b *Battle) RawReply(m *sdbot.Message, s string) {
	b.Room.RawReply(m, s)
}

// ChooseMove picks a move according to the battle logic and sends it to the
// server.
func (b *Battle) ChooseMove() {
	send(b, fmt.Sprintf("/choose move %s|%d", b.Logic.ChooseMove, b.Rqid))
}

// ChooseLead picks a lead pokemon according to the battle logic and sends
// it to the server.
func (b *Battle) ChooseLead() {
	send(b, fmt.Sprintf("/team %s|%d", b.Logic.ChooseLead, b.Rqid))
}

// ToggleTimer toggles the battle timer.
func (b *Battle) ToggleTimer() {
	send(b, "/timer")
}

func send(b *Battle, s string) {
	b.Bot.Send(fmt.Sprintf("%s|%s", b.Name, s))
}
