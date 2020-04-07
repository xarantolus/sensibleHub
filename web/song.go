package web

import (
	"fmt"
	"net/http"
	"xarantolus/sensiblehub/store"
	"xarantolus/sensiblehub/store/music"

	"github.com/gorilla/mux"
)

type SongPage struct {
	Title string

	music.Entry
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
		fmt.Sprintf("%s - %s", e.MusicData.Title, e.MusicData.Artist),
		e,
	})
}
