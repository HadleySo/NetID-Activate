package idm

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/viper"
)

// helper to craft JSON-RPC responses
func writeJSONRPCResponse(w http.ResponseWriter, result any, errObj any) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"result": result,
	}
	if errObj != nil {
		resp["error"] = errObj
	}
	_ = json.NewEncoder(w).Encode(resp)
}

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

	err = login(client, "admin", "Secret123")
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

func TestFindUserByEmail_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRPCResponse(w, map[string]any{"user": "test@example.com"}, nil)
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, err := newHTTPClient(false)
	if err != nil {
		t.Fatalf("newHTTPClient failed: %v", err)
	}

	result, err := findUserByEmail(client, "test@example.com")
	if err != nil {
		t.Errorf("findUserByEmail failed: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestFindUserByEmail_RPCError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRPCResponse(w, nil, map[string]any{"message": "user not found"})
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, _ := newHTTPClient(false)

	_, err := findUserByEmail(client, "missing@example.com")
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestFindUserByLogin_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRPCResponse(w, map[string]any{"uid": "jdoe"}, nil)
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, _ := newHTTPClient(false)

	result, err := findUserByLogin(client, "jdoe")
	if err != nil {
		t.Errorf("findUserByLogin failed: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestFindUserByLogin_RPCError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		writeJSONRPCResponse(w, nil, map[string]any{"message": "invalid uid"})
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, _ := newHTTPClient(false)

	_, err := findUserByLogin(client, "baduser")
	if err == nil {
		t.Error("expected error but got nil")
	}
}
