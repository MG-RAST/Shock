package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestViewACL(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	resp := doRequest(t, "GET", "/node/"+nodeID+"/acl/", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	ad := parseAclData(t, sr)
	assert.NotEmpty(t, ad.Owner, "ACL should have an owner")
}

func TestAddUserToReadACL(t *testing.T) {
	// User1 creates a node
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// User2 should not be able to read it yet
	resp := doRequest(t, "GET", "/node/"+nodeID, user2Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusUnauthorized, sr.Status)

	// User1 adds user2 to read ACL
	resp = doRequest(t, "PUT", "/node/"+nodeID+"/acl/read?users="+user2UUID, user1Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status, "errors: %v", sr.Error) {
		return
	}

	// User2 should now be able to read it
	resp = doRequest(t, "GET", "/node/"+nodeID, user2Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)
}

func TestRemoveUserFromACL(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Add user2 to read ACL
	resp := doRequest(t, "PUT", "/node/"+nodeID+"/acl/read?users="+user2UUID, user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status) {
		return
	}

	// Verify user2 can read
	resp = doRequest(t, "GET", "/node/"+nodeID, user2Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status) {
		return
	}

	// Remove user2 from read ACL
	resp = doRequest(t, "DELETE", "/node/"+nodeID+"/acl/read?users="+user2UUID, user1Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status, "errors: %v", sr.Error) {
		return
	}

	// User2 should no longer be able to read it
	resp = doRequest(t, "GET", "/node/"+nodeID, user2Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusUnauthorized, sr.Status)
}

func TestSetPublicRead(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Set public read
	resp := doRequest(t, "PUT", "/node/"+nodeID+"/acl/public_read", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status, "errors: %v", sr.Error) {
		return
	}

	// Anonymous user (no auth) should be able to read
	resp = doRequest(t, "GET", "/node/"+nodeID, "", nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	nd := parseNodeData(t, sr)
	assert.Equal(t, nodeID, nd.Id)
}

func TestOwnerTransfer(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Transfer ownership to user2
	resp := doRequest(t, "PUT", "/node/"+nodeID+"/acl/owner?users="+user2UUID, user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status, "errors: %v", sr.Error) {
		return
	}

	// Verify user2 is now the owner
	resp = doRequest(t, "GET", "/node/"+nodeID+"/acl/", user2Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	ad := parseAclData(t, sr)
	assert.Equal(t, user2UUID, ad.Owner)

	// Update cleanup to use user2 since they now own the node
	t.Cleanup(func() {
		deleteNode(t, user2Auth, nodeID)
	})
}

func TestNonOwnerCanOnlyRemoveSelf(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Add user2 to read and write ACLs
	resp := doRequest(t, "PUT", "/node/"+nodeID+"/acl/all?users="+user2UUID, user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status) {
		return
	}

	// User2 tries to remove user1 (the owner) from ACL - should fail
	resp = doRequest(t, "DELETE", "/node/"+nodeID+"/acl/read?users="+user1UUID, user2Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusBadRequest, sr.Status, "non-owner should not be able to remove other users")

	// User2 removes self from read ACL - should succeed
	resp = doRequest(t, "DELETE", "/node/"+nodeID+"/acl/read?users="+user2UUID, user2Auth, nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status, "user should be able to remove self from ACL, errors: %v", sr.Error)
}

func TestACLTypedGet(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Get read ACL
	resp := doRequest(t, "GET", "/node/"+nodeID+"/acl/read", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)
}

func TestACLInvalidType(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Invalid ACL type
	resp := doRequest(t, "GET", "/node/"+nodeID+"/acl/invalid", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusBadRequest, sr.Status)
}
