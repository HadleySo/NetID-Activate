package common

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hadleyso/netid-activate/src/config"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
)

// setupIDMTestServer creates a test HTTP server and configures viper to use it.
func setupIDMTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	server := httptest.NewServer(handler)
	viper.Set("IDM_HOST", server.URL)
	viper.Set("IDM_USERNAME", "admin")
	viper.Set("IDM_PASSWORD", "Secret123")
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

func TestGetOptionalGroupLimited(t *testing.T) {
	// Success handler
	successMux := http.NewServeMux()
	successMux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	successMux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
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

	// Login failure handler
	loginFailureMux := http.NewServeMux()
	loginFailureMux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	// RPC failure handler
	rpcFailureMux := http.NewServeMux()
	rpcFailureMux.HandleFunc("/ipa/session/login_password", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	rpcFailureMux.HandleFunc("/ipa/session/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error": {"message": "batch failed"}}`))
	})

	testCases := []struct {
		name             string
		user             *models.UserInfo
		mockGroups       map[string][]config.Group
		mockHandler      http.Handler
		expected         []config.Group
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "user has required group for unmanaged group and is manager of a managed group",
			user: &models.UserInfo{
				Groups:            []string{"required-group-1"},
				PreferredUsername: "employee",
			},
			mockGroups: map[string][]config.Group{
				"unmanaged-group":   {{RequiredGroup: "required-group-1", GroupName: "Unmanaged Group", MemberManager: false}},
				"another-unmanaged": {{RequiredGroup: "required-group-2", GroupName: "Another Unmanaged", MemberManager: false}},
				"managed-group":     {{RequiredGroup: "some-other-req", GroupName: "Managed Group", MemberManager: true}},
			},
			mockHandler: successMux,
			expected: []config.Group{
				{RequiredGroup: "required-group-1", GroupName: "Unmanaged Group", MemberManager: false, CN: "unmanaged-group"}},
			expectError: false,
		},
		{
			name: "user has no groups and not a manager",
			user: &models.UserInfo{
				Groups:            []string{},
				PreferredUsername: "employee",
			},
			mockGroups: map[string][]config.Group{
				"unmanaged-group": {{RequiredGroup: "required-group-1", GroupName: "Unmanaged Group", MemberManager: false}},
				"managed-group":   {{GroupName: "Managed Group", MemberManager: true}},
			},
			mockHandler: successMux,
			expected:    nil,
			expectError: false,
		},
		{
			name: "IDM login failure",
			user: &models.UserInfo{
				PreferredUsername: "employee",
			},
			mockGroups: map[string][]config.Group{
				"managed-group": {{GroupName: "Managed Group", MemberManager: true}},
			},
			mockHandler:      loginFailureMux,
			expected:         nil,
			expectError:      true,
			expectedErrorMsg: "RedHat IdM login failed",
		},
		{
			name: "IDM RPC failure",
			user: &models.UserInfo{
				PreferredUsername: "employee",
			},
			mockGroups: map[string][]config.Group{
				"managed-group": {{GroupName: "Managed Group", MemberManager: true}},
			},
			mockHandler:      rpcFailureMux,
			expected:         nil,
			expectError:      true,
			expectedErrorMsg: "RPC error: batch failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = setupIDMTestServer(t, tc.mockHandler)

			config.C.OptionalGroups = tc.mockGroups
			defer func() { config.C.OptionalGroups = nil }()

			result, err := GetOptionalGroupLimited(tc.user)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected an error, but got nil")
				}
				if tc.expectedErrorMsg != "" && !strings.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("expected error message to contain '%s', but got '%s'", tc.expectedErrorMsg, err.Error())
				}
				return // Don't check result if error is expected
			}

			if err != nil {
				t.Fatalf("did not expect an error, but got: %v", err)
			}

			sort.Slice(result, func(i, j int) bool {
				return result[i].CN < result[j].CN
			})
			sort.Slice(tc.expected, func(i, j int) bool {
				return tc.expected[i].CN < tc.expected[j].CN
			})

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected groups %+v, got %+v", tc.expected, result)
			}
		})
	}
}
