package db

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDBForInvite(t *testing.T) *gorm.DB {
	dbPath := "test_invite.db"
	viper.Set("DB_PATH", dbPath)

	db := DbConnect()
	err := db.AutoMigrate(&models.Invite{})
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

func TestHandleInvite(t *testing.T) {
	db := setupTestDBForInvite(t)

	// Invalid email
	success, err := HandleInvite("John", "Doe", "invalid-email", "CA", "USA", "Student", "inviter1", []string{"group1"})
	assert.NoError(t, err)
	assert.False(t, success)

	// Valid invite
	optionalGroups := []string{"group1", "group2"}
	success, err = HandleInvite("Jane", "Doe", "jane.doe@example.com", "NY", "USA", "Faculty", "inviter2", optionalGroups)
	assert.NoError(t, err)
	assert.True(t, success)

	var invite models.Invite
	result := db.Where("Email = ?", "jane.doe@example.com").First(&invite)
	assert.NoError(t, result.Error)
	assert.Equal(t, "Jane", invite.FirstName)
	assert.Equal(t, "Doe", invite.LastName)
	assert.Equal(t, "NY", invite.State)
	assert.Equal(t, "USA", invite.Country)
	assert.Equal(t, "Faculty", invite.Affiliation)
	assert.Equal(t, "inviter2", invite.Inviter)

	var retrievedGroups []string
	err = json.Unmarshal(invite.OptionalGroups, &retrievedGroups)
	assert.NoError(t, err)
	assert.Equal(t, optionalGroups, retrievedGroups)
}

func TestDeleteInviteEmail(t *testing.T) {
	db := setupTestDBForInvite(t)
	invite := models.Invite{Email: "test@example.com"}
	db.Create(&invite)

	DeleteInviteEmail("test@example.com")

	var resultInvite models.Invite
	result := db.Where("Email = ?", "test@example.com").First(&resultInvite)
	assert.ErrorIs(t, result.Error, gorm.ErrRecordNotFound)
}

func TestGetUserSent(t *testing.T) {
	db := setupTestDBForInvite(t)

	// No invites sent
	invites, err := GetUserSent("inviter1")
	assert.NoError(t, err)
	assert.Empty(t, invites)

	// Invites sent
	db.Create(&models.Invite{Email: "test1@example.com", Inviter: "inviter1"})
	db.Create(&models.Invite{Email: "test2@example.com", Inviter: "inviter1"})
	db.Create(&models.Invite{Email: "test3@example.com", Inviter: "inviter2"})

	invites, err = GetUserSent("inviter1")
	assert.NoError(t, err)
	assert.Len(t, invites, 2)

	emails := []string{invites[0].Email, invites[1].Email}
	assert.Contains(t, emails, "test1@example.com")
	assert.Contains(t, emails, "test2@example.com")
}
