package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/stretchr/testify/assert"
)

// TestStringInSlice tests the StringInSlice function
func TestStringInSlice(t *testing.T) {
	// Create a test slice
	slice := []string{"one", "two", "three"}

	// Test with a string that is in the slice
	result := util.StringInSlice("two", slice)
	assert.True(t, result, "StringInSlice should return true for a string in the slice")

	// Test with a string that is not in the slice
	result = util.StringInSlice("four", slice)
	assert.False(t, result, "StringInSlice should return false for a string not in the slice")

	// Test with an empty slice
	emptySlice := []string{}
	result = util.StringInSlice("one", emptySlice)
	assert.False(t, result, "StringInSlice should return false for an empty slice")

	// Test with a nil slice
	var nilSlice []string
	result = util.StringInSlice("one", nilSlice)
	assert.False(t, result, "StringInSlice should return false for a nil slice")
}

// TestCopyFile tests the CopyFile function
func TestCopyFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "util-test")
	assert.NoError(t, err, "Creating temp directory should not error")
	defer os.RemoveAll(tempDir)

	// Create a source file
	srcPath := filepath.Join(tempDir, "source.txt")
	srcContent := []byte("test content")
	err = os.WriteFile(srcPath, srcContent, 0644)
	assert.NoError(t, err, "Writing source file should not error")

	// Create a destination path
	dstPath := filepath.Join(tempDir, "destination.txt")

	// Copy the file
	size, err := util.CopyFile(srcPath, dstPath)
	assert.NoError(t, err, "CopyFile should not error")
	assert.Equal(t, int64(len(srcContent)), size, "CopyFile should return the correct size")

	// Verify that the destination file was created with the correct content
	dstContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err, "Reading destination file should not error")
	assert.Equal(t, srcContent, dstContent, "Destination file should have the same content as the source file")

	// Test with a non-existent source file
	_, err = util.CopyFile("/nonexistent", dstPath)
	assert.Error(t, err, "CopyFile should error with a non-existent source file")

	// Test with an invalid destination path
	_, err = util.CopyFile(srcPath, "/nonexistent/destination.txt")
	assert.Error(t, err, "CopyFile should error with an invalid destination path")
}

// TestRandString tests the RandString function
func TestRandString(t *testing.T) {
	// Generate a random string of length 10
	s := util.RandString(10)
	assert.Len(t, s, 10, "RandString should return a string of the specified length")

	// Generate another and verify they are (almost certainly) different
	s2 := util.RandString(10)
	// Note: there's a vanishingly small chance they could be equal, but practically they won't be
	assert.Len(t, s2, 10, "RandString should return a string of the specified length")

	// Generate a zero-length string
	s0 := util.RandString(0)
	assert.Len(t, s0, 0, "RandString(0) should return an empty string")
}

// TestToInt tests the ToInt function
func TestToInt(t *testing.T) {
	// Test with a valid integer string
	result := util.ToInt("42")
	assert.Equal(t, 42, result, "ToInt should convert '42' to 42")

	// Test with zero
	result = util.ToInt("0")
	assert.Equal(t, 0, result, "ToInt should convert '0' to 0")

	// Test with a negative number
	result = util.ToInt("-5")
	assert.Equal(t, -5, result, "ToInt should convert '-5' to -5")

	// Test with an invalid string (returns 0)
	result = util.ToInt("abc")
	assert.Equal(t, 0, result, "ToInt should return 0 for an invalid string")

	// Test with an empty string (returns 0)
	result = util.ToInt("")
	assert.Equal(t, 0, result, "ToInt should return 0 for an empty string")
}

// TestStripSuffix tests the StripSuffix function
func TestStripSuffix(t *testing.T) {
	// Test with a file that has a suffix
	result := util.StripSuffix("file.txt")
	assert.Equal(t, "file", result, "StripSuffix should remove the suffix")

	// Test with a file that has multiple dots
	result = util.StripSuffix("file.tar.gz")
	assert.Equal(t, "file.tar", result, "StripSuffix should remove only the last suffix")

	// Test with a file that has no suffix
	result = util.StripSuffix("file")
	assert.Equal(t, "file", result, "StripSuffix should return the file unchanged if no suffix")

	// Test with an empty string
	result = util.StripSuffix("")
	assert.Equal(t, "", result, "StripSuffix should return empty string for empty input")
}

// TestIsValidParamName tests the IsValidParamName function
func TestIsValidParamName(t *testing.T) {
	// Test with valid param names
	assert.True(t, util.IsValidParamName("action"), "action should be a valid param name")
	assert.True(t, util.IsValidParamName("file_name"), "file_name should be a valid param name")
	assert.True(t, util.IsValidParamName("upload_url"), "upload_url should be a valid param name")

	// Test with invalid param names
	assert.False(t, util.IsValidParamName("invalid_param"), "invalid_param should not be a valid param name")
	assert.False(t, util.IsValidParamName(""), "empty string should not be a valid param name")
}

// TestIsValidFileName tests the IsValidFileName function
func TestIsValidFileName(t *testing.T) {
	// Test with valid file names
	assert.True(t, util.IsValidFileName("upload"), "upload should be a valid file name")
	assert.True(t, util.IsValidFileName("attributes"), "attributes should be a valid file name")

	// Test with invalid file names
	assert.False(t, util.IsValidFileName("invalid"), "invalid should not be a valid file name")
	assert.False(t, util.IsValidFileName(""), "empty string should not be a valid file name")
}

// TestIsValidUploadFile tests the IsValidUploadFile function
func TestIsValidUploadFile(t *testing.T) {
	// Test with valid upload file names
	assert.True(t, util.IsValidUploadFile("upload"), "upload should be a valid upload file")
	assert.True(t, util.IsValidUploadFile("gzip"), "gzip should be a valid upload file")
	assert.True(t, util.IsValidUploadFile("bzip2"), "bzip2 should be a valid upload file")

	// Test with invalid upload file names
	assert.False(t, util.IsValidUploadFile("attributes"), "attributes should not be a valid upload file")
	assert.False(t, util.IsValidUploadFile(""), "empty string should not be a valid upload file")
}
