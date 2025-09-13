package db

import (
	"encoding/json"
	"log"

	"github.com/hadleyso/netid-activate/src/models"
)

// Sets login name options
// otherwise return nil
func SetLoginNames(logins []string, inviteID string) error {

	db := DbConnect()

	var userInvite models.Invite
	result := db.Where("ID = ?", inviteID).First(&userInvite)

	if result.Error != nil {
		log.Println("Error in EmailValid(): " + result.Error.Error())
		return result.Error
	}

	userInvite.LoginNames, _ = json.Marshal(logins)
	db.Save(userInvite)

	return nil

}
