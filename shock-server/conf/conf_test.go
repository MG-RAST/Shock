package conf_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/golib/goconfig/config"
	"github.com/stretchr/testify/assert"
)

// TestBool tests the Bool function
func TestBool(t *testing.T) {
	// Test with "true"
	assert.True(t, conf.Bool("true"), "Bool should return true for 'true'")

	// Test with "True"
	assert.True(t, conf.Bool("True"), "Bool should return true for 'True'")

	// Test with "TRUE"
	assert.True(t, conf.Bool("TRUE"), "Bool should return true for 'TRUE'")

	// Test with "1"
	assert.True(t, conf.Bool("1"), "Bool should return true for '1'")

	// Test with "false"
	assert.False(t, conf.Bool("false"), "Bool should return false for 'false'")

	// Test with "False"
	assert.False(t, conf.Bool("False"), "Bool should return false for 'False'")

	// Test with "FALSE"
	assert.False(t, conf.Bool("FALSE"), "Bool should return false for 'FALSE'")

	// Test with "0"
	assert.False(t, conf.Bool("0"), "Bool should return false for '0'")

	// Test with invalid value
	assert.False(t, conf.Bool("invalid"), "Bool should return false for invalid value")

	// Test with empty string
	assert.False(t, conf.Bool(""), "Bool should return false for empty string")
}

