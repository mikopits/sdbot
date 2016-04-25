package sdbot

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server            string
	Port              string
	Nick              string
	Password          string
	MessagesPerSecond float64
	Rooms             []string
	Avatar            int
	PluginPrefixes    []string
	PluginSuffixes    []string
	PluginPrefix      *regexp.Regexp
	PluginSuffix      *regexp.Regexp
	CaseInsensitive   bool
}

// Reads the config data from toml config file.
func ReadConfig() *Config {
	configfile := "config.toml"
	_, err := os.Stat(configfile)
	if err != nil {
		Fatal(&Log, fmt.Sprintf("Config file is missing: %s", configfile))
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		Error(&Log, err)
	}

	config.generatePluginPrefixRegexp()
	config.generatePluginSuffixRegexp()

	return &config
}

func (c *Config) generatePluginPrefixRegexp() {
	regStr := "^(" + strings.Join(c.PluginPrefixes, "|") + ")"
	reg, err := regexp.Compile(regStr)
	if err != nil {
		Error(&Log, err)
	}

	c.PluginPrefix = reg
}

func (c *Config) generatePluginSuffixRegexp() {
	regStr := "(" + strings.Join(c.PluginSuffixes, "|") + ")$"
	reg, err := regexp.Compile(regStr)
	if err != nil {
		Error(&Log, err)
	}

	c.PluginSuffix = reg
}
