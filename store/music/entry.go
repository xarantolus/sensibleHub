package music

import (
	"fmt"
	"path/filepath"
	"time"
)

// Entry represents the entry for an audio file
type Entry struct {
	// ID is the randomly-generated, unique ID for this file
	ID string `json:"id"`

	LastEdit time.Time `json:"last_edit"`
	Added    time.Time `json:"added"`

	// SourceURL is either the source input url where the audio file was downloaded from or
	// the `webpage_url` field from the youtube-dl info file. Both point to the website where one might listen to the song
	SourceURL string `json:"source_url"`

	// SyncSettings defines if this Entrys' MP3 file should be synced
	SyncSettings SyncSettings `json:"sync_settings"`

	// MetaFile is the `.info.json` metadata file from youtube-dl
	MetaFile MetaFile `json:"meta_file"`

	// FileData describes data about the original file that was downloaded
	FileData `json:"file_data"`

	// AudioSettings stores settings such as audio start and end time
	AudioSettings AudioSettings `json:"audio_settings"`

	// MusicData describes music metadata that is typically embedded into music files
	MusicData MusicData `json:"music_data"`

	// PictureData describes the picture file (cover) that should be embedded into the file
	PictureData PictureData `json:"picture_data"`
}

// CoverPath returns the path of the cover
func (e *Entry) CoverPath() string {
	if e.PictureData.Filename == "" {
		return ""
	}

	return filepath.Join("data", "songs", e.ID, e.PictureData.Filename)
}

func (e *Entry) AudioPath() string {
	return filepath.Join("data", "songs", e.ID, e.FileData.Filename)
}

func (e *Entry) DirPath() string {
	return filepath.Join("data", "songs", e.ID)
}

func (e *Entry) SongName() (out string) {
	if e.MusicData.Artist != "" {
		out = e.MusicData.Artist + " - "
	}

	return out + e.MusicData.Title
}

func (e Entry) FormatDuration() string {
	dur := time.Duration(e.MusicData.Duration) * time.Second

	hours := dur / time.Hour
	mins := (dur - hours*time.Hour) / time.Minute
	secs := (dur - mins*time.Minute) / time.Second
	msecs := (dur - secs*time.Second) / time.Second

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d.%02d", hours, mins, secs, msecs)
	}

	return fmt.Sprintf("%02d:%02d.%03d", mins, secs, msecs)
}

// Artist returns the artists' name or a fall back
func (e *Entry) Artist() string {
	if e.MusicData.Artist != "" {
		return e.MusicData.Artist
	}
	return "Unknown"
}

type SyncSettings struct {
	Should bool `json:"should"`
}

type MetaFile struct {
	Filename string `json:"filename"`
}

type FileData struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

type AudioSettings struct {
	// Start is the start time of the song in the given audio. If not set, it is < 0
	Start float64 `json:"start"`
	// Start is the end time of the song in the given audio. If not set, it is < 0
	End float64 `json:"end"`
}

type MusicData struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Year   int    `json:"year"`

	// Duration is the duration of the original file in seconds
	Duration float64 `json:"duration"`
}

type PictureData struct {
	// Filename is the name of the original cover for this song
	Filename string `json:"filename"`
}
