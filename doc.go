// Copyright 2016 Eric Furugori.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

// Package sdbot implements the Pokemon Showdown protocol and provides a
// library to interact with users.
//
// Usage
//
// The Bot type represents the bot and its state. To create a Bot, call the
// NewBot function. This currently assumes that your configuration toml file
// is both called config.toml and is located in the same directory as your
// bot script. Expect this to change in later versions.
//
// To connect to the server, call the bot's Connect method.
//
// To register plugins, write your plugins under a package and import them.
// Register them by calling the bot's RegisterPlugin and RegisterTimedPlugin
// methods. It is recommended to register your plugins before connecting to
// the server.
//
// Concurrency
//
// Each bot will spawn multiple goroutines for both reading and writing to the
// socket, as well as for plugin loops listening for matches from the server
// messages. The bot internal state is designed for concurrent access, so two
// goroutines looking to read the bot's state, such as the userlist, will get
// the same result.
//
// A bot will spawn a separate goroutine to run every Plugin event. For this
// reason, applications are responsible for ensuring that the plugin
// EventHandlers are safe for concurrent use. For this reason, Bot exports a
// Synchronize method that takes a pointer to a function and an identifying
// string that will run all functions corresponding to that identifying string
// on the same mutex.
//
// Consider you want a plugin that reads and writes to a map defined in its
// event handler. Maps cannot be read and written to concurrently, so you need
// to ensure that another plugin event must wait until the current one is done.
//
// func (eh *PluginEventHandler) HandleEvent(m *sdbot.Message, args []string) {
//     var readVal interface{}
//     rk := someKey()
//     wk := anotherKey()
//     rw := func() interface{} {
//         eh.Map[wk] = someVal()
//         readVal = eh.Map[rk]
//         return nil
//     }
//
//     m.Bot.Synchronize("maprw", &rw)
// }
package sdbot
