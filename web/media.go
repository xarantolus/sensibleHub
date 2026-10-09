package web

import (
	"bytes"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"xarantolus/sensibleHub/store"

	"github.com/gorilla/mux"
)

// HandleCover serves the cover image for the song with the `songID` given in the URL,
// or 404 if it has none. If the URL parameter `size` is "small", a preview is sent.
func (s *server) HandleCover(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return httpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := s.m.GetEntry(v["songID"])
	if !ok {
		return httpError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}

	// Browsers keep the image but revalidate it on every use, so an edited cover
	// shows up without changing the URL; unchanged covers cost a 304.
	sizeParam := r.URL.Query().Get("size")
	le := e.LastEdit.UTC().Format(http.TimeFormat)
	etag := fmt.Sprintf(`"%x-%s"`, e.LastEdit.UnixNano(), strings.ToLower(sizeParam))
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("ETag", etag)

	// ServeContent checks these too, but only after the preview has been generated,
	// which is the expensive part.
	if r.Header.Get("If-None-Match") == etag || (r.Header.Get("If-None-Match") == "" && r.Header.Get("If-Modified-Since") == le) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	cp := e.CoverPath()
	if cp == "" {
		return httpError{StatusCode: http.StatusNotFound, Message: "Song has no cover"}
	}

	switch {
	case strings.ToUpper(sizeParam) == "SMALL":
		coverBytes, format, err := e.CoverPreview()
		if err != nil {
			return err
		}

		w.Header().Set("Last-Modified", le)
		w.Header().Set("Content-Type", format)
		w.Header().Set("Content-Length", strconv.Itoa(len(coverBytes)))

		http.ServeContent(w, r, "cover-small.png", e.LastEdit, bytes.NewReader(coverBytes))
		return nil
	default:
		// WARNING: Using ServeContent is REQUIRED here, ServeFile does NOT work
		// This is because ServeFile overwrites the last-modified header to the last time the file has been edited and sends 304 Not modified
		// The browser will continue to use an old/cached and already deleted image if we use ServeFile, at least until
		// the website has been reloaded. Since that isn't good, we need to do it manually and correctly

		coverFile, err := os.Open(cp)
		if err != nil {
			return err
		}
		defer coverFile.Close()

		mtype := mime.TypeByExtension(strings.TrimPrefix(filepath.Ext(e.PictureData.Filename), "."))
		if mtype != "" {
			w.Header().Set("Content-Type", mtype)
		}

		info, err := coverFile.Stat()
		if err == nil {
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		}

		w.Header().Set("Last-Modified", le)

		fn := store.CleanName(e.Filename(filepath.Ext(cp)))

		w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", fn))

		http.ServeContent(w, r, fn, e.LastEdit, coverFile)
		return nil
	}
}

// HandleAudio serves the default audio for the song with the `songID` specified in the URL.
// This is different from the MP3 download handler.
func (s *server) HandleAudio(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return httpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := s.m.GetEntry(v["songID"])
	if !ok {
		return httpError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}
	cp := e.AudioPath()

	// A song's original audio file is never rewritten (trimming only changes playback).
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", store.CleanName(e.Filename(filepath.Ext(e.FileData.Filename)))))

	http.ServeFile(w, r, cp)

	return nil
}

// HandleMP3 returns the requested songs' audio as an MP3 stream.
// It creates the mp3 file from its associated data and caches the result until the song is edited
func (s *server) HandleMP3(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return httpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	e, ok := s.m.GetEntry(v["songID"])
	if !ok {
		return httpError{
			StatusCode: http.StatusNotFound,
			Message:    "Song not found",
		}
	}

	outName, err := e.MP3Path(s.m.GetConfig())
	if err != nil {
		return
	}

	// The MP3 is regenerated after edits; ServeFile's Last-Modified lets clients revalidate.
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", store.CleanName(e.Filename("mp3"))))

	http.ServeFile(w, r, outName)

	return nil
}
