package store

import (
	"os"
	"xarantolus/sensibleHub/store/music"
)

// DeleteEntry deletes the entry with the given ID and removes all files associated with it
func (m *Manager) DeleteEntry(id string) (err error) {
	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	entry, ok := m.Songs[id]
	if !ok {
		return songNotFound(id)
	}

	delete(m.Songs, id)

	if entry.ID != "" {
		err = os.RemoveAll(entry.DirPath())
		if err != nil && !os.IsNotExist(err) {
			return
		}
	}

	err = m.Save(false)
	if err != nil {
		return
	}

	m.publish(SongDeleted{ID: id})

	return nil
}

// DeleteCoverImage deletes the cover image for the given song
func (m *Manager) DeleteCoverImage(id string) (err error) {
	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	entry, ok := m.Songs[id]
	if !ok {
		return songNotFound(id)
	}

	// If we don't have a cover image, we cannot remove one
	if entry.PictureData.Filename == "" {
		return nil
	}

	err = os.Remove(entry.CoverPath())
	if err != nil && !os.IsNotExist(err) {
		return
	}

	entry.PictureData = music.PictureData{}

	_, err = m.commit(entry)
	return err
}
