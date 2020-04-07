package web

import (
	"fmt"
	"log"
	"net/http"
)

type HttpError struct {
	StatusCode int
	Message    string
}

func (e HttpError) Error() string {
	return fmt.Sprintf("Status code %d: %s", e.StatusCode, e.Message)
}

// ErrWrap wraps a http handler func that also returns an error and handles said error
func ErrWrap(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}

		log.Printf("[Web] Serving %s: %s\n", r.URL.Path, err.Error())

		// is it an http error?
		if h, ok := err.(HttpError); ok {
			http.Error(w, h.Message, h.StatusCode)
			return
		}

		// some other error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
