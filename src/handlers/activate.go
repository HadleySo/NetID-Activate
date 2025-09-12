package handlers

import (
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/otcode"
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

	// Send email
	if isValid {
		errorOTP := otcode.HandleSendOTP(activateEmail)
		if errorOTP != nil {
			log.Println("Call to SendOTP() in src/handlers/activate.go error")
			http.Redirect(w, r, "/500", http.StatusSeeOther)
		}
	}

	// Show template regardless of email valid or not
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/activate-otp.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base",
		struct {
			ActivateEmail string
			models.PageBase
		}{
			ActivateEmail: activateEmail,
			PageBase: models.PageBase{
				PageTitle:  os.Getenv("SITE_NAME"),
				FaviconURL: os.Getenv("FAVICON_URL"),
				LogoURL:    os.Getenv("LOGO_URL"),
			},
		},
	)

}
