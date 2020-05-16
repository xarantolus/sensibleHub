package store

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// CleanUp removes all unused directories in the data directory. They might not have been delted due to errors.
func (m *Manager) CleanUp() (n int) {
	songsList, err := ioutil.ReadDir(filepath.Dir(songDirTemplate))
	if err != nil {
		return
	}

	for _, song := range songsList {
		if !song.IsDir() {
			continue
		}

		dir := filepath.Base(song.Name())
		if _, ok := m.Songs[dir]; !ok {
			err := os.RemoveAll(fmt.Sprintf(songDirTemplate, dir))
			if err != nil {
				log.Printf("[Cleanup]: Error while removing %s: %s\n", song.Name(), err.Error())
			} else {
				n++
			}
		}
	}

	return
}
