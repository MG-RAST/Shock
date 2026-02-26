package archive_test

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/stretchr/testify/assert"
)

// TestIsValidToArchive tests the IsValidToArchive function
func TestIsValidToArchive(t *testing.T) {
	// Test valid archive formats
	assert.True(t, archive.IsValidToArchive("zip"), "zip should be a valid archive format")
	assert.True(t, archive.IsValidToArchive("tar"), "tar should be a valid archive format")

	// Test invalid archive formats
	assert.False(t, archive.IsValidToArchive("tar.gz"), "tar.gz should not be a valid to-archive format")
	assert.False(t, archive.IsValidToArchive("tar.bz2"), "tar.bz2 should not be a valid to-archive format")
	assert.False(t, archive.IsValidToArchive("invalid"), "invalid should not be a valid archive format")
	assert.False(t, archive.IsValidToArchive(""), "empty string should not be a valid archive format")
}

// TestIsValidArchive tests the IsValidArchive function
func TestIsValidArchive(t *testing.T) {
	// Test valid archive formats
	assert.True(t, archive.IsValidArchive("zip"), "zip should be a valid archive format")
	assert.True(t, archive.IsValidArchive("tar"), "tar should be a valid archive format")
	assert.True(t, archive.IsValidArchive("tar.gz"), "tar.gz should be a valid archive format")
	assert.True(t, archive.IsValidArchive("tar.bz2"), "tar.bz2 should be a valid archive format")

	// Test invalid archive formats
	assert.False(t, archive.IsValidArchive("invalid"), "invalid should not be a valid archive format")
	assert.False(t, archive.IsValidArchive(""), "empty string should not be a valid archive format")
}

// TestIsValidUncompress tests the IsValidUncompress function
func TestIsValidUncompress(t *testing.T) {
	// Test valid uncompress formats
	assert.True(t, archive.IsValidUncompress("gzip"), "gzip should be a valid uncompress format")
	assert.True(t, archive.IsValidUncompress("bzip2"), "bzip2 should be a valid uncompress format")

	// Test invalid uncompress formats
	assert.False(t, archive.IsValidUncompress("zip"), "zip should not be a valid uncompress format")
	assert.False(t, archive.IsValidUncompress("tar"), "tar should not be a valid uncompress format")
	assert.False(t, archive.IsValidUncompress("invalid"), "invalid should not be a valid uncompress format")
	assert.False(t, archive.IsValidUncompress(""), "empty string should not be a valid uncompress format")
}

// TestIsValidCompress tests the IsValidCompress function
func TestIsValidCompress(t *testing.T) {
	// Test valid compress formats
	assert.True(t, archive.IsValidCompress("gzip"), "gzip should be a valid compress format")
	assert.True(t, archive.IsValidCompress("zip"), "zip should be a valid compress format")

	// Test invalid compress formats
	assert.False(t, archive.IsValidCompress("bzip2"), "bzip2 should not be a valid compress format")
	assert.False(t, archive.IsValidCompress("tar"), "tar should not be a valid compress format")
	assert.False(t, archive.IsValidCompress("invalid"), "invalid should not be a valid compress format")
	assert.False(t, archive.IsValidCompress(""), "empty string should not be a valid compress format")
}

// TestUncompressReader tests the UncompressReader function
func TestUncompressReader(t *testing.T) {
	// Test with gzip format
	// Create a valid gzip compressed data
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	gzipWriter.Name = "test.txt"
	_, err := gzipWriter.Write([]byte("test text"))
	assert.NoError(t, err, "Writing to gzip writer should not error")
	err = gzipWriter.Close()
	assert.NoError(t, err, "Closing gzip writer should not error")

	gzipReader := bytes.NewReader(buf.Bytes())

	reader, err := archive.UncompressReader("gzip", gzipReader)
	assert.NoError(t, err, "UncompressReader with gzip should not error")

	// Read the uncompressed data
	uncompressed, err := ioutil.ReadAll(reader)
	assert.NoError(t, err, "Reading uncompressed data should not error")
	assert.Equal(t, "test text", string(uncompressed), "Uncompressed data should match")

	// Test with bzip2 format
	// Note: Creating a valid bzip2 file programmatically is complex, so we'll skip the actual decompression test

	// Test with invalid format
	reader, err = archive.UncompressReader("invalid", bytes.NewReader([]byte("test")))
	assert.NoError(t, err, "UncompressReader with invalid format should not error")
	assert.Equal(t, bytes.NewReader([]byte("test")), reader, "Reader should be returned unchanged")

	// Test with empty format
	reader, err = archive.UncompressReader("", bytes.NewReader([]byte("test")))
	assert.NoError(t, err, "UncompressReader with empty format should not error")
	assert.Equal(t, bytes.NewReader([]byte("test")), reader, "Reader should be returned unchanged")
}

// TestCompressReader tests the CompressReader function
func TestCompressReader(t *testing.T) {
	// Test with gzip format
	testData := "test compression data"
	inReader := ioutil.NopCloser(bytes.NewReader([]byte(testData)))

	compressedReader := archive.CompressReader("gzip", "test.txt", inReader)
	assert.NotNil(t, compressedReader, "CompressReader should return a reader")

	// Read the compressed data
	compressedData, err := ioutil.ReadAll(compressedReader)
	assert.NoError(t, err, "Reading compressed data should not error")
	assert.NotEmpty(t, compressedData, "Compressed data should not be empty")

	// Test with zip format
	inReader = ioutil.NopCloser(bytes.NewReader([]byte(testData)))
	compressedReader = archive.CompressReader("zip", "test.txt", inReader)
	assert.NotNil(t, compressedReader, "CompressReader should return a reader")

	// Read the compressed data
	compressedData, err = ioutil.ReadAll(compressedReader)
	assert.NoError(t, err, "Reading compressed data should not error")
	assert.NotEmpty(t, compressedData, "Compressed data should not be empty")

	// Test with invalid format
	inReader = ioutil.NopCloser(bytes.NewReader([]byte(testData)))
	compressedReader = archive.CompressReader("invalid", "test.txt", inReader)
	assert.Equal(t, inReader, compressedReader, "CompressReader should return the input reader unchanged")
}

