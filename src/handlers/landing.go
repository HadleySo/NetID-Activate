package handlers

import (
	"net/http"
	"text/template"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/scenes"
	"github.com/spf13/viper"
)

func Landing(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/landing.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base", models.PageBase{
		PageTitle:  viper.GetString("SITE_NAME"),
		FaviconURL: viper.GetString("FAVICON_URL"),
		LogoURL:    viper.GetString("LOGO_URL"),
	})
}
