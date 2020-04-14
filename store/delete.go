package store

import (
	"fmt"
	"os"
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
	if err != nil {
		return
	}

	return m.Save(false)
}
