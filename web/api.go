package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"xarantolus/sensibleHub/store"

	"github.com/gorilla/mux"
)

// HandleAPISongSearch is for the search API
func (s *server) HandleAPISongSearch(w http.ResponseWriter, r *http.Request) (err error) {
	type shortResult struct {
		Title string `json:"title"`
		ID    string `json:"id"`
	}

	type apiSearchResult struct {
		Query string `json:"query"`

		Results []shortResult `json:"results"`
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		return fmt.Errorf("empty query")
	}

	limit := 5
	if i, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		limit = i
	}

	res := s.m.Search(query)
	if len(res) > limit {
		res = res[:limit]
	}

	// TODO:  Suggest random song (title = 🔀 Random Song, id=random)

	var searchResult = apiSearchResult{Query: query, Results: make([]shortResult, len(res))}
	for i, r := range res {
		searchResult.Results[i] = shortResult{
			Title: r.SongName(),
			ID:    r.ID,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(searchResult)
}

// HandleAPIListing shows an API listing given via an mux URL variable
func (s *server) HandleAPIListing(w http.ResponseWriter, r *http.Request) (err error) {
	var possibleListings = map[string]func() []store.Group{
		"title":          s.m.GroupByTitle,
		"artist":         s.m.GroupByArtist,
		"year":           s.m.GroupByYear,
		"incomplete":     s.m.Incomplete,
		"unsynced":       s.m.Unsynced,
		"recentlyedited": s.m.RecentlyEdited,
	}

	vars := mux.Vars(r)
	if vars == nil {
		return fmt.Errorf("expected listing type in URL")
	}

	listingType := vars["listing"]

	listFunc, ok := possibleListings[listingType]
	if !ok {
		return fmt.Errorf("invalid listing type %q", listingType)
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(listFunc())
}
