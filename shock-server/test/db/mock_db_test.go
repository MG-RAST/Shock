package db_test

import (
	"testing"

	"github.com/MG-RAST/Shock/shock-server/test/db"
	"github.com/MG-RAST/Shock/shock-server/test/util"
	"github.com/stretchr/testify/assert"
)

// TestNewMockDB tests creating a new mock database
func TestNewMockDB(t *testing.T) {
	mockDB := db.NewMockDB()
	assert.NotNil(t, mockDB, "Mock database should not be nil")
}

// TestMockDBUserOperations tests adding and retrieving users
func TestMockDBUserOperations(t *testing.T) {
	mockDB := db.NewMockDB()

	// Create a test user
	testUser := util.TestUser()

	// Add the user
	mockDB.AddTestUser(testUser)

	// Retrieve the user by UUID
	retrieved, err := mockDB.GetUser(testUser.Uuid)
	assert.NoError(t, err, "GetUser should not error for existing user")
	assert.NotNil(t, retrieved, "Retrieved user should not be nil")
	assert.Equal(t, testUser.Uuid, retrieved.Uuid, "UUID should match")
	assert.Equal(t, testUser.Username, retrieved.Username, "Username should match")
	assert.Equal(t, testUser.Fullname, retrieved.Fullname, "Fullname should match")
	assert.Equal(t, testUser.Email, retrieved.Email, "Email should match")
	assert.Equal(t, testUser.Admin, retrieved.Admin, "Admin should match")

	// Test retrieving a non-existent user
	_, err = mockDB.GetUser("nonexistent")
	assert.Error(t, err, "GetUser should error for non-existent user")
}

// TestMockDBAdminUser tests adding an admin user
func TestMockDBAdminUser(t *testing.T) {
	mockDB := db.NewMockDB()

	adminUser := util.TestAdminUser()
	mockDB.AddTestUser(adminUser)

	retrieved, err := mockDB.GetUser(adminUser.Uuid)
	assert.NoError(t, err, "GetUser should not error")
	assert.True(t, retrieved.Admin, "Admin should be true")
}

// TestMockDBNodeOperations tests adding, retrieving, and deleting nodes
func TestMockDBNodeOperations(t *testing.T) {
	mockDB := db.NewMockDB()

	// Create a test node
	testNode := util.CreateTestNode()

	// Add the node
	mockDB.AddNode(testNode)

	// Retrieve the node
	retrieved, err := mockDB.GetNode(testNode.Id)
	assert.NoError(t, err, "GetNode should not error for existing node")
	assert.NotNil(t, retrieved, "Retrieved node should not be nil")
	assert.Equal(t, testNode.Id, retrieved.Id, "Node ID should match")
	assert.Equal(t, testNode.File.Name, retrieved.File.Name, "File name should match")
	assert.False(t, retrieved.CreatedOn.IsZero(), "CreatedOn should be set")
	assert.False(t, retrieved.LastModified.IsZero(), "LastModified should be set")

	// Test retrieving a non-existent node
	_, err = mockDB.GetNode("nonexistent")
	assert.Error(t, err, "GetNode should error for non-existent node")

	// Delete the node
	err = mockDB.DeleteNode(testNode.Id)
	assert.NoError(t, err, "DeleteNode should not error for existing node")

	// Verify deletion
	_, err = mockDB.GetNode(testNode.Id)
	assert.Error(t, err, "GetNode should error after deletion")

	// Test deleting a non-existent node
	err = mockDB.DeleteNode("nonexistent")
	assert.Error(t, err, "DeleteNode should error for non-existent node")
}

// TestMockDBUpdateNode tests updating a node
func TestMockDBUpdateNode(t *testing.T) {
	mockDB := db.NewMockDB()

	testNode := util.CreateTestNode()
	mockDB.AddNode(testNode)

	// Update the file name
	err := mockDB.UpdateNode(testNode.Id, map[string]interface{}{
		"file.name": "updated_file.txt",
	})
	assert.NoError(t, err, "UpdateNode should not error")

	// Verify the update
	retrieved, err := mockDB.GetNode(testNode.Id)
	assert.NoError(t, err, "GetNode should not error")
	assert.Equal(t, "updated_file.txt", retrieved.File.Name, "File name should be updated")

	// Update the file size
	err = mockDB.UpdateNode(testNode.Id, map[string]interface{}{
		"file.size": int64(999),
	})
	assert.NoError(t, err, "UpdateNode should not error for size update")

	retrieved, err = mockDB.GetNode(testNode.Id)
	assert.NoError(t, err, "GetNode should not error")
	assert.Equal(t, int64(999), retrieved.File.Size, "File size should be updated")

	// Update attributes
	attrs := map[string]string{"key": "value"}
	err = mockDB.UpdateNode(testNode.Id, map[string]interface{}{
		"attributes": attrs,
	})
	assert.NoError(t, err, "UpdateNode should not error for attributes update")

	// Test updating a non-existent node
	err = mockDB.UpdateNode("nonexistent", map[string]interface{}{
		"file.name": "test",
	})
	assert.Error(t, err, "UpdateNode should error for non-existent node")

	// Test updating with an unsupported field
	err = mockDB.UpdateNode(testNode.Id, map[string]interface{}{
		"unsupported_field": "value",
	})
	assert.Error(t, err, "UpdateNode should error for unsupported field")
}

// TestMockDBFindNodes tests finding nodes
func TestMockDBFindNodes(t *testing.T) {
	mockDB := db.NewMockDB()

	// Add multiple nodes
	node1 := util.CreateTestNode()
	node1.Id = "node_1"
	mockDB.AddNode(node1)

	node2 := util.CreateTestNode()
	node2.Id = "node_2"
	mockDB.AddNode(node2)

	// Find all nodes (default behavior returns all)
	nodes, err := mockDB.FindNodes(nil)
	assert.NoError(t, err, "FindNodes should not error")
	assert.Len(t, nodes, 2, "Should find 2 nodes")
}

// TestMockDBSetQueryFunc tests setting a custom query function
func TestMockDBSetQueryFunc(t *testing.T) {
	mockDB := db.NewMockDB()

	node1 := util.CreateTestNode()
	node1.Id = "node_1"
	mockDB.AddNode(node1)

	node2 := util.CreateTestNode()
	node2.Id = "node_2"
	mockDB.AddNode(node2)

	// Set a custom query function that returns only node_1
	mockDB.SetQueryFunc(func(query interface{}) (interface{}, error) {
		result := []*util.Node{node1}
		return result, nil
	})

	nodes, err := mockDB.FindNodes(nil)
	assert.NoError(t, err, "FindNodes with custom query should not error")
	assert.Len(t, nodes, 1, "Custom query should return 1 node")
	assert.Equal(t, "node_1", nodes[0].Id, "Should return node_1")
}

// TestMockDBReset tests resetting the mock database
func TestMockDBReset(t *testing.T) {
	mockDB := db.NewMockDB()

	// Add a user and a node
	mockDB.AddTestUser(util.TestUser())
	mockDB.AddNode(util.CreateTestNode())

	// Reset
	mockDB.Reset()

	// Verify everything is cleared
	_, err := mockDB.GetUser("test_user")
	assert.Error(t, err, "GetUser should error after reset")

	_, err = mockDB.GetNode("test_node_id")
	assert.Error(t, err, "GetNode should error after reset")

	nodes, err := mockDB.FindNodes(nil)
	assert.NoError(t, err, "FindNodes should not error after reset")
	assert.Len(t, nodes, 0, "Should find 0 nodes after reset")
}
