package globus

import (
	"github.com/MG-RAST/Shock/shock-server/user"
)

// Token is an exported version of token for testing
type Token token

// Exported version of authHeaderType for testing
func AuthHeaderType(header string) string {
	return authHeaderType(header)
}

// MockAuth is a mock function for testing
var MockAuth func(header string) (*user.User, error)

// MockFetchToken is a mock function for testing
var MockFetchToken func(username, password string) (*Token, error)

// FetchToken is an exported version of fetchToken for testing
func FetchToken(username, password string) (*Token, error) {
	if MockFetchToken != nil {
		return MockFetchToken(username, password)
	}
	t, err := fetchToken(username, password)
	if err != nil {
		return nil, err
	}
	return (*Token)(t), nil
}

// MockFetchProfile is a mock function for testing
var MockFetchProfile func(tokenString string) (*user.User, error)

// FetchProfile is an exported version of fetchProfile for testing
func FetchProfile(tokenString string) (*user.User, error) {
	if MockFetchProfile != nil {
		return MockFetchProfile(tokenString)
	}
	return fetchProfile(tokenString)
}

// MockClientId is a mock function for testing
var MockClientId func(tokenString string) (string, error)

// ClientId is an exported version of clientId for testing
func ClientId(tokenString string) (string, error) {
	if MockClientId != nil {
		return MockClientId(tokenString)
	}
	return clientId(tokenString)
}
