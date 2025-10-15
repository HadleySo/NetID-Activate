package models

import "github.com/spf13/viper"

type PageBase struct {
	PageTitle  string
	FaviconURL string
	LogoURL    string
	UserInfo   *UserInfo
}

// Constructor for PageBase with default title, favicon, and logo
//
// Default title pass in empty string "", otherwise pass title
func NewPageBase(PageTitle string) PageBase {
	if PageTitle == "" {
		PageTitle = viper.GetString("SITE_NAME")
	}
	return PageBase{
		PageTitle:  PageTitle,
		FaviconURL: viper.GetString("FAVICON_URL"),
		LogoURL:    viper.GetString("LOGO_URL"),
	}
}
