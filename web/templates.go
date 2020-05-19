package web

import (
	"html/template"
	"net/http"
	"strings"
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
		"have": func(s ...string) bool {
			for _, e := range s {
				if strings.TrimSpace(e) == "" {
					return false
				}
			}
			return true
		},
		"haveI": func(i ...int) bool {
			for _, e := range i {
				if e == 0 {
					return false
				}
			}
			return true
		},
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
