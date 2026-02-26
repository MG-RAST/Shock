package file_test

import (
	"io"
	"os"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/test/file"
	"github.com/stretchr/testify/assert"
)

// TestNewMockFileSystem tests creating a new mock file system
func TestNewMockFileSystem(t *testing.T) {
	fs := file.NewMockFileSystem()
	assert.NotNil(t, fs, "Mock file system should not be nil")
}

// TestMockFileSystemAddAndOpen tests adding a file and opening it
func TestMockFileSystemAddAndOpen(t *testing.T) {
	fs := file.NewMockFileSystem()
	testData := []byte("test data")

	// Add a file
	fs.AddFile("/test/file.txt", testData)

	// Open the file
	reader, err := fs.Open("/test/file.txt")
	assert.NoError(t, err, "Opening existing file should not error")
	assert.NotNil(t, reader, "Reader should not be nil")

	// Read the content
	buf := make([]byte, len(testData))
	n, err := reader.Read(buf)
	assert.NoError(t, err, "Reading should not error")
	assert.Equal(t, len(testData), n, "Should read the correct number of bytes")
	assert.Equal(t, testData, buf, "Read data should match")

	// Close the reader
	err = reader.Close()
	assert.NoError(t, err, "Closing should not error")
}

// TestMockFileSystemOpenNonexistent tests opening a non-existent file
func TestMockFileSystemOpenNonexistent(t *testing.T) {
	fs := file.NewMockFileSystem()

	_, err := fs.Open("/nonexistent")
	assert.ErrorIs(t, err, os.ErrNotExist, "Opening non-existent file should return ErrNotExist")
}

// TestMockFileSystemCreate tests creating a file via the Create method
func TestMockFileSystemCreate(t *testing.T) {
	fs := file.NewMockFileSystem()

	// Create a file
	writer, err := fs.Create("/test/new_file.txt")
	assert.NoError(t, err, "Creating file should not error")
	assert.NotNil(t, writer, "Writer should not be nil")

	// Write data
	testData := []byte("hello world")
	n, err := writer.Write(testData)
	assert.NoError(t, err, "Writing should not error")
	assert.Equal(t, len(testData), n, "Should write the correct number of bytes")

	// Close the writer (this persists the data)
	err = writer.Close()
	assert.NoError(t, err, "Closing writer should not error")

	// Read the file back
	content, err := fs.ReadFile("/test/new_file.txt")
	assert.NoError(t, err, "Reading created file should not error")
	assert.Equal(t, testData, content, "Content should match what was written")
}

// TestMockFileSystemRemove tests removing a file
func TestMockFileSystemRemove(t *testing.T) {
	fs := file.NewMockFileSystem()
	fs.AddFile("/test/file.txt", []byte("data"))

	// Remove the file
	err := fs.Remove("/test/file.txt")
	assert.NoError(t, err, "Removing existing file should not error")

	// Verify the file is gone
	_, err = fs.Open("/test/file.txt")
	assert.ErrorIs(t, err, os.ErrNotExist, "Opening removed file should return ErrNotExist")

	// Remove non-existent file
	err = fs.Remove("/nonexistent")
	assert.ErrorIs(t, err, os.ErrNotExist, "Removing non-existent file should return ErrNotExist")
}

// TestMockFileSystemStat tests getting file info
func TestMockFileSystemStat(t *testing.T) {
	fs := file.NewMockFileSystem()
	testData := []byte("test data content")
	fs.AddFile("/test/file.txt", testData)

	// Stat the file
	info, err := fs.Stat("/test/file.txt")
	assert.NoError(t, err, "Stat should not error")
	assert.NotNil(t, info, "FileInfo should not be nil")
	assert.Equal(t, "file.txt", info.Name(), "Name should match basename")
	assert.Equal(t, int64(len(testData)), info.Size(), "Size should match content length")
	assert.False(t, info.IsDir(), "Should not be a directory")
	assert.Equal(t, os.FileMode(0644), info.Mode(), "Mode should be 0644")

	// Stat a non-existent file
	_, err = fs.Stat("/nonexistent")
	assert.ErrorIs(t, err, os.ErrNotExist, "Stat on non-existent file should return ErrNotExist")
}

// TestMockFileSystemMkdirAll tests creating directories
func TestMockFileSystemMkdirAll(t *testing.T) {
	fs := file.NewMockFileSystem()

	err := fs.MkdirAll("/test/dir/subdir", 0755)
	assert.NoError(t, err, "MkdirAll should not error")
}

// TestMockFileSystemReadFile tests the ReadFile method
func TestMockFileSystemReadFile(t *testing.T) {
	fs := file.NewMockFileSystem()
	testData := []byte("file content")
	fs.AddFile("/test/file.txt", testData)

	// Read existing file
	content, err := fs.ReadFile("/test/file.txt")
	assert.NoError(t, err, "ReadFile should not error")
	assert.Equal(t, testData, content, "Content should match")

	// Read non-existent file
	_, err = fs.ReadFile("/nonexistent")
	assert.ErrorIs(t, err, os.ErrNotExist, "ReadFile on non-existent file should return ErrNotExist")
}

// TestMockFileSystemWriteFile tests the WriteFile method
func TestMockFileSystemWriteFile(t *testing.T) {
	fs := file.NewMockFileSystem()
	testData := []byte("written data")

	// Write a file
	err := fs.WriteFile("/test/file.txt", testData, 0644)
	assert.NoError(t, err, "WriteFile should not error")

	// Read it back
	content, err := fs.ReadFile("/test/file.txt")
	assert.NoError(t, err, "ReadFile should not error")
	assert.Equal(t, testData, content, "Content should match")

	// Overwrite the file
	newData := []byte("new data")
	err = fs.WriteFile("/test/file.txt", newData, 0644)
	assert.NoError(t, err, "WriteFile overwrite should not error")

	content, err = fs.ReadFile("/test/file.txt")
	assert.NoError(t, err, "ReadFile should not error")
	assert.Equal(t, newData, content, "Content should match the new data")
}

// TestMockFileSystemReset tests resetting the mock file system
func TestMockFileSystemReset(t *testing.T) {
	fs := file.NewMockFileSystem()
	fs.AddFile("/test/file1.txt", []byte("data1"))
	fs.AddFile("/test/file2.txt", []byte("data2"))

	// Reset
	fs.Reset()

	// Verify files are gone
	_, err := fs.Open("/test/file1.txt")
	assert.ErrorIs(t, err, os.ErrNotExist, "File should not exist after reset")

	_, err = fs.Open("/test/file2.txt")
	assert.ErrorIs(t, err, os.ErrNotExist, "File should not exist after reset")
}

// TestMockFileSystemOpenReadAll tests reading entire file content via Open
func TestMockFileSystemOpenReadAll(t *testing.T) {
	fs := file.NewMockFileSystem()
	testData := []byte("complete file content to read")
	fs.AddFile("/test/file.txt", testData)

	reader, err := fs.Open("/test/file.txt")
	assert.NoError(t, err, "Open should not error")

	content, err := io.ReadAll(reader)
	assert.NoError(t, err, "ReadAll should not error")
	assert.Equal(t, testData, content, "Content should match")

	err = reader.Close()
	assert.NoError(t, err, "Close should not error")
}
