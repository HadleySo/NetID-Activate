package routes

import (
	"github.com/hadleyso/netid-activate/src/auth"
	"github.com/hadleyso/netid-activate/src/handlers"
)

func invite() {
	authedRouter := Router.PathPrefix("/invite").Subrouter()
	authedRouter.Use(auth.MiddleValidateSession)

	Router.HandleFunc("/invite", handlers.InviteGet).Methods("GET")
	authedRouter.HandleFunc("/", handlers.InviteLandingPage).Methods("GET")
	authedRouter.HandleFunc("/", handlers.InviteSubmit).Methods("POST")
	authedRouter.HandleFunc("/sent", handlers.GetSent).Methods("GET")
	authedRouter.HandleFunc("/sent/delete", handlers.DeleteInvite).Methods("POST")
}
