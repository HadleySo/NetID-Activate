package routes

import (
	"github.com/hadleyso/netid-activate/src/handlers"
)

func activate() {
	Router.HandleFunc("/activate", handlers.ActivateEmailPost).Methods("POST")
	Router.HandleFunc("/activate", handlers.ActivateEmailGet).Methods("GET")
	Router.HandleFunc("/otp", handlers.ActivateOTPPost).Methods("POST")
	Router.HandleFunc("/login-name-select", handlers.CreateUser).Methods("POST")
}
