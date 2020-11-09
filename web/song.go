package web

import (
	"net/http"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/music"

	"github.com/gorilla/mux"
)

type songPage struct {
	Title string

	*music.Entry

	SimilarSongs []music.Entry
}

// HandleShowSong shows information about a song
func HandleShowSong(w http.ResponseWriter, r *http.Request) (err error) {
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

	similar := store.M.GetRelatedSongs(e)

	return renderTemplate(w, r, "song.html", songPage{
		e.SongName(),
		&e,
		similar,
	})
}

// HandleEditSong handles editing a song
func HandleEditSong(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return HTTPError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	isAjax := r.Header.Get("X-XHR") == "true"

	songID := v["songID"]

	err = r.ParseMultipartForm(250 << 20) // Limit: 250MB
	if err != nil {
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	if r.FormValue("delete-cover") == "delete-cover" {
		err = store.M.DeleteCoverImage(songID)
		if err != nil {
			return err
		}

		if isAjax {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"message": "Deleted"}`, http.StatusOK)
		} else {
			http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
		}
		return nil
	}
	// If the delete button was clicked
	if r.FormValue("delete") == "delete" {
		err = store.M.DeleteEntry(songID)
		if err != nil {
			return err
		}

		if isAjax {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"message": "Deleted"}`, http.StatusOK)
		} else {
			http.Redirect(w, r, "/songs", http.StatusFound)
		}
		return nil
	}

	coverFile, fh, err := r.FormFile("cover-upload-button")
	if err != nil && err != http.ErrMissingFile {
		// Any other error
		return
	}

	var coverName string
	if coverFile != nil {
		coverName = fh.Filename
	}

	newData := store.EditEntryData{
		CoverImage:    coverFile,
		CoverFilename: coverName,

		Title:  r.FormValue("song-title"),
		Artist: r.FormValue("song-artist"),
		Album:  r.FormValue("song-album"),
		Year:   r.FormValue("song-year"),

		Start: r.FormValue("audio-start"),
		End:   r.FormValue("audio-end"),

		Sync: r.FormValue("should-sync"),
	}

	err = store.M.EditEntry(songID, newData)
	if err != nil {
		if err == store.ErrAudioSameStartEnd {
			return HTTPError{
				StatusCode: http.StatusPreconditionFailed,
				Message:    err.Error(),
			}
		}
		return
	}

	if isAjax {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"message": "Updated"}`, http.StatusOK)
		return
	}

	http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
	return
}
