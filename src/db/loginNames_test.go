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

func setupTestDBForLoginNames(t *testing.T) *gorm.DB {
	dbPath := "test_login_names.db"
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

func TestSetLoginNames(t *testing.T) {
	db := setupTestDBForLoginNames(t)
	invite := models.Invite{Email: "test@example.com"}
	db.Create(&invite)

	loginNames := []string{"jdoe", "johndoe"}
	err := SetLoginNames(loginNames, invite.ID.String())
	assert.NoError(t, err)

	var updatedInvite models.Invite
	db.First(&updatedInvite, "id = ?", invite.ID)

	var retrievedNames []string
	err = json.Unmarshal(updatedInvite.LoginNames, &retrievedNames)
	assert.NoError(t, err)
	assert.Equal(t, loginNames, retrievedNames)
}

func TestCheckLoginNames(t *testing.T) {
	db := setupTestDBForLoginNames(t)
	loginNames := []string{"jdoe", "johndoe"}
	loginNamesJSON, _ := json.Marshal(loginNames)
	invite := models.Invite{Email: "test@example.com", LoginNames: loginNamesJSON}
	db.Create(&invite)

	// Name exists
	exists, err := CheckLoginNames(invite.ID.String(), "jdoe")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Name does not exist
	exists, err = CheckLoginNames(invite.ID.String(), "janedoe")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Error case - invalid invite ID
	exists, err = CheckLoginNames("invalid-uuid", "jdoe")
	assert.Error(t, err)
	assert.False(t, exists)
}
