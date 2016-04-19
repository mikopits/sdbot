package sdbot

import (
	"bytes"
	"log"
	"os"
	"regexp"

	"github.com/BurntSushi/toml"
)

// Information from config file
type Config struct {
	Server            string
	Port              string
	Nick              string
	Password          string
	MessagesPerSecond float64
	Rooms             []string
	Avatar            string
	PluginPrefixes    []string
	PluginPrefix      *regexp.Regexp
}

// Read the config data from toml
func ReadConfig() *Config {
	configfile := "config.toml"
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	config.generatePluginPrefixRegexp()
	log.Print(config)
	return &config
}

func (c *Config) generatePluginPrefixRegexp() {
	var buffer bytes.Buffer

	for _, prefix := range c.PluginPrefixes {
		buffer.WriteString(prefix)
	}

	c.PluginPrefix = regexp.MustCompile("^" + regexp.QuoteMeta(buffer.String()))
}
