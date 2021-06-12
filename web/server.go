package web

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"
)

type server struct {
	debug bool

	m         *store.Manager
	templates *template.Template

	router *mux.Router

	connectedSocketsLock sync.Mutex
	connectedSockets     map[*websocket.Conn]chan struct{}
}

// RunServer runs the web server on the port specified in `cfg`.
// `debugMode` sets whether to start the server in debug mode
func RunServer(manager *store.Manager, cfg config.Config, debugMode bool) (err error) {
	r := mux.NewRouter()
	r.StrictSlash(true)

	var server = server{
		debug:            debugMode,
		m:                manager,
		router:           r,
		connectedSockets: make(map[*websocket.Conn]chan struct{}),
	}

	err = server.parseTemplates()
	if err != nil {
		return
	}

	manager.SetEventFunc(server.AllSockets)

	// set up the file server that serves our data directory
	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data")))).Methods(http.MethodGet)

	// serve static assets and a favicon
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/fav/favicon.ico")
	}).Methods(http.MethodGet)

	// Index page
	server.route("/", server.HandleIndex).Methods(http.MethodGet)

	// Song submit form
	server.route("/add", server.HandleAddSong).Methods(http.MethodGet)
	server.route("/add", server.HandleDownloadSong).Methods(http.MethodPost)

	server.route("/abort", server.HandleAbortDownload).Methods(http.MethodPost)

	// Song listings
	server.route("/songs", server.HandleTitleListing).Methods(http.MethodGet)
	server.route("/artists", server.HandleArtistListing).Methods(http.MethodGet)
	server.route("/years", server.HandleYearListing).Methods(http.MethodGet)
	server.route("/incomplete", server.HandleIncompleteListing).Methods(http.MethodGet)
	server.route("/unsynced", server.HandleUnsyncedListing).Methods(http.MethodGet)
	server.route("/edits", server.HandleRecentlyEditedListing).Methods(http.MethodGet)

	// Search listing
	server.route("/search", server.HandleSearchListing).Methods(http.MethodGet)
	// Search API for search suggestions
	server.route("/api/v1/search", server.HandleAPISongSearch).Methods(http.MethodGet)

	// Song html page and handler for editing
	server.route("/song/{songID}", server.HandleShowSong).Methods(http.MethodGet)
	server.route("/song/{songID}", server.HandleEditSong).Methods(http.MethodPost)

	// Song Data retrieval
	server.route("/song/{songID}/cover", server.HandleCover).Methods(http.MethodGet)
	server.route("/song/{songID}/audio", server.HandleAudio).Methods(http.MethodGet)
	server.route("/song/{songID}/mp3", server.HandleMP3).Methods(http.MethodGet)

	// Redirects to a random song
	server.route("/songs/random", server.HandleRandomSong).Methods(http.MethodGet)

	// Album listing
	server.route("/album/{artist}/{album}", server.HandleShowAlbum).Methods(http.MethodGet)
	server.route("/album/{artist}/{album}", server.HandleEditAlbum).Methods(http.MethodPost)

	// Artist listing
	server.route("/artist/{artist}", server.HandleShowArtist).Methods(http.MethodGet)

	// Websocket
	server.route("/api/v1/events/ws", server.HandleWebsocket)

	log.Printf("[Web] Server listening on port %d\n", cfg.Port)
	return http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}

func (s *server) route(path string, f func(w http.ResponseWriter, r *http.Request) error) *mux.Route {
	return s.router.HandleFunc(path, s.errWrap(s.debugWrap(f)))
}

// debugWrap adds a wrapper that reloads templates before a route is processed when the server debug field is true
func (s *server) debugWrap(f func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	if !s.debug {
		return f
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		err := s.parseTemplates()
		if err != nil {
			return err
		}
		return f(w, r)
	}
}
