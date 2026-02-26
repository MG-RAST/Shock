package integration_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEmptyNode(t *testing.T) {
	resp := doRequest(t, "POST", "/node", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)
	assert.Empty(t, sr.Error)

	nd := parseNodeData(t, sr)
	assert.NotEmpty(t, nd.Id)
	cleanupNode(t, user1Auth, nd.Id)
}

func TestCreateNodeWithAttributes(t *testing.T) {
	attrs := map[string]interface{}{
		"project": "test_project",
		"sample":  "sample_001",
		"count":   42,
	}
	attrBytes, err := json.Marshal(attrs)
	if !assert.NoError(t, err) {
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("attributes", "attributes.json")
	if !assert.NoError(t, err) {
		return
	}
	_, err = part.Write(attrBytes)
	if !assert.NoError(t, err) {
		return
	}
	err = writer.Close()
	if !assert.NoError(t, err) {
		return
	}

	resp := doRequest(t, "POST", "/node", user1Auth, &buf, writer.FormDataContentType())
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)
	assert.Empty(t, sr.Error)

	nd := parseNodeData(t, sr)
	assert.NotEmpty(t, nd.Id)
	cleanupNode(t, user1Auth, nd.Id)

	// Verify attributes were stored
	assert.Equal(t, "test_project", nd.Attributes["project"])
	assert.Equal(t, "sample_001", nd.Attributes["sample"])
}

func TestCreateNodeUnauthenticated(t *testing.T) {
	// Config has ANON_WRITE=false, so unauthenticated POST should fail
	resp := doRequest(t, "POST", "/node", "", nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusUnauthorized, sr.Status)
}

func TestReadSingleNode(t *testing.T) {
	// Create a node first
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Read it back
	resp := doRequest(t, "GET", "/node/"+nodeID, user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)
	assert.Empty(t, sr.Error)

	nd := parseNodeData(t, sr)
	assert.Equal(t, nodeID, nd.Id)
}

func TestReadNodeNotFound(t *testing.T) {
	resp := doRequest(t, "GET", "/node/00000000-0000-0000-0000-000000000000", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusNotFound, sr.Status)
}

func TestUpdateNodeAttributes(t *testing.T) {
	// Create a node
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Update it with new attributes
	newAttrs := map[string]interface{}{
		"project": "updated_project",
		"version": 2,
	}
	attrBytes, err := json.Marshal(newAttrs)
	if !assert.NoError(t, err) {
		return
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("attributes", "attributes.json")
	if !assert.NoError(t, err) {
		return
	}
	_, err = part.Write(attrBytes)
	if !assert.NoError(t, err) {
		return
	}
	err = writer.Close()
	if !assert.NoError(t, err) {
		return
	}

	resp := doRequest(t, "PUT", "/node/"+nodeID, user1Auth, &buf, writer.FormDataContentType())
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status, "errors: %v", sr.Error)

	nd := parseNodeData(t, sr)
	assert.Equal(t, "updated_project", nd.Attributes["project"])
}

func TestDeleteNode(t *testing.T) {
	// Create a node
	nodeID := createEmptyNode(t, user1Auth)

	// Delete it
	resp := doRequest(t, "DELETE", "/node/"+nodeID, user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	// Verify it's gone
	resp = doRequest(t, "GET", "/node/"+nodeID, user1Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusNotFound, sr.Status)
}

func TestDeleteNodeUnauthorized(t *testing.T) {
	// User 1 creates a node
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// User 2 tries to delete it (should fail - no delete ACL)
	resp := doRequest(t, "DELETE", "/node/"+nodeID, user2Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusUnauthorized, sr.Status)
}

func TestReadNodeAnonymous(t *testing.T) {
	// Create a node (ANON_READ=true in config)
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Make the node publicly readable
	resp := doRequest(t, "PUT", "/node/"+nodeID+"/acl/public_read", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status, "setting public_read, errors: %v", sr.Error) {
		return
	}

	// Anonymous user should be able to read it
	resp = doRequest(t, "GET", "/node/"+nodeID, "", nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	nd := parseNodeData(t, sr)
	assert.Equal(t, nodeID, nd.Id)
}
