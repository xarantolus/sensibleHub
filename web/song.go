package web

import (
	"net/http"
	"xarantolus/sensiblehub/store"
	"xarantolus/sensiblehub/store/music"

	"github.com/gorilla/mux"
)

type SongPage struct {
	Title string

	*music.Entry
}

// HandleShowSong shows information about a song
func HandleShowSong(w http.ResponseWriter, r *http.Request) (err error) {
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

	return renderTemplate(w, r, "song.html", SongPage{
		e.SongName(),
		&e,
	})
}

// HandleEditSong handles editing a song
func HandleEditSong(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["songID"] == "" {
		return HttpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need a song ID",
		}
	}

	var isAjax = r.Header.Get("X-XHR") == "true"

	songID := v["songID"]

	err = r.ParseMultipartForm(10 << 20) // Limit: 10MB
	if err != nil {
		return
	}

	// If the delete button was clicked
	if r.FormValue("delete") == "delete" {
		err = store.M.DeleteEntry(songID)
		if err != nil {
			return err
		}

		if isAjax {
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
		return
	}

	if isAjax {
		http.Error(w, `{"message": "Updated"}`, http.StatusOK)
		return
	}

	http.Redirect(w, r, r.URL.String(), http.StatusSeeOther)
	return
}
