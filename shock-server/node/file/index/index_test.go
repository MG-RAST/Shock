package index_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/stretchr/testify/assert"
)

// TestNew tests creating a new index
func TestNew(t *testing.T) {
	idx := index.New()

	assert.NotNil(t, idx, "New index should not be nil")
	assert.Equal(t, "file", idx.Type(), "Default index type should be 'file'")
	assert.Equal(t, int64(0), idx.GetLength(), "Default index length should be 0")
}

// TestSet tests setting index properties
func TestSet(t *testing.T) {
	idx := index.New()

	// Set properties
	props := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	// The Set method is a no-op in the base Idx struct, so we're just testing that it doesn't panic
	idx.Set(props)
}

// TestType tests getting the index type
func TestType(t *testing.T) {
	idx := index.New()

	assert.Equal(t, "file", idx.Type(), "Index type should be 'file'")
}

// TestGetLength tests getting the index length
func TestGetLength(t *testing.T) {
	idx := index.New()

	assert.Equal(t, int64(0), idx.GetLength(), "Index length should be 0")
}

// TestPart tests the Part method
func TestPart(t *testing.T) {
	// Create a temporary index file
	tmpfile, err := ioutil.TempFile("", "index_test")
	assert.NoError(t, err, "Creating temp file should not error")
	defer os.Remove(tmpfile.Name())

	// Write test index data (2 records, each with offset and length)
	// Record 1: offset=100, length=50
	// Record 2: offset=200, length=75
	data := []byte{
		// Record 1
		100, 0, 0, 0, 0, 0, 0, 0, // offset (little-endian int64)
		50, 0, 0, 0, 0, 0, 0, 0, // length (little-endian int64)
		// Record 2
		200, 0, 0, 0, 0, 0, 0, 0, // offset (little-endian int64)
		75, 0, 0, 0, 0, 0, 0, 0, // length (little-endian int64)
	}
	_, err = tmpfile.Write(data)
	assert.NoError(t, err, "Writing to temp file should not error")
	tmpfile.Close()

	idx := index.New()

	// Test getting a single part
	pos, length, err := idx.Part("1", tmpfile.Name(), 2)
	assert.NoError(t, err, "Getting part should not error")
	assert.Equal(t, int64(100), pos, "Position should match first record offset")
	assert.Equal(t, int64(50), length, "Length should match first record length")

	// Test getting a range of parts
	pos, length, err = idx.Part("1-2", tmpfile.Name(), 2)
	assert.NoError(t, err, "Getting part range should not error")
	assert.Equal(t, int64(100), pos, "Position should match first record offset")
	assert.Equal(t, int64(175), length, "Length should include both records")

	// Test invalid part number
	_, _, err = idx.Part("3", tmpfile.Name(), 2)
	assert.Error(t, err, "Getting invalid part should error")
	assert.Contains(t, err.Error(), "Index record out of bounds", "Error message should indicate out of bounds")

	// Test invalid part format
	_, _, err = idx.Part("invalid", tmpfile.Name(), 2)
	assert.Error(t, err, "Getting part with invalid format should error")
	assert.Contains(t, err.Error(), "Index record out of bounds", "Error message should indicate out of bounds")

	// Test non-existent file
	_, _, err = idx.Part("1", "non_existent_file", 2)
	assert.Error(t, err, "Getting part from non-existent file should error")
	assert.Contains(t, err.Error(), "Index file is missing", "Error message should indicate no index file")
}

