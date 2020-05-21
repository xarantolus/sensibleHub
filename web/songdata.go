package web

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"xarantolus/sensiblehub/store"

	"github.com/gorilla/mux"
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

	sizeParam := r.URL.Query().Get("size")
	le := e.LastEdit.UTC().Format(http.TimeFormat)

	r.Header.Set("Cache-Control", "max-age=0, must-revalidate")

	// While ServeContent checks this too, the calls to coverGroup.Do are quite expensive and take long.
	// So if we are able to abort before getting to that point because the browser already has that image,
	// we can save some resources and make this request *a lot* faster
	lm := r.Header.Get("If-Modified-Since")

	if lm != "" && lm == le {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	switch strings.ToUpper(sizeParam) {
	case "SMALL":

		coverBytes, format, err := e.CoverPreview()
		if err != nil {
			return err
		}

		w.Header().Set("Last-Modified", le)
		w.Header().Set("Content-Type", format)

		http.ServeContent(w, r, "cover-small.png", e.LastEdit, bytes.NewReader(coverBytes))
		return nil
	default:

		w.Header().Set("Last-Modified", le)
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
