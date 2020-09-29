package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Port int `json:"port"`

	FTP struct {
		Port int `json:"port"`

		Users []struct {
			Name   string `json:"name"`
			Passwd string `json:"passwd"`
		} `json:"users"`
	} `json:"ftp"`

	KeepGeneratedDays int `json:"keep_generated_days"`

	Cover struct {
		MaxSize int `json:"max_size"`
	} `json:"cover"`

	AllowExternal struct {
		Apple bool `json:"apple"`
	} `json:"allow_external"`
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
	if err != nil {
		return
	}

	return
}
