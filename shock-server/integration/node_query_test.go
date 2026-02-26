package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadManyNodes(t *testing.T) {
	// Create 3 nodes
	ids := make([]string, 3)
	for i := 0; i < 3; i++ {
		ids[i] = createEmptyNode(t, user1Auth)
		cleanupNode(t, user1Auth, ids[i])
	}

	// GET /node should return paginated results
	resp := doRequest(t, "GET", "/node", user1Auth, nil, "")
	pr := parsePaginatedResponse(t, resp)
	assert.Equal(t, http.StatusOK, pr.Status)
	assert.GreaterOrEqual(t, pr.TotalCount, 3)

	nodes := parseNodeList(t, pr.Data)
	assert.GreaterOrEqual(t, len(nodes), 3)
}

func TestQueryByAttribute(t *testing.T) {
	// Create nodes with different attributes
	attrs1 := map[string]interface{}{"project": "alpha", "type": "query_test"}
	attrs2 := map[string]interface{}{"project": "beta", "type": "query_test"}

	id1 := createNodeWithAttributes(t, user1Auth, attrs1)
	cleanupNode(t, user1Auth, id1)
	id2 := createNodeWithAttributes(t, user1Auth, attrs2)
	cleanupNode(t, user1Auth, id2)

	// Query for project=alpha
	resp := doRequest(t, "GET", "/node?query&project=alpha&type=query_test", user1Auth, nil, "")
	pr := parsePaginatedResponse(t, resp)
	assert.Equal(t, http.StatusOK, pr.Status)

	nodes := parseNodeList(t, pr.Data)
	// Should find at least the node with project=alpha
	found := false
	for _, n := range nodes {
		if n.Id == id1 {
			found = true
		}
		// node with beta should not be in results
		assert.NotEqual(t, id2, n.Id, "beta node should not appear in alpha query")
	}
	assert.True(t, found, "expected to find node %s in query results", id1)
}

func TestQueryPagination(t *testing.T) {
	// Create several nodes to ensure we have enough for pagination
	ids := make([]string, 5)
	for i := 0; i < 5; i++ {
		ids[i] = createNodeWithAttributes(t, user1Auth, map[string]interface{}{
			"pagination_test": "true",
			"index":           i,
		})
		cleanupNode(t, user1Auth, ids[i])
	}

	// Request with limit=2, offset=0
	resp := doRequest(t, "GET", "/node?query&pagination_test=true&limit=2&offset=0", user1Auth, nil, "")
	pr := parsePaginatedResponse(t, resp)
	assert.Equal(t, http.StatusOK, pr.Status)
	assert.Equal(t, 2, pr.Limit)
	assert.Equal(t, 0, pr.Offset)
	assert.GreaterOrEqual(t, pr.TotalCount, 5)

	nodes := parseNodeList(t, pr.Data)
	assert.Equal(t, 2, len(nodes))

	// Request next page
	resp = doRequest(t, "GET", "/node?query&pagination_test=true&limit=2&offset=2", user1Auth, nil, "")
	pr2 := parsePaginatedResponse(t, resp)
	assert.Equal(t, http.StatusOK, pr2.Status)
	assert.Equal(t, 2, pr2.Limit)
	assert.Equal(t, 2, pr2.Offset)

	nodes2 := parseNodeList(t, pr2.Data)
	assert.Equal(t, 2, len(nodes2))

	// Ensure pages contain different nodes
	if len(nodes) > 0 && len(nodes2) > 0 {
		assert.NotEqual(t, nodes[0].Id, nodes2[0].Id, "different pages should have different nodes")
	}
}

func TestQueryDefaultPagination(t *testing.T) {
	// Default limit is 25, offset is 0
	resp := doRequest(t, "GET", "/node", user1Auth, nil, "")
	pr := parsePaginatedResponse(t, resp)
	assert.Equal(t, http.StatusOK, pr.Status)
	assert.Equal(t, 25, pr.Limit)
	assert.Equal(t, 0, pr.Offset)
}
