package sdbot

import (
	"testing"
)

// TestRenameDifferentName tests both the functionality of a user joining a
// room, and ensures that, if a user is renamed to a different name, the old
// user is removed from the room and the information is updated for the new
// nick.
func TestRenameDifferentName(t *testing.T) {
	b := initBot()
	joinMsg := ">testroom\n|J|+Tympy"
	b.Connection.parse(joinMsg)

	u := b.UserList["tympy"]
	r := b.RoomList["testroom"]

	// Test if "tympy" joined "testroom"
	if u == nil {
		t.Fatal(`user "tympy" not instantiated`)
	}
	if r == nil {
		t.Fatal(`room "testroom" not instantiated`)
	}
	if u.Name != "Tympy" {
		t.Errorf(`u.Name (%s) should == "Tympy"`, u.Name)
	}
	if u.Auths["testroom"] != Voiced {
		t.Errorf(`u.Auths["testroom"] (%s) should == "+"`, u.Auths["testroom"])
	}
	if len(r.Users) != 1 {
		t.Errorf(`len(r.Users) (%d) should == "1"`, len(r.Users))
	}
	if r.Users[0] != "tympy" {
		t.Errorf(`r.Users[0] (%s) should == "tympy"`, r.Users[0])
	}

	renameMsg := ">testroom\n|N|+Tympani|tympy"
	b.Connection.parse(renameMsg)

	newu := b.UserList["tympani"]
	newr := b.RoomList["testroom"]

	// Test if "tympy" was renamed to "tympani"
	if newu == nil {
		t.Fatal(`user "tympani" not instantiated`)
	}
	if newr == nil {
		t.Fatal(`room "testroom" not instantiated`)
	}
	if newu.Name != "Tympani" {
		t.Errorf(`newu.Name (%s) should == "Tympani"`, newu.Name)
	}
	if newu.Auths["testroom"] != Voiced {
		t.Errorf(`newu.Auths["testroom"] (%s) should == "+"`, newu.Auths["testroom"])
	}
	if len(newr.Users) != 1 {
		t.Errorf(`len(newr.Users) (%d) should == 1`, len(newr.Users))
	}
	if newr.Users[0] != "tympani" {
		t.Errorf(`newr.Users[0] (%s) should == "tympani"`, newr.Users[0])
	}
}

// TestRenameSameNames tests that the bot correctly updates the state of the
// user and room lists when a user changes nicks with a non-canonical change
// (ie. Tympy -> T%%%%%%ympy).
func TestRenameSameNames(t *testing.T) {
	b := initBot()
	joinMsg := ">testroom\n|J|+Tympy"
	b.Connection.parse(joinMsg)

	u := b.UserList["tympy"]
	r := b.RoomList["testroom"]

	// Test if "tympy" joined "testroom"
	if u == nil {
		t.Fatal(`user "tympy" not instantiated`)
	}
	if r == nil {
		t.Fatal(`room "testroom" not instantiated`)
	}
	if u.Name != "Tympy" {
		t.Errorf(`u.Name (%s) should == "Tympy"`, u.Name)
	}
	if u.Auths["testroom"] != Voiced {
		t.Errorf(`u.Auths["testroom"] (%s) should == "+"`, u.Auths["testroom"])
	}
	if len(r.Users) != 1 {
		t.Errorf(`len(r.Users) (%d) should == "1"`, len(r.Users))
	}
	if r.Users[0] != "tympy" {
		t.Errorf(`r.Users[0] (%s) should == "tympy"`, r.Users[0])
	}

	renameMsg := ">testroom\n|N|+T#ympy|tympy"
	b.Connection.parse(renameMsg)

	newu := b.UserList["tympy"]
	newr := b.RoomList["testroom"]

	// Test if "Tympy" was renamed to "T#ympy"
	if newu == nil {
		t.Fatal(`user "tympy" not instantiated`)
	}
	if newr == nil {
		t.Fatal(`room "testroom" not instantiated`)
	}
	if newu.Name != "T#ympy" {
		t.Errorf(`newu.Name (%s) should == "Tympy"`, newu.Name)
	}
	if newu.Auths["testroom"] != Voiced {
		t.Errorf(`newu.Auths["testroom"] (%s) should == "+"`, newu.Auths["testroom"])
	}
	if len(r.Users) != 1 {
		t.Errorf(`len(r.Users) (%d) should == "1"`, len(r.Users))
	}
	if r.Users[0] != "tympy" {
		t.Errorf(`r.Users[0] (%s) should == "tympy"`, r.Users[0])
	}
}

func initBot() *Bot {
	return NewBot("examples/config/config_example.toml")
}
