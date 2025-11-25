package idm

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"gorm.io/datatypes"
)

func TestRandPIN(t *testing.T) {
	pin := randPIN()
	match, _ := regexp.MatchString(`^[A-Z]{3}-[A-Z]{3}-[A-Z]{3}$`, pin)
	if !match {
		t.Errorf("PIN format is incorrect: got %s", pin)
	}
}

var testUser string = "testuser-" + uuid.New().String()[:8]

func TestHandleMakeUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		json.Unmarshal(body, &payload)
		method := payload["method"].(string)

		w.Header().Set("Content-Type", "application/json")
		if method == "user_add" {
			w.Write([]byte(`{"result": {"result": {}}}`)) // success
		} else if method == "batch" {
			w.Write([]byte(`{"result": {"count": 1, "results": [{}]}}`)) // success
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	_ = setupIDMTestServer(t, mux)

	groups := []string{"employees", "hpc_org_008bbc9505b0429cb20d531182a9cf7e"}
	jsonGroups, _ := json.Marshal(groups)

	invite := models.Invite{
		FirstName:      "Test",
		LastName:       "User",
		Email:          "test@example.com",
		Country:        "USA",
		Affiliation:    "staff",
		State:          "CA",
		Inviter:        "admin",
		OptionalGroups: datatypes.JSON(jsonGroups),
	}

	pin, err := HandleMakeUser(invite, testUser)
	if err != nil {
		t.Fatalf("HandleMakeUser failed: %v", err)
	}

	if pin == "" {
		t.Error("expected a PIN, but got empty string")
	}
	if !strings.Contains(pin, "-") {
		t.Errorf("expected PIN to contain '-', got %s", pin)
	}
}

func TestAddUserGroups_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		// Always return a successful batch response
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"result": {"count": 2, "results": [{}, {}]}}`))
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, err := newHTTPClient(false)
	if err != nil {
		t.Fatalf("newHTTPClient failed: %v", err)
	}

	err = addUserGroups(client, testUser, []string{"grp1", "grp2"})
	if err != nil {
		t.Errorf("addUserGroups failed: %v", err)
	}
}

func TestAddUserGroups_RPCError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		// Simulate RPC error object
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": {"message": "group not found"}}`))
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, _ := newHTTPClient(false)

	err := addUserGroups(client, testUser, []string{"missinggroup"})
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestAddUserGroups_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		// Return 500 to simulate server failure
		w.WriteHeader(http.StatusInternalServerError)
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, _ := newHTTPClient(false)

	err := addUserGroups(client, testUser, []string{"grp1"})
	if err == nil {
		t.Error("expected error due to HTTP 500 but got nil")
	}
}

func TestMakeUser_Success(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		// Simulate successful user_add
		writeJSONRPCResponse(w, map[string]any{"result": map[string]any{"uid": "jdoe"}}, nil)
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, _ := newHTTPClient(false)

	result, err := makeUser(client, "jdoe", "jdoe@example.com", "John", "Doe", "USA", "staff", "secret", "CA", "mgr001")
	if err != nil {
		t.Fatalf("makeUser failed: %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestMakeUser_RPCError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		// Simulate RPC error object
		writeJSONRPCResponse(w, nil, map[string]any{"message": "duplicate user"})
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, _ := newHTTPClient(false)

	_, err := makeUser(client, "jdoe", "jdoe@example.com", "John", "Doe", "USA", "staff", "secret", "CA", "mgr001")
	if err == nil {
		t.Error("expected error but got nil")
	}
}

func TestMakeUser_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		// Return 500 to simulate server failure
		w.WriteHeader(http.StatusInternalServerError)
	})
	ts := setupIDMTestServer(t, mux)
	viper.Set("IDM_HOST", ts.URL)

	client, _ := newHTTPClient(false)

	_, err := makeUser(client, "jdoe", "jdoe@example.com", "John", "Doe", "USA", "staff", "secret", "CA", "mgr001")
	if err == nil {
		t.Error("expected error due to HTTP 500 but got nil")
	}
}

func TestMakeUser_InvalidCountry(t *testing.T) {
	// Pass an invalid alpha-3 code to trigger country lookup error
	client, _ := newHTTPClient(false)

	_, err := makeUser(client, "jdoe", "jdoe@example.com", "John", "Doe", "XXX", "staff", "secret", "CA", "mgr001")
	if err == nil {
		t.Error("expected error due to invalid country code but got nil")
	}
}
