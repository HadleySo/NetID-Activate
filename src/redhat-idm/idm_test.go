package idm

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
