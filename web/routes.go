package web

import (
	"net/http"
	"xarantolus/sensibleHub/store/music"
)

type indexPage struct {
	Title string

	NewEntries []music.Entry
	// Whether all entries in `NewEntries` were added today
	NewEntriesToday bool
}

// HandleIndex shows the main/index page. It shows new songs from today or the most recently added songs
func (s *server) HandleIndex(w http.ResponseWriter, r *http.Request) (err error) {
	entries, today := s.m.Newest()

	return s.renderTemplate(w, r, "index.html", indexPage{
		Title:           "Sensible Hub",
		NewEntries:      entries,
		NewEntriesToday: today,
	})
}

type newPage struct {
	Title       string
	LastError   error
	Running     bool
	DownloadURL string

	NewestSong *music.Entry
}

// HandleAddSong displays the form for adding a song
func (s *server) HandleAddSong(w http.ResponseWriter, r *http.Request) (err error) {
	dl, okr := s.m.IsDownloading()

	var nsp *music.Entry
	ns, ok := s.m.NewestSong()
	if ok {
		nsp = &ns
	}

	return s.renderTemplate(w, r, "add.html", newPage{
		Title:       "Add a new song",
		LastError:   s.m.LastError(),
		Running:     okr,
		DownloadURL: dl,
		NewestSong:  nsp,
	})
}
