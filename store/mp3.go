package store

import (
	"log"
	"os"
	"time"
	"xarantolus/sensibleHub/store/music"
)

// Saving after every generated MP3 would rewrite the data file once per song
// while a client syncs.
const mp3SizeSaveInterval = 10 * time.Second

// MP3Path returns the generated MP3 file for e (see music.Entry.MP3Path) and
// remembers its size for FTP listings.
func (m *Manager) MP3Path(e music.Entry) (string, error) {
	p, err := e.MP3Path(m.GetConfig())
	if err != nil {
		return "", err
	}

	fi, err := os.Stat(p)
	if err != nil {
		return "", err
	}
	m.recordMP3Size(e.ID, e.LastEdit, fi.Size())

	return p, nil
}

func (m *Manager) recordMP3Size(id string, lastEdit time.Time, size int64) {
	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	entry, ok := m.Songs[id]
	if !ok || !entry.LastEdit.Equal(lastEdit) {
		return
	}
	if known, ok := entry.MP3Size(); ok && known == size {
		return
	}

	entry.MP3 = &music.GeneratedMP3{Size: size, LastEdit: lastEdit, Version: music.MP3Version}
	m.Songs[id] = entry

	m.mp3SizesUnsaved = true
	if time.Since(m.mp3SizesSaved) >= mp3SizeSaveInterval {
		m.saveMP3Sizes()
	}
}

// SaveMP3Sizes saves sizes recorded since the last save. Losing them only
// means clients download those songs once more.
func (m *Manager) SaveMP3Sizes() {
	m.SongsLock.Lock()
	defer m.SongsLock.Unlock()

	if m.mp3SizesUnsaved {
		m.saveMP3Sizes()
	}
}

// saveMP3Sizes saves the data file. m.SongsLock must be held.
func (m *Manager) saveMP3Sizes() {
	if err := m.Save(false); err != nil {
		log.Printf("[MP3] Saving generated sizes: %v\n", err)
		return
	}
	m.mp3SizesUnsaved = false
	m.mp3SizesSaved = time.Now()
}

// removeGeneratedMP3 deletes the MP3 generated for a song that was just edited.
// It is outdated anyway, and the cleanup marks files as used by their mtime,
// which must never make a file from before the edit look current.
func removeGeneratedMP3(id string) {
	err := os.Remove(music.MP3FilePath(id))
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Cannot remove outdated MP3 of %s: %v\n", id, err)
	}
}
