package globus_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/auth/globus"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/stretchr/testify/assert"
)

// TestAuthHeaderType tests the authHeaderType function
func TestAuthHeaderType(t *testing.T) {
	// Test valid header with bearer token
	validHeader := "Bearer token123"
	headerType := globus.AuthHeaderType(validHeader)
	assert.Equal(t, "bearer", headerType, "Header type should be 'bearer'")

	// Test valid header with basic auth
	validBasicHeader := "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ="
	headerType = globus.AuthHeaderType(validBasicHeader)
	assert.Equal(t, "basic", headerType, "Header type should be 'basic'")

	// Test valid header with globus token
	validGlobusHeader := "Globus-Goauthtoken token123"
	headerType = globus.AuthHeaderType(validGlobusHeader)
	assert.Equal(t, "globus-goauthtoken", headerType, "Header type should be 'globus-goauthtoken'")

	// Test invalid header (no space)
	invalidHeader := "Bearertoken123"
	headerType = globus.AuthHeaderType(invalidHeader)
	assert.Equal(t, "", headerType, "Header type should be empty for invalid header")

	// Test empty header
	emptyHeader := ""
	headerType = globus.AuthHeaderType(emptyHeader)
	assert.Equal(t, "", headerType, "Header type should be empty for empty header")
}

// TestAuth tests the Auth function
func TestAuth(t *testing.T) {
	// Save original mock function and restore it after the test
	originalMockAuth := globus.MockAuth
	defer func() {
		globus.MockAuth = originalMockAuth
	}()

	// Create a mock user for testing
	mockUser := &user.User{
		Username: "test_user",
		Fullname: "Test User",
		Email:    "test@example.com",
	}

	// Set up mock function
	globus.MockAuth = func(header string) (*user.User, error) {
		// Test basic auth header
		if header == "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ=" {
			return mockUser, nil
		}

		// Test globus token header
		if header == "Globus-Goauthtoken valid_token" {
			return mockUser, nil
		}

		// Test goauth token header
		if header == "Goauth valid_token" {
			return mockUser, nil
		}

		// Test invalid header type
		if header == "Invalid token123" {
			return nil, errors.New("(globus) Invalid authentication header, unknown bearer token: invalid")
		}

		// Test empty header
		if header == "" {
			return nil, errors.New("(globus) Invalid authentication header, missing bearer token.")
		}

		return nil, errors.New("Authentication failed")
	}

	// Test with basic auth header
	basicHeader := "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ=" // base64 of "test_user:test_password"
	u, err := globus.Auth(basicHeader)
	assert.NoError(t, err, "Auth with basic header should not error")
	assert.NotNil(t, u, "User should not be nil")
	assert.Equal(t, "test_user", u.Username, "Username should match")

	// Test with globus token header
	globusHeader := "Globus-Goauthtoken valid_token"
	u, err = globus.Auth(globusHeader)
	assert.NoError(t, err, "Auth with globus token should not error")
	assert.NotNil(t, u, "User should not be nil")
	assert.Equal(t, "test_user", u.Username, "Username should match")

	// Test with goauth token header
	goauthHeader := "Goauth valid_token"
	u, err = globus.Auth(goauthHeader)
	assert.NoError(t, err, "Auth with goauth token should not error")
	assert.NotNil(t, u, "User should not be nil")
	assert.Equal(t, "test_user", u.Username, "Username should match")

	// Test with invalid header type
	invalidHeader := "Invalid token123"
	u, err = globus.Auth(invalidHeader)
	assert.Error(t, err, "Auth with invalid header type should error")
	assert.Nil(t, u, "User should be nil")
	assert.Contains(t, err.Error(), "unknown bearer token", "Error message should indicate unknown bearer token")

	// Test with empty header
	emptyHeader := ""
	u, err = globus.Auth(emptyHeader)
	assert.Error(t, err, "Auth with empty header should error")
	assert.Nil(t, u, "User should be nil")
	assert.Contains(t, err.Error(), "missing bearer token", "Error message should indicate missing bearer token")
}

// TestFetchToken tests the fetchToken function
func TestFetchToken(t *testing.T) {
	// Save original function and restore it after the test
	originalMockFetchToken := globus.MockFetchToken
	defer func() { globus.MockFetchToken = originalMockFetchToken }()

	// Set up mock function
	globus.MockFetchToken = func(username, password string) (*globus.Token, error) {
		if username == "test_user" && password == "test_password" {
			return &globus.Token{
				AccessToken: "test_token",
				UserName:    "test_user",
			}, nil
		}
		return nil, errors.New("(globus) Authentication failed: Unexpected response status: 403 Forbidden")
	}

	// Test with valid credentials
	token, err := globus.FetchToken("test_user", "test_password")
	assert.NoError(t, err, "FetchToken with valid credentials should not error")
	assert.NotNil(t, token, "Token should not be nil")
	assert.Equal(t, "test_token", token.AccessToken, "Access token should match")
	assert.Equal(t, "test_user", token.UserName, "Username should match")

	// Test with invalid credentials
	token, err = globus.FetchToken("invalid", "invalid")
	assert.Error(t, err, "FetchToken with invalid credentials should error")
	assert.Nil(t, token, "Token should be nil")
	assert.Contains(t, err.Error(), "Authentication failed", "Error message should indicate authentication failure")

	// Test with server error
	globus.MockFetchToken = func(username, password string) (*globus.Token, error) {
		return nil, errors.New("(globus) HTTP GET: dial tcp: lookup invalid-server: no such host")
	}
	token, err = globus.FetchToken("test_user", "test_password")
	assert.Error(t, err, "FetchToken with server error should error")
	assert.Nil(t, token, "Token should be nil")
}

