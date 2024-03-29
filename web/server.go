package web

import (
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type server struct {
	debug bool

	m         *store.Manager
	templates *template.Template

	assetFS    fs.FS
	templateFS fs.FS

	router *mux.Router

	connectedSocketsLock sync.Mutex
	connectedSockets     map[*websocket.Conn]chan struct{}
}

// RunServer runs the web server on the port specified in `cfg`.
// `debugMode` sets whether to start the server in debug mode
func RunServer(manager *store.Manager, cfg config.Config, assetFS, templateFS fs.FS, debugMode bool) (err error) {
	r := mux.NewRouter()
	r.StrictSlash(true)

	var server = server{
		debug: debugMode,
		m:     manager,

		assetFS:    assetFS,
		templateFS: templateFS,

		router:           r,
		connectedSockets: make(map[*websocket.Conn]chan struct{}),
	}

	if debugMode {
		log.Printf("[Debug] Using local templates and assets because of debug mode")
		server.templateFS = os.DirFS(".")
		server.assetFS = server.templateFS
	}

	err = server.parseTemplates(templateFS)
	if err != nil {
		return
	}

	manager.SetEventFunc(server.AllSockets)

	// set up the file server that serves our data directory
	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data")))).Methods(http.MethodGet)

	// serve static assets and a favicon
	r.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assetFS))).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "assets/fav/favicon.ico", http.StatusMovedPermanently)
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
	server.route("/added", server.HandleSortedByAddDateListing).Methods(http.MethodGet)

	// Search listing
	server.route("/search", server.HandleSearchListing).Methods(http.MethodGet)
	// Search API for search suggestions

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

	// API
	server.route("/api/v1/listing/{listing}", server.HandleAPIListing).Methods(http.MethodGet)
	server.route("/api/v1/song/{songID}", server.HandleAPISong).Methods(http.MethodGet)

	server.route("/api/v1/search", server.HandleAPISongSearch).Methods(http.MethodGet)
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
		err := s.parseTemplates(s.templateFS)
		if err != nil {
			return err
		}
		return f(w, r)
	}
}
