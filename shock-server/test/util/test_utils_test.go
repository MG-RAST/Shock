package util_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/test/util"
	"github.com/stretchr/testify/assert"
)

// TestTestUser tests the TestUser function
func TestTestUser(t *testing.T) {
	u := util.TestUser()
	assert.NotNil(t, u, "User should not be nil")
	assert.Equal(t, "test_user", u.Uuid, "UUID should match")
	assert.Equal(t, "test_user", u.Username, "Username should match")
	assert.Equal(t, "Test User", u.Fullname, "Fullname should match")
	assert.Equal(t, "test@example.com", u.Email, "Email should match")
	assert.False(t, u.Admin, "Admin should be false")
	assert.NotNil(t, u.CustomFields, "CustomFields should not be nil")
}

// TestTestAdminUser tests the TestAdminUser function
func TestTestAdminUser(t *testing.T) {
	u := util.TestAdminUser()
	assert.NotNil(t, u, "Admin user should not be nil")
	assert.Equal(t, "test_admin", u.Uuid, "UUID should match")
	assert.Equal(t, "test_admin", u.Username, "Username should match")
	assert.Equal(t, "Test Admin", u.Fullname, "Fullname should match")
	assert.Equal(t, "admin@example.com", u.Email, "Email should match")
	assert.True(t, u.Admin, "Admin should be true")
}

// TestCreateTestNode tests the CreateTestNode function
func TestCreateTestNode(t *testing.T) {
	node := util.CreateTestNode()
	assert.NotNil(t, node, "Node should not be nil")
	assert.Equal(t, "test_node_id", node.Id, "Node ID should match")
	assert.Equal(t, "1.0", node.Version, "Version should match")

	// Verify file info
	assert.Equal(t, "test_file.txt", node.File.Name, "File name should match")
	assert.Equal(t, int64(12), node.File.Size, "File size should match")
	assert.Equal(t, "text", node.File.Format, "File format should match")
	assert.NotNil(t, node.File.Checksum, "Checksum map should not be nil")
	assert.Equal(t, "test_checksum", node.File.Checksum["md5"], "MD5 checksum should match")

	// Verify ACL
	assert.Equal(t, "test_user", node.Acl.Owner, "ACL owner should match")
	assert.Contains(t, node.Acl.Read, "test_user", "Read ACL should contain test_user")
	assert.Contains(t, node.Acl.Write, "test_user", "Write ACL should contain test_user")
	assert.Contains(t, node.Acl.Delete, "test_user", "Delete ACL should contain test_user")

	// Verify timestamps
	assert.False(t, node.CreatedOn.IsZero(), "CreatedOn should be set")
	assert.False(t, node.LastModified.IsZero(), "LastModified should be set")
}

// TestCreateTestFormFile tests the CreateTestFormFile function
func TestCreateTestFormFile(t *testing.T) {
	content := "test file content"
	formFile, err := util.CreateTestFormFile(content)
	assert.NoError(t, err, "CreateTestFormFile should not error")
	assert.Equal(t, "test_file.txt", formFile.Name, "File name should match")
	assert.NotEmpty(t, formFile.Path, "File path should not be empty")

	// Verify the file exists and has the correct content
	data, err := ioutil.ReadFile(formFile.Path)
	assert.NoError(t, err, "Reading the form file should not error")
	assert.Equal(t, content, string(data), "File content should match")

	// Clean up the temp directory
	dir := formFile.Path[:len(formFile.Path)-len("/test_file.txt")]
	os.RemoveAll(dir)
}

// TestNewInMemoryFile tests the NewInMemoryFile function
func TestNewInMemoryFile(t *testing.T) {
	testData := []byte("test data")
	f := util.NewInMemoryFile("test.txt", testData)
	assert.NotNil(t, f, "InMemoryFile should not be nil")

	// Test reading
	buf := make([]byte, len(testData))
	n, err := f.Read(buf)
	assert.NoError(t, err, "Read should not error")
	assert.Equal(t, len(testData), n, "Should read the correct number of bytes")
	assert.Equal(t, testData, buf, "Read data should match")

	// Reading again should return EOF
	n, err = f.Read(buf)
	assert.Equal(t, io.EOF, err, "Should return EOF")
	assert.Equal(t, 0, n, "Should read 0 bytes")
}

