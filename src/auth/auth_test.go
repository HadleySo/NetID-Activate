package auth

import (
	"encoding/gob"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/spf13/viper"

	"github.com/hadleyso/netid-activate/src/models"
)

// helper: reset global store and set session key
func setupStore(t *testing.T) {
	gob.Register(&models.UserInfo{})
	t.Helper()
	viper.Set("SESSION_KEY", "test-session-key")
	SessionCookieStore = nil
}

// helper: create a request that already contains a saved session cookie for `name`
// `values` should be a map[any]any matching gorilla/sessions expectations
func requestWithSession(t *testing.T, name string, values map[any]any) (*http.Request, *httptest.ResponseRecorder) {
	t.Helper()

	// ensure store exists
	if SessionCookieStore == nil {
		SessionCookieStore = sessions.NewCookieStore([]byte(viper.GetString("SESSION_KEY")))
	}

	// initial request/recorder to create cookie
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	sess, err := SessionCookieStore.Get(req, name)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	// ensure Values map exists
	if sess.Values == nil {
		sess.Values = make(map[interface{}]interface{})
	}
	for k, v := range values {
		sess.Values[k] = v
	}
	if err := sess.Save(req, rr); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	// extract cookie header and create a new request that carries it
	cookie := rr.Result().Header.Get("Set-Cookie")
	req2 := httptest.NewRequest("GET", "/", nil)
	if cookie != "" {
		req2.Header.Set("Cookie", cookie)
	}

	rr2 := httptest.NewRecorder()
	return req2, rr2
}

func TestUnauthorizedLogin_SetsFlashAndStatus(t *testing.T) {
	setupStore(t)

	req := httptest.NewRequest("GET", "/protected?foo=bar", nil)
	// set RequestURI to simulate original path
	req.RequestURI = "/protected?foo=bar"
	rr := httptest.NewRecorder()

	UnauthorizedLogin(rr, req)

	res := rr.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status %d got %d", http.StatusUnauthorized, res.StatusCode)
	}

	cookie := res.Header.Get("Set-Cookie")
	if !strings.Contains(cookie, "FLASH_PATH") {
		t.Fatalf("expected FLASH_PATH cookie to be set, got header: %q", cookie)
	}
}

func TestValidateSession_Unauthenticated(t *testing.T) {
	setupStore(t)

	// no session cookie present
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	ok := ValidateSession(rr, req)
	if ok {
		t.Fatalf("expected ValidateSession to return false for unauthenticated request")
	}

	// should have set unauthorized status via UnauthorizedLogin
	if rr.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected UnauthorizedLogin to set status %d, got %d", http.StatusUnauthorized, rr.Result().StatusCode)
	}
}

func TestValidateSession_Authenticated(t *testing.T) {
	setupStore(t)

	// create request with authenticated session
	req, rr := requestWithSession(t, "IDCLAIM_AUTH", map[interface{}]interface{}{
		"AUTHENTICATED": true,
	})

	ok := ValidateSession(rr, req)
	if !ok {
		t.Fatalf("expected ValidateSession to return true for authenticated request")
	}
}

func TestGetUser_SuccessAndFailure(t *testing.T) {
	setupStore(t)

	// success: store a *models.UserInfo pointer in session
	user := &models.UserInfo{}
	req, rr := requestWithSession(t, "IDCLAIM_IDENTITY", map[interface{}]interface{}{
		"IDP": user,
	})

	got, err := GetUser(rr, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil user")
	}

	// failure: store wrong type in session
	req2, rr2 := requestWithSession(t, "IDCLAIM_IDENTITY", map[interface{}]interface{}{
		"IDP": "not-a-user",
	})

	got2, err2 := GetUser(rr2, req2)
	if err2 == nil {
		t.Fatalf("expected error when IDP is wrong type, got nil and user=%v", got2)
	}
}

func TestMiddleValidateSession_Behavior(t *testing.T) {
	setupStore(t)

	// next handler that marks it was called
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	// case 1: unauthenticated -> should not call next and should set unauthorized
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler := MiddleValidateSession(next)
	handler.ServeHTTP(rr, req)

	if called {
		t.Fatalf("expected next not to be called for unauthenticated request")
	}
	if rr.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected UnauthorizedLogin to set status %d, got %d", http.StatusUnauthorized, rr.Result().StatusCode)
	}

	// case 2: authenticated -> should call next
	called = false
	req2, rr2 := requestWithSession(t, "IDCLAIM_AUTH", map[interface{}]interface{}{
		"AUTHENTICATED": true,
	})

	handler.ServeHTTP(rr2, req2)
	if !called {
		t.Fatalf("expected next to be called for authenticated request")
	}
	if rr2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected next to set status %d, got %d", http.StatusOK, rr2.Result().StatusCode)
	}
}
