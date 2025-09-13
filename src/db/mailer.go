package db

import (
	"errors"
	"log"
	"time"

	"github.com/hadleyso/netid-activate/src/models"
	"gorm.io/gorm"
)

func CanEmail(email string) bool {
	db := DbConnect()

	var mailerEntry models.EmailRate
	result := db.Where("Email = ?", email).First(&mailerEntry)

	// Has not been emailed
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		mailerEntry.Email = email
		mailerEntry.LastSend = time.Now()
		db.Create(&mailerEntry)
		return true
	}

	// DB error
	if result.Error != nil {
		log.Println("Error in CanEmail(): " + result.Error.Error())
		return false
	}

	// Emailed less than 5 minutes ago
	if time.Since(mailerEntry.LastSend) < 5*time.Minute {
		return false
	}

	mailerEntry.LastSend = time.Now()
	db.Save(&mailerEntry)
	return true
}
