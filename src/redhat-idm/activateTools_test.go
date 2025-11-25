package idm

import (
	"net/http"
	"strings"
	"testing"
)

func TestCheckUsernamesExists_LoginFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	_ = setupIDMTestServer(t, mux)

	_, err := CheckUsernamesExists([]string{"user1"})
	if err == nil {
		t.Fatal("expected an error for login failure, but got nil")
	}
	if !strings.Contains(err.Error(), "RedHat IdM login failed") {
		t.Errorf("expected error to contain 'RedHat IdM login failed', but got: %s", err.Error())
	}
}
