package handlers

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/sessions"
	"github.com/hadleyso/netid-activate/src/common"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/mailer"
	"github.com/hadleyso/netid-activate/src/models"
	idm "github.com/hadleyso/netid-activate/src/redhat-idm"
	"github.com/hadleyso/netid-activate/src/scenes"
)

var SessionCookieStore *sessions.CookieStore = nil

func ActivateEmailGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Check email and send OTP code
// Show OTP code form
func ActivateEmailPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	activateEmail := r.Form.Get("activateEmail")

	// Email has value
	if activateEmail == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	// Check if invite exists
	isValid, err := db.EmailValid(activateEmail)
	if err != nil {
		log.Println("Call to EmailValid() in src/handlers/activate.go error: ")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Check if recently emailed
	canEmail := db.CanEmail(activateEmail)

	// Send email
	if isValid && canEmail {
		errorOTP := mailer.HandleSendOTP(activateEmail)
		if errorOTP != nil {
			log.Println("Call to SendOTP() in src/handlers/activate.go error")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
	}

	// Show email message (did not resend)
	emailNotResent := false
	if isValid && !canEmail {
		emailNotResent = true
	}

	// Show template regardless of email valid or not
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/activate-otp.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base",
		struct {
			ActivateEmail string
			EmailNotReset bool
			models.PageBase
		}{
			ActivateEmail: activateEmail,
			EmailNotReset: emailNotResent,
			PageBase: models.PageBase{
				PageTitle:  os.Getenv("SITE_NAME"),
				FaviconURL: os.Getenv("FAVICON_URL"),
				LogoURL:    os.Getenv("LOGO_URL"),
			},
		},
	)

}

