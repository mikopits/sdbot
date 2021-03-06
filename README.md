SDBot - A Pokemon Showdown Bot Framework
==========================================

[![GoDoc](https://godoc.org/github.com/mikopits/sdbot?status.svg)](https://godoc.org/github.com/mikopits/sdbot)
[![Go Report Card](https://goreportcard.com/badge/github.com/mikopits/sdbot)](https://goreportcard.com/report/github.com/mikopits/sdbot)

Description
-----------

SDBot is a bot framework written in [Go](https://golang.org/) for [Pokemon Showdown](https://pokemonshowdown.com/) designed to take
advantage of Go's inherent concurrency.

Still in developmental stages. This means that the API is still likely to change at any time.

SDBot has the following [dependencies](https://godoc.org/github.com/mikopits/sdbot?imports).

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
  b := sdbot.NewBot("path/to/your/config.toml")
  b.RegisterPlugin(plugins.HelloWorldPlugin(), "hello world")
  b.RegisterPlugin(plugins.EchoPlugin(), "echo")
  b.RegisterTimedPlugin(plugins.CountPlugin(), "count")
  b.Connect()
}
```

And be sure to set your `config.toml` file in the same directory (See the [example](https://github.com/mikopits/sdbot/blob/master/examples/config/config_example.toml)).

Contribute
----------

If you've gotten panics or have any other problems, go ahead and open an issue.

Have an awesome feature you want to add? Looking at the source code and know
you can solve a problem much more elegantly? Fork the project, develop the
feature on its own branch, and submit a pull request. Or you can open an issue,
tell me how dumb I am, and I can do it myself, I suppose.

TODOs
-----

* Tests
* Battle logic
* More example plugins that sweep more features
