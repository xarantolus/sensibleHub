package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"xarantolus/sensiblehub/store"
)

type addAccept struct {
	URL string `json:"url"`
}

//
func HandleDownloadSong(w http.ResponseWriter, r *http.Request) (err error) {

	// For AJAX requests
	if strings.ToUpper(r.URL.Query().Get("format")) == "JSON" {
		acc := new(addAccept)

		err = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(acc)
		if err != nil {
			return
		}

		err = store.M.Enqueue(acc.URL)
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

	err = store.M.Enqueue(r.FormValue("url"))
	if err != nil {
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return
}
