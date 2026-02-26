package acl_test

import (
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/acl"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/stretchr/testify/assert"
)

// TestSetOwner tests setting the owner of an ACL
func TestSetOwner(t *testing.T) {
	// Create a new ACL
	a := &acl.Acl{}

	// Set the owner
	ownerID := "test_owner"
	a.SetOwner(ownerID)

	// Verify the owner was set
	assert.Equal(t, ownerID, a.Owner, "Owner should be set correctly")
}

// TestSet tests setting permissions for a user
func TestSet(t *testing.T) {
	// Create a new ACL
	a := &acl.Acl{}

	// Set permissions for a user
	userID := "test_user"
	rights := acl.Rights{
		"read":   true,
		"write":  true,
		"delete": false,
	}

	a.Set(userID, rights)

	// Verify the permissions were set
	assert.Contains(t, a.Read, userID, "User should have read permission")
	assert.Contains(t, a.Write, userID, "User should have write permission")
	assert.NotContains(t, a.Delete, userID, "User should not have delete permission")

	// Set additional permissions
	rights = acl.Rights{
		"read":   false, // This should not remove the existing permission
		"write":  false, // This should not remove the existing permission
		"delete": true,  // This should add the permission
	}

	a.Set(userID, rights)

	// Verify the permissions were updated
	assert.Contains(t, a.Read, userID, "User should still have read permission")
	assert.Contains(t, a.Write, userID, "User should still have write permission")
	assert.Contains(t, a.Delete, userID, "User should now have delete permission")
}

// TestUnSet tests removing permissions for a user
func TestUnSet(t *testing.T) {
	// Create a new ACL
	a := &acl.Acl{
		Read:   []string{"test_user", "other_user"},
		Write:  []string{"test_user"},
		Delete: []string{"test_user", "other_user"},
	}

	// Remove permissions for a user
	userID := "test_user"
	rights := acl.Rights{
		"read":   true,
		"write":  false,
		"delete": true,
	}

	a.UnSet(userID, rights)

	// Verify the permissions were removed
	assert.NotContains(t, a.Read, userID, "User should not have read permission")
	assert.Contains(t, a.Write, userID, "User should still have write permission")
	assert.NotContains(t, a.Delete, userID, "User should not have delete permission")

	// Verify other users' permissions were not affected
	assert.Contains(t, a.Read, "other_user", "Other user's read permission should not be affected")
	assert.Contains(t, a.Delete, "other_user", "Other user's delete permission should not be affected")
}

// TestCheck tests checking permissions for a user
func TestCheck(t *testing.T) {
	// Create a new ACL
	a := &acl.Acl{
		Read:   []string{"test_user", "public"},
		Write:  []string{"test_user"},
		Delete: []string{"admin_user"},
	}

	// Check permissions for a user with multiple permissions
	userID := "test_user"
	rights := a.Check(userID)

	assert.True(t, rights["read"], "User should have read permission")
	assert.True(t, rights["write"], "User should have write permission")
	assert.False(t, rights["delete"], "User should not have delete permission")

	// Check permissions for a user with one permission
	userID = "admin_user"
	rights = a.Check(userID)

	assert.False(t, rights["read"], "User should not have read permission")
	assert.False(t, rights["write"], "User should not have write permission")
	assert.True(t, rights["delete"], "User should have delete permission")

	// Check permissions for a user with no permissions
	userID = "unknown_user"
	rights = a.Check(userID)

	assert.False(t, rights["read"], "User should not have read permission")
	assert.False(t, rights["write"], "User should not have write permission")
	assert.False(t, rights["delete"], "User should not have delete permission")

	// Check permissions for public
	userID = "public"
	rights = a.Check(userID)

	assert.True(t, rights["read"], "Public should have read permission")
	assert.False(t, rights["write"], "Public should not have write permission")
	assert.False(t, rights["delete"], "Public should not have delete permission")
}

