package web

import (
	"net/http"
	"xarantolus/sensiblehub/store"
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
