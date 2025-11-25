package handlers

import (
	"encoding/gob"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDBForAdminHandlers(t *testing.T) *gorm.DB {
	dbPath := "test_handlers_admin.db"
	viper.Set("DB_PATH", dbPath)
	viper.Set("SESSION_KEY", "a-very-secret-key-for-admin")

	database := db.DbConnect()
	err := database.AutoMigrate(&models.Invite{}, &models.OTP{}, &models.EmailRate{})
	if err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	t.Cleanup(func() {
		dbInstance, _ := database.DB()
		dbInstance.Close()
		os.Remove(dbPath)
	})

	return database
}

// init is used to register types for gob encoding in sessions
func init() {
	gob.Register(&models.UserInfo{})
}

// Helper to create a request with a valid session cookie
func newRequestWithSession(t *testing.T, method, path string, body string, contentType string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	store := sessions.NewCookieStore([]byte(viper.GetString("SESSION_KEY")))

	session, _ := store.Get(req, "IDCLAIM_IDENTITY")
	user := &models.UserInfo{
		PreferredUsername: "testuser",
		Email:             "test@example.com",
		Groups:            []string{"some-group"},
	}
	session.Values["IDP"] = user

	rr := httptest.NewRecorder()
	err := session.Save(req, rr)
	if err != nil {
		t.Fatalf("Error saving session: %v", err)
	}

	for _, cookie := range rr.Result().Cookies() {
		req.AddCookie(cookie)
	}

	return req
}

func TestGetSent(t *testing.T) {
	database := setupTestDBForAdminHandlers(t)

	database.Create(&models.Invite{Inviter: "testuser", Email: "invite1@example.com", FirstName: "Invite", LastName: "One"})
	database.Create(&models.Invite{Inviter: "anotheruser", Email: "invite2@example.com", FirstName: "Invite", LastName: "Two"})

	req := newRequestWithSession(t, "GET", "/invite/sent", "", "")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(GetSent)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "invite1@example.com")
	assert.NotContains(t, rr.Body.String(), "invite2@example.com")
}

func TestDeleteInvite(t *testing.T) {
	database := setupTestDBForAdminHandlers(t)

	invite := models.Invite{Inviter: "testuser", Email: "delete@example.com"}
	database.Create(&invite)

	form := url.Values{}
	form.Add("email", "delete@example.com")

	req := newRequestWithSession(t, "POST", "/invite/sent/delete", form.Encode(), "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteInvite)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/invite/sent", rr.Header().Get("Location"))

	var count int64
	database.Model(&models.Invite{}).Where("email = ?", "delete@example.com").Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDeleteInvite_NotOwner(t *testing.T) {
	database := setupTestDBForAdminHandlers(t)

	invite := models.Invite{Inviter: "anotheruser", Email: "another@example.com"}
	database.Create(&invite)

	form := url.Values{}
	form.Add("email", "another@example.com")

	req := newRequestWithSession(t, "POST", "/invite/sent/delete", form.Encode(), "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(DeleteInvite)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Contains(t, rr.Header().Get("Location"), "/500?error=Not+your+invite")

	var count int64
	database.Model(&models.Invite{}).Where("email = ?", "another@example.com").Count(&count)
	assert.Equal(t, int64(1), count)
}
