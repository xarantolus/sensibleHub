package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"xarantolus/sensiblehub/store"

	"github.com/gorilla/mux"
	"golang.org/x/sync/singleflight"
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

	w.Header().Set("Content-Type", e.PictureData.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", e.MusicData.Artist+" - "+e.MusicData.Title+filepath.Ext(cp)))

	http.ServeFile(w, r, cp)

	return nil
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

	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", e.MusicData.Artist+" - "+e.MusicData.Title+filepath.Ext(e.FileData.Filename)))

	http.ServeFile(w, r, cp)

	return nil
}

var (
	mp3Group singleflight.Group
)

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

	ap, err := filepath.Abs(e.AudioPath())
	if err != nil {
		return
	}

	cp, err := filepath.Abs(e.CoverPath())
	if err != nil {
		return
	}

	var outName = filepath.Join("data", "songs", e.ID, "latest.mp3")

	// Re-create this mp3 file if it doesn't exist or doesn't have the latest details
	if fi, ferr := os.Stat(outName); os.IsNotExist(ferr) || fi.ModTime().Before(e.LastEdit) {
		_, err, _ = mp3Group.Do(e.ID, func() (res interface{}, err error) {
			td, err := ioutil.TempDir("", "sh-mp3")
			if err != nil {
				return
			}

			tmpFile := filepath.Join(td, "audio.mp3")

			cmd := exec.Command("ffmpeg",
				"-y",

				"-i", ap, // Audio input

				"-i", cp, // Cover input

				// somehow map these to each other
				"-map", "0", "-map", "1",

				"-c:v", "copy",
				"-f", "mp3",

				"-id3v2_version", "3",

				// set some idv3 tags
				"-metadata", "title="+e.MusicData.Title,
				"-metadata", "artist="+e.MusicData.Artist,
				"-metadata", "album="+e.MusicData.Album,
				"-metadata", "year="+strconv.Itoa(e.MusicData.Year),

				// image cover metadata
				"-metadata:s:v", "title=Album cover",
				"-metadata:s:v", "comment=Cover (front)",

				// "-flush_packets", "0",
				tmpFile)

			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				return
			}

			err = os.Rename(tmpFile, outName)
			if err != nil {
				return
			}

			mp3Group.Forget(e.ID)

			return outName, nil
		})

		if err != nil {
			return err
		}

	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", e.MusicData.Artist+" - "+e.MusicData.Title+".mp3"))

	http.ServeFile(w, r, outName)

	return nil
}
