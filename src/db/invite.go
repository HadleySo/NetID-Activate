package db

import (
	"encoding/json"
	"errors"
	"log"
	"net/mail"

	"github.com/hadleyso/netid-activate/src/models"
	"gorm.io/gorm"
)

// Add user to invited table
func HandleInvite(firstName string, lastName string, email string, state string, country string, affiliation string, inviter string, optionalGroups []string) (bool, error) {

	// Check if email formatted correctly
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false, nil
	}

	db := DbConnect()

	optionalGroupsJson, err := json.Marshal(optionalGroups)
	if err != nil {
		return false, errors.New("Error marshalling OptionalGroups")
	}
	userInvite := models.Invite{FirstName: firstName, LastName: lastName, Email: email, State: state, Country: country, Affiliation: affiliation, Inviter: inviter, OptionalGroups: optionalGroupsJson}
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

func GetUserSent(uid string) ([]models.Invite, error) {
	db := DbConnect()

	var invites []models.Invite
	result := db.Where("Inviter = ?", uid).Find(&invites)

	// Has not been emailed
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return invites, nil
	}

	// DB error
	if result.Error != nil {
		log.Println("Error in GetUserSent(): " + result.Error.Error())
		return invites, result.Error
	}

	return invites, nil

}
