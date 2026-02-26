package basic_test

import (
	"fmt"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/auth/basic"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/stretchr/testify/assert"
)

// TestDecodeHeader tests decoding the authorization header
func TestDecodeHeader(t *testing.T) {
	// Test valid basic auth header
	validHeader := "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ=" // base64 of "test_user:test_password"
	username, password, err := basic.DecodeHeader(validHeader)
	assert.NoError(t, err, "Decoding valid header should not error")
	assert.Equal(t, "test_user", username, "Username should match")
	assert.Equal(t, "test_password", password, "Password should match")

	// Test invalid header format (missing space)
	invalidHeader := "Basicinvalid"
	_, _, err = basic.DecodeHeader(invalidHeader)
	assert.Error(t, err, "Decoding invalid header should error")

	// Test invalid header format (wrong auth type)
	wrongTypeHeader := "Bearer dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ="
	_, _, err = basic.DecodeHeader(wrongTypeHeader)
	assert.Error(t, err, "Decoding header with wrong auth type should error")

	// Test invalid base64
	invalidBase64Header := "Basic invalid-base64"
	_, _, err = basic.DecodeHeader(invalidBase64Header)
	assert.Error(t, err, "Decoding header with invalid base64 should error")

	// Test invalid format (missing colon)
	invalidFormatHeader := "Basic dGVzdF91c2Vy" // base64 of "test_user"
	_, _, err = basic.DecodeHeader(invalidFormatHeader)
	assert.Error(t, err, "Decoding header with invalid format should error")
}

// TestAuth tests the Auth function
func TestAuth(t *testing.T) {
	// Save original MockFindByUsernamePassword function and restore it after the test
	originalMock := basic.MockFindByUsernamePassword
	defer func() { basic.MockFindByUsernamePassword = originalMock }()

	// Set the mock function
	basic.MockFindByUsernamePassword = func(username, password string) (*user.User, error) {
		if username == "test_user" && password == "test_password" {
			return &user.User{
				Uuid:     "test_uuid",
				Username: "test_user",
				Fullname: "Test User",
				Email:    "test@example.com",
			}, nil
		}
		return nil, assert.AnError
	}

	// Test valid auth
	validHeader := "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ=" // base64 of "test_user:test_password"
	u, err := basic.Auth(validHeader)
	assert.NoError(t, err, "Auth with valid credentials should not error")
	assert.NotNil(t, u, "User should not be nil")
	assert.Equal(t, "test_user", u.Username, "Username should match")

	// Test invalid auth
	invalidHeader := "Basic aW52YWxpZDppbnZhbGlk" // base64 of "invalid:invalid"
	u, err = basic.Auth(invalidHeader)
	assert.Error(t, err, "Auth with invalid credentials should error")
	assert.Nil(t, u, "User should be nil")

	// Test invalid header format
	invalidFormatHeader := "InvalidFormat"
	u, err = basic.Auth(invalidFormatHeader)
	assert.Error(t, err, "Auth with invalid header format should error")
	assert.Nil(t, u, "User should be nil")
}

// TestAuthWithDebug tests the Auth function with debug enabled
func TestAuthWithDebug(t *testing.T) {
	// Save original debug setting and restore it after the test
	originalDebug := conf.DEBUG_AUTH
	defer func() { conf.DEBUG_AUTH = originalDebug }()

	// Enable debug
	conf.DEBUG_AUTH = true

	// Save original MockFindByUsernamePassword function and restore it after the test
	originalMock := basic.MockFindByUsernamePassword
	defer func() { basic.MockFindByUsernamePassword = originalMock }()

	// Set the mock function
	basic.MockFindByUsernamePassword = func(username, password string) (*user.User, error) {
		if username == "test_user" && password == "test_password" {
			return &user.User{
				Uuid:     "test_uuid",
				Username: "test_user",
				Fullname: "Test User",
				Email:    "test@example.com",
			}, nil
		}
		// When DEBUG_AUTH is true, we need to include the debug info in the error
		if conf.DEBUG_AUTH {
			return nil, fmt.Errorf("(Basic/Auth) user.FindByUsernamePassword returned: %s", assert.AnError.Error())
		}
		return nil, assert.AnError
	}

	// Test valid auth
	validHeader := "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ=" // base64 of "test_user:test_password"
	u, err := basic.Auth(validHeader)
	assert.NoError(t, err, "Auth with valid credentials should not error")
	assert.NotNil(t, u, "User should not be nil")

	// Test invalid auth
	invalidHeader := "Basic aW52YWxpZDppbnZhbGlk" // base64 of "invalid:invalid"
	u, err = basic.Auth(invalidHeader)
	assert.Error(t, err, "Auth with invalid credentials should error")
	assert.Nil(t, u, "User should be nil")
	assert.Contains(t, err.Error(), "(Basic/Auth)", "Error message should contain debug info")

	// Test invalid header format
	invalidFormatHeader := "InvalidFormat"
	u, err = basic.Auth(invalidFormatHeader)
	assert.Error(t, err, "Auth with invalid header format should error")
	assert.Nil(t, u, "User should be nil")
	assert.Contains(t, err.Error(), "(Basic/Auth)", "Error message should contain debug info")
}

// TestDecodeHeaderWithDebug tests the DecodeHeader function with debug enabled
func TestDecodeHeaderWithDebug(t *testing.T) {
	// Save original debug setting and restore it after the test
	originalDebug := conf.DEBUG_AUTH
	defer func() { conf.DEBUG_AUTH = originalDebug }()

	// Enable debug
	conf.DEBUG_AUTH = true

	// Test valid basic auth header
	validHeader := "Basic dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ=" // base64 of "test_user:test_password"
	username, password, err := basic.DecodeHeader(validHeader)
	assert.NoError(t, err, "Decoding valid header should not error")
	assert.Equal(t, "test_user", username, "Username should match")
	assert.Equal(t, "test_password", password, "Password should match")

	// Test invalid header format (missing space)
	invalidHeader := "Basicinvalid"
	_, _, err = basic.DecodeHeader(invalidHeader)
	assert.Error(t, err, "Decoding invalid header should error")
	assert.Contains(t, err.Error(), "(basic/DecodeHeader)", "Error message should contain debug info")

	// Test invalid header format (wrong auth type)
	wrongTypeHeader := "Bearer dGVzdF91c2VyOnRlc3RfcGFzc3dvcmQ="
	_, _, err = basic.DecodeHeader(wrongTypeHeader)
	assert.Error(t, err, "Decoding header with wrong auth type should error")
	assert.Contains(t, err.Error(), "(basic/DecodeHeader)", "Error message should contain debug info")

	// Test invalid base64
	invalidBase64Header := "Basic invalid-base64"
	_, _, err = basic.DecodeHeader(invalidBase64Header)
	assert.Error(t, err, "Decoding header with invalid base64 should error")
	assert.Contains(t, err.Error(), "(basic/DecodeHeader)", "Error message should contain debug info")

	// Test invalid format (missing colon)
	invalidFormatHeader := "Basic dGVzdF91c2Vy" // base64 of "test_user"
	_, _, err = basic.DecodeHeader(invalidFormatHeader)
	assert.Error(t, err, "Decoding header with invalid format should error")
	assert.Contains(t, err.Error(), "(basic/DecodeHeader)", "Error message should contain debug info")
}
