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
	testBot := NewBot("examples/config/config_example.toml")
	FindUserEnsured("tympy", testBot)
	testRoom := FindRoomEnsured("test", testBot)
	testRoom.AddUser("tympy")
	Rename("tympy", "tympani", testRoom, testBot)

	if testBot.UserList["tympani"] == nil {
		t.Fatal(`user "tympani" was not created or stored`)
	}
	if len(testRoom.Users) != 1 {
		t.Fatalf(`number of users in room (%d) should == 1`, len(testRoom.Users))
	}
	if testRoom.Users[0] != "tympani" {
		t.Fatalf(`user in room (%s) should == tympani`, testRoom.Users[0])
	}

	FindUserEnsured("randomuser1", testBot)
	FindUserEnsured("randomuser2", testBot)
	testRoom.AddUser("randomuser1")
	testRoom.AddUser("randomuser2")
	Rename("tympani", "tympy", testRoom, testBot)

	if len(testRoom.Users) != 3 {
		t.Fatalf(`number of users in room (%d) should == 3`, len(testRoom.Users))
	}
	var oldFound bool
	var newFound bool
	for _, u := range testRoom.Users {
		if u == "tympy" {
			newFound = true
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

	FindUserEnsured("tympy", testBot).AddAuth("test", "+")
	testName := "T\\%\\%\\%\\%\\%ympy"
	Rename("tympy", testName, testRoom, testBot)
	testUser := FindUserEnsured("tympy", testBot)

	if len(testRoom.Users) != 3 {
		t.Fatalf(`number of users in room (%d) should == 3`, len(testRoom.Users))
	}
	if testUser.Name != testName {
		t.Fatalf(`username (%s) should == (%s)`, testUser.Name, testName)
	}
}