// TestFetchProfile tests the fetchProfile function
func TestFetchProfile(t *testing.T) {
	// Save original function and restore it after the test
	originalMockFetchProfile := globus.MockFetchProfile
	defer func() { globus.MockFetchProfile = originalMockFetchProfile }()

	// Create a mock user for testing
	mockUser := &user.User{
		Username: "test_user",
		Fullname: "Test User",
		Email:    "test@example.com",
	}

	// Set up mock function
	globus.MockFetchProfile = func(tokenString string) (*user.User, error) {
		if tokenString == "valid_token" {
			return mockUser, nil
		}
		return nil, errors.New("(globus) Authentication failed: Unexpected response status: 403 Forbidden")
	}

	// Test with valid token
	u, err := globus.FetchProfile("valid_token")
	assert.NoError(t, err, "FetchProfile with valid token should not error")
	assert.NotNil(t, u, "User should not be nil")
	assert.Equal(t, "test_user", u.Username, "Username should match")
	assert.Equal(t, "Test User", u.Fullname, "Fullname should match")
	assert.Equal(t, "test@example.com", u.Email, "Email should match")

	// Test with invalid token
	u, err = globus.FetchProfile("invalid_token")
	assert.Error(t, err, "FetchProfile with invalid token should error")
	assert.Nil(t, u, "User should be nil")
	assert.Contains(t, err.Error(), "Authentication failed", "Error message should indicate authentication failure")

	// Test with server error
	globus.MockFetchProfile = func(tokenString string) (*user.User, error) {
		return nil, errors.New("(globus) HTTP GET: dial tcp: lookup invalid-server: no such host")
	}
	u, err = globus.FetchProfile("valid_token")
	assert.Error(t, err, "FetchProfile with server error should error")
	assert.Nil(t, u, "User should be nil")
}

// TestClientId tests the clientId function
func TestClientId(t *testing.T) {
	// Save original function and restore it after the test
	originalMockClientId := globus.MockClientId
	defer func() { globus.MockClientId = originalMockClientId }()

	// Set up mock function
	globus.MockClientId = func(tokenString string) (string, error) {
		if tokenString == "client_id=test_client|other_stuff=value" {
			return "test_client", nil
		}
		if tokenString == "valid_token" {
			return "test_client", nil
		}
		if tokenString == "invalid_token" {
			return "", errors.New("(globus) Authentication failed: Unexpected response status: 403 Forbidden")
		}
		return "", errors.New("(globus) HTTP GET: dial tcp: lookup invalid-server: no such host")
	}

	// Test with old format token
	oldFormatToken := "client_id=test_client|other_stuff=value"
	clientId, err := globus.ClientId(oldFormatToken)
	assert.NoError(t, err, "ClientId with old format token should not error")
	assert.Equal(t, "test_client", clientId, "Client ID should match")

	// Test with new format token
	newFormatToken := "valid_token"
	clientId, err = globus.ClientId(newFormatToken)
	assert.NoError(t, err, "ClientId with new format token should not error")
	assert.Equal(t, "test_client", clientId, "Client ID should match")

	// Test with invalid token
	invalidToken := "invalid_token"
	clientId, err = globus.ClientId(invalidToken)
	assert.Error(t, err, "ClientId with invalid token should error")
	assert.Empty(t, clientId, "Client ID should be empty")

	// Test with server error
	serverErrorToken := "server_error_token"
	clientId, err = globus.ClientId(serverErrorToken)
	assert.Error(t, err, "ClientId with server error should error")
	assert.Empty(t, clientId, "Client ID should be empty")
}

// Helper function to set up a mock token server
func setupTokenServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a token request
		if r.Method == http.MethodGet {
			// Check for basic auth
			username, password, ok := r.BasicAuth()
			if ok && username == "test_user" && password == "test_password" {
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"access_token": "test_token",
					"token_id":     "test_token_id",
					"user_name":    "test_user",
					"client_id":    "test_client",
				})
				return
			}

			// Check for token validation
			token := r.Header.Get("X-Globus-Goauthtoken")
			if token == "valid_token" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"client_id": "test_client",
				})
				return
			}

			// Invalid credentials or token
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Invalid request method
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}

// Helper function to set up a mock profile server
func setupProfileServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is a profile request
		if r.Method == http.MethodGet {
			// Check for valid token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "Globus-Goauthtoken valid_token" {
				// Create a mock user for testing
				mockUser := &user.User{
					Username: "test_user",
					Fullname: "Test User",
					Email:    "test@example.com",
				}

				// Save original MockFetchProfile function and restore it after the test
				originalMockFetchProfile := globus.MockFetchProfile
				defer func() { globus.MockFetchProfile = originalMockFetchProfile }()

				// Set the mock function to return our mock user
				globus.MockFetchProfile = func(tokenString string) (*user.User, error) {
					if tokenString == "valid_token" {
						return mockUser, nil
					}
					return nil, errors.New("Authentication failed")
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"username": "test_user",
					"fullname": "Test User",
					"email":    "test@example.com",
				})
				return
			}

			// Invalid token
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Invalid request method
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
}
