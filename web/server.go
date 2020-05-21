package web

import (
	"log"
	"net/http"
	"runtime"
	"strconv"
	"xarantolus/sensiblehub/store"
	"xarantolus/sensiblehub/store/config"

	"github.com/gorilla/mux"
)

var (
	Debug = runtime.GOOS == "windows"
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

	r.HandleFunc("/abort", ErrWrap(debugWrap(HandleAbortDownload))).Methods(http.MethodPost)

	// Song listings
	r.HandleFunc("/songs", ErrWrap(debugWrap(HandleTitleListing))).Methods(http.MethodGet)
	r.HandleFunc("/artists", ErrWrap(debugWrap(HandleArtistListing))).Methods(http.MethodGet)
	r.HandleFunc("/years", ErrWrap(debugWrap(HandleYearListing))).Methods(http.MethodGet)
	r.HandleFunc("/incomplete", ErrWrap(debugWrap(HandleIncompleteListing))).Methods(http.MethodGet)

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

	// Album Listing
	// r.HandleFunc("/albums")
	r.HandleFunc("/album/{artist}/{album}", ErrWrap(debugWrap(HandleShowAlbum))).Methods(http.MethodGet)
	r.HandleFunc("/album/{artist}/{album}", ErrWrap(debugWrap(HandleEditAlbum))).Methods(http.MethodPost)

	// Artist listing
	r.HandleFunc("/artist/{artist}", ErrWrap(debugWrap(HandleShowArtist))).Methods(http.MethodGet)

	// Websocket
	r.HandleFunc("/api/v1/events/ws", ErrWrap(debugWrap(HandleWebsocket)))

	log.Printf("[Web] Server listening on port %d\n", cfg.Port)
	return http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}

// debugWrap parses templates every time they are requested if the debug mode is enabled
func debugWrap(f func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	if !Debug {
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
