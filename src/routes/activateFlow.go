package routes

import (
	"github.com/hadleyso/netid-activate/src/handlers"
)

func activate() {
	Router.HandleFunc("/activate", handlers.ActivateEmailPost).Methods("POST")
	Router.HandleFunc("/activate", handlers.ActivateEmailGet).Methods("GET")
}
