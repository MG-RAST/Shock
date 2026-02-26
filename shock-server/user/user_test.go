package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUserStruct tests the User struct field assignments
func TestUserStruct(t *testing.T) {
	u := User{
		Uuid:     "test-uuid-123",
		Username: "testuser",
		Fullname: "Test User",
		Email:    "test@example.com",
		Password: "secret",
		Admin:    true,
	}

	assert.Equal(t, "test-uuid-123", u.Uuid)
	assert.Equal(t, "testuser", u.Username)
	assert.Equal(t, "Test User", u.Fullname)
	assert.Equal(t, "test@example.com", u.Email)
	assert.Equal(t, "secret", u.Password)
	assert.True(t, u.Admin)
}

// TestUsersType tests that Users is a slice of User
func TestUsersType(t *testing.T) {
	users := Users{
		{Uuid: "uuid1", Username: "user1"},
		{Uuid: "uuid2", Username: "user2"},
	}

	assert.Len(t, users, 2)
	assert.Equal(t, "user1", users[0].Username)
	assert.Equal(t, "user2", users[1].Username)
}

// TestSetMongoInfoWithMock tests SetMongoInfo using the exported mock hook
func TestSetMongoInfoWithMock(t *testing.T) {
	// Save and restore the original mock
	originalMock := MockSetMongoInfo
	defer func() { MockSetMongoInfo = originalMock }()

	// Set up a mock that sets UUID and Admin
	MockSetMongoInfo = func(u *User) error {
		u.Uuid = "mock-uuid"
		u.Admin = true
		return nil
	}

	u := &User{Username: "testuser"}
	err := u.SetMongoInfo()
	assert.NoError(t, err)
	assert.Equal(t, "mock-uuid", u.Uuid)
	assert.True(t, u.Admin)
}

// TestSetMongoInfoMockError tests SetMongoInfo when mock returns an error
func TestSetMongoInfoMockError(t *testing.T) {
	originalMock := MockSetMongoInfo
	defer func() { MockSetMongoInfo = originalMock }()

	MockSetMongoInfo = func(u *User) error {
		return assert.AnError
	}

	u := &User{Username: "testuser"}
	err := u.SetMongoInfo()
	assert.Error(t, err)
}

// TestUserZeroValue tests the zero value of User struct
func TestUserZeroValue(t *testing.T) {
	var u User
	assert.Empty(t, u.Uuid)
	assert.Empty(t, u.Username)
	assert.Empty(t, u.Fullname)
	assert.Empty(t, u.Email)
	assert.Empty(t, u.Password)
	assert.False(t, u.Admin)
}
