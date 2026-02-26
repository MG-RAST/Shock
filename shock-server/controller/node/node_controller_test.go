package node_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/controller/node"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/acl"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"github.com/stretchr/testify/assert"
)

// setupTestController creates a test controller and mock dependencies
func setupTestController() (*node.NodeController, *httptest.Server) {
	// Create controller
	controller := &node.NodeController{}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a context for the request
		ctx := context.NewContext(nil, r, w)

		// Route the request based on the method and path
		switch r.Method {
		case "GET":
			if r.URL.Path == "/node" {
				controller.GetMany(ctx)
			} else {
				controller.Get(ctx)
			}
		case "POST":
			controller.Create(ctx)
		case "PUT":
			controller.Update(ctx)
		case "DELETE":
			controller.Delete(ctx)
		case "OPTIONS":
			controller.Options(ctx)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))

	return controller, server
}

// TestNodeControllerCreate tests the Create method of the NodeController
func TestNodeControllerCreate(t *testing.T) {
	// Set up controller and test server
	_, server := setupTestController()
	defer server.Close()

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

	// Create a multipart form request
	// Note: In a real test, we would create a multipart form with a file
	// For simplicity, we'll use a JSON request here
	requestBody := map[string]string{
		"file_name": "test_file.txt",
	}
	requestJSON, _ := json.Marshal(requestBody)

	// Create a request with basic auth
	req, err := http.NewRequest("POST", server.URL+"/node", bytes.NewBuffer(requestJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("test_user", "test_password")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code should be OK")

	// Parse response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Verify response contains a node ID
	var response map[string]interface{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok, "Response should contain data")
	assert.NotEmpty(t, data["id"], "Response should contain a node ID")
}

// TestNodeControllerGet tests the Get method of the NodeController
func TestNodeControllerGet(t *testing.T) {
	// Set up controller and test server
	_, server := setupTestController()
	defer server.Close()

	// Create a test node
	testNode := createTestNode()

	// Create a request
	req, err := http.NewRequest("GET", server.URL+"/node/"+testNode.Id, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetBasicAuth("test_user", "test_password")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code should be OK")

	// Parse response body
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Verify response contains the node
	var response map[string]interface{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok, "Response should contain data")
	assert.Equal(t, testNode.Id, data["id"], "Response should contain the correct node ID")
}

// TestNodeControllerDelete tests the Delete method of the NodeController
func TestNodeControllerDelete(t *testing.T) {
	// Set up controller and test server
	_, server := setupTestController()
	defer server.Close()

	// Create a test node
	testNode := createTestNode()

	// Create a request
	req, err := http.NewRequest("DELETE", server.URL+"/node/"+testNode.Id, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetBasicAuth("test_user", "test_password")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code should be OK")

	// Verify the node is deleted
	req, _ = http.NewRequest("GET", server.URL+"/node/"+testNode.Id, nil)
	req.SetBasicAuth("test_user", "test_password")
	resp, _ = client.Do(req)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Node should be deleted")
}

// TestNodeControllerUpdate tests the Update method of the NodeController
func TestNodeControllerUpdate(t *testing.T) {
	// Set up controller and test server
	_, server := setupTestController()
	defer server.Close()

	// Create a test node
	testNode := createTestNode()

	// Create update request body
	updateBody := map[string]string{
		"attributes_str": `{"project":"updated_project"}`,
	}
	updateJSON, _ := json.Marshal(updateBody)

	// Create a request
	req, err := http.NewRequest("PUT", server.URL+"/node/"+testNode.Id, bytes.NewBuffer(updateJSON))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("test_user", "test_password")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Status code should be OK")

	// Verify the node is updated
	req, _ = http.NewRequest("GET", server.URL+"/node/"+testNode.Id, nil)
	req.SetBasicAuth("test_user", "test_password")
	resp, _ = client.Do(req)

	responseBody, _ := ioutil.ReadAll(resp.Body)
	var response map[string]interface{}
	json.Unmarshal(responseBody, &response)

	data, _ := response["data"].(map[string]interface{})
	attributes, _ := data["attributes"].(map[string]interface{})
	assert.Equal(t, "updated_project", attributes["project"], "Node attributes should be updated")
}

// Helper function to create a test node
func createTestNode() *node.Node {
	// Create a test user
	testUser := &user.User{
		Uuid:     "test_user",
		Username: "test_user",
	}

	// Create a node
	n := node.New("")
	n.Type = "basic"
	n.Acl.SetOwner(testUser.Uuid)
	n.Acl.Set(testUser.Uuid, acl.Rights{"read": true, "write": true, "delete": true})

	// Set file info
	n.File.Name = "test_file.txt"
	n.File.Size = 12
	n.File.Checksum = map[string]string{"md5": "test_checksum"}

	// Save the node
	err := n.Save()
	if err != nil {
		return nil
	}

	return n
}
