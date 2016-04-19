package bot

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

// Information from config file
type Config struct {
	Server            string
	Port              string
	Nick              string
	Password          string
	MessagesPerSecond int
	Rooms             []string
	Avatar            int
}

// Read the config data
func ReadConfig() Config {
	configfile := "config.toml"
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	log.Print(config)
	return config
}
