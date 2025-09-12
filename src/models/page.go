package models

type PageBase struct {
	PageTitle  string
	FaviconURL string
	LogoURL    string
	UserInfo   *UserInfo
}
