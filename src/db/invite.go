package db

import (
	"log"
	"net/mail"

	"github.com/hadleyso/netid-activate/src/models"
)

// Add user to invited table
func HandleInvite(FirstName string, LastName string, Email string, State string, Country string, Affiliation string) (bool, error) {

	// Check if email formatted correctly
	_, err := mail.ParseAddress(Email)
	if err != nil {
		return false, nil
	}

	db := DbConnect()

	userInvite := models.Invite{FirstName: FirstName, LastName: LastName, Email: Email, State: State, Country: Country, Affiliation: Affiliation}
	result := db.Create(&userInvite)

	if result.Error != nil {
		log.Println("Error in HandleInvite(): " + result.Error.Error())
		return false, result.Error
	}

	return true, nil

}
