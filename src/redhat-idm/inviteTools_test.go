package idm

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/hadleyso/netid-activate/src/config"
	"github.com/hadleyso/netid-activate/src/models"
)

func TestCheckEmailExists_LoginFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})
	_ = setupIDMTestServerBroken(t, mux)

	_, err := CheckEmailExists("test@example.com")
	if err == nil {
		t.Fatal("expected an error for login failure, but got nil")
	}
	if !strings.Contains(err.Error(), "RedHat IdM login failed") {
		t.Errorf("expected error to contain 'RedHat IdM login failed', but got: %s", err.Error())
	}
}

func TestCheckManagedGroup(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		// Mock response for batch call
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"result": {
				"count": 1,
				"results": [
					{
						"result": {
							"cn": ["managed-group"],
							"membermanager_user": ["testuser"]
						}
					}
				]
			}
		}`))
	})
	_ = setupIDMTestServerReplicate(t, mux)

	user := &models.UserInfo{PreferredUsername: "testuser"}
	groups := map[string][]config.Group{
		"managed-group": {
			{MemberManager: true, GroupName: "Managed Group"},
		},
	}

	err, managedGroups := CheckManagedGroup(user, groups)
	if err != nil {
		t.Fatalf("CheckManagedGroup failed: %v", err)
	}

	expected := []config.Group{
		{MemberManager: true, GroupName: "Managed Group", CN: "managed-group"},
	}

	if !reflect.DeepEqual(managedGroups, expected) {
		t.Errorf("expected managed groups %+v, got %+v", expected, managedGroups)
	}
}
