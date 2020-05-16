package web

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"xarantolus/sensiblehub/store"

	"github.com/gorilla/mux"
	"golang.org/x/sync/singleflight"
)

var (
	// coverGroup manages the functions that generate cover previews.
	// They are started in HandleCover, and forgotten in HandleEditSong
	coverGroup singleflight.Group
)

func HandleCover(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return HttpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := store.M.GetEntry(v["songID"])
	if !ok {
		return HttpError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}

	cp := e.CoverPath()
	if cp == "" {
		http.ServeFile(w, r, "assets/image-missing.svg")
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	sizeParam := r.URL.Query().Get("size")
	switch strings.ToUpper(sizeParam) {
	case "SMALL":

		// While ServeContent checks this too, the calls to coverGroup.Do are quite expensive and take long.
		// So if we are able to abort before getting to that point because the browser already has that image,
		// we can save some resources and make this request *a lot* faster
		lm := r.Header.Get("If-Modified-Since")
		le := e.LastEdit.UTC().Format(http.TimeFormat)

		if lm != "" && lm == le {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Since we never call
		coverBytes, err, _ := coverGroup.Do(e.ID+"-small", func() (res interface{}, err error) {
			var b bytes.Buffer
			err = store.ResizeCover(cp, 60, &b)
			if err != nil {
				return
			}

			return b.Bytes(), nil
		})
		if err != nil {
			return err
		}
		w.Header().Set("Last-Modified", le)
		w.Header().Set("Content-Type", "image/png")

		http.ServeContent(w, r, "cover-small.png", e.LastEdit, bytes.NewReader(coverBytes.([]byte)))
		return nil

	default:
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", e.MusicData.Artist+" - "+e.MusicData.Title+filepath.Ext(cp)))
		http.ServeFile(w, r, cp)

		return nil
	}
}

func HandleAudio(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return HttpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := store.M.GetEntry(v["songID"])
	if !ok {
		return HttpError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}
	cp := e.AudioPath()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", e.MusicData.Artist+" - "+e.MusicData.Title+filepath.Ext(e.FileData.Filename)))

	http.ServeFile(w, r, cp)

	return nil
}

// HandleMP3 returns the requested songs' audio as an MP3 stream.
// It creates the mp3 file from the associated data and caches the result
func HandleMP3(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return HttpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := store.M.GetEntry(v["songID"])
	if !ok {
		return HttpError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}

	outName, err := e.MP3Path()
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", e.MusicData.Artist+" - "+e.MusicData.Title+".mp3"))

	http.ServeFile(w, r, outName)

	return nil
}
