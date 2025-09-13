package idm

import (
	"log"
	"os"

	"github.com/hadleyso/netid-activate/src/models"
)

func MakeUser(invite models.Invite) (string, error) {

	// Create client
	client, errClient := newHTTPClient(false)
	if errClient != nil {
		log.Println("MakeUser() unable to newHTTPClient() " + errClient.Error())
		return "", errClient
	}

	// Auth
	username := os.Getenv("IDM_USERNAME")
	password := os.Getenv("IDM_PASSWORD")
	errLogin := login(client, username, password)
	if errLogin != nil {
		log.Println("MakeUser() unable to login() with HTTPClient " + errLogin.Error())
		return "", errLogin
	}

	// Create user

	// Add to groups

	return "", nil
}
