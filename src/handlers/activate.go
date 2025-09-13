package handlers

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/hadleyso/netid-activate/src/common"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/mailer"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/scenes"
)

func ActivateEmailGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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
	}

	// Check if recently emailed
	canEmail := db.CanEmail(activateEmail)

	// Send email
	if isValid && canEmail {
		errorOTP := mailer.HandleSendOTP(activateEmail)
		if errorOTP != nil {
			log.Println("Call to SendOTP() in src/handlers/activate.go error")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
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

func ActivateOTPPost(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	activateEmail := r.Form.Get("activateEmail")
	activateOTP := r.Form.Get("activateOTP")

	inviteID, isValid, err := db.EmailOTPValid(activateEmail, activateOTP)
	// OTP Error
	if err != nil {
		log.Println("Call to EmailOTPValid() in ActivateOTPPost() src/handlers/activate.go error")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
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
				Tile:    " One Time Code ",
				PageBase: models.PageBase{
					PageTitle:  os.Getenv("SITE_NAME"),
					FaviconURL: os.Getenv("FAVICON_URL"),
					LogoURL:    os.Getenv("LOGO_URL"),
				},
			},
		)
	}

	// Get invite details
	invite, err := db.InviteDetails(inviteID)
	if err != nil {
		log.Println("Call to InviteDetails() in ActivateOTPPost() src/handlers/activate.go error")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
	}

	// Generate login names and save
	usernameOptions, err := common.GetLoginOptions(invite)
	if err != nil {
		log.Println("Call to InviteDetails() in ActivateOTPPost() src/handlers/activate.go error")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
	}
	db.SetLoginNames(usernameOptions, inviteID)

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
