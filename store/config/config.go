package config

import (
	"log"
	"os"
	"strings"

	"github.com/muhammadmuzzammil1998/jsonc"
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
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	err = jsonc.Unmarshal(content, &c)
	if err != nil {
		return
	}

	rid, ok := os.LookupEnv("RUNNING_IN_DOCKER")
	if ok && strings.ToLower(rid) == "true" {
		c.Alternatives.FFmpeg = "ffmpeg"
		c.Alternatives.FFprobe = "ffprobe"
		c.Alternatives.YoutubeDL = "yt-dlp"
		log.Println("[Info] Running in Docker, using local binaries. This means that the \"alternatives\" config part is ignored!")
	} else {
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
	}

	return
}
