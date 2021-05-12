package web

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"xarantolus/sensibleHub/store"
	"xarantolus/sensibleHub/store/music"
)

// listingPage defines a listing of grouped songs
type listingPage struct {
	Title string

	// Groups are the groups that should be displayed
	Groups []store.Group
}

// HandleTitleListing returns the song listing, sorted by titles
func (s *server) HandleTitleListing(w http.ResponseWriter, r *http.Request) (err error) {
	return s.renderTemplate(w, r, "listing.html", listingPage{
		Title:  "Songs",
		Groups: s.m.GroupByTitle(),
	})
}

// HandleArtistListing returns the artist listing, sorted by artist names
func (s *server) HandleArtistListing(w http.ResponseWriter, r *http.Request) (err error) {
	return s.renderTemplate(w, r, "listing.html", listingPage{
		Title:  "Artists",
		Groups: s.m.GroupByArtist(),
	})
}

// HandleYearListing returns the year listing
func (s *server) HandleYearListing(w http.ResponseWriter, r *http.Request) (err error) {
	return s.renderTemplate(w, r, "listing.html", listingPage{
		Title:  "Years",
		Groups: s.m.GroupByYear(),
	})
}

// HandleIncompleteListing returns all items with incomplete data
func (s *server) HandleIncompleteListing(w http.ResponseWriter, r *http.Request) (err error) {
	return s.renderTemplate(w, r, "listing.html", listingPage{
		Title:  "Incomplete",
		Groups: s.m.Incomplete(),
	})
}

// HandleUnsyncedListing returns all items with that are not synced
func (s *server) HandleUnsyncedListing(w http.ResponseWriter, r *http.Request) (err error) {
	return s.renderTemplate(w, r, "listing.html", listingPage{
		Title:  "Unsynced",
		Groups: s.m.Unsynced(),
	})
}

// HandleUnsyncedListing returns all items with that are not synced
func (s *server) HandleRecentlyEditedListing(w http.ResponseWriter, r *http.Request) (err error) {
	return s.renderTemplate(w, r, "listing.html", listingPage{
		Title:  "Recently edited",
		Groups: s.m.RecentlyEdited(),
	})
}

type searchListing struct {
	Title string
	Songs []music.Entry

	Query string
}

// HandleSearchListing returns a search listing
func (s *server) HandleSearchListing(w http.ResponseWriter, r *http.Request) (err error) {
	query := r.URL.Query().Get("q")
	if query == "" {
		return fmt.Errorf("Empty query")
	}

	res := s.m.Search(query)
	if len(res) == 1 {
		http.Redirect(w, r, "/song/"+res[0].ID, http.StatusTemporaryRedirect)
		return
	}

	return s.renderTemplate(w, r, "search.html", searchListing{
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
func (s *server) HandleShowAlbum(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["artist"] == "" || v["album"] == "" {
		return httpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need an artist and album",
		}
	}

	al, ok := s.m.GetAlbum(v["artist"], v["album"])
	if !ok {
		return httpError{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("Cannot find album %s for artist %s", v["album"], v["artist"]),
		}
	}

	al.Title = al.Artist + " - " + al.Title
	return s.renderTemplate(w, r, "album.html", albumPage{
		Title: al.Title,
		A:     al,
	})
}

type artistPage struct {
	Title string

	Info store.ArtistInfo
}

// HandleShowArtist shows the artist page for the artist given in the url
func (s *server) HandleShowArtist(w http.ResponseWriter, r *http.Request) (err error) {
	v := mux.Vars(r)
	if v == nil || v["artist"] == "" {
		return httpError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "Need an artist",
		}
	}

	artistInfo, ok := s.m.Artist(v["artist"])
	if !ok {
		return httpError{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf("Cannot find any albums for %s", v["artist"]),
		}
	}

	return s.renderTemplate(w, r, "artist.html", artistPage{
		Title: artistInfo.Name,
		Info:  artistInfo,
	})
}
