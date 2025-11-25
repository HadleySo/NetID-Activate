package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func setupTestDBForActivateHandlers(t *testing.T) (*gorm.DB, models.Invite) {
	dbPath := "test_handlers_activate.db"
	viper.Set("DB_PATH", dbPath)
	viper.Set("SESSION_KEY", "a-very-secret-key-for-activate")
	viper.Set("DEV", "true")

	database := db.DbConnect()
	err := database.AutoMigrate(&models.Invite{}, &models.OTP{}, &models.EmailRate{})
	assert.NoError(t, err)

	groups := []string{"employees", "hpc_org_008bbc9505b0429cb20d531182a9cf7e"}
	jsonGroups, _ := json.Marshal(groups)

	invite := models.Invite{
		FirstName:      "Test",
		LastName:       "User",
		Email:          "test@example.com",
		Country:        "USA",
		Affiliation:    "Guest-Test-setupTestDBForActivateHandlers",
		State:          "GA",
		Inviter:        "admin",
		OptionalGroups: datatypes.JSON(jsonGroups),
	}
	database.Create(&invite)

	t.Cleanup(func() {
		dbInstance, _ := database.DB()
		dbInstance.Close()
		os.Remove(dbPath)
	})

	return database, invite
}

func setupIDMTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	server := httptest.NewServer(handler)
	viper.Set("IDM_HOST", server.URL)
	viper.Set("IDM_USERNAME", "testuser")
	viper.Set("IDM_PASSWORD", "testpass")
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

func TestActivateEmailGet(t *testing.T) {
	req := httptest.NewRequest("GET", "/activate", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ActivateEmailGet)
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/", rr.Header().Get("Location"))
}

func TestActivateEmailPost(t *testing.T) {
	database, invite := setupTestDBForActivateHandlers(t)

	form := url.Values{}
	form.Add("activateEmail", invite.Email)
	req := httptest.NewRequest("POST", "/activate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ActivateEmailPost)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "One Time Code")

	var otp models.OTP
	result := database.Where("invite_id = ?", invite.ID.String()).First(&otp)
	assert.NoError(t, result.Error)
}

func TestActivateOTPPost(t *testing.T) {
	database, invite := setupTestDBForActivateHandlers(t)
	otp := models.OTP{Code: 123456, InviteID: invite.ID.String()}
	database.Create(&otp)

	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {"count": 0}}`))
	})
	setupIDMTestServer(t, mux)

	form := url.Values{}
	form.Add("activateEmail", invite.Email)
	form.Add("activateOTP", "123456")
	req := httptest.NewRequest("POST", "/otp", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ActivateOTPPost)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Please select your login name from the options below")

	var count int64
	database.Model(&models.OTP{}).Where("code = ?", 123456).Count(&count)
	assert.Equal(t, int64(0), count)
}

func newRequestWithActivationSession(t *testing.T, invite models.Invite, loginName string) *http.Request {
	form := url.Values{}
	form.Add("inviteID", invite.ID.String())
	form.Add("loginname", loginName)

	req := httptest.NewRequest("POST", "/login-name-select", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if SessionCookieStore == nil {
		SessionCookieStore = sessions.NewCookieStore([]byte(viper.GetString("SESSION_KEY")))
	}

	session, _ := SessionCookieStore.Get(req, "IDCLAIM_ACTIVATION")
	session.Values["activating"] = true
	session.Values["activateEmail"] = invite.Email
	session.Values["inviteID"] = invite.ID.String()

	rr := httptest.NewRecorder()
	err := session.Save(req, rr)
	assert.NoError(t, err)

	for _, cookie := range rr.Result().Cookies() {
		req.AddCookie(cookie)
	}
	return req
}

func TestCreateUser(t *testing.T) {
	database, invite := setupTestDBForActivateHandlers(t)
	loginNames := []string{"testuser", "tuser"}
	loginNamesJSON, _ := json.Marshal(loginNames)
	invite.LoginNames = loginNamesJSON
	database.Save(&invite)

	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {"count": 0}}`))
	})
	setupIDMTestServer(t, mux)

	req := newRequestWithActivationSession(t, invite, "testuser")
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(CreateUser)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
	assert.Equal(t, "/success/"+invite.ID.String(), rr.Header().Get("Location"))

	var count int64
	database.Model(&models.Invite{}).Where("id = ?", invite.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestCreateSuccess(t *testing.T) {
	setupTestDBForActivateHandlers(t)

	inviteID := uuid.New().String()
	req := httptest.NewRequest("GET", "/success/"+inviteID, nil)
	req = mux.SetURLVars(req, map[string]string{"inviteID": inviteID})

	if SessionCookieStore == nil {
		SessionCookieStore = sessions.NewCookieStore([]byte(viper.GetString("SESSION_KEY")))
	}
	session, _ := SessionCookieStore.Get(req, "IDCLAIM_SUCCESS")
	flashData := SuccessData{
		FirstName: "Test",
		LoginName: "testuser",
		Password:  "ABC-DEF-GHI",
	}
	session.AddFlash(flashData, inviteID)
	rr := httptest.NewRecorder()
	err := session.Save(req, rr)
	assert.NoError(t, err)
	for _, cookie := range rr.Result().Cookies() {
		req.AddCookie(cookie)
	}

	rr = httptest.NewRecorder()
	handler := http.HandlerFunc(CreateSuccess)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "testuser")
	assert.Contains(t, rr.Body.String(), "ABC-DEF-GHI")
}
