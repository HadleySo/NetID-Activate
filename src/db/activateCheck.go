package db

import (
	"errors"
	"log"
	"math/big"
	"net/mail"

	"github.com/hadleyso/netid-activate/src/models"
	"gorm.io/gorm"
)

// Get invite by email if exists
// otherwise return nil
func EmailValid(email string) (bool, error) {

	// Check if email formatted correctly
	_, err := mail.ParseAddress(email)
	if err != nil {
		return false, nil
	}

	db := DbConnect()

	var userInvite models.Invite
	result := db.Where("email = ?", email).First(&userInvite)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if result.Error != nil {
		log.Println("Error in EmailValid(): " + result.Error.Error())
		return false, result.Error
	}

	return true, nil

}

// Write OTP code to invite
func SaveOTP(email string, otpCode *big.Int) error {
	db := DbConnect()

	// Get invite
	var userInvite models.Invite
	result := db.Where("email = ?", email).First(&userInvite)
	if result.Error != nil {
		return result.Error
	}
	inviteID := userInvite.ID

	// Write OTP code
	otpEntry := models.OTP{Code: *otpCode, Invite: inviteID}
	db.Create(&otpEntry)

	return nil
}
