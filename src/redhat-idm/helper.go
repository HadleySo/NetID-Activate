package idm

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/ybbus/jsonrpc/v3"
)

func newHTTPClient(insecureSkipVerify bool) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
	if caPath := os.Getenv("CACERT_PATH"); caPath != "" {
		b, err := os.ReadFile(caPath)
		if err != nil {
			log.Println("newHTTPClient() could not open cert file")
			return nil, err
		}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(b); !ok {
			log.Printf("newHTTPClient: no certs appended from %s", caPath)
		}
		tlsConfig.RootCAs = pool
	}

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{
		Jar:       jar,
		Transport: transport,
	}, nil
}

func login(client *http.Client, user string, password string) error {
	IDM_HOST := os.Getenv("IDM_HOST")
	loginURL := IDM_HOST + "/ipa/session/login_password"
	form := url.Values{
		"user":     {user},
		"password": {password},
	}

	body := strings.NewReader(form.Encode())

	req, err := http.NewRequest("POST", loginURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", IDM_HOST+"/ipa")
	req.Header.Set("Accept", "text/plain")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("RedHat IdM login failed: " + resp.Status)
		return fmt.Errorf("RedHat IdM login failed: %s", resp.Status)
	}
	return nil
}

// Find user by email
// client must be authenticated
func findUserByEmail(client *http.Client, email string) (any, error) {
	rpcURL := os.Getenv("IDM_HOST") + "/ipa/session/json"
	rpcClient := jsonrpc.NewClientWithOpts(rpcURL,
		&jsonrpc.RPCClientOpts{
			AllowUnknownFields: true, // IdM returns principal
			CustomHeaders: map[string]string{
				"Referer":      os.Getenv("IDM_HOST") + "/ipa",
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
			HTTPClient: client,
		})

	// Params: 1st = query filters, 2nd = options
	params := []any{
		[]string{},
		map[string]any{"all": true, "sizelimit": 1, "mail": []string{email}},
	}

	resp, err := rpcClient.Call(context.Background(), "user_find", params...)
	if err != nil {
		log.Println("findUserByEmail() call error")
		return nil, err
	}
	if resp.Error != nil {
		log.Println("findUserByEmail() response error")
		return nil, fmt.Errorf("RPC error: %v", resp.Error.Message)
	}
	return resp.Result, nil
}

// Find user by login name
// client must be authenticated
func findUserByLogin(client *http.Client, loginName string) (any, error) {
	rpcURL := os.Getenv("IDM_HOST") + "/ipa/session/json"
	rpcClient := jsonrpc.NewClientWithOpts(rpcURL,
		&jsonrpc.RPCClientOpts{
			AllowUnknownFields: true, // IdM returns principal
			CustomHeaders: map[string]string{
				"Referer":      os.Getenv("IDM_HOST") + "/ipa",
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
			HTTPClient: client,
		})

	// Params: 1st = query filters, 2nd = options
	params := []any{
		[]string{},
		map[string]any{"all": true, "sizelimit": 1, "pkey_only": true, "uid": loginName},
	}

	resp, err := rpcClient.Call(context.Background(), "user_find", params...)
	if err != nil {
		log.Println("findUserByEmail() call error")
		return nil, err
	}
	if resp.Error != nil {
		log.Println("findUserByEmail() response error")
		return nil, fmt.Errorf("RPC error: %v", resp.Error.Message)
	}
	return resp.Result, nil
}

func getDN(client *http.Client) (string, error) {
	// Set connection
	rpcURL := os.Getenv("IDM_HOST") + "/ipa/session/json"
	rpcClient := jsonrpc.NewClientWithOpts(rpcURL,
		&jsonrpc.RPCClientOpts{
			AllowUnknownFields: true, // IdM returns principal
			CustomHeaders: map[string]string{
				"Referer":      os.Getenv("IDM_HOST") + "/ipa",
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
			HTTPClient: client,
		})

	// Params
	params := []any{
		[]string{},
		map[string]any{},
	}

	// Call RPC
	resp, err := rpcClient.Call(context.Background(), "config_show", params...)
	if err != nil {
		log.Println("getDN() call error " + err.Error())
		return "", err
	}
	if resp.Error != nil {
		log.Println("getDN() response error " + resp.Error.Message)
		return "", fmt.Errorf("RPC error: %v", resp.Error.Message)
	}

	// Unmarshall
	data, ok := resp.Result.(map[string]any)
	if !ok {
		fmt.Println("getDN() Error parsing RPC")
		return "", fmt.Errorf("getDN() Error parsing RPC")
	}

	// Extract "dn" entry
	if result, ok := data["result"].(map[string]any); ok {
		if dn, ok := result["dn"].(string); ok {

			parts := strings.Split(dn, ",")

			// Filter only the dc= components
			var baseDN []string
			for _, part := range parts {
				if strings.HasPrefix(part, "dc=") {
					baseDN = append(baseDN, part)
				}
			}
			return strings.Join(baseDN, ","), nil
		} else {
			log.Println("getDN() DN not found or not a string")
		}
	}

	fmt.Println("getDN() result not found or not a map")
	return "", fmt.Errorf("getDN() result not found or not a map")

}
