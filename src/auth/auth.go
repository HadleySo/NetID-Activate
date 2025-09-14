package auth

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/zitadel/oidc/v3/pkg/client/rp"
	"github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/scenes"
)

var (
	CallbackPath                             = "/auth/callback"
	SessionCookieStore *sessions.CookieStore = nil
)

// Generate HTTP error code and render login page to redirect
func UnauthorizedLogin(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/login.html"))
	http.SetCookie(w, &http.Cookie{Name: "FLASH_PATH", Value: r.RequestURI, Path: "/", MaxAge: 300})
	w.WriteHeader(http.StatusUnauthorized)
	tmpl.Execute(w, nil)
}

// Sets cookie with user data after pulling from OIDC
func MarshallUserInfo(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[*oidc.IDTokenClaims], state string, rp rp.RelyingParty, info *oidc.UserInfo) {
	if SessionCookieStore == nil {
		SessionCookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	}
	data, err := json.Marshal(info)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user models.UserInfo

	if err := json.Unmarshal(data, &user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session_IDCLAIM_IDENTITY, _ := SessionCookieStore.Get(r, "IDCLAIM_IDENTITY")
	session_IDCLAIM_IDENTITY.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	session_IDCLAIM_IDENTITY.Values["IDP"] = &user
	session_IDCLAIM_IDENTITY.Save(r, w)

	session_IDCLAIM_AUTH, _ := SessionCookieStore.Get(r, "IDCLAIM_AUTH")
	session_IDCLAIM_AUTH.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
	session_IDCLAIM_AUTH.Values["AUTHENTICATED"] = true
	session_IDCLAIM_AUTH.Save(r, w)

	FLASH_PATH, errCookie := r.Cookie("FLASH_PATH")
	if errCookie != nil {
		http.SetCookie(w, &http.Cookie{Name: "FLASH_PATH", Value: "", Path: "/", MaxAge: 0})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
	http.SetCookie(w, &http.Cookie{Name: "FLASH_PATH", Value: "", Path: "/", MaxAge: 0})
	http.Redirect(w, r, FLASH_PATH.Value, http.StatusSeeOther)

}

// Returns user data from existing session
func GetUser(w http.ResponseWriter, r *http.Request) (*models.UserInfo, error) {
	if SessionCookieStore == nil {
		SessionCookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	}

	// User data
	session_IDCLAIM_IDENTITY, _ := SessionCookieStore.Get(r, "IDCLAIM_IDENTITY")
	user, err := session_IDCLAIM_IDENTITY.Values["IDP"].(*models.UserInfo)
	if !err {
		http.Error(w, "Error getting user from session", http.StatusInternalServerError)
		return user, fmt.Errorf("Error getting user from session")
	}

	return user, nil
}

// Check if request has valid user session
func ValidateSession(w http.ResponseWriter, r *http.Request) bool {
	if SessionCookieStore == nil {
		SessionCookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	}

	session_IDCLAIM_AUTH, _ := SessionCookieStore.Get(r, "IDCLAIM_AUTH")

	if session_IDCLAIM_AUTH.Values["AUTHENTICATED"] != true {
		UnauthorizedLogin(w, r)
		return false
	}
	return true
}

// Check if request has valid session
func MiddleValidateSession(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if SessionCookieStore == nil {
			SessionCookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
		}

		session_HQ_AUTH, _ := SessionCookieStore.Get(r, "IDCLAIM_AUTH")

		if session_HQ_AUTH.Values["AUTHENTICATED"] != true {
			UnauthorizedLogin(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
