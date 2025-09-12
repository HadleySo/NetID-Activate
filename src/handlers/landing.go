package handlers

import (
	"net/http"
	"os"
	"text/template"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/scenes"
)

func Landing(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/landing.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base", models.PageBase{
		PageTitle:  os.Getenv("SITE_NAME"),
		FaviconURL: os.Getenv("FAVICON_URL"),
		LogoURL:    os.Getenv("LOGO_URL"),
	})
}
