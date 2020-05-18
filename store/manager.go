package store

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"xarantolus/sensiblehub/store/music"

	"github.com/gorilla/websocket"
)

const (
	managerDataFile = "data/manager.json"
	searchURL       = "https://www.youtube.com/results?search_query=%s"
)

// Manager is the struct that contains the application's data. It is only present *once* in the instance named `M`
type Manager struct {
	// Songs is a map[song.ID]song, it maps the ids to their songs
	Songs     map[string]music.Entry `json:"songs"`
	SongsLock *sync.RWMutex          `json:"-"`

	enqueuedURLs chan string

	evtFunc func(f func(c *websocket.Conn) error)

	isWorking    bool
	isWorkingMut sync.RWMutex

	lastErr error

	OnEvent func(ManagerEvent) `json:"-"`
}

// M is the global Manager instance
var M *Manager

// InitializeManager initializes the global Manager instance `M`. It must be called before using `M`
func InitializeManager() (err error) {
	// Initialize an empty manager
	M = &Manager{
		Songs:        make(map[string]music.Entry),
		SongsLock:    new(sync.RWMutex),
		enqueuedURLs: make(chan string, 25),
	}

	go M.serve()

	// Try reading from file
	f, err := os.Open(managerDataFile)
	if err != nil {
		// If the file doesn't exist, it will be created on next save
		if os.IsNotExist(err) {
			return nil
		}

		M = nil
		// This is a real error - might have to do with permissions
		return err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(M)
	if err != nil {
		M = nil
		return err
	}

	return
}

// Save saves the current state of the manager instance to its data file
func (m *Manager) Save(lock ...bool) (err error) {
	var shouldLock = true // Default: Lock m.SongsLock
	if len(lock) > 0 {
		shouldLock = lock[0]
	}

	if shouldLock {
		m.SongsLock.Lock()
		defer m.SongsLock.Unlock()
	}

	err = os.MkdirAll(filepath.Dir(managerDataFile), 0755) // https://stackoverflow.com/a/31151508
	if err != nil {
		return
	}

	var tmpFile = managerDataFile + ".tmp"

	f, err := os.Create(tmpFile)
	if err != nil {
		return
	}

	// Use a pretty format
	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")

	err = enc.Encode(m)
	if err != nil {
		f.Close()
		return
	}

	err = f.Close()
	if err != nil {
		return
	}

	// Overwrite old file, this is an "atomic update"
	return os.Rename(tmpFile, managerDataFile)
}

// Add adds an entry to the manager and saves it.
// It assumes that m.SongsLock is already locked
func (m *Manager) Add(e *music.Entry) (err error) {
	if _, ok := m.Songs[e.ID]; ok {
		return fmt.Errorf("ID %s already taken", e.ID)
	}

	m.Songs[e.ID] = *e

	err = m.Save(false)
	if err != nil {
		return
	}

	m.event("song-add", map[string]interface{}{
		"id":   e.ID,
		"song": *e,
	})

	return
}

// GetEntry returns the entry with the given ID
func (m *Manager) GetEntry(id string) (e music.Entry, ok bool) {
	m.SongsLock.RLock()
	e, ok = m.Songs[id]
	m.SongsLock.RUnlock()

	return
}

// Enqueue adds a new url to the queue of songs that should be downloaded
func (m *Manager) Enqueue(u string) (err error) {
	parsed, err := url.ParseRequestURI(u)
	if err == nil {
		if e, ok := m.hasLink(parsed); ok {
			return fmt.Errorf("%s has already been downloaded", e.SongName())
		}
	} else {
		// Search youtube music - these are auto generated videos that exist for *some* artists
		// Only the first item will be downloaded by m.download because of options passed to youtube-dl
		u = fmt.Sprintf(searchURL, url.QueryEscape(strings.TrimSpace(u)+" \"auto generated\""))
	}

	select {
	case m.enqueuedURLs <- u:
		return nil
	default:
		return fmt.Errorf("Cannot enqueue more songs at this time")
	}
}

func (m *Manager) hasLink(u *url.URL) (me music.Entry, ok bool) {
	u.Host = strings.TrimPrefix(u.Host, "www.")

	// make sure that resolved links are recognized
	if u.Host == "youtu.be" {
		u.Host = "youtube.com"
		u.Path = "/watch"

		u.Query().Set("v", strings.TrimPrefix(u.Path, "/"))
	}

	// clean other url parameters
	if u.Host == "youtube.com" {
		for key := range u.Query() {
			if key != "v" {
				u.Query().Del(key)
			}
		}
	}

	u.Scheme = ""

	s := u.String()
	for _, e := range m.AllEntries() {
		if e.SourceURL == "Imported" {
			continue
		}

		// This check is actually flawed.
		// If we have a song with the SourceURL `youtube.com/watch?v=000`, it will also match `youtube.com/watch?v=00`
		// This is however quite unlikely to be annoying in practice
		if strings.Contains(strings.Replace(e.SourceURL, "www.", "", 1), s) {
			return e, true
		}
	}

	return me, false
}

// generateID generates a new, unique ID.
// It assumes that m.SongsLock is already locked
func (m *Manager) generateID() (id string) {
	var counter int

	for counter < 10000 {
		id = randSeq(4) // len(letters)^4 = 7.311.616 => this is more than enough for now

		if _, ok := m.Songs[id]; !ok {
			return
		}

		counter++
	}

	panic("couldn't obtain a randomly-generated unique song ID after 10.000 tries")
}

// https://stackoverflow.com/a/22892986
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func init() {
	rand.Seed(time.Now().UnixNano())
}
func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (m *Manager) serve() {
	for newURL := range m.enqueuedURLs {
		m.setIsWorking(true)

		err := m.download(newURL)
		m.lastErr = err

		if err != nil {
			log.Printf("[Downloader]: %s\n", err.Error())
		}

		m.setIsWorking(false)
	}
}

func (m *Manager) setIsWorking(state bool) {
	m.isWorkingMut.Lock()
	m.isWorking = state
	m.isWorkingMut.Unlock()

	if state {
		m.event("progress-start", nil)
	} else {
		data := map[string]string{}
		if m.lastErr != nil {
			data["error"] = m.lastErr.Error()
		}
		m.event("progress-end", data)
	}
}

// LastError returns the last error encountered while downloading
func (m *Manager) LastError() error {
	return m.lastErr
}

// IsWorking returns whether the manager is currently doing work
func (m *Manager) IsWorking() bool {
	m.isWorkingMut.RLock()
	defer m.isWorkingMut.RUnlock()

	return m.isWorking
}
