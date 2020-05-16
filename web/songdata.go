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

	"github.com/bogem/id3v2"

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

	var outName = filepath.Join("data", "songs", e.ID, "latest.mp3")

	// Re-create this mp3 file if it doesn't exist or doesn't have the latest details
	if fi, ferr := os.Stat(outName); os.IsNotExist(ferr) || fi.ModTime().Before(e.LastEdit) {
		_, err, _ = mp3Group.Do(e.ID, func() (res interface{}, err error) {
			td, err := ioutil.TempDir("", "sh-mp3")
			if err != nil {
				return
			}
			defer os.RemoveAll(td)

			tempAudio := filepath.Join(td, "temp.mp3")

			// Convert the given audio to mp3
			err = exec.Command("ffmpeg", "-i", ap, "-f", "mp3", tempAudio).Run()
			if err != nil {
				return
			}

			// And open it
			tag, err := id3v2.Open(tempAudio, id3v2.Options{Parse: false})
			if err != nil {
				return
			}

			// Now we edit its tags

			// Set artist
			if have(&e.MusicData.Artist) {
				tag.SetArtist(e.MusicData.Artist)
			}

			// Set title
			if have(&e.MusicData.Title) {
				tag.SetTitle(e.MusicData.Title)
			}

			if have(&e.MusicData.Album) {
				tag.SetAlbum(e.MusicData.Album)
			}

			if e.MusicData.Year != 0 {
				tag.SetYear(strconv.Itoa(e.MusicData.Year))
			}

			// Set artwork
			if have(&e.PictureData.Filename) {
				b, err := ioutil.ReadFile(e.CoverPath())
				if err != nil {
					tag.Close()
					return nil, err
				}

				// Other mime types don't work as only jpeg and png images are accepted
				mimeType := "image/jpeg"
				if strings.ToUpper(filepath.Ext(e.PictureData.Filename)) == ".PNG" {
					mimeType = "image/png"
				}

				pic := id3v2.PictureFrame{
					Encoding:    id3v2.EncodingUTF8,
					MimeType:    mimeType,
					PictureType: id3v2.PTFrontCover,
					Description: "Front cover",
					Picture:     b,
				}

				tag.AddAttachedPicture(pic)
			}

			// and save it
			err = tag.Save()
			if err != nil {
				tag.Close()
				return
			}

			err = tag.Close()
			if err != nil {
				return
			}

			// if everything goes right, we can now move it to its destination
			err = os.Rename(tempAudio, outName)
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

func have(s *string) bool {
	return strings.TrimSpace(*s) != ""
}
