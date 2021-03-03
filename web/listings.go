package web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/music"
)

// Listing defines a listing of grouped songs
type Listing struct {
	Title string

	// Groups are the groups that should be displayed
	Groups []store.Group
}

// HandleTitleListing returns the song listing, sorted by titles
func HandleTitleListing(w http.ResponseWriter, r *http.Request) (err error) {
	return renderTemplate(w, r, "listing.html", Listing{
		Title:  "Songs",
		Groups: store.M.GroupByTitle(),
	})
}

// HandleArtistListing returns the artist listing, sorted by artist names
func HandleArtistListing(w http.ResponseWriter, r *http.Request) (err error) {
	return renderTemplate(w, r, "listing.html", Listing{
		Title:  "Artists",
		Groups: store.M.GroupByArtist(),
	})
}

// HandleYearListing returns the year listing
func HandleYearListing(w http.ResponseWriter, r *http.Request) (err error) {
	return renderTemplate(w, r, "listing.html", Listing{
		Title:  "Years",
		Groups: store.M.GroupByYear(),
	})
}

// HandleIncompleteListing returns all items with incomplete data
func HandleIncompleteListing(w http.ResponseWriter, r *http.Request) (err error) {
	return renderTemplate(w, r, "listing.html", Listing{
		Title:  "Incomplete",
		Groups: store.M.Incomplete(),
	})
}

// HandleUnsyncedListing returns all items with that are not synced
func HandleUnsyncedListing(w http.ResponseWriter, r *http.Request) (err error) {
	return renderTemplate(w, r, "listing.html", Listing{
		Title:  "Unsynced",
		Groups: store.M.Unsynced(),
	})
}

type searchListing struct {
	Title string
	Songs []music.Entry

	Query string
}

// HandleSearchListing returns a search listing
func HandleSearchListing(w http.ResponseWriter, r *http.Request) (err error) {
	query := r.URL.Query().Get("q")
	if query == "" {
		return fmt.Errorf("Empty query")
	}

	res := store.M.Search(query)
	if len(res) == 1 {
		http.Redirect(w, r, "/song/"+res[0].ID, http.StatusTemporaryRedirect)
		return
	}

	return renderTemplate(w, r, "search.html", searchListing{
		Title: "Search results",
		Songs: res,
		Query: query,
	})
}

type albumPage struct {
	Title string

	A store.Album
}

// HandleShowAlbum shows the album page for the artist and album that's given in the url
func HandleShowAlbum(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["artist"] == "" || v["album"] == "" {
		return HTTPError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need an artist and album",
		}
	}

	al, ok := store.M.GetAlbum(v["artist"], v["album"])
	if !ok {
		return HTTPError{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("Cannot find album %s for artist %s", v["album"], v["artist"]),
		}
	}

	al.Title = al.Artist + " - " + al.Title
	return renderTemplate(w, r, "album.html", albumPage{
		Title: al.Title,
		A:     al,
	})
}

type artistPage struct {
	Title string

	Info store.ArtistInfo
}

// HandleShowArtist shows the artist page for the artist given in the url
func HandleShowArtist(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["artist"] == "" {
		return HTTPError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need an artist",
		}
	}

	artistInfo, ok := store.M.Artist(v["artist"])
	if !ok {
		return HTTPError{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("Cannot find any albums for %s", v["artist"]),
		}
	}

	return renderTemplate(w, r, "artist.html", artistPage{
		Title: artistInfo.Name,
		Info:  artistInfo,
	})
}
