package db

import (
	"os"
	"testing"
	"time"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
)

func setupTestDBForMailer(t *testing.T) {
	dbPath := "test_mailer.db"
	viper.Set("DB_PATH", dbPath)

	db := DbConnect()
	err := db.AutoMigrate(&models.EmailRate{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	dbInstance, _ := db.DB()
	dbInstance.Close()

	t.Cleanup(func() {
		os.Remove(dbPath)
	})
}

func TestCanEmail(t *testing.T) {
	setupTestDBForMailer(t)

	email := "test@example.com"

	// First time, should be true
	if !CanEmail(email) {
		t.Errorf("CanEmail() should return true for the first time")
	}

	// Second time, immediately after, should be false
	if CanEmail(email) {
		t.Errorf("CanEmail() should return false when called immediately after")
	}

	// Update timestamp in DB to simulate time passing
	db := DbConnect()
	var mailerEntry models.EmailRate
	result := db.Where("Email = ?", email).First(&mailerEntry)
	if result.Error != nil {
		t.Fatalf("Could not find mailer entry: %v", result.Error)
	}

	mailerEntry.LastSend = time.Now().Add(-6 * time.Minute)
	db.Save(&mailerEntry)
	dbInstance, _ := db.DB()
	dbInstance.Close()

	// After 5 minutes, should be true again
	if !CanEmail(email) {
		t.Errorf("CanEmail() should return true after 5 minutes have passed")
	}

	// And check that it's now false again
	if CanEmail(email) {
		t.Errorf("CanEmail() should return false when called immediately after the successful 5-minute check")
	}
}
