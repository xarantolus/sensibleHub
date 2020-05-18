package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"xarantolus/sensiblehub/store"
)

type addAccept struct {
	SearchTerm string `json:"searchTerm"`
}

// HandleDownloadSong handles a song download request. This kind of request is done
// from the /add page, either using AJAX (with ?format=json) or a normal form submit
func HandleDownloadSong(w http.ResponseWriter, r *http.Request) (err error) {
	// For AJAX requests
	if strings.ToUpper(r.URL.Query().Get("format")) == "JSON" {
		acc := new(addAccept)

		err = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(acc)
		if err != nil {
			return
		}

		err = store.M.Enqueue(acc.SearchTerm)
		if err == nil {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
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

	err = store.M.Enqueue(r.FormValue("searchTerm"))
	if err != nil {
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return
}
