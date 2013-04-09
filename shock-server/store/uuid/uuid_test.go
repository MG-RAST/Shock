package uuid_test

import (
	"fmt"
	. "github.com/MG-RAST/Shock/shock-server/store/uuid"
	"testing"
)

func TestNew(t *testing.T) {
	newuuid := New()
	fmt.Printf("uuid: %v.\n", newuuid)
}
