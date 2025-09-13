package common

import (
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/hadleyso/netid-activate/src/models"
	idm "github.com/hadleyso/netid-activate/src/redhat-idm"
)

func GetLoginOptions(invite models.Invite) ([]string, error) {
	loginNames := LoginGenerator(invite)

	readyNames, err := idm.CheckUsernamesExists(loginNames)
	if err != nil {
		return []string{}, err
	}

	return readyNames, nil
}

func LoginGenerator(invite models.Invite) []string {
	firstName := strings.ToLower(invite.FirstName)
	lastName := strings.ToLower(invite.LastName)
	usernameNumber := strconv.Itoa(rand.IntN(100) + 1)

	options := []string{}

	// First name + last name (never truncated)
	options = append(options, firstName+lastName)

	// First three letters + last name
	options = append(options, firstName[:3]+lastName)

	// First two letters + last name
	options = append(options, firstName[:2]+lastName)

	// First name + last two letters
	options = append(options, firstName+lastName[:2])

	// First initial + last name + number
	options = append(options, firstName[:1]+lastName+usernameNumber)

	// build a new filtered slice
	var filtered []string
	for _, s := range options {
		if len(s) >= 5 && len(s) <= 18 {
			filtered = append(filtered, s)
		}
	}

	return filtered
}
