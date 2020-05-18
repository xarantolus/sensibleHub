package store

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type EditEntryData struct {
	CoverImage    io.ReadCloser
	CoverFilename string

	Title  string
	Artist string
	Album  string
	Year   string

	Start string
	End   string

	Sync string
}

// EditEntry edits the entry with the given `id`.
// All fields in `data` may be empty, only those that have values will be updated.
// As a special case, CoverImage will only be read if CoverFilename is also set.
// Also, `data.Title` will not lead to any change if empty
func (m *Manager) EditEntry(id string, data EditEntryData) (err error) {
	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	entry, ok := m.Songs[id]
	if !ok {

		return fmt.Errorf("Cannot edit entry with id %s as it doesn't exist", id)
	}

	var entryBefore = entry

	// Title must not be empty
	setValidS(&entry.MusicData.Title, data.Title)

	// These fields may be empty
	entry.MusicData.Artist = strings.TrimSpace(data.Artist)
	entry.MusicData.Album = strings.TrimSpace(data.Album)

	year, err := strconv.Atoi(data.Year)
	if err == nil {
		entry.MusicData.Year = &year
	} else {
		// allow clearing year value after it has been set
		entry.MusicData.Year = nil
	}

	// these floats must have a valid value that is between 0 and the length of the audio
	setValidF(&entry.AudioSettings.Start, data.Start, 0, entry.MusicData.Duration)
	setValidF(&entry.AudioSettings.End, data.End, 0, entry.MusicData.Duration)

	// Swap if end is before start
	if entry.AudioSettings.Start > entry.AudioSettings.End {
		entry.AudioSettings.Start, entry.AudioSettings.End = entry.AudioSettings.End, entry.AudioSettings.Start
	}

	// Sync must be a valid bool
	setValidB(&entry.SyncSettings.Should, data.Sync)

	var editedImage bool
	if data.CoverImage != nil && data.CoverFilename != "" {
		var oldCover, oldCoverPath = entry.PictureData.Filename, entry.CoverPath()

		// If we have no extension, it will be converted to a jpeg image
		ext := filepath.Ext(data.CoverFilename)
		if ext == "" {
			ext = ".jpg"
		}
		coverFN := "cover" + strings.ToLower(ext)

		err = CropCover(data.CoverImage, "", filepath.Join(entry.DirPath(), coverFN))
		if err != nil {
			return
		}

		if oldCover != coverFN && oldCover != "" {
			err = os.Remove(oldCoverPath)
			if err != nil {
				return
			}
		}

		entry.PictureData.Filename = coverFN
		editedImage = true
	}

	// If the image filename is the same, it will not be recognized. this is why we need the second check
	if entry == entryBefore && !editedImage {
		return nil // No edits have been made
	}

	entry.LastEdit = time.Now()

	m.Songs[id] = entry

	err = m.Save(false)
	if err != nil {
		return
	}

	m.event("song-edit", map[string]interface{}{
		"id":   id,
		"song": entry,
	})

	return nil
}

func setValidS(target *string, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		*target = value
	}
}

func setValidF(target *float64, value string, min, max float64) {
	f, err := strconv.ParseFloat(value, 64)
	if err == nil && f >= min && f <= max {
		*target = f
	}
}

func setValidB(target *bool, value string) {
	value = strings.ToUpper(value)

	if value == "ON" {
		*target = true
	} else {
		// HTML checkboxes are either "on" or "", which is quite a bad design in my book
		*target = false
	}
}
