package config

type Config struct {
	SessionKey          string              `mapstructure:"SESSION_KEY" yaml:"SESSION_KEY"`
	ServerPort          int                 `mapstructure:"SERVER_PORT" yaml:"SERVER_PORT"`
	ServerHostname      string              `mapstructure:"SERVER_HOSTNAME" yaml:"SERVER_HOSTNAME"`
	OIDCServerPort      int                 `mapstructure:"OIDC_SERVER_PORT" yaml:"OIDC_SERVER_PORT"`
	OIDCWellKnown       string              `mapstructure:"OIDC_WELL_KNOWN" yaml:"OIDC_WELL_KNOWN"`
	ClientID            string              `mapstructure:"CLIENT_ID" yaml:"CLIENT_ID"`
	ClientSecret        string              `mapstructure:"CLIENT_SECRET" yaml:"CLIENT_SECRET"`
	Scopes              string              `mapstructure:"SCOPES" yaml:"SCOPES"`
	DBPath              string              `mapstructure:"DB_PATH" yaml:"DB_PATH"`
	LogoURL             string              `mapstructure:"LOGO_URL" yaml:"LOGO_URL"`
	FaviconURL          string              `mapstructure:"FAVICON_URL" yaml:"FAVICON_URL"`
	SiteName            string              `mapstructure:"SITE_NAME" yaml:"SITE_NAME"`
	TenantName          string              `mapstructure:"TENANT_NAME" yaml:"TENANT_NAME"`
	Affiliation         []map[string]string `mapstructure:"AFFILIATION" yaml:"AFFILIATION"`
	LoginRedirect       string              `mapstructure:"LOGIN_REDIRECT" yaml:"LOGIN_REDIRECT"`
	LinkServiceProvider string              `mapstructure:"LINK_SERVICE_PROVIDER" yaml:"LINK_SERVICE_PROVIDER"`
	LinkPrivacyPolicy   string              `mapstructure:"LINK_PRIVACY_POLICY" yaml:"LINK_PRIVACY_POLICY"`
	EmailFrom           string              `mapstructure:"EMAIL_FROM" yaml:"EMAIL_FROM"`
	AWSRegion           string              `mapstructure:"AWS_REGION" yaml:"AWS_REGION"`
	AWSAccessKeyID      string              `mapstructure:"AWS_ACCESS_KEY_ID" yaml:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey  string              `mapstructure:"AWS_SECRET_ACCESS_KEY" yaml:"AWS_SECRET_ACCESS_KEY"`
	CACertPath          string              `mapstructure:"CACERT_PATH" yaml:"CACERT_PATH"`
	IDMHost             string              `mapstructure:"IDM_HOST" yaml:"IDM_HOST"`
	IDMUsername         string              `mapstructure:"IDM_USERNAME" yaml:"IDM_USERNAME"`
	IDMPassword         string              `mapstructure:"IDM_PASSWORD" yaml:"IDM_PASSWORD"`
	IDMAddGroup         string              `mapstructure:"IDM_ADD_GROUP" yaml:"IDM_ADD_GROUP"`
	OptionalGroups      map[string][]Group  `mapstructure:"OPTIONAL_GROUPS" yaml:"OPTIONAL_GROUPS"`
}

type Group struct {
	RequiredGroup string `mapstructure:"group_required" yaml:"group_required"`
	GroupName     string `mapstructure:"group_name" yaml:"group_name"`
	MemberManager bool   `mapstructure:"memberManager" yaml:"memberManager"`
}

var C Config
