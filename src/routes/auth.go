package routes

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hadleyso/netid-activate/src/auth"
	"github.com/spf13/viper"
	"github.com/zitadel/oidc/v3/pkg/client/rp"

	httphelper "github.com/zitadel/oidc/v3/pkg/http"
)

var (
	rpMutex        sync.Mutex
	hqRelyingParty rp.RelyingParty
	lastInitErr    error
	lastInitTime   time.Time
	retryDelay     = 30 * time.Second
)

func getRelyingParty(ctx context.Context, cookieHandler *httphelper.CookieHandler) (rp.RelyingParty, error) {

	SERVER_HOSTNAME := viper.GetString("SERVER_HOSTNAME")

	// OpenID Connect Client
	clientID := viper.GetString("CLIENT_ID")
	clientSecret := viper.GetString("CLIENT_SECRET")
	issuer := viper.GetString("OIDC_WELL_KNOWN")
	scopes := strings.Split(viper.GetString("SCOPES"), " ")
	port := viper.GetString("OIDC_SERVER_PORT")

	// OIDC URIs
	var redirectURI string
	if port != "" {
		redirectURI = fmt.Sprintf("%s:%v%v", SERVER_HOSTNAME, port, auth.CallbackPath)
	} else {
		redirectURI = fmt.Sprintf("%s%v", SERVER_HOSTNAME, auth.CallbackPath)
	}

	// Set Relying Party settings
	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}

	rpMutex.Lock()
	defer rpMutex.Unlock()

	// Have a working RP, return it
	if hqRelyingParty != nil {
		return hqRelyingParty, nil
	}

	// If last attempt failed, only retry after cooldown
	if time.Since(lastInitTime) < retryDelay {
		return nil, lastInitErr
	}

	// Try again
	lastInitTime = time.Now()
	rp, err := rp.NewRelyingPartyOIDC(ctx, issuer, clientID, clientSecret, redirectURI, scopes, options...)
	if err != nil {
		lastInitErr = err
		return nil, err
	}

	// Success
	hqRelyingParty = rp
	lastInitErr = nil
	return hqRelyingParty, nil
}

func authRoutes() {

	cookieHandler := httphelper.NewCookieHandler([]byte(viper.GetString("SESSION_KEY")), []byte(viper.GetString("SESSION_KEY")), httphelper.WithUnsecure())

	state := func() string {
		return uuid.New().String()
	}

	// Register routes
	Router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		relyingParty, err := getRelyingParty(ctx, cookieHandler)
		if err != nil {
			http.Error(w, "IdP unavailable, try again later", http.StatusServiceUnavailable)
			return
		}

		rp.AuthURLHandler(state, relyingParty, rp.WithPromptURLParam("force"))(w, r)
	}).Methods("GET")

	Router.HandleFunc(auth.CallbackPath, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		relyingParty, err := getRelyingParty(ctx, cookieHandler)
		if err != nil {
			http.Error(w, "IdP unavailable, try again later", http.StatusServiceUnavailable)
			return
		}

		rp.CodeExchangeHandler(
			rp.UserinfoCallback(auth.MarshallUserInfo),
			relyingParty,
		)(w, r)
	}).Methods("GET", "POST")

	Router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {

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

		ctx := r.Context()
		relyingParty, err := getRelyingParty(ctx, cookieHandler)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, relyingParty.GetEndSessionEndpoint(), http.StatusSeeOther)

	}).Methods("GET")

}
