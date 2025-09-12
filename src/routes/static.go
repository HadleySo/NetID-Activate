package routes

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func static() {
	staticContent, _ := fs.Sub(staticFiles, "static")
	fs := http.FileServer(http.FS(staticContent))
	Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs)).Methods("GET")

}
