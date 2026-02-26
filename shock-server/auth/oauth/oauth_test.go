package oauth_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/auth/oauth"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/stretchr/testify/assert"
)

// TestAuthHeaderType tests the authHeaderType function
func TestAuthHeaderType(t *testing.T) {
	// Test valid header with bearer token
	validHeader := "Bearer token123"
	headerType := oauth.AuthHeaderType(validHeader)
	assert.Equal(t, "bearer", headerType, "Header type should be 'bearer'")

	// Test valid header with basic auth
	validBasicHeader := "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ="
	headerType = oauth.AuthHeaderType(validBasicHeader)
	assert.Equal(t, "basic", headerType, "Header type should be 'basic'")

	// Test valid header with oauth token
	validOauthHeader := "OAuth token123"
	headerType = oauth.AuthHeaderType(validOauthHeader)
	assert.Equal(t, "oauth", headerType, "Header type should be 'oauth'")

	// Test invalid header (no space)
	invalidHeader := "Bearertoken123"
	headerType = oauth.AuthHeaderType(invalidHeader)
	assert.Equal(t, "", headerType, "Header type should be empty for invalid header")

	// Test empty header
	emptyHeader := ""
	headerType = oauth.AuthHeaderType(emptyHeader)
	assert.Equal(t, "", headerType, "Header type should be empty for empty header")
}

