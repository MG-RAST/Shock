package file_test

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/stretchr/testify/assert"
)

// TestFormFile tests the FormFile struct and its methods
func TestFormFile(t *testing.T) {
	// Create a temporary file for testing
	tempDir, err := ioutil.TempDir("", "shock-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testContent := []byte("test file content")
	testFilePath := tempDir + "/test_file.txt"
	err = ioutil.WriteFile(testFilePath, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a FormFile
	formFile := file.FormFile{
		Name: "test_file.txt",
		Path: testFilePath,
		Checksum: map[string]string{
			"md5": "test_checksum",
		},
	}

	// Verify FormFile properties
	assert.Equal(t, "test_file.txt", formFile.Name, "FormFile name should match")
	assert.Equal(t, testFilePath, formFile.Path, "FormFile path should match")
	assert.Equal(t, "test_checksum", formFile.Checksum["md5"], "FormFile checksum should match")

	// Test Remove method
	formFile.Remove()
	_, err = os.Stat(testFilePath)
	assert.True(t, os.IsNotExist(err), "File should be removed")
}

// TestFormFiles tests the FormFiles map and its methods
func TestFormFiles(t *testing.T) {
	// Create temporary files for testing
	tempDir, err := ioutil.TempDir("", "shock-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile1Path := tempDir + "/test_file1.txt"
	testFile2Path := tempDir + "/test_file2.txt"
	err = ioutil.WriteFile(testFile1Path, []byte("test file 1 content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file 1: %v", err)
	}
	err = ioutil.WriteFile(testFile2Path, []byte("test file 2 content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file 2: %v", err)
	}

	// Create FormFiles
	formFiles := file.FormFiles{
		"file1": file.FormFile{
			Name: "test_file1.txt",
			Path: testFile1Path,
		},
		"file2": file.FormFile{
			Name: "test_file2.txt",
			Path: testFile2Path,
		},
	}

	// Verify FormFiles
	assert.Equal(t, 2, len(formFiles), "FormFiles should have 2 entries")
	assert.Equal(t, "test_file1.txt", formFiles["file1"].Name, "FormFile 1 name should match")
	assert.Equal(t, "test_file2.txt", formFiles["file2"].Name, "FormFile 2 name should match")

	// Test RemoveAllFormFiles
	file.RemoveAllFormFiles(formFiles)
	_, err = os.Stat(testFile1Path)
	assert.True(t, os.IsNotExist(err), "File 1 should be removed")
	_, err = os.Stat(testFile2Path)
	assert.True(t, os.IsNotExist(err), "File 2 should be removed")
}

// TestMultiReaderAt tests the MultiReaderAt implementation
func TestMultiReaderAt(t *testing.T) {
	// Create test data
	data1 := []byte("first part of data")
	data2 := []byte("second part of data")
	data3 := []byte("third part of data")
	allData := append(append(append([]byte{}, data1...), data2...), data3...)

	// Test Read method - reads sequentially through all readers via io.ReadAll
	// Note: Read() consumes readers, so we create a separate MultiReaderAt for this
	t.Run("ReadAll", func(t *testing.T) {
		r1 := file.NewTestReaderAt("file1.txt", data1)
		r2 := file.NewTestReaderAt("file2.txt", data2)
		r3 := file.NewTestReaderAt("file3.txt", data3)
		mr := file.MultiReaderAt(r1, r2, r3)

		allBytes, err := io.ReadAll(mr)
		assert.NoError(t, err, "ReadAll should not error")
		assert.Equal(t, allData, allBytes, "ReadAll should return all concatenated data")
	})

	// Test ReadAt method (does not consume readers like Read does)
	t.Run("ReadAt", func(t *testing.T) {
		r1 := file.NewTestReaderAt("file1.txt", data1)
		r2 := file.NewTestReaderAt("file2.txt", data2)
		r3 := file.NewTestReaderAt("file3.txt", data3)
		mr := file.MultiReaderAt(r1, r2, r3)

		// Read from the beginning
		buffer := make([]byte, 10)
		n, err := mr.ReadAt(buffer, 0)
		assert.NoError(t, err, "ReadAt should not error")
		assert.Equal(t, 10, n, "ReadAt should return requested amount")
		assert.Equal(t, data1[:10], buffer, "ReadAt data should match")

		// Read across boundaries
		buffer = make([]byte, 10)
		n, err = mr.ReadAt(buffer, int64(len(data1)-5))
		assert.NoError(t, err, "ReadAt should not error")
		assert.Equal(t, 10, n, "ReadAt should return requested amount")
		expected := append(append([]byte{}, data1[len(data1)-5:]...), data2[:5]...)
		assert.Equal(t, expected, buffer, "ReadAt data should match")

		// Read from the end
		buffer = make([]byte, 5)
		n, err = mr.ReadAt(buffer, int64(len(data1)+len(data2)+len(data3)-5))
		assert.NoError(t, err, "ReadAt should not error")
		assert.Equal(t, 5, n, "ReadAt should return requested amount")
		assert.Equal(t, data3[len(data3)-5:], buffer, "ReadAt data should match")

		// Read beyond the end
		buffer = make([]byte, 10)
		n, err = mr.ReadAt(buffer, int64(len(data1)+len(data2)+len(data3)))
		assert.Equal(t, io.EOF, err, "ReadAt beyond end should return EOF")
		assert.Equal(t, 0, n, "ReadAt beyond end should return 0 bytes")
	})
}

// Note: testReaderAt was moved to file.go as an exported function NewTestReaderAt.
// Tests above use file.NewTestReaderAt directly.
