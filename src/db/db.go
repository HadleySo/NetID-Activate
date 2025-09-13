package db

import (
	"log"
	"os"

	"github.com/hadleyso/netid-activate/src/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Connect to db and return connection object
func DbConnect() *gorm.DB {
	var DB_PATH string = os.Getenv("DB_PATH")

	db, err := gorm.Open(sqlite.Open(DB_PATH), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database")

	}
	return db
}

// Migrate dbs
func MigrateDb() error {
	db := DbConnect()

	defer func() {
		dbInstance, _ := db.DB()
		_ = dbInstance.Close()
	}()

	// Migrate the schema
	if err := db.AutoMigrate(&models.Invite{}, &models.OTP{}, &models.EmailRate{}); err != nil {
		return err
	}

	return nil
}
