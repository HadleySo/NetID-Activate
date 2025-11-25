package db

import (
	"math/big"
	"os"
	"testing"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDBForActivateCheck(t *testing.T) *gorm.DB {
	dbPath := "test_activate_check.db"
	viper.Set("DB_PATH", dbPath)

	db := DbConnect()
	err := db.AutoMigrate(&models.Invite{}, &models.OTP{}, &models.EmailRate{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	t.Cleanup(func() {
		dbInstance, _ := db.DB()
		dbInstance.Close()
		os.Remove(dbPath)
	})

	return db
}

func TestEmailValid(t *testing.T) {
	db := setupTestDBForActivateCheck(t)

	// Invalid email format
	valid, err := EmailValid("invalid-email")
	assert.NoError(t, err)
	assert.False(t, valid)

	// Email not in DB
	valid, err = EmailValid("test@example.com")
	assert.NoError(t, err)
	assert.False(t, valid)

	// Email in DB
	db.Create(&models.Invite{Email: "test@example.com"})
	valid, err = EmailValid("test@example.com")
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestSaveOTP(t *testing.T) {
	db := setupTestDBForActivateCheck(t)
	invite := models.Invite{Email: "test@example.com"}
	db.Create(&invite)

	otpCode := big.NewInt(123456)
	err := SaveOTP("test@example.com", otpCode)
	assert.NoError(t, err)

	var otpEntry models.OTP
	result := db.Where("invite_id = ?", invite.ID.String()).First(&otpEntry)
	assert.NoError(t, result.Error)
	assert.Equal(t, int(otpCode.Int64()), otpEntry.Code)
}

func TestEmailOTPValid(t *testing.T) {
	db := setupTestDBForActivateCheck(t)
	invite := models.Invite{Email: "test@example.com"}
	db.Create(&invite)
	otpEntry := models.OTP{Code: 123456, InviteID: invite.ID.String()}
	db.Create(&otpEntry)

	// Invalid OTP
	inviteID, valid, err := EmailOTPValid("test@example.com", "654321")
	assert.NoError(t, err)
	assert.False(t, valid)
	assert.Equal(t, "", inviteID)

	// Valid OTP
	inviteID, valid, err = EmailOTPValid("test@example.com", "123456")
	assert.NoError(t, err)
	assert.True(t, valid)
	assert.Equal(t, invite.ID.String(), inviteID)
}

func TestClaimOTP(t *testing.T) {
	db := setupTestDBForActivateCheck(t)
	otpEntry := models.OTP{Code: 123456}
	db.Create(&otpEntry)
	emailRateEntry := models.EmailRate{Email: "test@example.com"}
	db.Create(&emailRateEntry)

	ClaimOTP("test@example.com", "123456")

	var otp models.OTP
	result := db.Where("Code = ?", "123456").First(&otp)
	assert.ErrorIs(t, result.Error, gorm.ErrRecordNotFound)

	var emailRate models.EmailRate
	result = db.Where("Email = ?", "test@example.com").First(&emailRate)
	assert.ErrorIs(t, result.Error, gorm.ErrRecordNotFound)
}

func TestInviteDetails(t *testing.T) {
	db := setupTestDBForActivateCheck(t)
	invite := models.Invite{FirstName: "John", LastName: "Doe", Email: "john.doe@example.com"}
	db.Create(&invite)

	retrievedInvite, err := InviteDetails(invite.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, "John", retrievedInvite.FirstName)
	assert.Equal(t, "Doe", retrievedInvite.LastName)
	assert.Equal(t, "john.doe@example.com", retrievedInvite.Email)
}

func TestInviteDetailsEmail(t *testing.T) {
	db := setupTestDBForActivateCheck(t)
	invite := models.Invite{FirstName: "Jane", LastName: "Doe", Email: "jane.doe@example.com"}
	db.Create(&invite)

	retrievedInvite, err := InviteDetailsEmail("jane.doe@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "Jane", retrievedInvite.FirstName)
	assert.Equal(t, "Doe", retrievedInvite.LastName)
	assert.Equal(t, "jane.doe@example.com", retrievedInvite.Email)
}
