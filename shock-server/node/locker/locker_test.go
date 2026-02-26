package locker_test

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// Initialize logger to prevent nil pointer panics in locker functions
	conf.LOG_OUTPUT = "console"
	logger.Initialize()
	os.Exit(m.Run())
}

// TestNodeLockerLockUnlock tests locking and unlocking nodes
func TestNodeLockerLockUnlock(t *testing.T) {
	// Create a new node locker
	nodeLockMgr := locker.NewNodeLocker()

	// Test locking a node
	nodeID := "test_node_1"
	err := nodeLockMgr.LockNode(nodeID)
	assert.NoError(t, err, "Locking node should not error")

	// Verify the node is locked
	lockedNodes := nodeLockMgr.GetLocked()
	assert.Len(t, lockedNodes, 1, "Should have 1 locked node")
	assert.Equal(t, nodeID, lockedNodes[0].Id, "Locked node ID should match")
	assert.True(t, lockedNodes[0].IsLocked, "Node should be locked")

	// Test unlocking the node
	nodeLockMgr.UnlockNode(nodeID)

	// Verify the node is unlocked
	lockedNodes = nodeLockMgr.GetLocked()
	assert.Len(t, lockedNodes, 0, "Should have 0 locked nodes after unlocking")
}

// TestNodeLockerConcurrentAccess tests concurrent access to the node locker
func TestNodeLockerConcurrentAccess(t *testing.T) {
	// Create a new node locker
	nodeLockMgr := locker.NewNodeLocker()

	nodeID := "test_node_2"

	// Lock the node
	err := nodeLockMgr.LockNode(nodeID)
	assert.NoError(t, err, "Locking node should not error")

	// Verify it's locked
	lockedNodes := nodeLockMgr.GetLocked()
	assert.Len(t, lockedNodes, 1, "Should have 1 locked node")

	// Start a goroutine that will try to acquire the lock (it will block until we unlock)
	acquired := make(chan bool, 1)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := nodeLockMgr.LockNode(nodeID)
		assert.NoError(t, err, "Locking node after unlock should not error")
		acquired <- true
	}()

	// Give the goroutine time to start and block on the lock
	time.Sleep(100 * time.Millisecond)

	// Unlock the node so the goroutine can acquire it
	nodeLockMgr.UnlockNode(nodeID)

	// Wait for the goroutine to acquire the lock
	select {
	case <-acquired:
		// Expected - goroutine acquired the lock after we released it
	case <-time.After(5 * time.Second):
		t.Fatal("Timed out waiting for goroutine to acquire lock")
	}

	wg.Wait()

	// Verify the node is locked again (by the goroutine)
	lockedNodes = nodeLockMgr.GetLocked()
	assert.Len(t, lockedNodes, 1, "Should have 1 locked node")
	assert.Equal(t, nodeID, lockedNodes[0].Id, "Locked node ID should match")

	// Clean up
	nodeLockMgr.UnlockNode(nodeID)
}

// TestNodeLockerGetAll tests getting all nodes
func TestNodeLockerGetAll(t *testing.T) {
	// Create a new node locker
	nodeLockMgr := locker.NewNodeLocker()

	// Add some nodes
	nodeID1 := "test_node_3"
	nodeID2 := "test_node_4"

	nodeLockMgr.Add(nodeID1)
	nodeLockMgr.Add(nodeID2)

	// Lock one of the nodes
	err := nodeLockMgr.LockNode(nodeID1)
	assert.NoError(t, err, "Locking node should not error")

	// Get all nodes
	allNodes := nodeLockMgr.GetAll()
	assert.Len(t, allNodes, 2, "Should have 2 nodes")

	// Verify the nodes
	var node1, node2 *locker.NodeLock
	for _, node := range allNodes {
		if node.Id == nodeID1 {
			node1 = node
		} else if node.Id == nodeID2 {
			node2 = node
		}
	}

	assert.NotNil(t, node1, "Node 1 should be in the list")
	assert.NotNil(t, node2, "Node 2 should be in the list")
	assert.True(t, node1.IsLocked, "Node 1 should be locked")
	assert.False(t, node2.IsLocked, "Node 2 should not be locked")

	// Clean up
	nodeLockMgr.UnlockNode(nodeID1)
}

