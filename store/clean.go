package store

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
	"xarantolus/sensibleHub/store/config"
)

// CleanUp removes all unused directories in the data directory. They might not have been deleted due to errors.
func (m *Manager) CleanUp(cfg config.Config) (n int) {
	songsList, err := ioutil.ReadDir(filepath.Dir(songDirTemplate))
	if err != nil {
		return
	}

	var existingSongs = make(map[string]bool)
	for _, song := range songsList {
		if !song.IsDir() {
			log.Printf("File %s is out of place, but we're not doing anything with it\n", song.Name())
			continue
		}

		dir := filepath.Base(song.Name())
		existingSongs[dir] = true
		if _, ok := m.Songs[dir]; !ok && dir != "" {
			err := os.RemoveAll(fmt.Sprintf(songDirTemplate, dir))
			if err != nil {
				log.Printf("[Cleanup]: Error while removing %s: %s\n", song.Name(), err.Error())
				continue
			}
			n++
		}
	}

	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	for _, e := range m.Songs {
		if existingSongs[e.ID] {
			continue
		}

		err := m.DeleteEntry(e.ID)
		if err != nil {
			log.Printf("[Cleanup]: Error while deleting %s (%s): %s\n", e.SongName(), e.ID, err.Error())
			continue
		}
		n++
	}

	m.DeleteGeneratedFiles(cfg.KeepGeneratedDays)

	return
}

// runCleanJob deletes all unused songs at 0:00
func (m *Manager) runCleanJob(maxAgeDays int) {
	// since maxAgeDays never changes while running, we don't need to start this at all
	if maxAgeDays < 0 {
		return
	}

	for {
		now := time.Now()
		nextTime := now.Round(24 * time.Hour)
		if nextTime.Before(now) {
			nextTime = nextTime.Add(24 * time.Hour)
		}

		time.Sleep(time.Until(nextTime))

		deleted := m.DeleteGeneratedFiles(maxAgeDays)
		if deleted > 0 {
			log.Printf("[Cleanup]: Removed %d unused generated files in\n", deleted)
		}
	}
}

// DeleteGeneratedFiles deletes generates files older than maxAgeDays * 24 * Hour.
// if maxAgeDays is less than 0, no files will be deleted
// if maxAgeDays is 0, all generated files will be deleted
func (m *Manager) DeleteGeneratedFiles(maxAgeDays int) (n int) {
	// If it's less than 0, don't delete any songs
	if maxAgeDays < 0 {
		return
	}

	// Delete all generated files that were created before this date
	maxDate := time.Now().Add(time.Duration(-maxAgeDays) * 24 * time.Hour)

	m.SongsLock.RLock()
	defer m.SongsLock.RUnlock()

	for _, song := range m.Songs {
		if song.LastEdit.After(maxDate) {
			continue
		}

		// this is the mp3 file
		var outName = filepath.Join("data", "songs", song.ID, "latest.mp3")

		err := os.Remove(outName)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("Cannot clean up MP3 file for %s (%s): %s\n", song.SongName(), song.ID, err.Error())
			}
			continue
		}

		n++
	}

	return
}
