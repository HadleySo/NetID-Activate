package routes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hadleyso/netid-activate/src/auth"
	"github.com/spf13/viper"
	"github.com/zitadel/logging"
	"github.com/zitadel/oidc/v3/pkg/client/rp"

	httphelper "github.com/zitadel/oidc/v3/pkg/http"
)

var hqRelyingParty rp.RelyingParty

func authRoutes() {

	// App config
	SERVER_HOSTNAME := viper.GetString("SERVER_HOSTNAME")

	// OpenID Connect Client
	clientID := viper.GetString("CLIENT_ID")
	clientSecret := viper.GetString("CLIENT_SECRET")
	issuer := viper.GetString("OIDC_WELL_KNOWN")
	port := viper.GetString("OIDC_SERVER_PORT")
	scopes := strings.Split(viper.GetString("SCOPES"), " ")

	// OIDC URIs
	var redirectURI string
	if port != "" {
		redirectURI = fmt.Sprintf("%s:%v%v", SERVER_HOSTNAME, port, auth.CallbackPath)
	} else {
		redirectURI = fmt.Sprintf("%s%v", SERVER_HOSTNAME, auth.CallbackPath)
	}

	cookieHandler := httphelper.NewCookieHandler([]byte(viper.GetString("SESSION_KEY")), []byte(viper.GetString("SESSION_KEY")), httphelper.WithUnsecure())

	// Set Relying Party settings
	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}

	// Set logging
	logger := slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelInfo,
		}),
	)

	state := func() string {
		return uuid.New().String()
	}

	// OIDC RelyingParty Create
	ctx := logging.ToContext(context.TODO(), logger)
	RelyingParty, err := rp.NewRelyingPartyOIDC(ctx, issuer, clientID, clientSecret, redirectURI, scopes, options...)
	if err != nil {
		fmt.Printf("error creating provider %s", err.Error()) // TODO: add logging
	}
	hqRelyingParty = RelyingParty

	// Register routes
	Router.HandleFunc("/login", rp.AuthURLHandler(state, hqRelyingParty, rp.WithPromptURLParam("Welcome back!"))).Methods("GET")
	Router.HandleFunc(auth.CallbackPath, rp.CodeExchangeHandler(rp.UserinfoCallback(auth.MarshallUserInfo), hqRelyingParty)).Methods("GET", "POST")

	Router.HandleFunc("/logout", handleLogout).Methods("GET")

}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		http.SetCookie(w, &http.Cookie{
			Name:     cookie.Name,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			HttpOnly: true,
		})
	}

	http.Redirect(w, r, hqRelyingParty.GetEndSessionEndpoint(), http.StatusSeeOther)

}
