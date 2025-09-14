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
	status()
	authRoutes()
	landing()
	activate()
	invite()
	log.Println("Routes registered [src/routes/routes]")
}
