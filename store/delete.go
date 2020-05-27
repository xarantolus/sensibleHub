package store

import (
	"fmt"
	"os"
	"time"
	"xarantolus/sensibleHub/store/music"
)

// DeleteEntry deletes the entry with the given ID and removes all files associated with it
func (m *Manager) DeleteEntry(id string) (err error) {
	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	entry, ok := m.Songs[id]
	if !ok {
		return fmt.Errorf("Cannot edit entry with id %s as it doesn't exist", id)
	}

	delete(m.Songs, id)

	err = os.RemoveAll(entry.DirPath())
	if err != nil && !os.IsNotExist(err) {
		return
	}

	err = m.Save(false)
	if err != nil {
		return
	}

	m.event("song-delete", map[string]interface{}{
		"id": id,
	})

	return nil
}

// DeleteCoverImage deletes the cover image for the given song
func (m *Manager) DeleteCoverImage(id string) (err error) {
	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	entry, ok := m.Songs[id]
	if !ok {
		return fmt.Errorf("Cannot edit entry with id %s as it doesn't exist", id)
	}

	// If we don't have a cover image, we cannot remove one
	if entry.PictureData.Filename == "" {
		return nil
	}

	err = os.Remove(entry.CoverPath())
	if err != nil {
		return
	}

	// Clear PictureData
	entry.PictureData = music.PictureData{}
	entry.LastEdit = time.Now()

	// Save this entry
	m.Songs[id] = entry

	err = m.Save(false)
	if err != nil {
		return
	}

	m.event("song-edit", map[string]interface{}{
		"id":   id,
		"song": entry,
	})

	return
}