// TestAuth tests the Auth function
func TestAuth(t *testing.T) {
	// Set up mock for SetMongoInfo
	originalSetMongoInfo := user.MockSetMongoInfo
	user.MockSetMongoInfo = func(u *user.User) error {
		return nil
	}
	defer func() { user.MockSetMongoInfo = originalSetMongoInfo }()

	// Set up test server
	oauthServer := setupOAuthServer(t)
	defer oauthServer.Close()

	// Set configuration
	originalOAuthMap := conf.AUTH_OAUTH
	originalOAuthDefault := conf.OAUTH_DEFAULT
	defer func() {
		conf.AUTH_OAUTH = originalOAuthMap
		conf.OAUTH_DEFAULT = originalOAuthDefault
	}()

	conf.AUTH_OAUTH = map[string]string{
		"oauth":  oauthServer.URL,
		"custom": oauthServer.URL + "/custom",
	}
	conf.OAUTH_DEFAULT = oauthServer.URL

	// Test with oauth header
	oauthHeader := "OAuth valid_token"
	user, err := oauth.Auth(oauthHeader)
	assert.NoError(t, err, "Auth with oauth header should not error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, "test_user", user.Username, "Username should match")

	// Test with custom bearer header
	customHeader := "Custom valid_token"
	user, err = oauth.Auth(customHeader)
	assert.NoError(t, err, "Auth with custom header should not error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, "test_user", user.Username, "Username should match")

	// Test with basic auth header (should error)
	basicHeader := "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ="
	user, err = oauth.Auth(basicHeader)
	assert.Error(t, err, "Auth with basic header should error")
	assert.Nil(t, user, "User should be nil")
	assert.Contains(t, err.Error(), "does not support username/password", "Error message should indicate no support for username/password")

	// Test with invalid header type
	invalidHeader := "Invalid token123"
	user, err = oauth.Auth(invalidHeader)
	assert.Error(t, err, "Auth with invalid header type should error")
	assert.Nil(t, user, "User should be nil")
	assert.Contains(t, err.Error(), "unknown bearer token", "Error message should indicate unknown bearer token")

	// Test with empty header
	emptyHeader := ""
	user, err = oauth.Auth(emptyHeader)
	assert.Error(t, err, "Auth with empty header should error")
	assert.Nil(t, user, "User should be nil")
	assert.Contains(t, err.Error(), "missing bearer token", "Error message should indicate missing bearer token")
}

// TestAuthToken tests the authToken function
func TestAuthToken(t *testing.T) {
	// Set up mock for SetMongoInfo
	originalSetMongoInfo := user.MockSetMongoInfo
	user.MockSetMongoInfo = func(u *user.User) error {
		return nil
	}
	defer func() { user.MockSetMongoInfo = originalSetMongoInfo }()

	// Set up test server
	oauthServer := setupOAuthServer(t)
	defer oauthServer.Close()

	// Test with valid token
	user, err := oauth.AuthToken("valid_token", oauthServer.URL)
	assert.NoError(t, err, "AuthToken with valid token should not error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, "test_user", user.Username, "Username should match")
	assert.Equal(t, "Test User", user.Fullname, "Fullname should match")
	assert.Equal(t, "test@example.com", user.Email, "Email should match")

	// Test with invalid token
	user, err = oauth.AuthToken("invalid_token", oauthServer.URL)
	assert.Error(t, err, "AuthToken with invalid token should error")
	assert.Nil(t, user, "User should be nil")
	assert.Contains(t, err.Error(), "Invalid Auth", "Error message should indicate authentication failure")

	// Test with server error
	user, err = oauth.AuthToken("server_error", oauthServer.URL)
	assert.Error(t, err, "AuthToken with server error should error")
	assert.Nil(t, user, "User should be nil")
	assert.Contains(t, err.Error(), "Unexpected response status", "Error message should indicate unexpected response")

	// Test with invalid URL
	user, err = oauth.AuthToken("valid_token", "http://invalid-server")
	assert.Error(t, err, "AuthToken with invalid URL should error")
	assert.Nil(t, user, "User should be nil")
}

// TestAuthTokenWithEmptyFields tests the authToken function with empty fields in the response
func TestAuthTokenWithEmptyFields(t *testing.T) {
	// Set up mock for SetMongoInfo
	originalSetMongoInfo := user.MockSetMongoInfo
	user.MockSetMongoInfo = func(u *user.User) error {
		return nil
	}
	defer func() { user.MockSetMongoInfo = originalSetMongoInfo }()

	// Set up test server
	emptyFieldsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an OAuth request
		if r.Method == http.MethodGet {
			// Check for valid token
			authHeader := r.Header.Get("Auth")
			if authHeader == "valid_token" {
				// Return response with empty username
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"login": "",
					"name":  "Test User",
					"email": "test@example.com",
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
	defer emptyFieldsServer.Close()

	// Test with valid token but empty username
	user, err := oauth.AuthToken("valid_token", emptyFieldsServer.URL)
	assert.Error(t, err, "AuthToken with empty username should error")
	assert.Nil(t, user, "User should be nil")
}

// TestAuthTokenWithNameVariations tests the authToken function with different name field variations
func TestAuthTokenWithNameVariations(t *testing.T) {
	// Set up mock for SetMongoInfo
	originalSetMongoInfo := user.MockSetMongoInfo
	user.MockSetMongoInfo = func(u *user.User) error {
		return nil
	}
	defer func() { user.MockSetMongoInfo = originalSetMongoInfo }()

	// Set up test server with name field
	nameServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an OAuth request
		if r.Method == http.MethodGet {
			// Check for valid token
			authHeader := r.Header.Get("Auth")
			if authHeader == "valid_token" {
				// Return response with name field
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"login": "test_user",
					"name":  "Full Name",
					"email": "test@example.com",
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
	defer nameServer.Close()

	// Set up test server with firstname/lastname fields
	firstLastServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an OAuth request
		if r.Method == http.MethodGet {
			// Check for valid token
			authHeader := r.Header.Get("Auth")
			if authHeader == "valid_token" {
				// Return response with firstname/lastname fields
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"login":     "test_user",
					"firstname": "First",
					"lastname":  "Last",
					"email":     "test@example.com",
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
	defer firstLastServer.Close()

	// Test with name field
	user, err := oauth.AuthToken("valid_token", nameServer.URL)
	assert.NoError(t, err, "AuthToken with name field should not error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, "Full Name", user.Fullname, "Fullname should match name field")

	// Test with firstname/lastname fields
	user, err = oauth.AuthToken("valid_token", firstLastServer.URL)
	assert.NoError(t, err, "AuthToken with firstname/lastname fields should not error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, "First Last", user.Fullname, "Fullname should be combination of firstname and lastname")
}

// Helper function to set up a mock OAuth server
func setupOAuthServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an OAuth request
		if r.Method == http.MethodGet {
			// Check for valid token
			authHeader := r.Header.Get("Auth")
			if authHeader == "valid_token" {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"login": "test_user",
					"name":  "Test User",
					"email": "test@example.com",
				})
				return
			} else if authHeader == "server_error" {
				w.WriteHeader(http.StatusInternalServerError)
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
