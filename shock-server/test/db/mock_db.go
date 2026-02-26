// Package db provides mock database implementations for testing
package db

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MG-RAST/Shock/shock-server/test/util"
)

// MockDB is a mock implementation of the MongoDB database for testing
type MockDB struct {
	nodes     map[string]*util.Node
	users     map[string]*util.User
	nodeLock  sync.RWMutex
	userLock  sync.RWMutex
	queryFunc func(query interface{}) (interface{}, error)
}

// NewMockDB creates a new mock database for testing
func NewMockDB() *MockDB {
	return &MockDB{
		nodes: make(map[string]*util.Node),
		users: make(map[string]*util.User),
	}
}

// AddTestUser adds a test user to the mock database
func (db *MockDB) AddTestUser(user *util.User) {
	db.userLock.Lock()
	defer db.userLock.Unlock()
	db.users[user.Uuid] = user
}

// GetUser retrieves a user from the mock database
func (db *MockDB) GetUser(id string) (*util.User, error) {
	db.userLock.RLock()
	defer db.userLock.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// AddNode adds a node to the mock database
func (db *MockDB) AddNode(node *util.Node) {
	db.nodeLock.Lock()
	defer db.nodeLock.Unlock()

	// Set creation time if not already set
	if node.CreatedOn.IsZero() {
		node.CreatedOn = time.Now()
	}

	// Set last modified time
	node.LastModified = time.Now()

	db.nodes[node.Id] = node
}

// GetNode retrieves a node from the mock database
func (db *MockDB) GetNode(id string) (*util.Node, error) {
	db.nodeLock.RLock()
	defer db.nodeLock.RUnlock()

	node, exists := db.nodes[id]
	if !exists {
		return nil, errors.New("node not found")
	}
	return node, nil
}

// DeleteNode removes a node from the mock database
func (db *MockDB) DeleteNode(id string) error {
	db.nodeLock.Lock()
	defer db.nodeLock.Unlock()

	if _, exists := db.nodes[id]; !exists {
		return errors.New("node not found")
	}

	delete(db.nodes, id)
	return nil
}

// UpdateNode updates a node in the mock database
func (db *MockDB) UpdateNode(id string, updates map[string]interface{}) error {
	db.nodeLock.Lock()
	defer db.nodeLock.Unlock()

	node, exists := db.nodes[id]
	if !exists {
		return errors.New("node not found")
	}

	// Apply updates (simplified for testing)
	for key, value := range updates {
		switch key {
		case "attributes":
			node.Attributes = value
		case "file.name":
			if file, ok := value.(string); ok {
				node.File.Name = file
			}
		case "file.size":
			if size, ok := value.(int64); ok {
				node.File.Size = size
			}
		// Add more update cases as needed
		default:
			return fmt.Errorf("update field not supported: %s", key)
		}
	}

	// Update last modified time
	node.LastModified = time.Now()

	return nil
}

// FindNodes finds nodes matching the query criteria
func (db *MockDB) FindNodes(query interface{}) ([]*util.Node, error) {
	db.nodeLock.RLock()
	defer db.nodeLock.RUnlock()

	// If a custom query function is set, use it
	if db.queryFunc != nil {
		result, err := db.queryFunc(query)
		if err != nil {
			return nil, err
		}
		if nodes, ok := result.([]*util.Node); ok {
			return nodes, nil
		}
		return nil, errors.New("invalid query result type")
	}

	// Default implementation: return all nodes (simplified)
	// In a real implementation, we would filter based on the query
	var result []*util.Node
	for _, node := range db.nodes {
		result = append(result, node)
	}

	return result, nil
}

// SetQueryFunc sets a custom query function for testing specific scenarios
func (db *MockDB) SetQueryFunc(fn func(query interface{}) (interface{}, error)) {
	db.queryFunc = fn
}

// Reset clears all data in the mock database
func (db *MockDB) Reset() {
	db.nodeLock.Lock()
	db.userLock.Lock()
	defer db.nodeLock.Unlock()
	defer db.userLock.Unlock()

	db.nodes = make(map[string]*util.Node)
	db.users = make(map[string]*util.User)
	db.queryFunc = nil
}
