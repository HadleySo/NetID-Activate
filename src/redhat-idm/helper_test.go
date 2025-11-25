package idm

import (
	"net/http"
	"testing"
)

func TestLogin(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	_ = setupIDMTestServer(t, mux)

	client, err := newHTTPClient(false)
	if err != nil {
		t.Fatalf("newHTTPClient failed: %v", err)
	}

	err = login(client, "user", "pass")
	if err != nil {
		t.Errorf("login failed: %v", err)
	}
}

func TestLogin_Failure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	_ = setupIDMTestServer(t, mux)

	client, err := newHTTPClient(false)
	if err != nil {
		t.Fatalf("newHTTPClient failed: %v", err)
	}

	err = login(client, "user", "pass")
	if err == nil {
		t.Error("login should have failed but it succeeded")
	}
}
