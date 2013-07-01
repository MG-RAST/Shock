package conf

import (
	"flag"
	"fmt"
	"github.com/jaredwilkening/goconfig/config"
	"os"
	"strconv"
	"strings"
)

type idxOpts struct {
	unique   bool
	dropDups bool
	sparse   bool
}

// Setup conf variables
var (
	Conf = map[string]string{}

	// Reload
	RELOAD = ""

	// Config File
	CONFIG_FILE = ""

	// Node Indices
	NODE_IDXS map[string]idxOpts = nil

	// Default Chunksize for size virtual index
	CHUNK_SIZE int64 = 1048576
)

func Initialize() {
	flag.StringVar(&CONFIG_FILE, "conf", "/usr/local/shock/conf/shock.cfg", "path to config file")
	flag.StringVar(&RELOAD, "reload", "", "path or url to shock data. WARNING this will drop all current data.")
	flag.Parse()
	c, err := config.ReadDefault(CONFIG_FILE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error reading conf file: %v\n", err)
		os.Exit(1)
	}

	// Ports
	Conf["site-port"], _ = c.String("Ports", "site-port")
	Conf["api-port"], _ = c.String("Ports", "api-port")

	// URLs
	Conf["site-url"], _ = c.String("External", "site-url")
	Conf["api-url"], _ = c.String("External", "api-url")

	// SSL
	Conf["ssl"], _ = c.String("SSL", "enable")
	if Bool(Conf["ssl"]) {
		Conf["ssl-key"], _ = c.String("SSL", "key")
		Conf["ssl-cert"], _ = c.String("SSL", "cert")
	}

	// Access-Control
	Conf["anon-write"], _ = c.String("Anonymous", "write")
	Conf["anon-read"], _ = c.String("Anonymous", "read")
	Conf["anon-user"], _ = c.String("Anonymous", "create-user")

	// Auth
	Conf["auth-type"], _ = c.String("Auth", "type")
	switch Conf["auth-type"] {
	case "globus":
		Conf["globus_token_url"], _ = c.String("Auth", "globus_token_url")
		Conf["globus_profile_url"], _ = c.String("Auth", "globus_profile_url")
	case "oauth":
		Conf["oauth_request_token_url"], _ = c.String("Auth", "oauth_request_token_url")
		Conf["oauth_auth_token_url"], _ = c.String("Auth", "oauth_auth_token_url")
		Conf["oauth_access_token_url"], _ = c.String("Auth", "oauth_access_token_url")
	case "basic":
		// nothing yet
	}

	// Admin
	Conf["admin-email"], _ = c.String("Admin", "email")
	Conf["admin-secret"], _ = c.String("Admin", "secretkey")

	// Directories
	Conf["site-path"], _ = c.String("Directories", "site")
	Conf["data-path"], _ = c.String("Directories", "data")
	Conf["logs-path"], _ = c.String("Directories", "logs")
	Conf["local-paths"], _ = c.String("Directories", "local_paths")

	// Mongodb
	Conf["mongodb-hosts"], _ = c.String("Mongodb", "hosts")
	if Conf["mongodb-database"], err = c.String("Mongodb", "database"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Mongodb database must be set in config file.")
		os.Exit(1)
	}
	Conf["mongodb-user"], _ = c.String("Mongodb", "user")
	Conf["mongodb-password"], _ = c.String("Mongodb", "password")

	// parse Node-Indices
	NODE_IDXS = map[string]idxOpts{}
	nodeIdx, _ := c.Options("Node-Indices")
	for _, opt := range nodeIdx {
		val, _ := c.String("Node-Indices", opt)
		opts := idxOpts{}
		for _, parts := range strings.Split(val, ",") {
			p := strings.Split(parts, ":")
			if p[0] == "unique" {
				if p[1] == "true" {
					opts.unique = true
				} else {
					opts.unique = false
				}
			} else if p[0] == "dropDups" {
				if p[1] == "true" {
					opts.dropDups = true
				} else {
					opts.dropDups = false
				}
			} else if p[0] == "sparse" {
				if p[1] == "true" {
					opts.sparse = true
				} else {
					opts.sparse = false
				}
			}
		}
		NODE_IDXS[opt] = opts
	}

	Conf["perf-log"], _ = c.String("Log", "perf_log")
}

func Bool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

func Print() {
	fmt.Printf("##### Admin #####\nemail:\t%s\nsecretkey:\t%s\n\n", Conf["admin-email"], Conf["admin-secret"])
	fmt.Printf("####### Anonymous ######\nread:\t%s\nwrite:\t%s\ncreate-user:\t%s\n\n", Conf["anon-read"], Conf["anon-write"], Conf["anon-user"])
	if Conf["auth-type"] == "basic" {
		fmt.Printf("##### Auth #####\ntype:\tbasic\n\n")
	} else if Conf["auth-type"] == "globus" {
		fmt.Printf("##### Auth #####\ntype:\tglobus\ntoken_url:\t%s\nprofile_url:\t%s\n\n", Conf["globus_token_url"], Conf["globus_profile_url"])
	}
	fmt.Printf("##### Directories #####\nsite:\t%s\ndata:\t%s\nlogs:\t%s\nlocal_paths:\t%s\n\n", Conf["site-path"], Conf["data-path"], Conf["logs-path"], Conf["local-paths"])
	if Bool(Conf["ssl"]) {
		fmt.Printf("##### SSL #####\nenabled:\t%s\nkey:\t%s\ncert:\t%s\n\n", Conf["ssl"], Conf["ssl-key"], Conf["ssl-cert"])
	} else {
		fmt.Printf("##### SSL #####\nenabled:\t%s\n\n", Conf["ssl"])
	}
	fmt.Printf("##### Mongodb #####\nhost(s):\t%s\ndatabase:\t%s\n\n", Conf["mongodb-hosts"], Conf["mongodb-database"])
	fmt.Printf("##### Ports #####\nsite:\t%s\napi:\t%s\n\n", Conf["site-port"], Conf["api-port"])
	if Bool(Conf["perf-log"]) {
		fmt.Printf("##### PerfLog enabled #####\n\n")
	}
}
