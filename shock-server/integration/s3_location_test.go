package integration_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// requireS3 skips the test if the Shock server has no "s3" location configured.
// This allows S3 tests to coexist in the same package with non-S3 tests.
func requireS3(t *testing.T) {
	t.Helper()
	resp := doRequest(t, "GET", "/location/s3/info", adminAuth, nil, "")
	sr := parseStandardResponse(t, resp)
	if sr.Status != http.StatusOK {
		t.Skip("skipping: server has no S3 location configured")
	}
}

// ---------------------------------------------------------------------------
// Location Config tests
// ---------------------------------------------------------------------------

func TestLocationS3ConfigInfo(t *testing.T) {
	requireS3(t)
	resp := doRequest(t, "GET", "/location/s3/info", adminAuth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)
	assert.Empty(t, sr.Error)

	lcd := parseLocationConfigData(t, sr)
	assert.Equal(t, "s3", lcd.ID)
	assert.Equal(t, "S3", lcd.Type)
	assert.Equal(t, "shock-data", lcd.Bucket)
}

func TestLocationInfoRequiresAdmin(t *testing.T) {
	requireS3(t)
	resp := doRequest(t, "GET", "/location/s3/info", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	// Server returns 500 for non-admin users (see location.go:49)
	assert.Equal(t, http.StatusInternalServerError, sr.Status)
	assert.NotEmpty(t, sr.Error)
}

// ---------------------------------------------------------------------------
// Auto-Upload tests
// ---------------------------------------------------------------------------

func TestAutoUploadCreatesLocation(t *testing.T) {
	requireS3(t)
	content := []byte("auto-upload integration test data")
	nodeID := createNodeWithFile(t, user1Auth, "auto_upload_test.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	nd, found := waitForLocation(t, user1Auth, nodeID, "s3", 30*time.Second)
	if !assert.True(t, found, "expected location 's3' with Stored=true within timeout") {
		return
	}
	assert.Equal(t, nodeID, nd.Id)

	// Verify the location entry
	var s3Loc *nodeLocation
	for i := range nd.Locations {
		if nd.Locations[i].ID == "s3" {
			s3Loc = &nd.Locations[i]
			break
		}
	}
	if !assert.NotNil(t, s3Loc) {
		return
	}
	assert.True(t, s3Loc.Stored)
}

func TestAutoUploadMultipleFiles(t *testing.T) {
	requireS3(t)
	files := []struct {
		name    string
		content []byte
	}{
		{"multi_1.txt", []byte("first file for multi-upload test")},
		{"multi_2.txt", []byte("second file for multi-upload test")},
		{"multi_3.txt", []byte("third file for multi-upload test")},
	}

	var nodeIDs []string
	for _, f := range files {
		nid := createNodeWithFile(t, user1Auth, f.name, f.content)
		cleanupNode(t, user1Auth, nid)
		nodeIDs = append(nodeIDs, nid)
	}

	// Wait for all three to get their S3 location
	for i, nid := range nodeIDs {
		_, found := waitForLocation(t, user1Auth, nid, "s3", 30*time.Second)
		assert.True(t, found, "node %d (%s) did not get s3 location within timeout", i, nid)
	}
}

// ---------------------------------------------------------------------------
// Location CRUD tests
// ---------------------------------------------------------------------------

func TestViewNodeLocations(t *testing.T) {
	requireS3(t)
	content := []byte("view locations test data")
	nodeID := createNodeWithFile(t, user1Auth, "view_locations.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	_, found := waitForLocation(t, user1Auth, nodeID, "s3", 30*time.Second)
	if !assert.True(t, found, "expected s3 location to appear") {
		return
	}

	// GET /node/{nid}/locations/
	resp := doRequest(t, "GET", "/node/"+nodeID+"/locations/", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	locs := parseLocationList(t, sr)
	if !assert.NotEmpty(t, locs) {
		return
	}

	var hasS3 bool
	for _, loc := range locs {
		if loc.ID == "s3" {
			hasS3 = true
			assert.True(t, loc.Stored)
		}
	}
	assert.True(t, hasS3, "expected 's3' in location list")
}

func TestViewSpecificLocation(t *testing.T) {
	requireS3(t)
	content := []byte("view specific location test data")
	nodeID := createNodeWithFile(t, user1Auth, "view_specific_loc.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	_, found := waitForLocation(t, user1Auth, nodeID, "s3", 30*time.Second)
	if !assert.True(t, found, "expected s3 location to appear") {
		return
	}

	// GET /node/{nid}/locations/s3
	resp := doRequest(t, "GET", "/node/"+nodeID+"/locations/s3", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	loc := parseSingleLocation(t, sr)
	assert.Equal(t, "s3", loc.ID)
	assert.True(t, loc.Stored)
	assert.NotNil(t, loc.RequestedDate, "expected RequestedDate to be set")
}

func TestAdminDeleteLocation(t *testing.T) {
	requireS3(t)
	content := []byte("delete location test data")
	nodeID := createNodeWithFile(t, user1Auth, "delete_location.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	_, found := waitForLocation(t, user1Auth, nodeID, "s3", 30*time.Second)
	if !assert.True(t, found, "expected s3 location to appear before deletion") {
		return
	}

	// DELETE /node/{nid}/locations/s3 with admin
	resp := doRequest(t, "DELETE", "/node/"+nodeID+"/locations/s3", adminAuth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	// Verify location is gone
	resp = doRequest(t, "GET", "/node/"+nodeID+"/locations/", adminAuth, nil, "")
	sr = parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	locs := parseLocationList(t, sr)
	for _, loc := range locs {
		assert.NotEqual(t, "s3", loc.ID, "expected s3 location to be deleted")
	}
}

func TestDeleteLocationRequiresAdmin(t *testing.T) {
	requireS3(t)
	content := []byte("delete auth test data")
	nodeID := createNodeWithFile(t, user1Auth, "delete_auth_test.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	_, found := waitForLocation(t, user1Auth, nodeID, "s3", 30*time.Second)
	if !assert.True(t, found, "expected s3 location to appear") {
		return
	}

	// DELETE with non-admin should fail
	resp := doRequest(t, "DELETE", "/node/"+nodeID+"/locations/s3", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusUnauthorized, sr.Status)
	assert.NotEmpty(t, sr.Error)
}

// ---------------------------------------------------------------------------
// Edge Case tests
// ---------------------------------------------------------------------------

func TestEmptyNodeHasNoLocations(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	resp := doRequest(t, "GET", "/node/"+nodeID+"/locations/", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusOK, sr.Status)

	locs := parseLocationList(t, sr)
	assert.Empty(t, locs)
}

func TestLocationNotFoundOnNode(t *testing.T) {
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// GET /node/{nid}/locations/nonexistent
	resp := doRequest(t, "GET", "/node/"+nodeID+"/locations/nonexistent", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	// Server returns 500 when location not found on node (see locations.go:77-79)
	assert.Equal(t, http.StatusInternalServerError, sr.Status)
	assert.NotEmpty(t, sr.Error)
}
