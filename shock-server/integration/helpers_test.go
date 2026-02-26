package integration_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/db"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/goconfig/config"
)

var (
	serverURL string

	// Test user credentials (seeded in TestMain)
	adminAuth string // Basic auth header for admin
	user1Auth string // Basic auth header for test_user1
	user2Auth string // Basic auth header for test_user2

	// User UUIDs (set after seeding)
	user1UUID string
	user2UUID string
)

// standardResponse mirrors the Shock API response format
type standardResponse struct {
	Status int             `json:"status"`
	Data   json.RawMessage `json:"data"`
	Error  []string        `json:"error"`
}

// paginatedResponse mirrors the Shock paginated API response format
type paginatedResponse struct {
	Status     int             `json:"status"`
	Data       json.RawMessage `json:"data"`
	Error      []string        `json:"error"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	TotalCount int             `json:"total_count"`
}

// nodeLocation represents a storage location entry on a node
type nodeLocation struct {
	ID            string  `json:"id"`
	Stored        bool    `json:"stored,omitempty"`
	RequestedDate *string `json:"requestedDate,omitempty"`
}

// locationConfigData represents fields from GET /location/{loc}/info
type locationConfigData struct {
	ID          string `json:"ID"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	Bucket      string `json:"bucket"`
	Region      string `json:"region"`
	Persistent  bool   `json:"persistent"`
	Description string `json:"Description"`
	Priority    int    `json:"priority"`
	MinPriority int    `json:"minpriority"`
	Tier        int    `json:"tier"`
	Cost        int    `json:"cost"`
}

// nodeData represents the essential fields of a node returned by the API
type nodeData struct {
	Id           string                 `json:"id"`
	File         nodeFile               `json:"file"`
	Attributes   map[string]interface{} `json:"attributes"`
	Tags         []string               `json:"tags"`
	Locations    []nodeLocation         `json:"locations"`
	CreatedOn    string                 `json:"created_on"`
	LastModified string                 `json:"last_modified"`
	Type         string                 `json:"type"`
}

type nodeFile struct {
	Name     string            `json:"name"`
	Size     int64             `json:"size"`
	Checksum map[string]string `json:"checksum"`
}

// aclData represents the ACL fields returned by /node/{id}/acl/
type aclData struct {
	Owner  string   `json:"owner"`
	Read   []string `json:"read"`
	Write  []string `json:"write"`
	Delete []string `json:"delete"`
}

// preAuthResponseData represents the preauth response
type preAuthResponseData struct {
	Url       string `json:"url"`
	ValidTill string `json:"validtill"`
	Format    string `json:"format"`
	Filename  string `json:"filename"`
	Files     int    `json:"files"`
	Size      int64  `json:"size"`
}

func TestMain(m *testing.M) {
	// 1. Get server URL from env (set by docker-compose)
	serverURL = os.Getenv("SHOCK_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:7445"
	}

	// 2. Wait for server to be ready (poll GET / up to 30s)
	if !waitForServer(serverURL, 30*time.Second) {
		fmt.Fprintf(os.Stderr, "ERROR: Shock server at %s did not become ready\n", serverURL)
		os.Exit(1)
	}

	// 3. Connect to MongoDB and seed test users.
	// Read settings from the server config file (shared single source of truth)
	// when SHOCK_CONFIG is set; otherwise fall back to env vars / defaults.
	if configFile := os.Getenv("SHOCK_CONFIG"); configFile != "" {
		c, err := config.ReadDefault(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: reading config %s: %s\n", configFile, err.Error())
			os.Exit(1)
		}
		conf.MONGODB_HOSTS, _ = c.String("Mongodb", "hosts")
		conf.MONGODB_DATABASE, _ = c.String("Mongodb", "database")
		conf.LOG_OUTPUT, _ = c.String("Log", "logoutput")
		conf.PATH_LOGS, _ = c.String("Paths", "logs")
	} else {
		mongoHost := os.Getenv("MONGO_HOST")
		if mongoHost == "" {
			mongoHost = "shock-mongo-test"
		}
		conf.MONGODB_HOSTS = mongoHost
		conf.MONGODB_DATABASE = "shock_integration_test"
		conf.LOG_OUTPUT = "console"
		conf.PATH_LOGS = "/var/log/shock"
	}

	if err := db.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: db.Initialize: %s\n", err.Error())
		os.Exit(1)
	}

	// Seed test users with known passwords
	u1, err := user.New("test_user1", "password1", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: creating test_user1: %s\n", err.Error())
		os.Exit(1)
	}
	user1UUID = u1.Uuid

	u2, err := user.New("test_user2", "password2", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: creating test_user2: %s\n", err.Error())
		os.Exit(1)
	}
	user2UUID = u2.Uuid

	// Build auth headers
	adminAuth = basicAuth("test_admin", "")
	user1Auth = basicAuth("test_user1", "password1")
	user2Auth = basicAuth("test_user2", "password2")

	// 4. Run tests
	code := m.Run()

	// 5. Cleanup: drop test database
	db.Drop()
	os.Exit(code)
}

