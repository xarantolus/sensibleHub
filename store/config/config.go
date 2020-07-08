package config

import (
	"encoding/json"
	"log"
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

	AllowExternal struct {
		Apple bool `json:"apple"`
	} `json:"allow_external"`

	MP3Settings struct {
		// must be 3 or 4 => either ID3v2.3 or ID3v2.4. (default is 4)
		TagVersion byte `json:"tag_version"`
		JPEGOnly   bool `json:"jpeg_only"`
	} `json:"mp3_settings"`
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

	// Valid versions - unset, 3 and 4
	if c.MP3Settings.TagVersion != 0 && c.MP3Settings.TagVersion != 3 && c.MP3Settings.TagVersion != 4 {
		log.Printf("[Warn] Config: invalid value %v for MP3Settings.TagVersion (must be 3 or 4), using ID3v2.4 as fallback\n", c.MP3Settings.TagVersion)
	}

	return
}
