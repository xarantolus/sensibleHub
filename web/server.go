package web

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/config"
	"xarantolus/sensibleHub/web/api"

	"github.com/gorilla/mux"
)

type server struct {
	m      *store.Manager
	router *mux.Router
}

// RunServer serves the API, song media and the single-page app in spa on the port from cfg.
func RunServer(manager *store.Manager, cfg config.Config, spa fs.FS) error {
	r := mux.NewRouter()
	s := server{m: manager, router: r}

	s.route("/media/songs/{songID}/cover", s.HandleCover).Methods(http.MethodGet)
	s.route("/media/songs/{songID}/audio", s.HandleAudio).Methods(http.MethodGet)
	s.route("/media/songs/{songID}/mp3", s.HandleMP3).Methods(http.MethodGet)

	r.Use(compressAPI)
	api.New(r, manager)

	r.PathPrefix("/").Handler(spaHandler(spa)).Methods(http.MethodGet, http.MethodHead)

	log.Printf("[Web] Server listening on port %d\n", cfg.Port)
	return http.ListenAndServe(":"+strconv.Itoa(cfg.Port), r)
}

func (s *server) route(path string, f func(w http.ResponseWriter, r *http.Request) error) *mux.Route {
	return s.router.HandleFunc(path, errWrap(f))
}

// spaHandler serves the built frontend. Paths that are not files are client-side
// routes and get index.html; unknown API paths still get a 404.
func spaHandler(spa fs.FS) http.Handler {
	files := http.FileServerFS(spa)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/media/") {
			http.NotFound(w, r)
			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name != "" {
			if info, err := fs.Stat(spa, name); err == nil && !info.IsDir() {
				if strings.HasPrefix(name, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "no-cache")
				}
				files.ServeHTTP(w, r)
				return
			} else if err != nil && !errors.Is(err, fs.ErrNotExist) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, spa, "index.html")
	})
}