// TestNodeLockerRemove tests removing nodes
func TestNodeLockerRemove(t *testing.T) {
	// Create a new node locker
	nodeLockMgr := locker.NewNodeLocker()

	// Add a node
	nodeID := "test_node_5"
	nodeLockMgr.Add(nodeID)

	// Verify the node was added
	allNodes := nodeLockMgr.GetAll()
	assert.Len(t, allNodes, 1, "Should have 1 node")

	// Remove the node
	nodeLockMgr.Remove(nodeID)

	// Verify the node was removed
	allNodes = nodeLockMgr.GetAll()
	assert.Len(t, allNodes, 0, "Should have 0 nodes after removal")
}

// TestNodeLockerRemoveOld tests removing old nodes
func TestNodeLockerRemoveOld(t *testing.T) {
	// Create a new node locker
	nodeLockMgr := locker.NewNodeLocker()

	// Add some nodes
	nodeID1 := "test_node_6"
	nodeID2 := "test_node_7"

	nodeLockMgr.Add(nodeID1)
	nodeLockMgr.Add(nodeID2)

	// Lock one of the nodes
	err := nodeLockMgr.LockNode(nodeID1)
	assert.NoError(t, err, "Locking node should not error")

	// Wait a bit to ensure the nodes have different update times
	time.Sleep(100 * time.Millisecond)

	// Unlock the node
	nodeLockMgr.UnlockNode(nodeID1)

	// Remove old nodes (with a very short expiration time)
	nodeLockMgr.RemoveOld(0)

	// Verify the nodes were removed
	allNodes := nodeLockMgr.GetAll()
	assert.Len(t, allNodes, 0, "Should have 0 nodes after removing old nodes")
}

// TestFileLockerGetAll tests getting all file locks
func TestFileLockerGetAll(t *testing.T) {
	// Create a new file locker
	fileLockMgr := locker.NewFileLocker()

	// Add some file locks
	fileID1 := "test_file_1"
	fileID2 := "test_file_2"

	fileLockMgr.Add(fileID1)
	fileLockMgr.Add(fileID2)

	// Get all file locks
	allLocks := fileLockMgr.GetAll()
	assert.Len(t, allLocks, 2, "Should have 2 file locks")
	assert.Contains(t, allLocks, fileID1, "File lock 1 should be in the map")
	assert.Contains(t, allLocks, fileID2, "File lock 2 should be in the map")
}

// TestFileLockerGet tests getting a specific file lock
func TestFileLockerGet(t *testing.T) {
	// Create a new file locker
	fileLockMgr := locker.NewFileLocker()

	// Add a file lock
	fileID := "test_file_3"
	fileLockMgr.Add(fileID)

	// Get the file lock
	lock := fileLockMgr.Get(fileID)
	assert.NotNil(t, lock, "File lock should not be nil")
	assert.WithinDuration(t, time.Now(), lock.CreatedOn, 1*time.Second, "CreatedOn should be recent")

	// Get a non-existent file lock
	nonExistentLock := fileLockMgr.Get("non_existent_file")
	assert.Nil(t, nonExistentLock, "Non-existent file lock should be nil")
}

// TestFileLockerError tests setting an error on a file lock
func TestFileLockerError(t *testing.T) {
	// Create a new file locker
	fileLockMgr := locker.NewFileLocker()

	// Add a file lock
	fileID := "test_file_4"
	fileLockMgr.Add(fileID)

	// Set an error
	fileLockMgr.Error(fileID, assert.AnError)

	// Verify the error was set
	lock := fileLockMgr.Get(fileID)
	assert.NotNil(t, lock, "File lock should not be nil")
	assert.NotEmpty(t, lock.Error, "Error should not be empty")

	// Test with nil error
	fileLockMgr.Error(fileID, nil)

	// Test with non-existent file
	fileLockMgr.Error("non_existent_file", assert.AnError)
}

// TestFileLockerRemove tests removing a file lock
func TestFileLockerRemove(t *testing.T) {
	// Create a new file locker
	fileLockMgr := locker.NewFileLocker()

	// Add a file lock
	fileID := "test_file_5"
	fileLockMgr.Add(fileID)

	// Verify the file lock was added
	lock := fileLockMgr.Get(fileID)
	assert.NotNil(t, lock, "File lock should not be nil")

	// Remove the file lock
	fileLockMgr.Remove(fileID)

	// Verify the file lock was removed
	lock = fileLockMgr.Get(fileID)
	assert.Nil(t, lock, "File lock should be nil after removal")
}

// TestFileLockerRemoveOld tests removing old file locks
func TestFileLockerRemoveOld(t *testing.T) {
	// Create a new file locker
	fileLockMgr := locker.NewFileLocker()

	// Add a file lock
	fileID := "test_file_6"
	fileLockMgr.Add(fileID)

	// Verify the file lock was added
	lock := fileLockMgr.Get(fileID)
	assert.NotNil(t, lock, "File lock should not be nil")

	// Remove old file locks (with a very short expiration time)
	fileLockMgr.RemoveOld(0)

	// Verify the file lock was removed
	lock = fileLockMgr.Get(fileID)
	assert.Nil(t, lock, "File lock should be nil after removing old locks")
}