// TestArchiveReader tests the ArchiveReader function
func TestArchiveReader(t *testing.T) {
	// Create test files
	file1 := &file.FileInfo{
		Name:     "file1.txt",
		Size:     10,
		ModTime:  testTime,
		Body:     ioutil.NopCloser(bytes.NewReader([]byte("file1 data"))),
		Checksum: "file1checksum",
	}

	file2 := &file.FileInfo{
		Name:     "file2.txt",
		Size:     10,
		ModTime:  testTime,
		Body:     ioutil.NopCloser(bytes.NewReader([]byte("file2 data"))),
		Checksum: "file2checksum",
	}

	files := []*file.FileInfo{file1, file2}

	// Test with tar format
	tarReader := archive.ArchiveReader("tar", files)
	assert.NotNil(t, tarReader, "ArchiveReader should return a reader")

	// Read the archive data
	tarData, err := ioutil.ReadAll(tarReader)
	assert.NoError(t, err, "Reading tar archive should not error")
	assert.NotEmpty(t, tarData, "Tar archive data should not be empty")

	// Test with zip format
	// Reset the file readers
	file1.Body = ioutil.NopCloser(bytes.NewReader([]byte("file1 data")))
	file2.Body = ioutil.NopCloser(bytes.NewReader([]byte("file2 data")))

	zipReader := archive.ArchiveReader("zip", files)
	assert.NotNil(t, zipReader, "ArchiveReader should return a reader")

	// Read the archive data
	zipData, err := ioutil.ReadAll(zipReader)
	assert.NoError(t, err, "Reading zip archive should not error")
	assert.NotEmpty(t, zipData, "Zip archive data should not be empty")

	// Test with invalid format
	// Reset the file readers
	file1.Body = ioutil.NopCloser(bytes.NewReader([]byte("file1 data")))
	file2.Body = ioutil.NopCloser(bytes.NewReader([]byte("file2 data")))

	invalidReader := archive.ArchiveReader("invalid", files)
	assert.NotNil(t, invalidReader, "ArchiveReader should return a reader")

	// Read the data
	invalidData, err := ioutil.ReadAll(invalidReader)
	assert.NoError(t, err, "Reading with invalid format should not error")
	assert.Equal(t, "file1 datafile2 data", string(invalidData), "Invalid format should concatenate file data")
}

// TestFilesFromArchive tests the FilesFromArchive function
func TestFilesFromArchive(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := ioutil.TempDir("", "archive-test")
	assert.NoError(t, err, "Creating temp directory should not error")
	defer os.RemoveAll(tempDir)

	// Create a test zip file
	zipPath := filepath.Join(tempDir, "test.zip")
	createTestZipFile(t, zipPath)

	// Test extracting files from zip
	fileList, unpackDir, err := archive.FilesFromArchive("zip", zipPath)
	assert.NoError(t, err, "FilesFromArchive with zip should not error")
	defer os.RemoveAll(unpackDir) // Clean up the unpack directory

	assert.NotEmpty(t, unpackDir, "Unpack directory should not be empty")
	assert.Len(t, fileList, 2, "Should extract 2 files from zip")

	// Verify the extracted files
	var file1, file2 file.FormFile
	for _, f := range fileList {
		if f.Name == "file1.txt" {
			file1 = f
		} else if f.Name == "file2.txt" {
			file2 = f
		}
	}

	assert.NotEmpty(t, file1.Path, "File1 path should not be empty")
	assert.NotEmpty(t, file2.Path, "File2 path should not be empty")

	// Read the extracted files
	file1Content, err := ioutil.ReadFile(file1.Path)
	assert.NoError(t, err, "Reading file1 should not error")
	assert.Equal(t, "file1 content", string(file1Content), "File1 content should match")

	file2Content, err := ioutil.ReadFile(file2.Path)
	assert.NoError(t, err, "Reading file2 should not error")
	assert.Equal(t, "file2 content", string(file2Content), "File2 content should match")

	// Test with invalid format
	_, _, err = archive.FilesFromArchive("invalid", zipPath)
	assert.Error(t, err, "FilesFromArchive with invalid format should error")
	assert.Contains(t, err.Error(), "invalid archive format", "Error message should indicate invalid format")
}

// Helper function to create a test zip file
func createTestZipFile(t *testing.T, path string) {
	// Create a new zip file
	zipFile, err := os.Create(path)
	assert.NoError(t, err, "Creating zip file should not error")
	defer zipFile.Close()

	// Create a zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add file1.txt
	file1, err := zipWriter.Create("file1.txt")
	assert.NoError(t, err, "Creating file1 in zip should not error")
	_, err = file1.Write([]byte("file1 content"))
	assert.NoError(t, err, "Writing to file1 should not error")

	// Add file2.txt
	file2, err := zipWriter.Create("file2.txt")
	assert.NoError(t, err, "Creating file2 in zip should not error")
	_, err = file2.Write([]byte("file2 content"))
	assert.NoError(t, err, "Writing to file2 should not error")
}

// Helper variables
var testTime = time.Now()
