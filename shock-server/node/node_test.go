package node_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/stretchr/testify/assert"
)

// TestNodeCreation tests the creation of a new node
func TestNodeCreation(t *testing.T) {
	// Create a test user
	u := &user.User{
		Uuid:         "test_user",
		Username:     "test_user",
		Fullname:     "Test User",
		Email:        "test@example.com",
		Admin:        false,
		CustomFields: map[string][]string{},
	}

	// Create a temporary file for testing
	tempDir, err := ioutil.TempDir("", "shock-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testContent := []byte("test file content")
	testFile := file.FormFile{
		Name: "test_file.txt",
		Path: tempDir + "/test_file.txt",
	}
	err = ioutil.WriteFile(testFile.Path, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a map of form files
	files := file.FormFiles{
		"upload": testFile,
	}

	// Create parameters for node creation
	params := map[string]string{
		"file_name": "test_file.txt",
	}

	// Create a node
	n, err := node.CreateNodeUpload(u, params, files)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	// Verify node properties
	assert.NotEmpty(t, n.Id, "Node ID should not be empty")
	assert.Equal(t, "basic", n.Type, "Node type should be 'basic'")
	assert.Equal(t, u.Uuid, n.Acl.Owner, "Node owner should be the test user")
	assert.Equal(t, "test_file.txt", n.File.Name, "File name should match")
	assert.Equal(t, int64(len(testContent)), n.File.Size, "File size should match content length")
	assert.NotEmpty(t, n.File.Checksum, "File checksum should not be empty")
	assert.False(t, n.CreatedOn.IsZero(), "Created time should be set")

	// Clean up
	deleted, err := n.Delete()
	assert.NoError(t, err, "Node deletion should not error")
	assert.True(t, deleted, "Node should be deleted")
}

// TestNodeFileOperations tests file operations on a node
func TestNodeFileOperations(t *testing.T) {
	// Create a test user
	u := &user.User{
		Uuid:         "test_user",
		Username:     "test_user",
		Fullname:     "Test User",
		Email:        "test@example.com",
		Admin:        false,
		CustomFields: map[string][]string{},
	}

	// Create a new node without a file
	n := node.New("")
	n.Type = "basic"
	n.Acl.SetOwner(u.Uuid)
	n.Acl.Set(u.Uuid, map[string]bool{"read": true, "write": true, "delete": true})

	// Create node directory
	err := n.Mkdir()
	assert.NoError(t, err, "Node directory creation should not error")

	// Create a test file
	tempDir, err := ioutil.TempDir("", "shock-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testContent := []byte("test file content for file operations")
	testFilePath := tempDir + "/test_file.txt"
	err = ioutil.WriteFile(testFilePath, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Set file on node
	formFile := file.FormFile{
		Name: "test_file.txt",
		Path: testFilePath,
	}
	err = n.SetFile(formFile)
	assert.NoError(t, err, "Setting file should not error")
	assert.Equal(t, "test_file.txt", n.File.Name, "File name should match")
	assert.Equal(t, int64(len(testContent)), n.File.Size, "File size should match content length")
	assert.NotEmpty(t, n.File.Checksum, "File checksum should not be empty")

	// Save the node
	err = n.Save()
	assert.NoError(t, err, "Node save should not error")

	// Read file from node
	reader, err := n.FileReader()
	assert.NoError(t, err, "Getting file reader should not error")
	defer reader.Close()

	buffer := new(bytes.Buffer)
	_, err = buffer.ReadFrom(reader)
	assert.NoError(t, err, "Reading file should not error")
	assert.Equal(t, testContent, buffer.Bytes(), "File content should match original content")

	// Clean up
	deleted, err := n.Delete()
	assert.NoError(t, err, "Node deletion should not error")
	assert.True(t, deleted, "Node should be deleted")
}

// TestNodeAttributes tests setting and getting node attributes
func TestNodeAttributes(t *testing.T) {
	// Create a test user
	u := &user.User{
		Uuid:         "test_user",
		Username:     "test_user",
		Fullname:     "Test User",
		Email:        "test@example.com",
		Admin:        false,
		CustomFields: map[string][]string{},
	}

	// Create a new node
	n := node.New("")
	n.Type = "basic"
	n.Acl.SetOwner(u.Uuid)
	n.Acl.Set(u.Uuid, map[string]bool{"read": true, "write": true, "delete": true})

	// Create node directory
	err := n.Mkdir()
	assert.NoError(t, err, "Node directory creation should not error")

	// Create a temporary file for attributes
	tempDir, err := ioutil.TempDir("", "shock-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create attributes JSON
	attrJSON := []byte(`{"project":"test","sample":"sample1","metadata":{"field1":"value1","field2":42}}`)
	attrFile := tempDir + "/attr.json"
	err = ioutil.WriteFile(attrFile, attrJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write attributes file: %v", err)
	}

	// Set attributes
	formFile := file.FormFile{
		Name: "attr.json",
		Path: attrFile,
	}
	err = n.SetAttributes(formFile)
	assert.NoError(t, err, "Setting attributes should not error")

	// Save the node
	err = n.Save()
	assert.NoError(t, err, "Node save should not error")

	// Load the node to verify attributes
	loadedNode, err := node.Load(n.Id)
	assert.NoError(t, err, "Loading node should not error")

	// Verify attributes (as a map)
	attrMap, ok := loadedNode.Attributes.(map[string]interface{})
	assert.True(t, ok, "Attributes should be a map")
	assert.Equal(t, "test", attrMap["project"], "Project attribute should match")
	assert.Equal(t, "sample1", attrMap["sample"], "Sample attribute should match")

	// Verify nested attributes
	metadataMap, ok := attrMap["metadata"].(map[string]interface{})
	assert.True(t, ok, "Metadata should be a map")
	assert.Equal(t, "value1", metadataMap["field1"], "Metadata field1 should match")
	assert.Equal(t, float64(42), metadataMap["field2"], "Metadata field2 should match")

	// Clean up
	deleted, err := n.Delete()
	assert.NoError(t, err, "Node deletion should not error")
	assert.True(t, deleted, "Node should be deleted")
}

// TestNodeExpiration tests setting and checking node expiration
func TestNodeExpiration(t *testing.T) {
	// Create a test user
	u := &user.User{
		Uuid:         "test_user",
		Username:     "test_user",
		Fullname:     "Test User",
		Email:        "test@example.com",
		Admin:        false,
		CustomFields: map[string][]string{},
	}

	// Create a new node
	n := node.New("")
	n.Type = "basic"
	n.Acl.SetOwner(u.Uuid)
	n.Acl.Set(u.Uuid, map[string]bool{"read": true, "write": true, "delete": true})

	// Create node directory
	err := n.Mkdir()
	assert.NoError(t, err, "Node directory creation should not error")

	// Save the node
	err = n.Save()
	assert.NoError(t, err, "Node save should not error")

	// Set expiration to 1 day from now
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	err = n.SetExpiration(tomorrow)
	assert.NoError(t, err, "Setting expiration should not error")

	// Save the node again
	err = n.Save()
	assert.NoError(t, err, "Node save should not error")

	// Load the node to verify expiration
	loadedNode, err := node.Load(n.Id)
	assert.NoError(t, err, "Loading node should not error")

	// Verify expiration date (should be within 24 hours of tomorrow)
	expectedTime, _ := time.Parse("2006-01-02", tomorrow)
	assert.WithinDuration(t, expectedTime, loadedNode.Expiration, 24*time.Hour, "Expiration should be set to tomorrow")

	// Clean up
	deleted, err := n.Delete()
	assert.NoError(t, err, "Node deletion should not error")
	assert.True(t, deleted, "Node should be deleted")
}
