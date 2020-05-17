package web

import (
	"fmt"
	"net/http"
	"xarantolus/sensiblehub/store"
	"xarantolus/sensiblehub/store/music"
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
