package routes

import (
	"github.com/hadleyso/netid-activate/src/handlers"
)

func landing() {
	Router.HandleFunc("/", handlers.Landing).Methods("GET")
}