// TestFormatDisplayAcl tests formatting ACL for display
func TestFormatDisplayAcl(t *testing.T) {
	// Create a new ACL
	a := &acl.Acl{
		Owner:  "owner_user",
		Read:   []string{"test_user", "public"},
		Write:  []string{"test_user"},
		Delete: []string{"admin_user"},
	}

	// Mock the user.FindByUuid function
	originalFindByUuid := user.MockFindByUuid
	defer func() { user.MockFindByUuid = originalFindByUuid }()

	user.MockFindByUuid = func(uuid string) (*user.User, error) {
		switch uuid {
		case "owner_user":
			return &user.User{
				Uuid:     "owner_user",
				Username: "owner",
				Fullname: "Owner User",
				Email:    "owner@example.com",
			}, nil
		case "test_user":
			return &user.User{
				Uuid:     "test_user",
				Username: "tester",
				Fullname: "Test User",
				Email:    "test@example.com",
			}, nil
		case "admin_user":
			return &user.User{
				Uuid:     "admin_user",
				Username: "admin",
				Fullname: "Admin User",
				Email:    "admin@example.com",
			}, nil
		default:
			return nil, assert.AnError
		}
	}

	// Test minimal verbosity (default)
	minimalDisplay := a.FormatDisplayAcl("")

	// Verify the minimal display
	minimalAcl, ok := minimalDisplay.(*acl.DisplayAcl)
	assert.True(t, ok, "Minimal display should be a DisplayAcl")
	assert.Equal(t, "owner_user", minimalAcl.Owner, "Owner should match")
	assert.Contains(t, minimalAcl.Read, "test_user", "Read permissions should include test_user")
	assert.Contains(t, minimalAcl.Write, "test_user", "Write permissions should include test_user")
	assert.Contains(t, minimalAcl.Delete, "admin_user", "Delete permissions should include admin_user")
	assert.True(t, minimalAcl.Public.Read, "Public read should be true")
	assert.False(t, minimalAcl.Public.Write, "Public write should be false")
	assert.False(t, minimalAcl.Public.Delete, "Public delete should be false")

	// Test full verbosity
	fullDisplay := a.FormatDisplayAcl("full")

	// Verify the full display
	fullAcl, ok := fullDisplay.(*acl.DisplayVerboseAcl)
	assert.True(t, ok, "Full display should be a DisplayVerboseAcl")
	assert.Equal(t, "owner_user", fullAcl.Owner.Uuid, "Owner UUID should match")
	assert.Equal(t, "owner", fullAcl.Owner.Username, "Owner username should match")
	assert.Equal(t, "Owner User", fullAcl.Owner.Fullname, "Owner fullname should match")

	// Verify read permissions
	assert.Len(t, fullAcl.Read, 1, "Should have 1 read user")
	assert.Equal(t, "test_user", fullAcl.Read[0].Uuid, "Read user UUID should match")
	assert.Equal(t, "tester", fullAcl.Read[0].Username, "Read user username should match")

	// Verify write permissions
	assert.Len(t, fullAcl.Write, 1, "Should have 1 write user")
	assert.Equal(t, "test_user", fullAcl.Write[0].Uuid, "Write user UUID should match")

	// Verify delete permissions
	assert.Len(t, fullAcl.Delete, 1, "Should have 1 delete user")
	assert.Equal(t, "admin_user", fullAcl.Delete[0].Uuid, "Delete user UUID should match")

	// Verify public permissions
	assert.True(t, fullAcl.Public.Read, "Public read should be true")
	assert.False(t, fullAcl.Public.Write, "Public write should be false")
	assert.False(t, fullAcl.Public.Delete, "Public delete should be false")
}

// TestInsert tests the insert helper function
func TestInsert(t *testing.T) {
	// Test inserting into an empty array
	arr := []string{}
	result := acl.Insert(arr, "test")
	assert.Len(t, result, 1, "Result should have 1 element")
	assert.Equal(t, "test", result[0], "Element should be inserted")

	// Test inserting a duplicate
	result = acl.Insert(result, "test")
	assert.Len(t, result, 1, "Result should still have 1 element")
	assert.Equal(t, "test", result[0], "Element should not be duplicated")

	// Test inserting multiple elements
	result = acl.Insert(result, "another")
	assert.Len(t, result, 2, "Result should have 2 elements")
	assert.Equal(t, "test", result[0], "First element should be preserved")
	assert.Equal(t, "another", result[1], "Second element should be inserted")
}

// TestInsertUser tests the insertUser helper function
func TestInsertUser(t *testing.T) {
	// Create test users
	user1 := user.User{Uuid: "user1", Username: "user1"}
	user2 := user.User{Uuid: "user2", Username: "user2"}

	// Test inserting into an empty array
	arr := []user.User{}
	result := acl.InsertUser(arr, user1)
	assert.Len(t, result, 1, "Result should have 1 element")
	assert.Equal(t, user1, result[0], "User should be inserted")

	// Test inserting a duplicate
	result = acl.InsertUser(result, user1)
	assert.Len(t, result, 1, "Result should still have 1 element")
	assert.Equal(t, user1, result[0], "User should not be duplicated")

	// Test inserting multiple users
	result = acl.InsertUser(result, user2)
	assert.Len(t, result, 2, "Result should have 2 elements")
	assert.Equal(t, user1, result[0], "First user should be preserved")
	assert.Equal(t, user2, result[1], "Second user should be inserted")
}

// TestDel tests the del helper function
func TestDel(t *testing.T) {
	// Test deleting from an array with one element
	arr := []string{"test"}
	result := acl.Del(arr, "test")
	assert.Len(t, result, 0, "Result should be empty")

	// Test deleting from an array with multiple elements
	arr = []string{"test1", "test2", "test3"}
	result = acl.Del(arr, "test2")
	assert.Len(t, result, 2, "Result should have 2 elements")
	assert.Equal(t, "test1", result[0], "First element should be preserved")
	assert.Equal(t, "test3", result[1], "Third element should be preserved")

	// Test deleting a non-existent element
	result = acl.Del(arr, "non_existent")
	assert.Len(t, result, 3, "Result should have 3 elements")
	assert.Equal(t, arr, result, "Array should be unchanged")

	// Test deleting from an empty array
	arr = []string{}
	result = acl.Del(arr, "test")
	assert.Len(t, result, 0, "Result should be empty")
}
