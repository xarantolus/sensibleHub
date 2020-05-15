package web

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	sizeParam := r.URL.Query().Get("size")
	switch strings.ToUpper(sizeParam) {
	case "SMALL":

		// While ServeContent checks this too, the calls to coverGroup.Do are quite expensive and take long.
		// So if we are able to abort before getting to that point because the browser already has that image,
		// we can save some resources and make this request *a lot* faster
		lm := r.Header.Get("If-Modified-Since")
		if lm != "" && lm == e.LastEdit.UTC().Format(http.TimeFormat) {
			http.Error(w, "", http.StatusNotModified)
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

			// ffmpeg options are quite complicated. Depending on whether on not we have a picture, we
			// need some parameters to embed that etc.
			cmd := exec.Command("ffmpeg",
				"-y", // don't ever ask anything

				"-i", ap, // Audio input
			)

			if e.PictureData.Filename != "" {
				cmd.Args = append(cmd.Args,
					"-i", cp, // Cover input

					// somehow map these to each other
					"-map", "0", "-map", "1",

					// image cover metadata
					"-metadata:s:v", "title=Album cover",
					"-metadata:s:v", "comment=Cover (front)")
			}

			cmd.Args = append(cmd.Args,
				// set some idv3 tags
				"-metadata", "title="+e.MusicData.Title,
				"-metadata", "artist="+e.MusicData.Artist,
				"-metadata", "album="+e.MusicData.Album,
				"-metadata", "date="+strconv.Itoa(e.MusicData.Year),

				"-f", "mp3",
				"-id3v2_version", "3")

			// Audio settings: start and end
			if e.AudioSettings.Start != -1 {
				cmd.Args = append(cmd.Args, "-ss", strconv.FormatFloat(e.AudioSettings.Start, 'f', 3, 64))
			}

			if e.AudioSettings.End != -1 {
				cmd.Args = append(cmd.Args, "-to", strconv.FormatFloat(e.AudioSettings.End, 'f', 3, 64))
			}

			// Set output file
			cmd.Args = append(cmd.Args, tmpFile)

			if debug {
				cmd.Stderr = os.Stderr
			}

			// Start running command
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
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", e.MusicData.Artist+" - "+e.MusicData.Title+".mp3"))

	http.ServeFile(w, r, outName)

	return nil
}
