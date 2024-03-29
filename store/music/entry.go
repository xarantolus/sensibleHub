package music

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
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

func (e *Entry) IsImported() bool {
	return e.SourceURL == "Import"
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

func (e *Entry) Filename(extension string) string {
	return strings.TrimSuffix(e.SongName(), ".") + "." + strings.TrimPrefix(extension, ".")
}

func (e *Entry) SongName() (out string) {
	if e.MusicData.Artist != "" {
		out = e.MusicData.Artist + " - "
	}

	out += e.MusicData.Title
	if out == "" {
		return "Unknown"
	}

	return out
}

func (e *Entry) AlbumName() string {
	if e.MusicData.Album == "" {
		return "Unknown"
	}

	return e.MusicData.Album
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

// TimeRange returns the playback range for HTML media elements
// See https://developer.mozilla.org/en-US/docs/Web/Guide/Audio_and_video_delivery#specifying_playback_range
func (e *Entry) PlaybackRange() string {
	// if end time == duration
	if e.AudioSettings.End <= 0 || math.Abs(e.AudioSettings.End-e.MusicData.Duration) < 0.001 {
		// if we also have no start time, then neither was set -- return nothing
		if e.AudioSettings.Start <= 0.001 {
			return ""
		}

		// Only the start time is set
		return fmt.Sprintf("#t=%f", e.AudioSettings.Start)
	}

	if e.AudioSettings.Start <= 0.001 {
		// Only the end time is set
		return fmt.Sprintf("#t=0,%f", e.AudioSettings.End)
	}

	// both are set
	return fmt.Sprintf("#t=%f,%f", e.AudioSettings.Start, e.AudioSettings.End)
}

type MusicData struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Year   *int   `json:"year,omitempty"` // optional, may be nil

	// Duration is the duration of the original file in seconds
	Duration float64 `json:"duration"`
}

type PictureData struct {
	// Filename is the name of the original cover for this song
	Filename string `json:"filename"`

	// DominantColorHEX is the dominant color hex without #-prefix
	DominantColorHEX Color `json:"dominant_color"`

	// Size is the width and height of the cover. We use one attribute only because both must be the same
	Size int `json:"size"`
}