// TestIndexLockerGetAll tests getting all index locks
func TestIndexLockerGetAll(t *testing.T) {
	// Create a new index locker
	indexLockMgr := locker.NewIndexLocker()

	// Add some index locks
	nodeID1 := "test_node_8"
	indexName1 := "test_index_1"
	nodeID2 := "test_node_9"
	indexName2 := "test_index_2"

	indexLockMgr.Add(nodeID1, indexName1)
	indexLockMgr.Add(nodeID2, indexName2)

	// Get all index locks
	allLocks := indexLockMgr.GetAll()
	assert.Len(t, allLocks, 2, "Should have 2 nodes with index locks")
	assert.Contains(t, allLocks, nodeID1, "Node 1 should be in the map")
	assert.Contains(t, allLocks, nodeID2, "Node 2 should be in the map")
	assert.Contains(t, allLocks[nodeID1], indexName1, "Index 1 should be in the map")
	assert.Contains(t, allLocks[nodeID2], indexName2, "Index 2 should be in the map")
}

// TestIndexLockerGet tests getting a specific index lock
func TestIndexLockerGet(t *testing.T) {
	// Create a new index locker
	indexLockMgr := locker.NewIndexLocker()

	// Add an index lock
	nodeID := "test_node_10"
	indexName := "test_index_3"
	indexLockMgr.Add(nodeID, indexName)

	// Get the index lock
	lock := indexLockMgr.Get(nodeID, indexName)
	assert.NotNil(t, lock, "Index lock should not be nil")
	assert.WithinDuration(t, time.Now(), lock.CreatedOn, 1*time.Second, "CreatedOn should be recent")

	// Get a non-existent index lock
	nonExistentLock := indexLockMgr.Get("non_existent_node", "non_existent_index")
	assert.Nil(t, nonExistentLock, "Non-existent index lock should be nil")
}

// TestIndexLockerError tests setting an error on an index lock
func TestIndexLockerError(t *testing.T) {
	// Create a new index locker
	indexLockMgr := locker.NewIndexLocker()

	// Add an index lock
	nodeID := "test_node_11"
	indexName := "test_index_4"
	indexLockMgr.Add(nodeID, indexName)

	// Set an error
	indexLockMgr.Error(nodeID, indexName, assert.AnError)

	// Verify the error was set
	lock := indexLockMgr.Get(nodeID, indexName)
	assert.NotNil(t, lock, "Index lock should not be nil")
	assert.NotEmpty(t, lock.Error, "Error should not be empty")

	// Test with nil error
	indexLockMgr.Error(nodeID, indexName, nil)

	// Test with non-existent index
	indexLockMgr.Error("non_existent_node", "non_existent_index", assert.AnError)
}

// TestIndexLockerRemove tests removing an index lock
func TestIndexLockerRemove(t *testing.T) {
	// Create a new index locker
	indexLockMgr := locker.NewIndexLocker()

	// Add an index lock
	nodeID := "test_node_12"
	indexName := "test_index_5"
	indexLockMgr.Add(nodeID, indexName)

	// Verify the index lock was added
	lock := indexLockMgr.Get(nodeID, indexName)
	assert.NotNil(t, lock, "Index lock should not be nil")

	// Remove the index lock
	indexLockMgr.Remove(nodeID, indexName)

	// Verify the index lock was removed
	lock = indexLockMgr.Get(nodeID, indexName)
	assert.Nil(t, lock, "Index lock should be nil after removal")
}

// TestIndexLockerRemoveOld tests removing old index locks
func TestIndexLockerRemoveOld(t *testing.T) {
	// Create a new index locker
	indexLockMgr := locker.NewIndexLocker()

	// Add an index lock
	nodeID := "test_node_13"
	indexName := "test_index_6"
	indexLockMgr.Add(nodeID, indexName)

	// Verify the index lock was added
	lock := indexLockMgr.Get(nodeID, indexName)
	assert.NotNil(t, lock, "Index lock should not be nil")

	// Remove old index locks (with a very short expiration time)
	indexLockMgr.RemoveOld(0)

	// Verify the index lock was removed
	lock = indexLockMgr.Get(nodeID, indexName)
	assert.Nil(t, lock, "Index lock should be nil after removing old locks")
}
