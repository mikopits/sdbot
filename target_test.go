package sdbot

import (
	"testing"
)

// TestRename performs two independent tests on the Rename function.
//
// First it tries to rename to a name considered to be a different user.
// Second, it tries to rename to a name that is considered to be canonically
// the same by the server (same alphanumeric characters).
func TestRename(t *testing.T) {
	testBot := initBot()
	testUser := FindUserEnsured("tympy", testBot)
	testBot.UserList[testUser.Name].AddAuth("test", Voiced)
	testRoom := FindRoomEnsured("test", testBot)
	testRoom.AddUser("tympy")
	Rename("tympy", "tympani", testRoom, testBot, Voiced)

	if testBot.UserList["tympani"] == nil {
		t.Fatal(`user "tympani" was not created or stored`)
	}
	if len(testRoom.Users) != 1 {
		t.Fatalf(`number of users in room (%d) should == 1`, len(testRoom.Users))
	}
	if testRoom.Users[0] != "tympani" {
		t.Fatalf(`user in room (%s) should == tympani`, testRoom.Users[0])
	}
	if FindUserEnsured(testRoom.Users[0], testBot).Auths["test"] != Voiced {
		t.Fatalf(`user "tympani" has auth (%s) should == %s`,
			FindUserEnsured(testRoom.Users[0], testBot).Auths["test"], Voiced)
	}

	FindUserEnsured("randomuser1", testBot)
	FindUserEnsured("randomuser2", testBot)
	testRoom.AddUser("randomuser1")
	testRoom.AddUser("randomuser2")
	Rename("tympani", "tympy", testRoom, testBot, Driver)
	var testRenamedReauthedUser *User

	if len(testRoom.Users) != 3 {
		t.Fatalf(`number of users in room (%d) should == 3`, len(testRoom.Users))
	}
	var oldFound bool
	var newFound bool
	for _, u := range testRoom.Users {
		if u == "tympy" {
			newFound = true
			testRenamedReauthedUser = FindUserEnsured(u, testBot)
		} else if u == "tympani" {
			oldFound = true
		}
	}
	if oldFound {
		t.Fatalf(`bool "oldFound" (%t) should == false`, oldFound)
	}
	if !newFound {
		t.Fatalf(`bool "newFound" (%t) should == true`, newFound)
	}
	if testRenamedReauthedUser.Auths["test"] != Driver {
		t.Fatalf(`renamed user "%s" has auth (%s) should == %s`,
			testRenamedReauthedUser.Name, testRenamedReauthedUser.Auths["test"], Driver)
	}

	FindUserEnsured("tympy", testBot)
	testName := "T\\%\\%\\%\\%\\%ympy"
	Rename("tympy", testName, testRoom, testBot, Voiced)
	testUser = FindUserEnsured("tympy", testBot)

	if len(testRoom.Users) != 3 {
		t.Fatalf(`number of users in room (%d) should == 3`, len(testRoom.Users))
	}
	if testUser.Name != testName {
		t.Fatalf(`username (%s) should == (%s)`, testUser.Name, testName)
	}
}

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

func initBot() *Bot {
	return NewBot("examples/config/config_example.toml")
}
