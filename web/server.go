package web

import (
	"log"
	"net/http"
	"strconv"
	"xarantolus/sensiblehub/store/config"
)

// RunServer runs the web server on the port specified in `cfg`
func RunServer(cfg config.Config) (err error) {
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))

	// Static assets
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/fav/favicon.ico")
	})

	log.Printf("Server listening on port %d\n", cfg.Port)
	return http.ListenAndServe(":"+strconv.Itoa(cfg.Port), nil)
}
