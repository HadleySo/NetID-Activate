package db

import (
	"log"
	"net/mail"

	"github.com/hadleyso/netid-activate/src/models"
)

// Add user to invited table
func HandleInvite(firstName string, lastName string, email string, state string, country string, affiliation string, inviter string) (bool, error) {

	// Check if email formatted correctly
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false, nil
	}

	db := DbConnect()

	userInvite := models.Invite{FirstName: firstName, LastName: lastName, Email: email, State: state, Country: country, Affiliation: affiliation, Inviter: inviter}
	result := db.Create(&userInvite)

	if result.Error != nil {
		log.Println("Error in HandleInvite(): " + result.Error.Error())
		return false, result.Error
	}

	return true, nil

}

// Delete invite by email
func DeleteInviteEmail(email string) {
	db := DbConnect()
	db.Where("Email = ?", email).Delete(&models.Invite{})
}
