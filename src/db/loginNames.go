package db

import (
	"encoding/json"
	"log"
	"slices"

	"github.com/hadleyso/netid-activate/src/models"
)

// Sets login name options
// otherwise return nil
func SetLoginNames(logins []string, inviteID string) error {

	db := DbConnect()

	var userInvite models.Invite
	result := db.Where("ID = ?", inviteID).First(&userInvite)

	if result.Error != nil {
		log.Println("Error in SetLoginNames(): " + result.Error.Error())
		return result.Error
	}

	userInvite.LoginNames, _ = json.Marshal(logins)
	db.Save(userInvite)

	return nil

}

// Check if login name is in saved login names
func CheckLoginNames(inviteID string, loginName string) (bool, error) {
	db := DbConnect()

	// Get from db and check err
	var userInvite models.Invite
	result := db.Where("ID = ?", inviteID).First(&userInvite)
	if result.Error != nil {
		log.Println("Error in CheckLoginNames(): " + result.Error.Error())
		return false, result.Error
	}

	// Unmarshall
	var names []string
	if err := json.Unmarshal(userInvite.LoginNames, &names); err != nil {
		return false, err
	}

	if slices.Contains(names, loginName) {
		return true, nil
	}

	return false, nil
}
