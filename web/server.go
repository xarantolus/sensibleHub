package web

import (
	"log"
	"net/http"
	"strconv"
	"xarantolus/sensiblehub/store"
	"xarantolus/sensiblehub/store/config"

	"github.com/gorilla/mux"
)

const (
	debug = true
)

// RunServer runs the web server on the port specified in `cfg`
func RunServer(cfg config.Config) (err error) {
	err = parseTemplates()
	if err != nil {
		return
	}

	store.M.SetEventFunc(AllSockets)

	r := mux.NewRouter()

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data")))).Methods(http.MethodGet)

	// Static assets
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/fav/favicon.ico")
	}).Methods(http.MethodGet)

	// Index page
	r.HandleFunc("/", ErrWrap(debugWrap(HandleIndex))).Methods(http.MethodGet)

	// Song submit form
	r.HandleFunc("/add", ErrWrap(debugWrap(HandleAddSong))).Methods(http.MethodGet)
	r.HandleFunc("/add", ErrWrap(debugWrap(HandleDownloadSong))).Methods(http.MethodPost)

	// Song listings
	r.HandleFunc("/songs", ErrWrap(debugWrap(HandleTitleListing))).Methods(http.MethodGet)
	r.HandleFunc("/artists", ErrWrap(debugWrap(HandleArtistListing))).Methods(http.MethodGet)
	r.HandleFunc("/years", ErrWrap(debugWrap(HandleYearListing))).Methods(http.MethodGet)

	// Search listing
	r.HandleFunc("/search", ErrWrap(debugWrap(HandleSearchListing))).Methods(http.MethodGet)

	// Song html page
	r.HandleFunc("/song/{songID}", ErrWrap(debugWrap(HandleShowSong))).Methods(http.MethodGet)

	// Song edit handler
	r.HandleFunc("/song/{songID}", ErrWrap(debugWrap(HandleEditSong))).Methods(http.MethodPost)

	// Song Data
	r.HandleFunc("/song/{songID}/cover", ErrWrap(HandleCover)).Methods(http.MethodGet)
	r.HandleFunc("/song/{songID}/audio", ErrWrap(HandleAudio)).Methods(http.MethodGet)
	r.HandleFunc("/song/{songID}/mp3", ErrWrap(HandleMP3)).Methods(http.MethodGet)

	// Websocket
	r.HandleFunc("/api/v1/events/ws", ErrWrap(debugWrap(HandleWebsocket)))

	log.Printf("[Web] Server listening on port %d\n", cfg.Port)
	return http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}

// debugWrap parses templates every time they are requested if the debug mode is enabled
func debugWrap(f func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	if !debug {
		return f
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		err := parseTemplates()
		if err != nil {
			return err
		}
		return f(w, r)
	}
}
