package web

import (
	"errors"
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

// errWrap turns an error returned by a media handler into a plain-text response.
// Unexpected errors are logged and answered without internal details.
func errWrap(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}

		var h httpError
		if errors.As(err, &h) {
			if h.StatusCode >= 500 {
				log.Printf("[Web] %s %s: %s\n", r.Method, r.URL.Path, err.Error())
			}
			http.Error(w, h.Message, h.StatusCode)
			return
		}

		log.Printf("[Web] %s %s: %s\n", r.Method, r.URL.Path, err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