// TestInitialize tests the Initialize function
func TestInitialize(t *testing.T) {
	// Save original command line arguments and restore them after the test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "conf-test")
	assert.NoError(t, err, "Creating temp directory should not error")
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configFile := filepath.Join(tempDir, "shock-server.conf")
	configContent := `
[Admin]
email=admin@example.com
users=admin1,admin2

[Anonymous]
read=true
write=false
delete=false

[Address]
api-ip=127.0.0.1
api-port=7445

[Auth]
basic=true
globus_token_url=https://nexus.api.globusonline.org/goauth/token
globus_profile_url=https://nexus.api.globusonline.org/users
oauth_urls=https://api.example.com/user
oauth_bearers=oauth
cache_timeout=60
use_auth=true

[Runtime]
expire_wait=60
GOMAXPROCS=4
max_revisions=3

[Log]
perf_log=false
rotate=true
logoutput=both
trace=false
debuglevel=0

[Mongodb]
attribute_indexes=
database=ShockDB
hosts=localhost
password=
user=

[Node-Indices]
size=unique:true,dropDups:false,sparse:false

[Paths]
site=/usr/local/shock/site
data=/usr/local/shock/data
logs=/var/log/shock
local_paths=/var/tmp
pidfile=

[Cache]
cache_path=

[Migrate]
min_replica_count=2
node_migration=false
node_data_removal=false

[SSL]
enable=false
key=
cert=
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	assert.NoError(t, err, "Writing test config file should not error")

	// Create a test Locations.yaml file
	locationsFile := filepath.Join(tempDir, "Locations.yaml")
	locationsContent := `
Locations:
  - ID: local
    Type: local
    URL: file:///
    Description: Local file system
    Persistent: true
    Priority: 0
    MinPriority: 0
    Tier: 0
    Cost: 0
`
	err = os.WriteFile(locationsFile, []byte(locationsContent), 0644)
	assert.NoError(t, err, "Writing test Locations.yaml file should not error")

	// Create a test Types.yaml file
	typesFile := filepath.Join(tempDir, "Types.yaml")
	typesContent := `
Types:
  - ID: default
    Description: Default type
    Priority: 0
`
	err = os.WriteFile(typesFile, []byte(typesContent), 0644)
	assert.NoError(t, err, "Writing test Types.yaml file should not error")

	// Set command line arguments to use the test config file
	os.Args = []string{"shock-server", "-conf=" + configFile}

	// Initialize the configuration
	err = conf.Initialize()
	assert.NoError(t, err, "Initialize should not error")

	// Verify that the configuration was loaded correctly
	assert.Equal(t, "admin@example.com", conf.ADMIN_EMAIL, "ADMIN_EMAIL should be set correctly")
	assert.Equal(t, "admin1,admin2", conf.ADMIN_USERS, "ADMIN_USERS should be set correctly")
	assert.Equal(t, true, conf.ANON_READ, "ANON_READ should be set correctly")
	assert.Equal(t, false, conf.ANON_WRITE, "ANON_WRITE should be set correctly")
	assert.Equal(t, false, conf.ANON_DELETE, "ANON_DELETE should be set correctly")
	assert.Equal(t, "127.0.0.1", conf.API_IP, "API_IP should be set correctly")
	assert.Equal(t, 7445, conf.API_PORT, "API_PORT should be set correctly")
	assert.Equal(t, true, conf.AUTH_BASIC, "AUTH_BASIC should be set correctly")
	assert.Equal(t, "https://nexus.api.globusonline.org/goauth/token", conf.AUTH_GLOBUS_TOKEN_URL, "AUTH_GLOBUS_TOKEN_URL should be set correctly")
	assert.Equal(t, "https://nexus.api.globusonline.org/users", conf.AUTH_GLOBUS_PROFILE_URL, "AUTH_GLOBUS_PROFILE_URL should be set correctly")
	assert.Equal(t, "https://api.example.com/user", conf.AUTH_OAUTH_URL_STR, "AUTH_OAUTH_URL_STR should be set correctly")
	assert.Equal(t, "oauth", conf.AUTH_OAUTH_BEARER_STR, "AUTH_OAUTH_BEARER_STR should be set correctly")
	assert.Equal(t, 60, conf.AUTH_CACHE_TIMEOUT, "AUTH_CACHE_TIMEOUT should be set correctly")
	assert.Equal(t, true, conf.USE_AUTH, "USE_AUTH should be set correctly")
	assert.Equal(t, 60, conf.EXPIRE_WAIT, "EXPIRE_WAIT should be set correctly")
	assert.Equal(t, "4", conf.GOMAXPROCS, "GOMAXPROCS should be set correctly")
	assert.Equal(t, 3, conf.MAX_REVISIONS, "MAX_REVISIONS should be set correctly")
	assert.Equal(t, false, conf.LOG_PERF, "LOG_PERF should be set correctly")
	assert.Equal(t, true, conf.LOG_ROTATE, "LOG_ROTATE should be set correctly")
	assert.Equal(t, "both", conf.LOG_OUTPUT, "LOG_OUTPUT should be set correctly")
	assert.Equal(t, false, conf.LOG_TRACE, "LOG_TRACE should be set correctly")
	assert.Equal(t, 0, conf.DEBUG_LEVEL, "DEBUG_LEVEL should be set correctly")
	assert.Equal(t, "", conf.MONGODB_ATTRIBUTE_INDEXES, "MONGODB_ATTRIBUTE_INDEXES should be set correctly")
	assert.Equal(t, "ShockDB", conf.MONGODB_DATABASE, "MONGODB_DATABASE should be set correctly")
	assert.Equal(t, "localhost", conf.MONGODB_HOSTS, "MONGODB_HOSTS should be set correctly")
	assert.Equal(t, "", conf.MONGODB_PASSWORD, "MONGODB_PASSWORD should be set correctly")
	assert.Equal(t, "", conf.MONGODB_USER, "MONGODB_USER should be set correctly")
	assert.Equal(t, "/usr/local/shock/site", conf.PATH_SITE, "PATH_SITE should be set correctly")
	assert.Equal(t, "/usr/local/shock/data", conf.PATH_DATA, "PATH_DATA should be set correctly")
	assert.Equal(t, "/var/log/shock", conf.PATH_LOGS, "PATH_LOGS should be set correctly")
	assert.Equal(t, "/var/tmp", conf.PATH_LOCAL, "PATH_LOCAL should be set correctly")
	assert.Equal(t, "", conf.PATH_PIDFILE, "PATH_PIDFILE should be set correctly")
	assert.Equal(t, "", conf.PATH_CACHE, "PATH_CACHE should be set correctly")
	assert.Equal(t, 2, conf.MIN_REPLICA_COUNT, "MIN_REPLICA_COUNT should be set correctly")
	assert.Equal(t, false, conf.NODE_MIGRATION, "NODE_MIGRATION should be set correctly")
	assert.Equal(t, false, conf.NODE_DATA_REMOVAL, "NODE_DATA_REMOVAL should be set correctly")
	assert.Equal(t, false, conf.SSL, "SSL should be set correctly")
	assert.Equal(t, "", conf.SSL_KEY, "SSL_KEY should be set correctly")
	assert.Equal(t, "", conf.SSL_CERT, "SSL_CERT should be set correctly")

	// Verify that the Locations and Types were loaded correctly
	assert.NotNil(t, conf.LocationsMap, "LocationsMap should not be nil")
	assert.Contains(t, conf.LocationsMap, "local", "LocationsMap should contain 'local'")
	assert.NotNil(t, conf.TypesMap, "TypesMap should not be nil")
	assert.Contains(t, conf.TypesMap, "default", "TypesMap should contain 'default'")
}

// TestPrint tests the Print function
func TestPrint(t *testing.T) {
	// This is a simple test to ensure that Print doesn't panic
	conf.Print()
}

// TestCleanPath tests the cleanPath function
func TestCleanPath(t *testing.T) {
	// Test with a path with trailing slash
	path := "/path/to/dir/"
	cleanedPath := conf.CleanPath(path)
	assert.Equal(t, "/path/to/dir", cleanedPath, "Path should be cleaned correctly")

	// Test with a path without trailing slash
	path = "/path/to/dir"
	cleanedPath = conf.CleanPath(path)
	assert.Equal(t, "/path/to/dir", cleanedPath, "Path should be cleaned correctly")

	// Test with a relative path (CleanPath uses filepath.Abs, so relative paths become absolute)
	path = "path/to/dir"
	cleanedPath = conf.CleanPath(path)
	expectedAbs, _ := filepath.Abs("path/to/dir")
	assert.Equal(t, expectedAbs, cleanedPath, "Relative path should be converted to absolute")

	// Test with an empty path
	path = ""
	cleanedPath = conf.CleanPath(path)
	assert.Equal(t, "", cleanedPath, "Path should be cleaned correctly")
}

// TestParseConfiguration tests the parseConfiguration function
func TestParseConfiguration(t *testing.T) {
	// Save original values and restore them after the test
	originalAdminUsers := conf.ADMIN_USERS
	originalAuthOAuthURLStr := conf.AUTH_OAUTH_URL_STR
	originalAuthOAuthBearerStr := conf.AUTH_OAUTH_BEARER_STR
	originalLogOutput := conf.LOG_OUTPUT
	originalPathSite := conf.PATH_SITE
	originalPathData := conf.PATH_DATA
	originalPathLogs := conf.PATH_LOGS
	originalPathLocal := conf.PATH_LOCAL
	originalPathPidfile := conf.PATH_PIDFILE
	originalPathCache := conf.PATH_CACHE

	defer func() {
		conf.ADMIN_USERS = originalAdminUsers
		conf.AUTH_OAUTH_URL_STR = originalAuthOAuthURLStr
		conf.AUTH_OAUTH_BEARER_STR = originalAuthOAuthBearerStr
		conf.LOG_OUTPUT = originalLogOutput
		conf.PATH_SITE = originalPathSite
		conf.PATH_DATA = originalPathData
		conf.PATH_LOGS = originalPathLogs
		conf.PATH_LOCAL = originalPathLocal
		conf.PATH_PIDFILE = originalPathPidfile
		conf.PATH_CACHE = originalPathCache
	}()

	// Clear global state that may have been set by previous tests
	conf.AdminUsers = nil
	conf.AUTH_OAUTH = make(map[string]string)
	conf.OAUTH_DEFAULT = ""

	// Set up test values
	conf.ADMIN_USERS = "admin1,admin2, admin3"
	conf.AUTH_OAUTH_URL_STR = "https://api1.example.com/user,https://api2.example.com/user"
	conf.AUTH_OAUTH_BEARER_STR = "oauth1,oauth2"
	conf.LOG_OUTPUT = "both"
	conf.PATH_SITE = "/path/to/site/"
	conf.PATH_DATA = "/path/to/data/"
	conf.PATH_LOGS = "/path/to/logs/"
	conf.PATH_LOCAL = "/path/to/local/"
	conf.PATH_PIDFILE = "/path/to/pidfile/"
	conf.PATH_CACHE = "/path/to/cache/"

	// Parse the configuration
	err := conf.ParseConfiguration()
	assert.NoError(t, err, "ParseConfiguration should not error")

	// Verify that the configuration was parsed correctly
	assert.Len(t, conf.AdminUsers, 3, "AdminUsers should have 3 elements")
	assert.Contains(t, conf.AdminUsers, "admin1", "AdminUsers should contain 'admin1'")
	assert.Contains(t, conf.AdminUsers, "admin2", "AdminUsers should contain 'admin2'")
	assert.Contains(t, conf.AdminUsers, "admin3", "AdminUsers should contain 'admin3'")

	assert.Len(t, conf.AUTH_OAUTH, 2, "AUTH_OAUTH should have 2 elements")
	assert.Equal(t, "https://api1.example.com/user", conf.AUTH_OAUTH["oauth1"], "AUTH_OAUTH['oauth1'] should be set correctly")
	assert.Equal(t, "https://api2.example.com/user", conf.AUTH_OAUTH["oauth2"], "AUTH_OAUTH['oauth2'] should be set correctly")
	assert.Equal(t, "https://api1.example.com/user", conf.OAUTH_DEFAULT, "OAUTH_DEFAULT should be set correctly")

	assert.Equal(t, "/path/to/site", conf.PATH_SITE, "PATH_SITE should be cleaned")
	assert.Equal(t, "/path/to/data", conf.PATH_DATA, "PATH_DATA should be cleaned")
	assert.Equal(t, "/path/to/logs", conf.PATH_LOGS, "PATH_LOGS should be cleaned")
	assert.Equal(t, "/path/to/local", conf.PATH_LOCAL, "PATH_LOCAL should be cleaned")
	assert.Equal(t, "/path/to/pidfile", conf.PATH_PIDFILE, "PATH_PIDFILE should be cleaned")
	assert.Equal(t, "/path/to/cache", conf.PATH_CACHE, "PATH_CACHE should be cleaned")
}

// TestGetConfiguration tests the getConfiguration function
func TestGetConfiguration(t *testing.T) {
	// Save and restore os.Args (GetConfiguration parses command-line flags)
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"shock-server"}

	// Create a test config
	c := config.NewDefault()

	// Get the configuration
	c_store, err := conf.GetConfiguration(c)
	assert.NoError(t, err, "GetConfiguration should not error")
	assert.NotNil(t, c_store, "Config store should not be nil")
}
