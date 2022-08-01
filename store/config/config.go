package config

import (
	"encoding/json"
	"os"
	"strings"
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

	Alternatives struct {
		FFmpeg    string `json:"ffmpeg"`
		FFprobe   string `json:"ffprobe"`
		YoutubeDL string `json:"youtube-dl"`
	} `json:"alternatives"`

	GenerateOnStartup bool `json:"generate_on_startup"`
}

const (
	DefaultConfigFile = "config.json"
)

func Parse(path string) (c Config, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&c)
	if err != nil {
		return
	}

	// Set default paths/names if none are set
	c.Alternatives.FFmpeg = strings.TrimSpace(c.Alternatives.FFmpeg)
	if c.Alternatives.FFmpeg == "" {
		c.Alternatives.FFmpeg = "ffmpeg"
	}
	c.Alternatives.FFprobe = strings.TrimSpace(c.Alternatives.FFprobe)
	if c.Alternatives.FFprobe == "" {
		c.Alternatives.FFprobe = "ffprobe"
	}
	c.Alternatives.YoutubeDL = strings.TrimSpace(c.Alternatives.YoutubeDL)
	if c.Alternatives.YoutubeDL == "" {
		c.Alternatives.YoutubeDL = "youtube-dl"
	}

	return
}
