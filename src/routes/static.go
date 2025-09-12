package routes

import (
	"embed"
	"io/fs"
	"net/http"
	"text/template"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/scenes"
)

//go:embed static/*
var staticFiles embed.FS

func static() {
	staticContent, _ := fs.Sub(staticFiles, "static")
	fs := http.FileServer(http.FS(staticContent))
	Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs)).Methods("GET")

}

func errorRoutes() {

	Router.HandleFunc("/500",
		func(w http.ResponseWriter, r *http.Request) {
			tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/500.html", "scenes/base.html"))
			tmpl.ExecuteTemplate(w, "base", models.PageBase{PageTitle: "Error", FaviconURL: "", LogoURL: ""})
		},
	).Methods("GET")

}
