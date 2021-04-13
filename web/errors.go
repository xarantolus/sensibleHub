package web

import (
	"fmt"
	"log"
	"net/http"
)

// httpError is an error type that also contains an appropriate HTTP status code
type httpError struct {
	StatusCode int
	Message    string
}

func (e httpError) Error() string {
	return fmt.Sprintf("Status code %d: %s", e.StatusCode, e.Message)
}

// errWrap wraps a http handler func that also returns an error and handles said error
func (s *server) errWrap(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}

		// is it an http error?
		if h, ok := err.(httpError); ok {
			// Log all errors that aren't caused by the client directly
			if h.StatusCode < 400 || h.StatusCode >= 500 {
				log.Printf("[Web] %s %s: %s\n", r.Method, r.URL.Path, err.Error())
			}

			http.Error(w, h.Message, h.StatusCode)
			return
		}

		log.Printf("[Web] %s %s: %s\n", r.Method, r.URL.Path, err.Error())

		// some other error

		// there is the possibility that we leak internal details here, but it doesn't really matter in this case
		// as no http requests (with secret tokens etc.) are performed on the back-end
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
}
