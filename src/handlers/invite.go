package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"text/template"

	"github.com/hadleyso/netid-activate/src/auth"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/mailer"
	"github.com/hadleyso/netid-activate/src/models"
	idm "github.com/hadleyso/netid-activate/src/redhat-idm"
	"github.com/hadleyso/netid-activate/src/scenes"
)

func InviteGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/invite/", http.StatusSeeOther)
}

func InviteLandingPage(w http.ResponseWriter, r *http.Request) {
	authenticated := auth.ValidateSession(w, r)
	if !authenticated {
		return
	}

	// Get affiliation
	raw := os.Getenv("AFFILIATION")
	var affiliationMap map[string]string
	if err := json.Unmarshal([]byte(raw), &affiliationMap); err != nil {
		log.Println("InviteLandingPage() unable to parse AFFILIATION env")
		http.Redirect(w, r, "/500", http.StatusSeeOther)
	}

	// Render template
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/invite-form.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base",
		struct {
			Affiliation map[string]string
			models.PageBase
		}{
			Affiliation: affiliationMap,
			PageBase: models.PageBase{
				PageTitle:  os.Getenv("SITE_NAME"),
				FaviconURL: os.Getenv("FAVICON_URL"),
				LogoURL:    os.Getenv("LOGO_URL"),
			},
		},
	)

}

func InviteSubmit(w http.ResponseWriter, r *http.Request) {

	authenticated := auth.ValidateSession(w, r)
	if !authenticated {
		return
	}

	r.ParseForm()
	firstName := strings.TrimSpace(r.Form.Get("firstName"))
	lastName := strings.TrimSpace(r.Form.Get("lastName"))
	email := strings.TrimSpace(r.Form.Get("email"))
	state := strings.TrimSpace(r.Form.Get("state"))
	country := strings.TrimSpace(r.Form.Get("country"))
	affiliation := strings.TrimSpace(r.Form.Get("affiliation"))

	// Check if email formatted correctly
	_, err := mail.ParseAddress(email)

	// Check filled out
	if firstName == "" || lastName == "" || email == "" || state == "" || country == "" || affiliation == "" || err != nil {
		tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/400.html", "scenes/base.html"))
		tmpl.ExecuteTemplate(w, "base",
			struct {
				Tile    string
				Message string
				models.PageBase
			}{
				Message: "Please complete the form fully",
				Tile:    "Invite Form",
				PageBase: models.PageBase{
					PageTitle:  os.Getenv("SITE_NAME"),
					FaviconURL: os.Getenv("FAVICON_URL"),
					LogoURL:    os.Getenv("LOGO_URL"),
				},
			},
		)
		return
	}

	// Check if already invited
	isInvited, _ := db.EmailValid(email)
	if isInvited {
		tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/400.html", "scenes/base.html"))
		tmpl.ExecuteTemplate(w, "base",
			struct {
				Tile    string
				Message string
				models.PageBase
			}{
				Message: "User already invited",
				Tile:    "Invite Form",
				PageBase: models.PageBase{
					PageTitle:  os.Getenv("SITE_NAME"),
					FaviconURL: os.Getenv("FAVICON_URL"),
					LogoURL:    os.Getenv("LOGO_URL"),
				},
			},
		)
		return
	}

	// Check if email in directory
	emailExists, err := idm.CheckEmailExists(email)
	if err != nil {
		http.Redirect(w, r, "/500?error=error+in+CheckEmailExists", http.StatusSeeOther)
		return
	}

	if emailExists {
		tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/400.html", "scenes/base.html"))
		tmpl.ExecuteTemplate(w, "base",
			struct {
				Tile    string
				Message string
				models.PageBase
			}{
				Message: "User already has an account",
				Tile:    "Invite Form",
				PageBase: models.PageBase{
					PageTitle:  os.Getenv("SITE_NAME"),
					FaviconURL: os.Getenv("FAVICON_URL"),
					LogoURL:    os.Getenv("LOGO_URL"),
				},
			},
		)
		return
	}

	// Add to DB
	dbSuccess, err := db.HandleInvite(firstName, lastName, email, state, country, affiliation)
	if dbSuccess == false {
		http.Redirect(w, r, "/500?error=DB+HandleInvite+error", http.StatusSeeOther)
		return
	}

	// Send email
	errMail := mailer.HandleSendInvite(email)
	if errMail != nil {
		http.Redirect(w, r, "/500?error=mail+HandleSendInvite+error", http.StatusSeeOther)
		return
	}

	successMessage := "Success, " + "an email has been sent to " + firstName + "'s " + email + " inbox."

	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/400.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base",
		struct {
			Tile    string
			Message string
			models.PageBase
		}{
			Message: successMessage,
			Tile:    "Invite Form",
			PageBase: models.PageBase{
				PageTitle:  os.Getenv("SITE_NAME"),
				FaviconURL: os.Getenv("FAVICON_URL"),
				LogoURL:    os.Getenv("LOGO_URL"),
			},
		},
	)
}
