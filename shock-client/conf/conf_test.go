package conf_test

import (
	"fmt"
	. "github.com/MG-RAST/Shock/shock-client/conf"
	"testing"
)

func TestDefaultPath(t *testing.T) {
	fmt.Println(DefaultPath())
}