// TestInMemoryFileSeek tests seeking in an in-memory file
func TestInMemoryFileSeek(t *testing.T) {
	testData := []byte("test data")
	f := util.NewInMemoryFile("test.txt", testData)

	// Seek to the beginning
	pos, err := f.Seek(0, io.SeekStart)
	assert.NoError(t, err, "Seek to start should not error")
	assert.Equal(t, int64(0), pos, "Position should be 0")

	// Read after seeking
	buf := make([]byte, 4)
	n, err := f.Read(buf)
	assert.NoError(t, err, "Read should not error")
	assert.Equal(t, 4, n, "Should read 4 bytes")
	assert.Equal(t, []byte("test"), buf, "Data should match")

	// Seek from current position
	pos, err = f.Seek(1, io.SeekCurrent)
	assert.NoError(t, err, "Seek from current should not error")
	assert.Equal(t, int64(5), pos, "Position should be 5")

	// Seek from end
	pos, err = f.Seek(-4, io.SeekEnd)
	assert.NoError(t, err, "Seek from end should not error")
	assert.Equal(t, int64(5), pos, "Position should be 5")

	// Seek with invalid whence
	_, err = f.Seek(0, 3)
	assert.Error(t, err, "Seek with invalid whence should error")
}

// TestInMemoryFileReadAt tests reading at a specific offset
func TestInMemoryFileReadAt(t *testing.T) {
	testData := []byte("test data")
	f := util.NewInMemoryFile("test.txt", testData)

	// Read at offset 0
	buf := make([]byte, 4)
	n, err := f.ReadAt(buf, 0)
	assert.NoError(t, err, "ReadAt offset 0 should not error")
	assert.Equal(t, 4, n, "Should read 4 bytes")
	assert.Equal(t, []byte("test"), buf, "Data should match")

	// Read at offset 5
	buf = make([]byte, 4)
	n, err = f.ReadAt(buf, 5)
	assert.NoError(t, err, "ReadAt offset 5 should not error")
	assert.Equal(t, 4, n, "Should read 4 bytes")
	assert.Equal(t, []byte("data"), buf, "Data should match")

	// Read beyond end
	buf = make([]byte, 20)
	n, err = f.ReadAt(buf, 5)
	assert.Equal(t, io.EOF, err, "ReadAt beyond end should return EOF")
	assert.Equal(t, 4, n, "Should read remaining bytes")
}

// TestInMemoryFileClose tests closing an in-memory file
func TestInMemoryFileClose(t *testing.T) {
	testData := []byte("test data")
	f := util.NewInMemoryFile("test.txt", testData)

	// Close the file
	err := f.Close()
	assert.NoError(t, err, "Close should not error")

	// Reading after close should error
	buf := make([]byte, 10)
	_, err = f.Read(buf)
	assert.ErrorIs(t, err, os.ErrClosed, "Read after close should return ErrClosed")

	// Seeking after close should error
	_, err = f.Seek(0, io.SeekStart)
	assert.ErrorIs(t, err, os.ErrClosed, "Seek after close should return ErrClosed")
}

// TestInMemoryFileStat tests getting file info
func TestInMemoryFileStat(t *testing.T) {
	testData := []byte("test data")
	f := util.NewInMemoryFile("test.txt", testData)

	info, err := f.Stat()
	assert.NoError(t, err, "Stat should not error")
	assert.NotNil(t, info, "FileInfo should not be nil")
	assert.Equal(t, "test.txt", info.Name(), "Name should match")
	assert.Equal(t, int64(len(testData)), info.Size(), "Size should match")
	assert.False(t, info.IsDir(), "Should not be a directory")
	assert.Equal(t, os.FileMode(0644), info.Mode(), "Mode should be 0644")
	assert.Nil(t, info.Sys(), "Sys should be nil")
}

// TestCleanupTestDir tests the CleanupTestDir function
func TestCleanupTestDir(t *testing.T) {
	// Create a temp directory
	tempDir, err := ioutil.TempDir("", "test-cleanup")
	assert.NoError(t, err, "Creating temp dir should not error")

	// Verify it exists
	_, err = os.Stat(tempDir)
	assert.NoError(t, err, "Temp dir should exist")

	// Clean it up
	err = util.CleanupTestDir(tempDir)
	assert.NoError(t, err, "CleanupTestDir should not error")

	// Verify it's gone
	_, err = os.Stat(tempDir)
	assert.True(t, os.IsNotExist(err), "Temp dir should no longer exist")
}
