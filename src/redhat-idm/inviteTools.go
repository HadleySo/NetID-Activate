package idm

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"

	"github.com/hadleyso/netid-activate/src/config"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
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
	batchParams := []string{}

	for cn, group := range groups {
		for _, g := range group {
			// Special handling for MemberManager groups.
			if g.MemberManager {
				batchParams = append(batchParams, cn)
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

	// Login complete, get group info

	err, response := getGroupBatch(client, batchParams)
	if err != nil {
		log.Println("CheckEmailExists() unable to getGroupBatch() " + err.Error())
		return err, nil
	}

	// Got group result info with Member Managers

	cachedGetGroupBatch := makeCachedGetGroupBatch(client)

	for _, r := range response.Results {

		if len(r.Result.CN) < 1 {
			continue
		}

		cn := r.Result.CN[0]
		group := groups[cn][0]
		group.CN = cn

		// No manager
		if r.Result.MemberManagerUser == nil && r.Result.MemberManagerGroup == nil {
			continue
		}

		// Direct member‑manager user
		if r.Result.MemberManagerUser != nil && userInSlice(user.PreferredUsername, r.Result.MemberManagerUser) {
			filterGroup = append(filterGroup, group)
			log.Printf("CheckManagedGroup() user %s direct manager %s\n", user.PreferredUsername, group.CN)
			continue
		}

		// Don't run if not available
		if r.Result.MemberManagerGroup == nil {
			continue
		}

		// Check if user is in the group that can manage
		err, mgrGroupResponse := cachedGetGroupBatch(r.Result.MemberManagerGroup)
		if err != nil {
			log.Println("CheckEmailExists() unable to cachedGetGroupBatch() " + err.Error())
			return errClient, nil
		}

		for _, managedResponseResult := range mgrGroupResponse.Results {
			if managedResponseResult.Result.MemberUser != nil && userInSlice(user.PreferredUsername, managedResponseResult.Result.MemberUser) {
				log.Printf("CheckManagedGroup() user %s group manager %s\n", user.PreferredUsername, group.CN)
				filterGroup = append(filterGroup, group)
			}
		}

	}

	return nil, filterGroup
}

func userInSlice(username string, list []string) bool {
	return slices.Contains(list, username)
}