// Validate OTP code
// Generate login name and login name select show form
func ActivateOTPPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	activateEmail := r.Form.Get("activateEmail")
	activateOTP := r.Form.Get("activateOTP")

	inviteID, isValid, err := db.EmailOTPValid(activateEmail, activateOTP)
	// OTP Error
	if err != nil {
		log.Println("Call to EmailOTPValid() in ActivateOTPPost() src/handlers/activate.go error")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// OTP invalid
	if isValid == false {
		tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/400.html", "scenes/base.html"))
		tmpl.ExecuteTemplate(w, "base",
			struct {
				Tile    string
				Message string
				models.PageBase
			}{
				Message: "Your OTP code has expired or is invalid",
				Tile:    "One Time Code ",
				PageBase: models.PageBase{
					PageTitle:  os.Getenv("SITE_NAME"),
					FaviconURL: os.Getenv("FAVICON_URL"),
					LogoURL:    os.Getenv("LOGO_URL"),
				},
			},
		)
		return
	}

	// Get invite details
	invite, err := db.InviteDetails(inviteID)
	if err != nil {
		log.Println("Call to InviteDetails() in ActivateOTPPost() src/handlers/activate.go error")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Generate login names and save
	usernameOptions, err := common.GetLoginOptions(invite)
	if err != nil {
		log.Println("Call to InviteDetails() in ActivateOTPPost() src/handlers/activate.go error")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	db.SetLoginNames(usernameOptions, inviteID)

	// Set auth cookie
	if SessionCookieStore == nil {
		SessionCookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	}
	session_IDCLAIM_ACTIVATION, _ := SessionCookieStore.Get(r, "IDCLAIM_ACTIVATION")
	session_IDCLAIM_ACTIVATION.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	session_IDCLAIM_ACTIVATION.Values["activateEmail"] = activateEmail
	session_IDCLAIM_ACTIVATION.Values["activateOTP"] = activateOTP
	session_IDCLAIM_ACTIVATION.Values["inviteID"] = inviteID
	session_IDCLAIM_ACTIVATION.Values["activating"] = true
	session_IDCLAIM_ACTIVATION.Save(r, w)

	// Invalidate OTP
	db.ClaimOTP(activateEmail, activateOTP)

	// Return username selection form
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/activate-username.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base",
		struct {
			models.PageBase
			LoginNames []string
			InviteID   string
		}{
			LoginNames: usernameOptions,
			InviteID:   inviteID,
			PageBase: models.PageBase{
				PageTitle:  os.Getenv("SITE_NAME"),
				FaviconURL: os.Getenv("FAVICON_URL"),
				LogoURL:    os.Getenv("LOGO_URL"),
			},
		},
	)

}

// Validate cookie
// Check not created
// Call creator
func CreateUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	inviteID := r.Form.Get("inviteID")
	loginName := r.Form.Get("loginname")

	// Get invite details
	invite, err := db.InviteDetails(inviteID)
	if err != nil {
		log.Println("Call to InviteDetails() in CreateUser() src/handlers/activate.go error")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Get cookie
	if SessionCookieStore == nil {
		SessionCookieStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	}
	session_IDCLAIM_ACTIVATION, _ := SessionCookieStore.Get(r, "IDCLAIM_ACTIVATION")

	// Check cookie
	cookieCheckOk := true

	if session_IDCLAIM_ACTIVATION.Values["activating"] != true {
		log.Println("CreateUser() src/handlers/activate.go IDCLAIM_ACTIVATION cookie failed activating ")
		cookieCheckOk = false
	}
	if session_IDCLAIM_ACTIVATION.Values["activateEmail"] != invite.Email {
		log.Println("CreateUser() src/handlers/activate.go IDCLAIM_ACTIVATION cookie failed activateEmail")
		cookieCheckOk = false
	}
	if session_IDCLAIM_ACTIVATION.Values["inviteID"] != invite.ID.String() {
		log.Println("CreateUser() src/handlers/activate.go IDCLAIM_ACTIVATION cookie failed inviteID")
		cookieCheckOk = false
	}

	// Give error on bad check
	if cookieCheckOk == false {
		tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/400.html", "scenes/base.html"))
		tmpl.ExecuteTemplate(w, "base",
			struct {
				Tile    string
				Message string
				models.PageBase
			}{
				Message: "There has been an error processing your request, either a security flag or other condition.",
				Tile:    "Activation",
				PageBase: models.PageBase{
					PageTitle:  os.Getenv("SITE_NAME"),
					FaviconURL: os.Getenv("FAVICON_URL"),
					LogoURL:    os.Getenv("LOGO_URL"),
				},
			},
		)
		return
	}

	// Check user doesn't exist (email)
	emailExists, err := idm.CheckEmailExists(invite.Email)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Check user doesn't exist (username)
	var usernameUsed bool
	if emailExists == false {
		readyNames, err := idm.CheckUsernamesExists([]string{loginName})
		if err != nil {
			http.Redirect(w, r, "/500", http.StatusSeeOther)
			return
		}
		usernameUsed = len(readyNames) == 0
	}

	if emailExists || usernameUsed {
		tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/400.html", "scenes/base.html"))
		tmpl.ExecuteTemplate(w, "base",
			struct {
				Tile    string
				Message string
				models.PageBase
			}{
				Message: "Your account already exists, please login at: " + os.Getenv("LOGIN_REDIRECT"),
				Tile:    "Activation",
				PageBase: models.PageBase{
					PageTitle:  os.Getenv("SITE_NAME"),
					FaviconURL: os.Getenv("FAVICON_URL"),
					LogoURL:    os.Getenv("LOGO_URL"),
				},
			},
		)
		return
	}

	// Call maker
	passwd, err := idm.HandleMakeUser(invite, loginName)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Delete invite
	db.DeleteInviteEmail(invite.Email)

	// Show success message
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/activate-success.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base",
		struct {
			FirstName     string
			Tenant        string
			LoginName     string
			Password      string
			LoginRedirect string
			models.PageBase
		}{
			FirstName:     invite.FirstName,
			Tenant:        os.Getenv("TENANT_NAME"),
			LoginName:     loginName,
			Password:      passwd,
			LoginRedirect: os.Getenv("LOGIN_REDIRECT"),
			PageBase: models.PageBase{
				PageTitle:  os.Getenv("SITE_NAME"),
				FaviconURL: os.Getenv("FAVICON_URL"),
				LogoURL:    os.Getenv("LOGO_URL"),
			},
		},
	)

}
