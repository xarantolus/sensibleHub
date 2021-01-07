package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"xarantolus/sensibleHub/store"
)

type apiSearchResult struct {
	Query string `json:"query"`

	Results []shortResult `json:"results"`
}

type shortResult struct {
	Title string `json:"title"`
	ID    string `json:"id"`
}

// HandleAPISongSearch is for the search API
func HandleAPISongSearch(w http.ResponseWriter, r *http.Request) (err error) {
	query := r.URL.Query().Get("q")
	if query == "" {
		return fmt.Errorf("Empty query")
	}

	limit := 5
	if i, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil {
		limit = i
	}

	res := store.M.Search(query)
	if len(res) > limit {
		res = res[:limit]
	}

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
