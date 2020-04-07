package web

import (
	"html/template"
	"net/http"
)

var (
	funcMap   = map[string]interface{}{}
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
