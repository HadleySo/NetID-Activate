package common

import (
	"slices"
	"strings"
	"testing"

	"github.com/hadleyso/netid-activate/src/models"
)

func TestLoginGenerator(t *testing.T) {
	invite := models.Invite{
		FirstName: "Test",
		LastName:  "User",
	}

	options := LoginGenerator(invite)

	expectedOptions := []string{
		"testuser",
		"tesuser",
		"teuser",
		"testus",
	}

	if len(options) != 5 {
		t.Errorf("Expected 5 login options, but got %d", len(options))
	}

	for _, expected := range expectedOptions {
		if !slices.Contains(options, expected) {
			t.Errorf("Expected to find option '%s' in %v, but it was not present", expected, options)
		}
	}

	// check for random part
	foundRandom := false
	for _, opt := range options {
		if strings.HasPrefix(opt, "tuser") {
			foundRandom = true
			break
		}
	}
	if !foundRandom {
		t.Errorf("Expected to find an option with prefix 'tuser', but it was not present")
	}
}

func TestLoginGenerator_LengthFilter(t *testing.T) {
	invite := models.Invite{
		FirstName: "longfirstname",
		LastName:  "longlastname",
	}

	options := LoginGenerator(invite)

	// "longfirstnamelonglastname" is 25 chars, should be filtered out.
	unexpectedOption := "longfirstnamelonglastname"
	if slices.Contains(options, unexpectedOption) {
		t.Errorf("Unexpected option '%s' found, it should have been filtered out due to length", unexpectedOption)
	}

	expectedOptions := []string{
		"lonlonglastname", // 15
		"lolonglastname",  // 14
		"longfirstnamelo", // 15
	}

	if len(options) != 4 {
		t.Errorf("Expected 4 login options, but got %d. Options: %v", len(options), options)
	}

	for _, expected := range expectedOptions {
		if !slices.Contains(options, expected) {
			t.Errorf("Expected to find option '%s' in %v, but it was not present", expected, options)
		}
	}
	// and random one "llonglastname" + num

	foundRandom := false
	for _, opt := range options {
		if strings.HasPrefix(opt, "llonglastname") {
			foundRandom = true
			break
		}
	}
	if !foundRandom {
		t.Errorf("Expected to find an option with prefix 'llonglastname', but it was not present")
	}
}
