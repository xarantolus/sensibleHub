package store

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"xarantolus/sensibleHub/store/music"

	"github.com/vitali-fedulov/images4"
)

// ErrAudioSameStartEnd is returned while editing a song if the Start and End properties are
// the same because having a zero-second song doesn't make sense
var ErrAudioSameStartEnd = fmt.Errorf("Audio start/end must not be the same")

// EditEntryData is used for editing an entry.
// Not all fields must be set, most are optional
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

	entryBefore := entry

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

	if entry.AudioSettings.Start == entry.AudioSettings.End {
		return ErrAudioSameStartEnd
	}

	// Sync must be a valid bool
	setValidB(&entry.SyncSettings.Should, data.Sync)

	var editedImage bool
	if data.CoverImage != nil && data.CoverFilename != "" {
		oldCover, oldCoverPath := entry.PictureData.Filename, entry.CoverPath()

		// If we have no extension, it will be converted to a jpeg image
		ext := filepath.Ext(data.CoverFilename)
		if ext == "" {
			ext = ".jpg"
		}
		coverFN := "cover" + strings.ToLower(ext)

		covDest := filepath.Join(entry.DirPath(), coverFN)
		err = cropCover(data.CoverImage, "", covDest)
		if err != nil {
			return
		}
		_ = data.CoverImage.Close()

		if oldCover != coverFN && oldCover != "" {
			err = os.Remove(oldCoverPath)
			if err != nil {
				return
			}
		}

		hex, _ := music.CalculateDominantColor(covDest)
		entry.PictureData.DominantColorHEX = music.Color(hex)

		i, err := images4.Open(covDest)
		if err == nil {
			entry.PictureData.Size = i.Bounds().Dx()
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

// EditAlbumCover edits all songs in the album identified by `artist` and `album` to
// have the cover given by `coverImage`
func (m *Manager) EditAlbumCover(artist, album string, coverName string, coverImage io.ReadCloser) (err error) {
	artist, album = CleanName(artist), CleanName(album)

	ext := filepath.Ext(coverName)
	if ext == "" {
		ext = ".jpg"
	}
	coverFN := "cover" + strings.ToLower(ext)

	tmpDir, err := ioutil.TempDir("", "shub-")
	if err != nil {
		return
	}
	tmpCoverPath := filepath.Join(tmpDir, coverFN)

	err = cropCover(coverImage, "", tmpCoverPath)
	if err != nil {
		return
	}
	_ = coverImage.Close()

	hex, _ := music.CalculateDominantColor(tmpCoverPath)

	var imageSize int
	i, err := images4.Open(tmpCoverPath)
	if err == nil {
		imageSize = i.Bounds().Dx()
	}

	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	for sid, e := range m.Songs {
		// Wrong artist?
		if !strings.EqualFold(CleanName(e.Artist()), artist) {
			continue
		}

		// Wrong album?
		if !strings.EqualFold(CleanName(e.AlbumName()), album) {
			continue
		}

		oldCover, oldCoverPath := e.PictureData.Filename, e.CoverPath()
		newPath := filepath.Join(e.DirPath(), coverFN)

		// Move the new cover to its place
		err = copyOverwrite(tmpCoverPath, newPath)
		if err != nil {
			return
		}

		// and delete the old cover if it wasn't overwritten anyways
		if oldCover != coverFN && oldCover != "" {
			err = os.Remove(oldCoverPath)
			if err != nil {
				return
			}
		}

		// Update song info
		e.PictureData.Filename = coverFN
		e.PictureData.DominantColorHEX = music.Color(hex)
		e.PictureData.Size = imageSize
		e.LastEdit = time.Now()

		m.Songs[sid] = e

		defer func(id string, s music.Entry) {
			m.event("song-edit", map[string]interface{}{
				"id":   id,
				"song": s,
			})
		}(sid, e)
	}

	return m.Save(false)
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

func copyOverwrite(src, dest string) (err error) {
	f, err := os.Open(src)
	if err != nil {
		return
	}
	defer f.Close()

	d, err := os.Create(dest)
	if err != nil {
		return
	}
	defer d.Close()

	_, err = io.Copy(d, f)

	return
}
