package idm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// Checks if email exists in IdM
// Errors true if unable to determine
func CheckEmailExists(email string) (bool, error) {
	client, errClient := newHTTPClient(false)

	if errClient != nil {
		log.Println("CheckEmailExists() unable to newHTTPClient() " + errClient.Error())
		return true, errClient
	}

	username := os.Getenv("IDM_USERNAME")
	password := os.Getenv("IDM_PASSWORD")
	errLogin := login(client, username, password)
	if errLogin != nil {
		log.Println("CheckEmailExists() unable to login() with HTTPClient " + errLogin.Error())
		return true, errLogin
	}

	rpcResponse, errRPC := findUserByEmail(client, email)
	if errRPC != nil {
		log.Println("CheckEmailExists() unable to findUserByEmail() with authenticated HTTPClient " + errRPC.Error())
		return true, errRPC
	}

	data, ok := rpcResponse.(map[string]interface{})
	if !ok {
		return true, fmt.Errorf("CheckEmailExists() unexpected response type: %T", rpcResponse)
	}

	// extract "count"
	rawCount, exists := data["count"]
	if !exists {
		return true, fmt.Errorf(`CheckEmailExists() key "count" not found`)
	}

	// Handle JSON.Number
	num, ok := rawCount.(json.Number)
	if !ok {
		return true, fmt.Errorf("count is not JSON.Number: %T", rawCount)
	}

	count64, err := num.Int64()
	if err != nil {
		return true, fmt.Errorf("invalid count value %q: %w", num, err)
	}

	count := int(count64)

	return count != 0, nil
}
