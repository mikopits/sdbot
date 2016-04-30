SDBot - A Pokemon Showdown Bot Framework
==========================================

Description
-----------

SDBot is a bot framework written in Go for Pokemon Showdown designed to take
advantage of Go's inherent concurrency.

Still in developmental stages.

Installation
------------

You can `go get` it.

```
go get github.com/mikopits/sdbot
```

Or you can clone the latest GitHub repository.

```
git clone http://github.com/mikopits/sdbot
```

Example
-------

To get the bot up and running and with a loaded example plugin:

```go
package main

import (
  "github.com/mikopits/sdbot"
  "github.com/mikopits/sdbot/examples/plugins"
)

func main() {
  b := sdbot.NewBot()
  b.RegisterPlugin(plugins.HelloWorldPlugin(), "hello world")
  b.RegisterPlugin(plugins.EchoPlugin(), "echo")
  b.RegisterTimedPlugin(plugins.CountPlugin(), "count")
  b.Connection.Connect()
}
```

And be sure to set your `config.toml` file in the same directory as
`package main`.
