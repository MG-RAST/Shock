package integration_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultipartFileUpload(t *testing.T) {
	content := []byte("Hello, Shock! This is test file content.")
	nodeID := createNodeWithFile(t, user1Auth, "test_file.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	// Verify the node has the file metadata
	resp := doRequest(t, "GET", "/node/"+nodeID, user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if !assert.Equal(t, http.StatusOK, sr.Status) {
		return
	}

	nd := parseNodeData(t, sr)
	assert.Equal(t, "test_file.txt", nd.File.Name)
	assert.Equal(t, int64(len(content)), nd.File.Size)
	assert.NotEmpty(t, nd.File.Checksum["md5"])
}

func TestFileDownload(t *testing.T) {
	content := []byte("Download me! This is the file content for testing.")
	nodeID := createNodeWithFile(t, user1Auth, "download_test.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	// Download the file
	resp := doRequest(t, "GET", "/node/"+nodeID+"?download", user1Auth, nil, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check Content-Disposition header
	contentDisposition := resp.Header.Get("Content-Disposition")
	assert.Contains(t, contentDisposition, "download_test.txt")

	// Verify content matches
	body, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, content, body)
}

func TestFileDownloadPartial(t *testing.T) {
	content := []byte("0123456789abcdefghij")
	nodeID := createNodeWithFile(t, user1Auth, "partial_test.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	// Download partial content with seek and length
	resp := doRequest(t, "GET", "/node/"+nodeID+"?download&seek=5&length=10", user1Auth, nil, "")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "56789abcde", string(body))
}

func TestUploadToExistingNode(t *testing.T) {
	// Create an empty node
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Upload a file to the existing node via PUT
	content := []byte("File added to existing node.")
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("upload", "added_file.txt")
	if !assert.NoError(t, err) {
		return
	}
	_, err = part.Write(content)
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
	assert.Equal(t, "added_file.txt", nd.File.Name)
	assert.Equal(t, int64(len(content)), nd.File.Size)
}

func TestDownloadURL(t *testing.T) {
	content := []byte("Preauth download test content.")
	nodeID := createNodeWithFile(t, user1Auth, "preauth_download.txt", content)
	cleanupNode(t, user1Auth, nodeID)

	// Request a download URL (preauth)
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

	assert.NotEmpty(t, preauthResp.Url, "preauth URL should not be empty")
	assert.Equal(t, int64(len(content)), preauthResp.Size)
}

func TestDownloadNoFile(t *testing.T) {
	// Create an empty node (no file)
	nodeID := createEmptyNode(t, user1Auth)
	cleanupNode(t, user1Auth, nodeID)

	// Try to download - should fail
	resp := doRequest(t, "GET", "/node/"+nodeID+"?download", user1Auth, nil, "")
	sr := parseStandardResponse(t, resp)
	assert.Equal(t, http.StatusBadRequest, sr.Status)
}

func TestUploadLargerFile(t *testing.T) {
	// Create a 1KB file
	content := make([]byte, 1024)
	for i := range content {
		content[i] = byte(i % 256)
	}
	nodeID := createNodeWithFile(t, user1Auth, "large_test.bin", content)
	cleanupNode(t, user1Auth, nodeID)

	// Download and verify
	resp := doRequest(t, "GET", "/node/"+nodeID+"?download", user1Auth, nil, "")
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, content, body)
}
