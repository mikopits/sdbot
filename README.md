SDBot - A Pokemon Showdown Bot Framework
==========================================

Description
-----------

SDBot is a bot framework written in Go for Pokemon Showdown designed to take
advantage of Go's inherent concurrency.

Still in developmental stages. This means that the API is still likely to
change at any time.

[Package API Documentation](https://godoc.org/github.com/mikopits/sdbot)
[Imported Packages and Dependencies](https://godoc.org/github.com/mikopits/sdbot?imports)

Installation
------------

You can `go get` it.

```
go get github.com/mikopits/sdbot
```

To install/update the package dependencies:

```
go get -u -v github.com/mikopits/sdbot
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
  b.Connect()
}
```

And be sure to set your `config.toml` file in the same directory as
`package main`.
