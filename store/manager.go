package store

import (
	"context"
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

	"xarantolus/sensibleHub/store/config"
	"xarantolus/sensibleHub/store/music"

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

	// enqueuedURLs is a queue where all urls that should be downloaded are put in.
	// They will be processed sequentially
	enqueuedURLs chan string

	// evtFunc is called whenever a websocket event should be written to all sockets
	// It should be set before using the manager / starting the server
	evtFunc func(f func(c *websocket.Conn) error)

	// isWorking indicates if the manager is currently downloading something.
	// State changes are accompanied by the "progress-start" and "progress-end" websocket events
	isWorking    bool
	isWorkingMut sync.RWMutex

	// lastErr is the last error encountered while running youtube-dl, might be nil
	lastErr error

	downloadContextLock sync.Mutex
	// currentDownload contains the url that is currently processed by youtube-dl
	currentDownload string
	downloadContext context.Context
	// downloadCancelFunc can be called to cancel the currently running download process
	downloadCancelFunc context.CancelFunc

	// cfg is the configuration
	cfg config.Config
}

// NewManager initializes the global Manager instance `M`. It must be called before using `M`
func NewManager(cfg config.Config) (m *Manager, err error) {
	// Initialize an empty manager
	m = &Manager{
		Songs:        make(map[string]music.Entry),
		SongsLock:    new(sync.RWMutex),
		enqueuedURLs: make(chan string, 25), // Allow up to 25 items to be queued
		cfg:          cfg,
	}

	go m.serve()

	// Try reading from file
	f, err := os.Open(managerDataFile)
	if err != nil {
		// If the file doesn't exist, it will be created on next save
		if os.IsNotExist(err) {
			return m, nil
		}

		// This is a real error - might have to do with permissions
		return m, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(m)
	if err != nil {
		return m, err
	}

	return
}

// Save saves the current state of the manager instance to its data file
func (m *Manager) Save(lock ...bool) (err error) {
	shouldLock := true // Default: Lock m.SongsLock
	if len(lock) > 0 {
		shouldLock = lock[0]
	}

	if shouldLock {
		m.SongsLock.Lock()
		defer m.SongsLock.Unlock()
	}

	err = os.MkdirAll(filepath.Dir(managerDataFile), 0o755) // https://stackoverflow.com/a/31151508
	if err != nil {
		return
	}

	tmpFile := managerDataFile + ".tmp"

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
		u = fmt.Sprintf("ytsearch:%s %q", strings.TrimSpace(u), "auto generated")
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
		q := make(url.Values)
		q.Set("v", strings.TrimPrefix(u.Path, "/"))
		u.RawQuery = q.Encode()

		u.Host = "youtube.com"
		u.Path = "/watch"
	}

	// music.youtube.com links will later be displayed as ...youtube.com/watch... (due to youtube-dls webpage_url field),
	// so we should compare them like that
	if u.Host == "music.youtube.com" {
		q := make(url.Values)
		q.Set("v", u.Query().Get("v"))
		u.RawQuery = q.Encode()

		u.Host = "youtube.com"
		u.Path = "/watch"
	}

	// clean other url parameters
	if u.Host == "youtube.com" {
		q := make(url.Values)
		q.Set("v", u.Query().Get("v"))
		u.RawQuery = q.Encode()
	}

	u.Scheme = ""

	s := u.String()
	for _, e := range m.AllEntries() {
		if e.IsImported() {
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
// It assumes that m.SongsLock is already locked for reading
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
			log.Printf("[Downloader] %s\n", err.Error())
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

// GetConfig returns the config used by this Manager
func (m *Manager) GetConfig() config.Config {
	return m.cfg
}
