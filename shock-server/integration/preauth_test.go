package integration_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPreAuthDownload(t *testing.T) {
	content := []byte("Preauth test content for download.")
	nodeID := createNodeWithFile(t, user1Auth, "preauth_test.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	// Get a preauth download URL
	resp := doRequest(t, "GET", "/node/"+nodeID+"?download_url", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status, "errors: %v", sr.Error) {
		return
	}

	var preauthResp preAuthResponseData
	err := json.Unmarshal(sr.Data, &preauthResp)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.NotEmpty(t, preauthResp.Url) {
		return
	}

	// Use the preauth URL to download without authentication
	client := &http.Client{Timeout: 10 * time.Second}
	preauthReq, err := http.NewRequest("GET", preauthResp.Url, nil)
	if !assert.NoError(t, err) {
		return
	}

	preauthHTTPResp, err := client.Do(preauthReq)
	if !assert.NoError(t, err) {
		return
	}
	defer preauthHTTPResp.Body.Close()

	assert.Equal(t, http.StatusOK, preauthHTTPResp.StatusCode)

	body, err := ioutil.ReadAll(preauthHTTPResp.Body)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, content, body)
}

func TestPreAuthInvalidToken(t *testing.T) {
	// Try to access a preauth with a non-existent token
	resp := doRequest(t, "GET", "/preauth/nonexistent_token_12345", "", nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusInternalServerError, sr.Status)
	assert.NotEmpty(t, sr.Error)
}

func TestPreAuthAutoDelete(t *testing.T) {
	content := []byte("Single use preauth content.")
	nodeID := createNodeWithFile(t, user1Auth, "single_use.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	// Get a preauth download URL
	resp := doRequest(t, "GET", "/node/"+nodeID+"?download_url", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status) {
		return
	}

	var preauthResp preAuthResponseData
	err := json.Unmarshal(sr.Data, &preauthResp)
	if !assert.NoError(t, err) {
		return
	}

	// First use should succeed
	client := &http.Client{Timeout: 10 * time.Second}
	req1, err := http.NewRequest("GET", preauthResp.Url, nil)
	if !assert.NoError(t, err) {
		return
	}
	resp1, err := client.Do(req1)
	if !assert.NoError(t, err) {
		return
	}
	resp1.Body.Close()
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	// Second use should fail (preauth is deleted after first use)
	req2, err := http.NewRequest("GET", preauthResp.Url, nil)
	if !assert.NoError(t, err) {
		return
	}
	resp2, err := client.Do(req2)
	if !assert.NoError(t, err) {
		return
	}
	defer resp2.Body.Close()

	// The preauth token should be gone - server returns an error
	body, _ := ioutil.ReadAll(resp2.Body)
	var sr2 standardResponse
	if json.Unmarshal(body, &sr2) == nil {
		assert.NotEqual(t, http.StatusOK, sr2.Status, "second preauth use should fail")
	}
}
