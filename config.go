package sdbot

import (
	"os"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config holds the configuration information read from the config.toml file.
type Config struct {
	Server                string
	Port                  string
	Nick                  string
	Password              string
	MessagesPerSecond     float64
	Rooms                 []string
	Avatar                int
	PluginPrefixes        []string
	PluginSuffixes        []string
	PluginPrefix          *regexp.Regexp
	PluginSuffix          *regexp.Regexp
	CaseInsensitive       bool
	IgnorePrivateMessages bool
	IgnoreChatMessages    bool
}

// Reads the config data from toml config file.
func readConfig(path string) *Config {
	_, err := os.Stat(path)
	if err != nil {
		Fatalf("Config file is missing: %s", path)
	}

	var config Config
	_, err = toml.DecodeFile(path, &config)
	CheckErr(err)

	config.generatePluginPrefixRegexp()
	config.generatePluginSuffixRegexp()

	if config.MessagesPerSecond == 0 {
		config.MessagesPerSecond = 3
	}

	return &config
}

func (c *Config) generatePluginPrefixRegexp() {
	var prefixes []string
	for _, prefix := range c.PluginPrefixes {
		prefixes = append(prefixes, regexp.QuoteMeta(prefix))
	}
	regStr := "^(" + strings.Join(prefixes, "|") + ")"
	reg, err := regexp.Compile(regStr)
	CheckErr(err)

	c.PluginPrefix = reg
}

func (c *Config) generatePluginSuffixRegexp() {
	var suffixes []string
	for _, suffix := range c.PluginSuffixes {
		suffixes = append(suffixes, regexp.QuoteMeta(suffix))
	}
	regStr := "(" + strings.Join(suffixes, "|") + ")$"
	reg, err := regexp.Compile(regStr)
	CheckErr(err)

	c.PluginSuffix = reg
}
