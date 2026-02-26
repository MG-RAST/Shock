package cache_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MG-RAST/Shock/shock-server/cache"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Initialize logger to prevent nil pointer panics in cache functions
	conf.LOG_OUTPUT = "console"
	logger.Initialize()
	os.Exit(m.Run())
}

// TestItem tests the Item struct
func TestItem(t *testing.T) {
	// Create a new Item
	now := time.Now()
	item := cache.Item{
		UUID:      "test_uuid",
		Access:    now,
		Type:      "file",
		Size:      1024,
		CreatedOn: now,
	}

	// Verify the fields
	assert.Equal(t, "test_uuid", item.UUID, "UUID should match")
	assert.Equal(t, now, item.Access, "Access time should match")
	assert.Equal(t, "file", item.Type, "Type should match")
	assert.Equal(t, int64(1024), item.Size, "Size should match")
	assert.Equal(t, now, item.CreatedOn, "CreatedOn should match")
}

// TestPath2uuid tests the path2uuid function
func TestPath2uuid(t *testing.T) {
	// Test with a simple path
	path := "/path/to/uuid.data"
	uuid := cache.Path2uuid(path)
	assert.Equal(t, "uuid", uuid, "UUID should be extracted correctly")

	// Test with a path containing multiple extensions
	path = "/path/to/uuid.tar.gz.data"
	uuid = cache.Path2uuid(path)
	assert.Equal(t, "uuid.tar.gz", uuid, "UUID should be extracted correctly")

	// Test with a path containing no extension
	path = "/path/to/uuid"
	uuid = cache.Path2uuid(path)
	assert.Equal(t, "uuid", uuid, "UUID should be extracted correctly")

	// Test with a path containing only a filename
	path = "uuid.data"
	uuid = cache.Path2uuid(path)
	assert.Equal(t, "uuid", uuid, "UUID should be extracted correctly")
}

// TestInitialize tests the Initialize function
func TestInitialize(t *testing.T) {
	// Save original PATH_CACHE and restore it after the test
	originalPathCache := conf.PATH_CACHE
	defer func() { conf.PATH_CACHE = originalPathCache }()

	// Test with empty PATH_CACHE
	conf.PATH_CACHE = ""
	err := cache.Initialize()
	assert.NoError(t, err, "Initialize with empty PATH_CACHE should not error")

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cache-test")
	assert.NoError(t, err, "Creating temp directory should not error")
	defer os.RemoveAll(tempDir)

	// Create a test cache structure
	cacheDir := filepath.Join(tempDir, "0", "0", "0", "0")
	err = os.MkdirAll(cacheDir, 0755)
	assert.NoError(t, err, "Creating cache directory should not error")

	// Create a test data file
	dataFile := filepath.Join(cacheDir, "test_uuid.data")
	err = os.WriteFile(dataFile, []byte("test data"), 0644)
	assert.NoError(t, err, "Creating test data file should not error")

	// Set PATH_CACHE to the temp directory
	conf.PATH_CACHE = tempDir

	// Initialize the cache
	err = cache.Initialize()
	assert.NoError(t, err, "Initialize with valid PATH_CACHE should not error")

	// Verify that the cache map contains the test UUID
	assert.Contains(t, cache.CacheMap, "test_uuid", "CacheMap should contain the test UUID")
	assert.Equal(t, "test_uuid", cache.CacheMap["test_uuid"].UUID, "UUID in CacheMap should match")
	assert.Equal(t, int64(9), cache.CacheMap["test_uuid"].Size, "Size in CacheMap should match")
}

// TestAdd tests the Add function
func TestAdd(t *testing.T) {
	// Initialize the cache map
	cache.CacheMap = make(map[string]*cache.Item)

	// Add an item to the cache
	uuid := "test_add_uuid"
	size := int64(1024)
	cache.Add(uuid, size)

	// Verify that the item was added
	assert.Contains(t, cache.CacheMap, uuid, "CacheMap should contain the added UUID")
	assert.Equal(t, uuid, cache.CacheMap[uuid].UUID, "UUID in CacheMap should match")
	assert.Equal(t, size, cache.CacheMap[uuid].Size, "Size in CacheMap should match")
	assert.WithinDuration(t, time.Now(), cache.CacheMap[uuid].CreatedOn, time.Second, "CreatedOn should be recent")
}

// TestRemove tests the Remove function
func TestRemove(t *testing.T) {
	// Save original PATH_CACHE and restore it after the test
	originalPathCache := conf.PATH_CACHE
	defer func() { conf.PATH_CACHE = originalPathCache }()

	// Test with empty PATH_CACHE
	conf.PATH_CACHE = ""
	err := cache.Remove("test_uuid")
	assert.NoError(t, err, "Remove with empty PATH_CACHE should not error")

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "cache-test")
	assert.NoError(t, err, "Creating temp directory should not error")
	defer os.RemoveAll(tempDir)

	// Set PATH_CACHE and PATH_DATA to the temp directory
	conf.PATH_CACHE = tempDir
	conf.PATH_DATA = tempDir

	// Initialize the cache map
	cache.CacheMap = make(map[string]*cache.Item)

	// Add an item to the cache
	uuid := "test_remove_uuid"
	size := int64(1024)
	cache.Add(uuid, size)

	// Verify that the item was added
	assert.Contains(t, cache.CacheMap, uuid, "CacheMap should contain the added UUID")

	// Remove the item
	err = cache.Remove(uuid)
	assert.NoError(t, err, "Remove should not error")

	// Verify that the item was removed
	assert.NotContains(t, cache.CacheMap, uuid, "CacheMap should not contain the removed UUID")
}

// TestTouch tests the Touch function
func TestTouch(t *testing.T) {
	// Initialize the cache map
	cache.CacheMap = make(map[string]*cache.Item)

	// Add an item to the cache
	uuid := "test_touch_uuid"
	size := int64(1024)
	cache.Add(uuid, size)

	// Set the access time to a past time
	pastTime := time.Now().Add(-time.Hour)
	cache.CacheMap[uuid].Access = pastTime

	// Touch the item
	cache.Touch(uuid)

	// Verify that the access time was updated
	assert.WithinDuration(t, time.Now(), cache.CacheMap[uuid].Access, time.Second, "Access time should be updated")

	// Test touching a non-existent item
	cache.Touch("non_existent_uuid")
	// This should not error or panic
}

// TestConcurrentAccess tests concurrent access to the cache
func TestConcurrentAccess(t *testing.T) {
	// Initialize the cache map
	cache.CacheMap = make(map[string]*cache.Item)

	// Add an item to the cache
	uuid := "test_concurrent_uuid"
	size := int64(1024)
	cache.Add(uuid, size)

	// Concurrently touch and remove the item
	done := make(chan bool)
	go func() {
		cache.Touch(uuid)
		done <- true
	}()
	go func() {
		cache.Remove(uuid)
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// This test is mainly to ensure that there are no race conditions or panics
}