// TestRange tests the Range method
func TestRange(t *testing.T) {
	// Create a temporary index file
	tmpfile, err := ioutil.TempFile("", "index_test")
	assert.NoError(t, err, "Creating temp file should not error")
	defer os.Remove(tmpfile.Name())

	// Write test index data (3 records, each with offset and length)
	// Record 1: offset=100, length=50
	// Record 2: offset=150, length=50 (contiguous with record 1)
	// Record 3: offset=300, length=75 (not contiguous)
	data := []byte{
		// Record 1
		100, 0, 0, 0, 0, 0, 0, 0, // offset (little-endian int64)
		50, 0, 0, 0, 0, 0, 0, 0, // length (little-endian int64)
		// Record 2
		150, 0, 0, 0, 0, 0, 0, 0, // offset (little-endian int64)
		50, 0, 0, 0, 0, 0, 0, 0, // length (little-endian int64)
		// Record 3: 300 = 0x12C = 44 + 1*256
		44, 1, 0, 0, 0, 0, 0, 0, // offset (little-endian int64)
		75, 0, 0, 0, 0, 0, 0, 0, // length (little-endian int64)
	}
	_, err = tmpfile.Write(data)
	assert.NoError(t, err, "Writing to temp file should not error")
	tmpfile.Close()

	idx := index.New()

	// Test getting a single part
	ranges, err := idx.Range("1", tmpfile.Name(), 3)
	assert.NoError(t, err, "Getting range should not error")
	assert.Len(t, ranges, 1, "Should return 1 range for a single part")
	assert.Equal(t, []int64{100, 50}, ranges[0], "Range should match first record")

	// Test getting a range of contiguous parts
	ranges, err = idx.Range("1-2", tmpfile.Name(), 3)
	assert.NoError(t, err, "Getting range of contiguous parts should not error")
	assert.Len(t, ranges, 1, "Should return 1 range for contiguous parts")
	assert.Equal(t, []int64{100, 100}, ranges[0], "Range should combine contiguous parts")

	// Test getting a range of non-contiguous parts
	ranges, err = idx.Range("1-3", tmpfile.Name(), 3)
	assert.NoError(t, err, "Getting range of non-contiguous parts should not error")
	assert.Len(t, ranges, 2, "Should return 2 ranges for non-contiguous parts")
	assert.Equal(t, []int64{100, 100}, ranges[0], "First range should combine contiguous parts")
	assert.Equal(t, []int64{300, 75}, ranges[1], "Second range should match third record")

	// Test invalid part number (out of range)
	_, err = idx.Range("4", tmpfile.Name(), 3)
	assert.Error(t, err, "Getting invalid range should error")
	assert.Contains(t, err.Error(), "Index record out of bounds", "Error message should indicate out of bounds")

	// Test invalid part format
	_, err = idx.Range("invalid", tmpfile.Name(), 3)
	assert.Error(t, err, "Getting range with invalid format should error")
	assert.Contains(t, err.Error(), "Index record out of bounds", "Error message should indicate out of bounds")

	// Test invalid range (end exceeds bounds)
	_, err = idx.Range("1-4", tmpfile.Name(), 3)
	assert.Error(t, err, "Getting range exceeding bounds should error")
	assert.Contains(t, err.Error(), "Invalid index record range", "Error message should indicate invalid range")

	// Test non-existent file
	_, err = idx.Range("1", "non_existent_file", 3)
	assert.Error(t, err, "Getting range from non-existent file should error")
	assert.Contains(t, err.Error(), "Index file is missing", "Error message should indicate no index file")
}

// TestIndexers tests the Indexers map
func TestIndexers(t *testing.T) {
	// Verify that the Indexers map contains the expected indexers
	assert.Contains(t, index.Indexers, "chunkrecord", "Indexers map should contain chunkrecord")
	assert.Contains(t, index.Indexers, "line", "Indexers map should contain line")
	assert.Contains(t, index.Indexers, "record", "Indexers map should contain record")
	assert.Contains(t, index.Indexers, "size", "Indexers map should contain size")

	// Verify that the indexer functions return non-nil values
	tmpfile, err := ioutil.TempFile("", "indexer_test")
	assert.NoError(t, err, "Creating temp file should not error")
	defer os.Remove(tmpfile.Name())

	for name, indexerFunc := range index.Indexers {
		indexer := indexerFunc(tmpfile, "basic", "", "")
		assert.NotNil(t, indexer, "Indexer function %s should return non-nil value", name)
	}
}
