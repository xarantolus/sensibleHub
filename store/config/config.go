package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port int

	FTP struct {
		Port int

		User   string
		Passwd string
	}
}

const (
	configFile = "config.json"
)

func Parse() (c Config, err error) {
	f, err := os.Open(configFile)
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&c)

	return
}
