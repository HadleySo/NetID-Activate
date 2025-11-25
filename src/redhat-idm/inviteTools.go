package idm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"slices"

	"github.com/hadleyso/netid-activate/src/config"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc/v3"
)

// Checks if email exists in IdM
// Errors true if unable to determine
func CheckEmailExists(email string) (bool, error) {
	client, errClient := newHTTPClient(false)

	if errClient != nil {
		log.Println("CheckEmailExists() unable to newHTTPClient() " + errClient.Error())
		return true, errClient
	}

	username := viper.GetString("IDM_USERNAME")
	password := viper.GetString("IDM_PASSWORD")
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

// Retrieve optional groups that give user has
// permissions for based on MemberManager LDAP
func CheckManagedGroup(user *models.UserInfo, groups map[string][]config.Group) (error, []config.Group) {

	var filterGroup []config.Group

	// Params
	batchParams := []any{}

	for cn, group := range groups {
		for _, g := range group {
			// Special handling for MemberManager groups.
			if g.MemberManager {
				entry := map[string]any{
					"method": "group_show",
					"params": []any{
						[]string{cn},
						map[string]any{"no_members": false},
					},
				}
				batchParams = append(batchParams, entry)
			}
		}
	}

	if len(batchParams) == 0 {
		return nil, filterGroup
	}

	client, errClient := newHTTPClient(false)

	if errClient != nil {
		log.Println("CheckEmailExists() unable to newHTTPClient() " + errClient.Error())
		return errClient, nil
	}

	username := viper.GetString("IDM_USERNAME")
	password := viper.GetString("IDM_PASSWORD")
	errLogin := login(client, username, password)
	if errLogin != nil {
		log.Println("CheckEmailExists() unable to login() with HTTPClient " + errLogin.Error())
		return errLogin, nil
	}

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

	resp, err := rpcClient.Call(context.Background(), "batch", batchParams, map[string]any{})
	if err != nil {
		log.Println("CheckManagedGroup() call error")
		return err, nil
	}
	if resp.Error != nil {
		log.Println("CheckManagedGroup() response error: " + resp.Error.Message)
		return error(fmt.Errorf("RPC error: %v", resp.Error.Message)), nil
	}

	var response models.BatchResponse
	data, err := json.Marshal(resp.Result)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &response); err != nil {
		panic(err)
	}

	for _, r := range response.Results {

		if len(r.Result.CN) < 1 {
			continue
		}

		cn := r.Result.CN[0]
		if r.Result.MemberManagerUser == nil {
			continue
		}
		if slices.Contains(r.Result.MemberManagerUser, user.PreferredUsername) {
			group := groups[cn][0]
			group.CN = cn
			filterGroup = append(filterGroup, group)
		}
	}

	return nil, filterGroup
}
