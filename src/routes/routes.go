package routes

import (
	"log"

	"github.com/gorilla/mux"
)

// Routing
var Router = mux.NewRouter()

func Main() {
	static()
	errorRoutes()
	authRoutes()
	landing()
	activate()
	log.Println("Routes registered [src/routes/routes]")
}
