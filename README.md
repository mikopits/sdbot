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

To get the bare bones bot up and running:

```go
package main

import "github.com/mikopits/sdbot"

func main() {
  bot := sdbot.NewBot()
  bot.Connection.Connect()
}
```

And be sure to set your config.toml file.
