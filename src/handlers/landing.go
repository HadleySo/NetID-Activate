package handlers

import (
	"net/http"
	"text/template"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/scenes"
)

func Landing(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/landing.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base", models.NewPageBase(""))
}
