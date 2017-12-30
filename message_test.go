package sdbot

import (
	"testing"
)

// TestParseChatMessage tests that the parseMessage function can correctly
// parse the details of an incoming chat message and properly unload them
// into a Message struct.
func TestParseChatMessage(t *testing.T) {
	b := NewBot("examples/config/config_example.toml")
	chatMsg := ">testroom\n|c:|100|+Mystifi|ayylmao"
	m := NewMessage(chatMsg, b)

	// Ensure that the message was instantiated
	if m == nil {
		t.Fatal(`NewMessage (m) not instantiated`)
	}

	// Ensure that the command is "c:" (chat message with timestamp)
	if m.Command != "c:" {
		t.Errorf(`m.Command (%s) should == "c:"`, m.Command)
	}

	// Ensure that the parameters are parsed correctly
	if len(m.Params) != 3 {
		t.Errorf(`len(m.Params) (%d) should == "3"`, len(m.Params))
	}
	if m.Params[0] != "100" {
		t.Errorf(`m.Params[0] (%s) should == "100"`, m.Params[0])
	}
	if m.Params[1] != "+Mystifi" {
		t.Errorf(`m.Params[1] (%s) should == "+Mystifi"`, m.Params[1])
	}
	if m.Params[2] != "ayylmao" {
		t.Errorf(`m.Params[2] (%s) should == "ayylmao"`, m.Params[2])
	}

	// Ensure that the timestamp is parsed correctly and set to zero
	if m.Timestamp != 100 {
		t.Errorf(`m.Timestamp (%d) should == "100"`, m.Timestamp)
	}

	// Ensure that the room and user were parsed correctly
	if m.Room.Name != "testroom" {
		t.Errorf(`m.Room.Name (%s) should == "testroom"`, m.Room.Name)
	}
	if m.User.Name != "Mystifi" {
		t.Errorf(`m.User.Name (%s) should == "Mystifi"`, m.User.Name)
	}

	// Ensure that the auth is Voiced ("+")
	if m.Auth != Voiced {
		t.Errorf(`m.Auth (%s) should == "+"`, m.Auth)
	}

	// Ensure that the message is parsed properly, and that it matches the final
	// parameter in m.Params
	if m.Message != "ayylmao" {
		t.Errorf(`m.Message (%s) should == "ayylmao"`, m.Message)
	}
	if m.Message != m.Params[2] {
		t.Errorf(`m.Message (%s) should == m.Params[2] (%s)`, m.Message, m.Params[2])
	}
}
