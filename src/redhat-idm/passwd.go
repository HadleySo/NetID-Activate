package idm

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc/v3"
)

func HandleResetPasswd(invite models.Invite, loginName string) (string, error) {

	// Create client
	client, errClient := newHTTPClient(false)
	if errClient != nil {
		log.Println("MakeUser() unable to newHTTPClient() " + errClient.Error())
		return "", errClient
	}

	// Auth
	username := viper.GetString("IDM_USERNAME")
	password := viper.GetString("IDM_PASSWORD")
	errLogin := login(client, username, password)
	if errLogin != nil {
		log.Println("MakeUser() unable to login() with HTTPClient " + errLogin.Error())
		return "", errLogin
	}

	// Generate password
	pin := randPIN()

	// Create user
	_, err := setPassword(client, loginName, pin)
	if err != nil {
		log.Println("MakeUser() unable to makeUser() " + err.Error())
		return "", err
	}

	return pin, nil
}

func setPassword(client *http.Client, uid string, password string) (any, error) {

	// Set connection
	rpcURL := viper.GetString("IDM_HOST") + "/ipa/session/json"
	rpcClient := jsonrpc.NewClientWithOpts(rpcURL,
		&jsonrpc.RPCClientOpts{
			AllowUnknownFields: true, // IdM returns principal
			CustomHeaders: map[string]string{
				"Referer":      viper.GetString("IDM_HOST") + "/ipa",
				"Content-Type": "application/json",
				"Accept":       "application/json",
			},
			HTTPClient: client,
		})

	// Params
	params := []any{
		[]string{uid},
		map[string]any{
			"password": password,
		},
	}

	resp, err := rpcClient.Call(context.Background(), "passwd", params...)
	if err != nil {
		log.Println("setPassword() call error " + err.Error())
		return nil, err
	}
	if resp.Error != nil {
		log.Println("setPassword() response error " + resp.Error.Message)
		return nil, fmt.Errorf("RPC error: %v", resp.Error.Message)
	}

	return resp.Result, nil

}
