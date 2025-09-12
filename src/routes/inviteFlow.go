package routes

import (
	"github.com/hadleyso/netid-activate/src/handlers"
)

func invite() {
	Router.HandleFunc("/invite", handlers.InviteGet).Methods("GET")
	Router.HandleFunc("/invite/", handlers.InviteLandingPage).Methods("GET")
	Router.HandleFunc("/invite/", handlers.InviteSubmit).Methods("POST")
}
