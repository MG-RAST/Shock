package mgrast_test

import (
	"fmt"
	. "github.com/MG-RAST/Shock/shock-server/auth/mgrast"
	"testing"
)

var (
	valid   = "gCyVPMHcAR9mvfKnyeshDWDj8"
	invalid = "ghaioehgaoei"
)

func TestAuthToken(t *testing.T) {
	user, err := AuthToken(valid)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("%#v\n", user)
	}
	user, err = AuthToken(invalid)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		fmt.Printf("%#v\n", user)
	}
}
