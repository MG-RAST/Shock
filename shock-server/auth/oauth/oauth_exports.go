package oauth

import (
	"github.com/MG-RAST/Shock/shock-server/user"
)

// Exported version of authHeaderType for testing
func AuthHeaderType(header string) string {
	return authHeaderType(header)
}

// MockAuthToken is a mock function for testing
var MockAuthToken func(token string, url string) (*user.User, error)

// AuthToken is an exported version of authToken for testing
func AuthToken(token string, url string) (*user.User, error) {
	if MockAuthToken != nil {
		return MockAuthToken(token, url)
	}
	return authToken(token, url)
}
