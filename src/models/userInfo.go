package models

type UserInfo struct {
	Name              string   `json:"name"`
	Groups            []string `json:"groups"`
	PreferredUsername string   `json:"preferred_username"`
	GivenName         string   `json:"given_name"`
	FamilyName        string   `json:"family_name"`
	Email             string   `json:"email"`
}
