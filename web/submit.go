package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type addAccept struct {
	SearchTerm string `json:"searchTerm"`
}

// HandleDownloadSong handles a song download request. This kind of request is done
// from the /add page, either using AJAX (with ?format=json) or a normal form submit
func (s *server) HandleDownloadSong(w http.ResponseWriter, r *http.Request) (err error) {
	// For AJAX requests
	if strings.ToUpper(r.URL.Query().Get("format")) == "JSON" {
		acc := new(addAccept)

		err = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(acc)
		if err != nil {
			return
		}

		err = s.m.Enqueue(acc.SearchTerm)
		if err == nil {
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte(`{}`))
		} else {
			w.WriteHeader(http.StatusPreconditionFailed)
			err = json.NewEncoder(w).Encode(map[string]string{
				"message": err.Error(),
			})
		}

		return err
	}

	err = r.ParseForm()
	if err != nil {
		return
	}

	err = s.m.Enqueue(r.FormValue("searchTerm"))
	if err != nil {
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return
}

// HandleAbortDownload aborts a currently running download
func (s *server) HandleAbortDownload(w http.ResponseWriter, r *http.Request) (err error) {
	// For AJAX requests
	if strings.ToUpper(r.URL.Query().Get("format")) == "JSON" {
		err = s.m.AbortDownload()
		if err == nil {
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte(`{}`))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(map[string]string{
				"message": err.Error(),
			})
		}

		return err
	}

	err = s.m.AbortDownload()
	if err != nil {
		return
	}

	http.Redirect(w, r, "/add", http.StatusSeeOther)

	return
}

// HandleEditAlbum edits an album, only accepts an image. This way, one image can quickly be set for all songs in one album
func (s *server) HandleEditAlbum(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["artist"] == "" || v["album"] == "" {
		return HTTPError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need an artist and album",
		}
	}

	err = r.ParseMultipartForm(250 << 20) // Limit: 250MB
	if err != nil {
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	coverFile, fh, err := r.FormFile("cover-upload-button")
	if err != nil {
		if err == http.ErrMissingFile {
			return HTTPError{
				StatusCode: http.StatusBadRequest,
				Message:    "Must include image",
			}
		}
		// Any other error
		return err
	}

	err = s.m.EditAlbumCover(v["artist"], v["album"], fh.Filename, coverFile)
	if err != nil {
		return
	}

	http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)

	return
}
