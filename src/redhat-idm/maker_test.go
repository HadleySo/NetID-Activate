package idm

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/hadleyso/netid-activate/src/models"
	"gorm.io/datatypes"
)

func TestRandPIN(t *testing.T) {
	pin := randPIN()
	match, _ := regexp.MatchString(`^[A-Z]{3}-[A-Z]{3}-[A-Z]{3}$`, pin)
	if !match {
		t.Errorf("PIN format is incorrect: got %s", pin)
	}
}

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

	pin, err := HandleMakeUser(invite, "testuser")
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
