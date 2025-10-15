package common

import (
	"errors"
	"log"

	"github.com/hadleyso/netid-activate/src/models"
	"github.com/spf13/viper"
)

type optionalGroups struct {
	GroupName     string
	RequiredGroup string
	DisplayName   string
}

// Get optional groups that user can add to
func GetOptionalGroupLimited(user *models.UserInfo) ([]optionalGroups, error) {
	groupReturn, err := GetOptionalGroup()
	if err != nil {
		return groupReturn, err
	}

	// Quick search
	inviterGroups := make(map[string]int)
	for _, val := range user.Groups {
		inviterGroups[val] = 0
	}

	// Allow empty to be included
	inviterGroups[""] = 0

	var filterGroup []optionalGroups
	for _, item := range groupReturn {
		if _, ok := inviterGroups[item.RequiredGroup]; ok {
			filterGroup = append(filterGroup, item)
		}
	}
	return filterGroup, nil
}

// Get all optional groups
func GetOptionalGroup() ([]optionalGroups, error) {

	// Not set in config
	if !viper.IsSet("OPTIONAL_GROUPS") {
		return []optionalGroups{}, nil
	}

	groupsRaw := viper.Get("OPTIONAL_GROUPS")

	groupsMap, ok := groupsRaw.(map[string]any)
	if !ok {
		log.Printf("InviteLandingPage() Expected OPTIONAL_GROUPS to be a slice, but got %T", groupsRaw)
		return []optionalGroups{}, errors.New("Expected OPTIONAL_GROUPS to be a slice")
	}

	groupReturn := []optionalGroups{}
	for group, groupSetting := range groupsMap {
		// For each group

		sliceSettings, ok := groupSetting.([]any)
		if !ok {
			log.Printf("InviteLandingPage() Expected OPTIONAL_GROUPS items to be a slice, but got %T", groupSetting)
			return []optionalGroups{}, errors.New("Expected OPTIONAL_GROUPS items to be a slice")
		}

		optGroup := optionalGroups{}
		for _, settingItem := range sliceSettings {
			// For settings in the group

			settingItemMap, ok := settingItem.(map[string]any)
			if !ok {
				log.Printf("InviteLandingPage() Expected OPTIONAL_GROUPS item settings to be a map string, but got %T", settingItem)
				return []optionalGroups{}, errors.New("Expected OPTIONAL_GROUPS item settings to be a map string")
			}
			for settingKey, settingValue := range settingItemMap {
				switch settingKey {
				case "group_required":
					optGroup.RequiredGroup = settingValue.(string)
				case "group_name":
					optGroup.DisplayName = settingValue.(string)
				}
			}
		}

		optGroup.GroupName = group
		groupReturn = append(groupReturn, optGroup)

	}

	return groupReturn, nil

}
