package web

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"xarantolus/sensibleHub/store"

	"github.com/gorilla/mux"
)

// HandleCover displays the cover image for the song with the `songID` given in the URL.
// If the song doesn't have a cover image, it will serve a placeholder image (svg) with an 404 status code.
// If the URL parameter `size` is "small", a cover preview image will be generated and sent.
func HandleCover(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return HTTPError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := store.M.GetEntry(v["songID"])
	if !ok {
		return HTTPError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}

	// Store the response, but always re-validate. That way, edited images will be seen
	w.Header().Set("Cache-Control", "no-cache, max-age=0, must-revalidate")
	w.Header().Set("Pragma", "no-cache")

	le := e.LastEdit.UTC().Format(http.TimeFormat)

	// While ServeContent checks this too, the calls to coverGroup.Do are quite expensive and take long.
	// So if we are able to abort before getting to that point because the browser already has that image,
	// we can save some resources and make this request *a lot* faster
	lm := r.Header.Get("If-Modified-Since")

	if lm != "" && lm == le {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	var isMissing bool

	cp := e.CoverPath()
	if cp == "" {
		cp = "assets/image-missing.svg"
		isMissing = true
	}

	sizeParam := r.URL.Query().Get("size")

	switch {
	case strings.ToUpper(sizeParam) == "SMALL" && !isMissing:
		coverBytes, format, err := e.CoverPreview()
		if err != nil {
			return err
		}

		w.Header().Set("Last-Modified", le)
		w.Header().Set("Content-Type", format)

		http.ServeContent(w, r, "cover-small.png", e.LastEdit, bytes.NewReader(coverBytes))
		return nil
	default:
		// WARNING: Using ServeContent is REQUIRED here, ServeFile does NOT work
		// This is because ServeFile overwrites the last-modified header to the last time the file has been edited and sends 304 Not modified
		// The browser will continue to use an old/cached and already deleted image if we use ServeFile, at least until
		// the website has been reloaded. Since that isn't good, we need to do it manually and correctly
		coF, err := os.Open(cp)
		if err != nil {
			return err
		}
		defer coF.Close()

		mtype := mime.TypeByExtension(strings.TrimPrefix(filepath.Ext(e.PictureData.Filename), "."))
		if mtype != "" {
			w.Header().Set("Content-Type", mtype)
		}

		w.Header().Set("Last-Modified", le)

		fn := e.SongName() + filepath.Ext(cp)
		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", fn))

		http.ServeContent(w, r, fn, e.LastEdit, coF)

		return nil
	}
}

// HandleAudio serves the default audio for the song with the `songID` specified in the URL.
// This is different from the MP3 download handler.
func HandleAudio(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return HTTPError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := store.M.GetEntry(v["songID"])
	if !ok {
		return HTTPError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}
	cp := e.AudioPath()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", e.SongName()+filepath.Ext(e.FileData.Filename)))

	http.ServeFile(w, r, cp)

	return nil
}

// HandleMP3 returns the requested songs' audio as an MP3 stream.
// It creates the mp3 file from its associated data and caches the result until the song is edited
func HandleMP3(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return HTTPError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := store.M.GetEntry(v["songID"])
	if !ok {
		return HTTPError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}

	outName, err := e.MP3Path()
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", e.SongName()+".mp3"))

	http.ServeFile(w, r, outName)

	return nil
}
