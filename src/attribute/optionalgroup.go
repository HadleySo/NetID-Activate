package attribute

import (
	"github.com/hadleyso/netid-activate/src/config"
	"github.com/hadleyso/netid-activate/src/models"
	idm "github.com/hadleyso/netid-activate/src/redhat-idm"
)

// Get optional groups that user can add to
func GetOptionalGroupLimited(user *models.UserInfo) ([]config.Group, error) {
	optionalGroups := config.C.OptionalGroups

	// Quick search
	inviterGroups := make(map[string]int)
	for _, val := range user.Groups {
		inviterGroups[val] = 0
	}

	// Allow empty to be included
	inviterGroups[""] = 0

	var filterGroup []config.Group
	for cn, groups := range optionalGroups {
		for _, g := range groups {
			if g.MemberManager {
				// Skip if managed
				continue
			}

			if _, ok := inviterGroups[g.RequiredGroup]; ok {
				if g.MemberManager != true {
					g.CN = cn
					filterGroup = append(filterGroup, g)
				}

			}
		}
	}

	err, managedGroups := idm.CheckManagedGroup(user, optionalGroups)
	if err != nil {
		return nil, err
	}

	filterGroup = append(filterGroup, managedGroups...)

	return filterGroup, nil
}
