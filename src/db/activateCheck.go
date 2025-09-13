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
	otpEntry := models.OTP{Code: int(otpCode.Int64()), InviteID: inviteID.String()}
	db.Create(&otpEntry)

	return nil
}

// Check if email OTP combo valid
// Returns InviteID, isValid, err
func EmailOTPValid(activateEmail string, activateOTP string) (string, bool, error) {
	db := DbConnect()

	// Get invite
	var otpEntry models.OTP
	result := db.Where("Code = ?", activateOTP).First(&otpEntry)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", false, nil
	}

	if result.Error != nil {
		return "", false, result.Error
	}

	return otpEntry.InviteID, true, nil
}

// Get invite details
func InviteDetails(inviteID string) (models.Invite, error) {
	db := DbConnect()

	// Get invite
	var userInvite models.Invite
	result := db.Where("id = ?", inviteID).First(&userInvite)

	return userInvite, result.Error
}