// waitForServer polls the server root endpoint until it responds or timeout
func waitForServer(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "/")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

// basicAuth returns "Basic <base64(user:pass)>" header value
func basicAuth(username, password string) string {
	creds := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(creds))
}

// doRequest makes an HTTP request to the running Shock server
func doRequest(t *testing.T, method, path, authHeader string, body io.Reader, contentType string) *http.Response {
	t.Helper()
	url := serverURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("executing request: %v", err)
	}
	return resp
}

// parseStandardResponse reads the JSON response into a standardResponse
func parseStandardResponse(t *testing.T, resp *http.Response) standardResponse {
	t.Helper()
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response body: %v", err)
	}

	var sr standardResponse
	err = json.Unmarshal(body, &sr)
	if err != nil {
		t.Fatalf("unmarshalling response: %s: %v", string(body), err)
	}
	return sr
}

// parsePaginatedResponse reads the JSON response into a paginatedResponse
func parsePaginatedResponse(t *testing.T, resp *http.Response) paginatedResponse {
	t.Helper()
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading response body: %v", err)
	}

	var pr paginatedResponse
	err = json.Unmarshal(body, &pr)
	if err != nil {
		t.Fatalf("unmarshalling paginated response: %s: %v", string(body), err)
	}
	return pr
}

// parseNodeData extracts node data from a standard response
func parseNodeData(t *testing.T, sr standardResponse) nodeData {
	t.Helper()
	var nd nodeData
	err := json.Unmarshal(sr.Data, &nd)
	if err != nil {
		t.Fatalf("unmarshalling node data: %s: %v", string(sr.Data), err)
	}
	return nd
}

// parseNodeList extracts a list of nodes from raw JSON data
func parseNodeList(t *testing.T, data json.RawMessage) []nodeData {
	t.Helper()
	var nodes []nodeData
	err := json.Unmarshal(data, &nodes)
	if err != nil {
		t.Fatalf("unmarshalling node list: %s: %v", string(data), err)
	}
	return nodes
}

// parseAclData extracts ACL data from a standard response
func parseAclData(t *testing.T, sr standardResponse) aclData {
	t.Helper()
	var ad aclData
	err := json.Unmarshal(sr.Data, &ad)
	if err != nil {
		t.Fatalf("unmarshalling acl data: %s: %v", string(sr.Data), err)
	}
	return ad
}

// createEmptyNode creates an empty node with auth, returns node ID
func createEmptyNode(t *testing.T, auth string) string {
	t.Helper()
	resp := doRequest(t, "POST", "/node", auth, nil, "")
	sr := parseStandardResponse(t, resp)
	if sr.Status != http.StatusOK {
		t.Fatalf("expected 200 creating empty node, got %d, errors: %v", sr.Status, sr.Error)
	}
	nd := parseNodeData(t, sr)
	if nd.Id == "" {
		t.Fatal("expected non-empty node ID")
	}
	return nd.Id
}

// createNodeWithFile uploads a file via multipart POST, returns node ID
func createNodeWithFile(t *testing.T, auth, filename string, content []byte) string {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("upload", filename)
	if err != nil {
		t.Fatalf("creating form file: %v", err)
	}
	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("writing content to form file: %v", err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatalf("closing multipart writer: %v", err)
	}

	resp := doRequest(t, "POST", "/node", auth, &buf, writer.FormDataContentType())
	sr := parseStandardResponse(t, resp)
	if sr.Status != http.StatusOK {
		t.Fatalf("expected 200 creating node with file, got %d, errors: %v", sr.Status, sr.Error)
	}
	nd := parseNodeData(t, sr)
	if nd.Id == "" {
		t.Fatal("expected non-empty node ID")
	}
	return nd.Id
}

