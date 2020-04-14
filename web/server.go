package web

import (
	"log"
	"net/http"
	"strconv"
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

	r := mux.NewRouter()

	r.PathPrefix("/data/").Handler(http.StripPrefix("/data/", http.FileServer(http.Dir("data")))).Methods(http.MethodGet)

	// Static assets
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("assets")))).Methods(http.MethodGet)
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/fav/favicon.ico")
	}).Methods(http.MethodGet)

	// Song listings
	r.HandleFunc("/songs", ErrWrap(debugWrap(HandleTitleListing))).Methods(http.MethodGet)
	r.HandleFunc("/artists", ErrWrap(debugWrap(HandleArtistListing))).Methods(http.MethodGet)

	// Song html page
	r.HandleFunc("/song/{songID}", ErrWrap(debugWrap(HandleShowSong))).Methods(http.MethodGet)

	// Song edit handler
	r.HandleFunc("/song/{songID}", ErrWrap(debugWrap(HandleEditSong))).Methods(http.MethodPost)

	// Song Data
	r.HandleFunc("/song/{songID}/cover", ErrWrap(HandleCover)).Methods(http.MethodGet)
	r.HandleFunc("/song/{songID}/audio", ErrWrap(HandleAudio)).Methods(http.MethodGet)
	r.HandleFunc("/song/{songID}/mp3", ErrWrap(HandleMP3)).Methods(http.MethodGet)

	log.Printf("Server listening on port %d\n", cfg.Port)
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
