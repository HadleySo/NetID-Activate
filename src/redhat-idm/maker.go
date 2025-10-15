package idm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/hadleyso/netid-activate/src/countries"
	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
	"github.com/ybbus/jsonrpc/v3"
)

func HandleMakeUser(invite models.Invite, loginName string) (string, error) {

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
	_, err := makeUser(client, loginName, invite.Email, invite.FirstName, invite.LastName, invite.Country, invite.Affiliation, pin, invite.State, invite.Inviter)
	if err != nil {
		return "", err

	}

	// Add to groups
	var groups []string
	if err := json.Unmarshal(invite.OptionalGroups, &groups); err != nil {
		return "", err
	}
	groups = append(groups, strings.Split(viper.GetString("IDM_ADD_GROUP"), ",")...)

	addUserGroups(client, loginName, groups) // TODO: handle errors
	return pin, nil
}

// Client must be authenticated
func makeUser(client *http.Client, uid string, email string, firstName string, lastName string, country string, affiliation string, password string, st string, managerUIN string) (any, error) {
	// Combine variables
	cn := firstName + " " + lastName
	initials := strings.ToUpper(firstName[:1] + lastName[:1])
	gecos := cn
	if viper.GetString("IDM_GECOS") == "true" {
		gecos = cn + " (" + country + " " + affiliation + ")"
	}

	// Get value
	alpha2, errCountry := countries.GetAlpha2FromAlpha3(country)
	if errCountry == false {
		log.Println("makeUser() GetAlpha2FromAlpha3 error")
		return nil, fmt.Errorf("makeUser() error in GetAlpha2FromAlpha3()")
	}
	countryName, errCountry := countries.GetNameFromAlpha3(country)
	if errCountry == false {
		log.Println("makeUser() GetAlpha2FromAlpha3 error")
		return nil, fmt.Errorf("makeUser() error in GetAlpha2FromAlpha3()")
	}

	// Set st
	st = st + ", " + countryName

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
			"all":          true,
			"cn":           cn,
			"displayname":  cn,
			"gecos":        gecos,
			"givenname":    firstName,
			"sn":           lastName,
			"initials":     initials,
			"mail":         []string{email},
			"st":           st,
			"userpassword": password,
			"manager":      managerUIN,
			"pager":        []string{alpha2},
		},
	}

	resp, err := rpcClient.Call(context.Background(), "user_add", params...)
	if err != nil {
		log.Println("makeUser() call error " + err.Error())
		return nil, err
	}
	if resp.Error != nil {
		log.Println("makeUser() response error " + resp.Error.Message)
		return nil, fmt.Errorf("RPC error: %v", resp.Error.Message)
	}

	return resp.Result, nil

}

func addUserGroups(client *http.Client, uid string, groups []string) error {

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

	var subRequests []any
	for _, grp := range groups {
		subRequests = append(subRequests, map[string]any{
			"method": "group_add_member",
			"params": []any{
				[]string{grp}, // group name array
				map[string]any{
					"user": []string{uid}, // single user string
				},
			},
		})
	}

	// Batch param wrap
	batchParams := []any{
		subRequests, // first param: array of sub-request dicts
		map[string]any{
			"version": "2.253", // second param: options
		},
	}

	// Call the batch method
	resp, err := rpcClient.Call(context.Background(), "batch", batchParams...)
	if err != nil {
		log.Println("addUserGroups() call error " + err.Error())
		return err
	}
	if resp.Error != nil {
		log.Println("addUserGroups() response error " + resp.Error.Message)
		return fmt.Errorf("RPC error: %v", resp.Error.Message)
	}

	return nil

}

func randPIN() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	parts := make([]string, 3)

	for i := range parts {
		var sb strings.Builder
		for j := 0; j < 3; j++ {
			idx := r.Intn(len(letters))
			sb.WriteByte(letters[idx])
		}
		parts[i] = sb.String()
	}

	return strings.Join(parts, "-")
}
