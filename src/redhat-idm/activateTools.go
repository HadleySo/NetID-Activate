package idm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Checks if login names exists in IdM
// Returns names that don't exist
func CheckUsernamesExists(loginNames []string) ([]string, error) {

	okUsername := []string{}

	client, errClient := newHTTPClient(false)

	if errClient != nil {
		log.Println("CheckUsernamesExists() unable to newHTTPClient() " + errClient.Error())
		return okUsername, errClient
	}

	username := os.Getenv("IDM_USERNAME")
	password := os.Getenv("IDM_PASSWORD")
	errLogin := login(client, username, password)
	if errLogin != nil {
		log.Println("CheckUsernamesExists() unable to login() with HTTPClient " + errLogin.Error())
		return okUsername, errLogin
	}

	for _, v := range loginNames {
		rpcResponse, errRPC := findUserByLogin(client, v)
		if errRPC != nil {
			log.Println("CheckUsernamesExists() unable to findUserByEmail() with authenticated HTTPClient " + errRPC.Error())
			return okUsername, errRPC
		}

		data, ok := rpcResponse.(map[string]interface{})
		if !ok {
			return okUsername, fmt.Errorf("CheckUsernamesExists() unexpected response type: %T", rpcResponse)
		}

		// extract "count"
		rawCount, exists := data["count"]
		if !exists {
			return okUsername, fmt.Errorf(`CheckUsernamesExists() key "count" not found`)
		}

		// Handle JSON.Number
		num, ok := rawCount.(json.Number)
		if !ok {
			return okUsername, fmt.Errorf("count is not JSON.Number: %T", rawCount)
		}

		count64, err := num.Int64()
		if err != nil {
			return okUsername, fmt.Errorf("invalid count value %q: %w", num, err)
		}

		count := int(count64)

		if count == 0 {
			okUsername = append(okUsername, v)
		}
	}

	return okUsername, nil
}
