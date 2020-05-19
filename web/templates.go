package web

import (
	"html/template"
	"net/http"
	"xarantolus/sensiblehub/store"
)

var (
	funcMap = map[string]interface{}{
		"query": func(struc interface{}) string {
			v, ok := struc.(searchListing)
			if !ok {
				return ""
			}

			return v.Query
		},
		"count": func(i int) int {
			return i + 1
		},
		"clean": store.CleanName,
	}
	templates *template.Template
)

func parseTemplates() (err error) {
	temp, err := template.New("base").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		return
	}

	templates = temp
	return nil
}

func renderTemplate(w http.ResponseWriter, r *http.Request, tmplName string, p interface{}) error {
	w.Header().Set("Content-Type", "text/html")

	return templates.ExecuteTemplate(w, tmplName, p)
}