// createNodeWithAttributes creates a node with JSON attributes via multipart
func createNodeWithAttributes(t *testing.T, auth string, attrs map[string]interface{}) string {
	t.Helper()
	attrBytes, err := json.Marshal(attrs)
	if err != nil {
		t.Fatalf("marshalling attributes: %v", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("attributes", "attributes.json")
	if err != nil {
		t.Fatalf("creating form file: %v", err)
	}
	_, err = part.Write(attrBytes)
	if err != nil {
		t.Fatalf("writing attributes to form file: %v", err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatalf("closing multipart writer: %v", err)
	}

	resp := doRequest(t, "POST", "/node", auth, &buf, writer.FormDataContentType())
	sr := parseStandardResponse(t, resp)
	if sr.Status != http.StatusOK {
		t.Fatalf("expected 200 creating node with attributes, got %d, errors: %v", sr.Status, sr.Error)
	}
	nd := parseNodeData(t, sr)
	if nd.Id == "" {
		t.Fatal("expected non-empty node ID")
	}
	return nd.Id
}

// deleteNode sends DELETE /node/{id}
func deleteNode(t *testing.T, auth, nodeID string) {
	t.Helper()
	resp := doRequest(t, "DELETE", "/node/"+nodeID, auth, nil, "")
	resp.Body.Close()
}

// cleanupNode registers a cleanup function to delete a node after the test
func cleanupNode(t *testing.T, auth, nodeID string) {
	t.Helper()
	t.Cleanup(func() {
		deleteNode(t, auth, nodeID)
	})
}

// parseLocationConfigData extracts location config from a standard response
func parseLocationConfigData(t *testing.T, sr standardResponse) locationConfigData {
	t.Helper()
	var lcd locationConfigData
	err := json.Unmarshal(sr.Data, &lcd)
	if err != nil {
		t.Fatalf("unmarshalling location config data: %s: %v", string(sr.Data), err)
	}
	return lcd
}

// parseLocationList extracts a list of node locations from a standard response
func parseLocationList(t *testing.T, sr standardResponse) []nodeLocation {
	t.Helper()
	var locs []nodeLocation
	err := json.Unmarshal(sr.Data, &locs)
	if err != nil {
		t.Fatalf("unmarshalling location list: %s: %v", string(sr.Data), err)
	}
	return locs
}

// parseSingleLocation extracts a single node location from a standard response
func parseSingleLocation(t *testing.T, sr standardResponse) nodeLocation {
	t.Helper()
	var loc nodeLocation
	err := json.Unmarshal(sr.Data, &loc)
	if err != nil {
		t.Fatalf("unmarshalling single location: %s: %v", string(sr.Data), err)
	}
	return loc
}

// waitForLocation polls GET /node/{nid} with exponential backoff until the
// specified location appears with Stored: true. Returns the final nodeData and
// whether the location was found within the timeout (30s).
func waitForLocation(t *testing.T, auth, nodeID, locationID string, timeout time.Duration) (nodeData, bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	interval := 500 * time.Millisecond
	maxInterval := 4 * time.Second

	for time.Now().Before(deadline) {
		resp := doRequest(t, "GET", "/node/"+nodeID, auth, nil, "")
		sr := parseStandardResponse(t, resp)
		if sr.Status != http.StatusOK {
			t.Fatalf("waitForLocation: expected 200 getting node, got %d", sr.Status)
		}
		nd := parseNodeData(t, sr)

		for _, loc := range nd.Locations {
			if loc.ID == locationID && loc.Stored {
				return nd, true
			}
		}

		time.Sleep(interval)
		interval *= 2
		if interval > maxInterval {
			interval = maxInterval
		}
	}

	// Final attempt
	resp := doRequest(t, "GET", "/node/"+nodeID, auth, nil, "")
	sr := parseStandardResponse(t, resp)
	nd := parseNodeData(t, sr)
	for _, loc := range nd.Locations {
		if loc.ID == locationID && loc.Stored {
			return nd, true
		}
	}
	return nd, false
}
