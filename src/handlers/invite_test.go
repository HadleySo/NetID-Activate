package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/hadleyso/netid-activate/src/config"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDBForInviteHandlers(t *testing.T) *gorm.DB {
	dbPath := "test_handlers_invite.db"
	viper.Set("DB_PATH", dbPath)
	viper.Set("SESSION_KEY", "a-very-secret-key-for-invite")

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

func TestInviteGet(t *testing.T) {
	req := httptest.NewRequest("GET", "/invite", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(InviteGet)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/invite/", rr.Header().Get("Location"))
}

func TestInviteLandingPage(t *testing.T) {
	affiliations := []any{
		map[string]any{"student": "Student"},
		map[string]any{"faculty": "Faculty"},
	}
	viper.Set("AFFILIATION", affiliations)
	config.C.OptionalGroups = map[string][]config.Group{
		"test-group": {{GroupName: "Test Group", MemberManager: false, RequiredGroup: "some-group"}},
	}
	defer func() { config.C.OptionalGroups = nil }()

	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"result": {"count": 0, "results": []}}`))
	})
	setupIDMTestServer(t, mux)

	req := newRequestWithSession(t, "GET", "/invite/", "", "")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(InviteLandingPage)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Student")
	assert.Contains(t, rr.Body.String(), "Faculty")
	assert.Contains(t, rr.Body.String(), "Test Group")
}

func TestInviteSubmit_Success(t *testing.T) {
	database := setupTestDBForInviteHandlers(t)
	viper.Set("DEV", "true")

	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {"count": 0}}`))
	})
	setupIDMTestServer(t, mux)

	config.C.OptionalGroups = map[string][]config.Group{}
	defer func() { config.C.OptionalGroups = nil }()

	form := url.Values{}
	form.Add("firstName", "Test")
	form.Add("lastName", "User")
	form.Add("email", "test@example.com")
	form.Add("state", "CA")
	form.Add("country", "USA")
	form.Add("affiliation", "student")

	req := newRequestWithSession(t, "POST", "/invite/", form.Encode(), "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(InviteSubmit)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Success, an email has been sent")

	var invite models.Invite
	result := database.Where("email = ?", "test@example.com").First(&invite)
	assert.NoError(t, result.Error)
	assert.Equal(t, "Test", invite.FirstName)
}

func TestInviteSubmit_AlreadyInvited(t *testing.T) {
	database := setupTestDBForInviteHandlers(t)
	database.Create(&models.Invite{Email: "test@example.com"})

	form := url.Values{}
	form.Add("firstName", "Test")
	form.Add("lastName", "User")
	form.Add("email", "test@example.com")
	form.Add("state", "CA")
	form.Add("country", "USA")
	form.Add("affiliation", "student")

	req := newRequestWithSession(t, "POST", "/invite/", form.Encode(), "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(InviteSubmit)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "User already invited")
}

func TestInviteSubmit_EmailExistsInIDM(t *testing.T) {
	setupTestDBForInviteHandlers(t)

	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {"count": 1}}`))
	})
	setupIDMTestServer(t, mux)

	config.C.OptionalGroups = map[string][]config.Group{}
	defer func() { config.C.OptionalGroups = nil }()

	form := url.Values{}
	form.Add("firstName", "Test")
	form.Add("lastName", "User")
	form.Add("email", "test@example.com")
	form.Add("state", "CA")
	form.Add("country", "USA")
	form.Add("affiliation", "student")

	req := newRequestWithSession(t, "POST", "/invite/", form.Encode(), "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(InviteSubmit)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "User already has an account")
}

func TestInviteSubmit_InvalidForm(t *testing.T) {
	req := newRequestWithSession(t, "POST", "/invite/", "", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(InviteSubmit)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Please complete the form fully")
}
