package battle

import (
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
	Name         string
	Format       string
	Id           string
	Logic        *BattleLogic
	Type         int
	Gen          int
	FirstPlayer  *Player
	SecondPlayer *Player
	Turn         int
	Rqid         int
	Room         *sdbot.Room
}

// NewBattle creates a new Battle with a given name. The name must be of the
// format "battle-[battleformat]-[battleid]" as assigned by the server. The
// gen must be the generation of the battle, an int in [1, 6].
func NewBattle(name string, gen int) *Battle {
	b := &Battle{Name: name}
	fmtReg := regexp.MustCompile("battle-(.+)-([0-9]+)")
	sm := fmtReg.FindStringSubmatch(name)
	b.Format = sm[1]
	b.Id = sm[2]
	b.Room = &sdbot.Room{Name: name}
	b.Gen = gen
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

// ChooseMove

// TODO placeholder
type BattleLogic struct{}
