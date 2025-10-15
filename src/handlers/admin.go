package handlers

import (
	"log"
	"net/http"
	"text/template"

	"github.com/hadleyso/netid-activate/src/auth"
	"github.com/hadleyso/netid-activate/src/db"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/hadleyso/netid-activate/src/scenes"
)

func GetSent(w http.ResponseWriter, r *http.Request) {
	// Get inviter
	user, errUser := auth.GetUser(w, r)
	if errUser != nil {
		http.Redirect(w, r, "/500?error=GetUser+error", http.StatusSeeOther)
		return
	}

	invites, err := db.GetUserSent(user.PreferredUsername)
	if err != nil {
		log.Println("GetUserSent() error in handler GetSent()")
		http.Redirect(w, r, "/500?error=GetUser+error", http.StatusSeeOther)
		return
	}

	tmpl := template.Must(template.ParseFS(scenes.TemplateFS, "scenes/invite-sent.html", "scenes/base.html"))
	tmpl.ExecuteTemplate(w, "base",
		struct {
			Invites []models.Invite
			models.PageBase
		}{
			Invites:  invites,
			PageBase: models.NewPageBase(""),
		},
	)
}

func DeleteInvite(w http.ResponseWriter, r *http.Request) {
	// Get requestor
	user, errUser := auth.GetUser(w, r)
	if errUser != nil {
		http.Redirect(w, r, "/500?error=GetUser+error", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	email := r.Form.Get("email")

	// Check owner
	invite, err := db.InviteDetailsEmail(email)
	if err != nil {
		http.Redirect(w, r, "/500?error=GetUser+error", http.StatusSeeOther)
		return
	}

	if invite.Inviter != user.PreferredUsername {
		http.Redirect(w, r, "/500?error=Not+your+invite", http.StatusSeeOther)
		return
	}

	db.DeleteInviteEmail(email)
	http.Redirect(w, r, "/invite/sent", http.StatusSeeOther)
}
