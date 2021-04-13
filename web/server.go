package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type server struct {
	m         *store.Manager
	templates *template.Template

	connectedSocketsLock sync.Mutex
	connectedSockets     map[*websocket.Conn]chan struct{}
}

// RunServer runs the web server on the port specified in `cfg`.
// `debugMode` sets whether to start the server in debug mode
func RunServer(manager *store.Manager, cfg config.Config, debugMode bool) (err error) {
	var server = server{
		m:                manager,
		connectedSockets: make(map[*websocket.Conn]chan struct{}),
	}

	err = server.parseTemplates()
	if err != nil {
		return
	}

	// debugWrap parses templates every time they are requested if the debug mode is enabled
	var debugWrap = func(f func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
		if !debugMode {
			return f
		}

		return func(w http.ResponseWriter, r *http.Request) error {
			err := server.parseTemplates()
			if err != nil {
				return err
			}
			return f(w, r)
		}
	}

	manager.SetEventFunc(server.AllSockets)

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data")))).Methods(http.MethodGet)

	// Static assets
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/fav/favicon.ico")
	}).Methods(http.MethodGet)

	// Index page
	r.HandleFunc("/", ErrWrap(debugWrap(server.HandleIndex))).Methods(http.MethodGet)

	// Song submit form
	r.HandleFunc("/add", ErrWrap(debugWrap(server.HandleAddSong))).Methods(http.MethodGet)
	r.HandleFunc("/add", ErrWrap(debugWrap(server.HandleDownloadSong))).Methods(http.MethodPost)

	r.HandleFunc("/abort", ErrWrap(debugWrap(server.HandleAbortDownload))).Methods(http.MethodPost)

	// Song listings
	r.HandleFunc("/songs", ErrWrap(debugWrap(server.HandleTitleListing))).Methods(http.MethodGet)
	r.HandleFunc("/artists", ErrWrap(debugWrap(server.HandleArtistListing))).Methods(http.MethodGet)
	r.HandleFunc("/years", ErrWrap(debugWrap(server.HandleYearListing))).Methods(http.MethodGet)
	r.HandleFunc("/incomplete", ErrWrap(debugWrap(server.HandleIncompleteListing))).Methods(http.MethodGet)
	r.HandleFunc("/unsynced", ErrWrap(debugWrap(server.HandleUnsyncedListing))).Methods(http.MethodGet)

	// Search listing
	r.HandleFunc("/search", ErrWrap(debugWrap(server.HandleSearchListing))).Methods(http.MethodGet)

	r.HandleFunc("/api/v1/search", ErrWrap(debugWrap(server.HandleAPISongSearch))).Methods(http.MethodGet)

	// Song html page
	r.HandleFunc("/song/{songID}", ErrWrap(debugWrap(server.HandleShowSong))).Methods(http.MethodGet)

	// Song edit handler
	r.HandleFunc("/song/{songID}", ErrWrap(debugWrap(server.HandleEditSong))).Methods(http.MethodPost)

	// Song Data
	r.HandleFunc("/song/{songID}/cover", ErrWrap(server.HandleCover)).Methods(http.MethodGet)
	r.HandleFunc("/song/{songID}/audio", ErrWrap(server.HandleAudio)).Methods(http.MethodGet)
	r.HandleFunc("/song/{songID}/mp3", ErrWrap(server.HandleMP3)).Methods(http.MethodGet)

	r.HandleFunc("/songs/random", ErrWrap(debugWrap(server.HandleRandomSong))).Methods(http.MethodGet)

	// Album Listing
	// r.HandleFunc("/albums")
	r.HandleFunc("/album/{artist}/{album}", ErrWrap(debugWrap(server.HandleShowAlbum))).Methods(http.MethodGet)
	r.HandleFunc("/album/{artist}/{album}", ErrWrap(debugWrap(server.HandleEditAlbum))).Methods(http.MethodPost)

	// Artist listing
	r.HandleFunc("/artist/{artist}", ErrWrap(debugWrap(server.HandleShowArtist))).Methods(http.MethodGet)

	// Websocket
	r.HandleFunc("/api/v1/events/ws", ErrWrap(debugWrap(server.HandleWebsocket)))

	log.Printf("[Web] Server listening on port %d\n", cfg.Port)
	return http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}
