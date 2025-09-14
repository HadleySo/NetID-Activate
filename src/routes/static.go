package routes

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"text/template"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/scenes"
)

var (
	Version   string
	GitCommit string
)

type Status struct {
	Version   string `json:"version"`
	GitCommit string `json:"hash"`
	Status    string `json:"status"`
}

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

func status() {
	Router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := Status{
			Version:   Version,
			GitCommit: GitCommit,
			Status:    "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})
}
