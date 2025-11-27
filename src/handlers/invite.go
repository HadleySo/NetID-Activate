package handlers

import (
	"log"
	"net/http"
	"net/mail"
	"strings"
	"text/template"

	"github.com/hadleyso/netid-activate/src/attribute"
	"github.com/hadleyso/netid-activate/src/auth"
	"github.com/hadleyso/netid-activate/src/countries"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/mailer"
	"github.com/hadleyso/netid-activate/src/models"
	idm "github.com/hadleyso/netid-activate/src/redhat-idm"
	"github.com/hadleyso/netid-activate/src/scenes"
	"github.com/spf13/viper"
)

func InviteGet(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/invite/", http.StatusSeeOther)
}

func InviteLandingPage(w http.ResponseWriter, r *http.Request) {
	// Get affiliation
	affiliationsRaw := viper.Get("AFFILIATION")
	affiliationList, ok := affiliationsRaw.([]any)
	if !ok {
		log.Printf("InviteLandingPage() Expected affiliation to be a slice, but got %T", affiliationsRaw)
		http.Redirect(w, r, "/500", http.StatusSeeOther)
	}

	affiliationMap := make(map[string]string)
	for _, item := range affiliationList {
		// Type assert each item to a map
		affiliation, ok := item.(map[string]any)
		if ok {
			// Convert key-value pairs to string
			for key, value := range affiliation {
				// Convert to string
				if strValue, ok := value.(string); ok {
					affiliationMap[strings.ToUpper(key)] = strValue
				}
			}
		}

	}

	// Get inviter
	user, errUser := auth.GetUser(w, r)
	if errUser != nil {
		http.Redirect(w, r, "/500?error=GetUser+error", http.StatusSeeOther)
		return
	}

	// Optional Group
	rawGroups, err := attribute.GetOptionalGroupLimited(user)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}
	optionalGroup := make(map[string]string)
	for _, group := range rawGroups {
		optionalGroup[group.CN] = group.GroupName
	}

	// Render template
	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/invite-form.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base",
		struct {
			Affiliation   map[string]string
			OptionalGroup map[string]string
			Countries     []countries.Country
			models.PageBase
		}{
			Affiliation:   affiliationMap,
			OptionalGroup: optionalGroup,
			Countries:     countries.Countries,
			PageBase:      models.NewPageBase(""),
		},
	)

}

func InviteSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	firstName := strings.TrimSpace(r.Form.Get("firstName"))
	lastName := strings.TrimSpace(r.Form.Get("lastName"))
	email := strings.ToLower(strings.TrimSpace(r.Form.Get("email")))
	state := strings.TrimSpace(r.Form.Get("state"))
	country := strings.TrimSpace(r.Form.Get("country"))
	affiliation := strings.TrimSpace(r.Form.Get("affiliation"))

	// Check if email formatted correctly
	_, err := mail.ParseAddress(email)

	// Check filled out
	if firstName == "" || lastName == "" || email == "" || state == "" || country == "" || affiliation == "" || err != nil || countries.Alpha3Exists(country) == false {
		tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/400.html", "scenes/base.html"))
		tmpl.ExecuteTemplate(w, "base",
			struct {
				Tile    string
				Message string
				models.PageBase
			}{
				Message:  "Please complete the form fully",
				Tile:     "Invite Form",
				PageBase: models.NewPageBase(""),
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
				Message:  "User already invited",
				Tile:     "Invite Form",
				PageBase: models.NewPageBase(""),
			},
		)
		return
	}

	// Get inviter
	user, errUser := auth.GetUser(w, r)
	if errUser != nil {
		http.Redirect(w, r, "/500?error=GetUser+error", http.StatusSeeOther)
		return
	}

	// Optional Group
	rawGroups, err := attribute.GetOptionalGroupLimited(user)
	if err != nil {
		http.Redirect(w, r, "/500", http.StatusSeeOther)
		return
	}

	// Get selected form Optional Groups
	optionalGroups := []string{}
	for _, group := range rawGroups {
		if r.Form.Get(group.CN) == "yes" {
			optionalGroups = append(optionalGroups, group.CN)
		}
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
				Message:  "User already has an account",
				Tile:     "Invite Form",
				PageBase: models.NewPageBase(""),
			},
		)
		return
	}

	// Get inviter info
	inviter := user.PreferredUsername

	// Add to DB
	dbSuccess, err := db.HandleInvite(firstName, lastName, email, state, country, affiliation, inviter, optionalGroups)
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
			Message:  successMessage,
			Tile:     "Invite Form",
			PageBase: models.NewPageBase(""),
		},
	)
}
