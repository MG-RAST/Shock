// Package conf parses start up args and config file
package conf

import (
	"flag"
	"fmt"
	"github.com/MG-RAST/golib/goconfig/config"
	"os"
	"strconv"
	"strings"
)

type idxOpts struct {
	unique   bool
	dropDups bool
	sparse   bool
}

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

	ANON_READ   = true
	ANON_WRITE  = false
	ANON_DELETE = false

	// Default is true, will rotate logs daily
	LOG_ROTATE = true
)

// Initialize is an explicit init. Enables outside use
// of shock-server packages. Parses config and populates
// the Conf variable.
func Initialize() {
	gopath := os.Getenv("GOPATH")
	flag.StringVar(&CONFIG_FILE, "conf", gopath+"/src/github.com/MG-RAST/Shock/shock-server.conf.template", "path to config file")
	flag.StringVar(&RELOAD, "reload", "", "path or url to shock data. WARNING this will drop all current data.")
	flag.Parse()
	c, err := config.ReadDefault(CONFIG_FILE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: error reading conf file: %v\n", err)
		os.Exit(1)
	}

	// Address
	Conf["api-ip"], _ = c.String("Address", "api-ip")
	Conf["api-port"], _ = c.String("Address", "api-port")

	// URLs
	Conf["api-url"], _ = c.String("External", "api-url")

	// SSL
	Conf["ssl"], _ = c.String("SSL", "enable")
	if Bool(Conf["ssl"]) {
		Conf["ssl-key"], _ = c.String("SSL", "key")
		Conf["ssl-cert"], _ = c.String("SSL", "cert")
	}

	// Access-Control
	ANON_READ, _ = c.Bool("Anonymous", "read")
	ANON_WRITE, _ = c.Bool("Anonymous", "write")
	ANON_DELETE, _ = c.Bool("Anonymous", "delete")

	// Auth
	Conf["basic_auth"], _ = c.String("Auth", "basic")
	Conf["globus_token_url"], _ = c.String("Auth", "globus_token_url")
	Conf["globus_profile_url"], _ = c.String("Auth", "globus_profile_url")
	Conf["mgrast_oauth_url"], _ = c.String("Auth", "mgrast_oauth_url")

	// Admin
	Conf["admin-email"], _ = c.String("Admin", "email")
	Conf["admin-users"], _ = c.String("Admin", "users")

	// Paths
	Conf["site-path"], _ = c.String("Paths", "site")
	Conf["data-path"], _ = c.String("Paths", "data")
	Conf["logs-path"], _ = c.String("Paths", "logs")
	Conf["local-paths"], _ = c.String("Paths", "local_paths")
	Conf["pidfile"], _ = c.String("Paths", "pidfile")

	// Runtime
	Conf["GOMAXPROCS"], _ = c.String("Runtime", "GOMAXPROCS")

	// Mongodb
	Conf["mongodb-hosts"], _ = c.String("Mongodb", "hosts")
	if Conf["mongodb-database"], err = c.String("Mongodb", "database"); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Mongodb database must be set in config file.")
		os.Exit(1)
	}
	Conf["mongodb-user"], _ = c.String("Mongodb", "user")
	Conf["mongodb-password"], _ = c.String("Mongodb", "password")
	Conf["mongodb-attribute-indexes"], _ = c.String("Mongodb", "attribute_indexes")

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
	LOG_ROTATE, _ = c.Bool("Log", "rotate")

}

// Bool is a convenience wrapper around strconv.ParseBool
func Bool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

// Print prints the configuration loads to stdout
func Print() {
	fmt.Printf("####### Anonymous ######\nread:\t%v\nwrite:\t%v\ndelete:\t%v\n\n", ANON_READ, ANON_WRITE, ANON_DELETE)
	if (Conf["globus_token_url"] != "" && Conf["globus_profile_url"] != "") || Conf["mgrast_oauth_url"] != "" {
		fmt.Printf("##### Auth #####\n")
		if Conf["globus_token_url"] != "" && Conf["globus_profile_url"] != "" {
			fmt.Printf("type:\tglobus\ntoken_url:\t%s\nprofile_url:\t%s\n\n", Conf["globus_token_url"], Conf["globus_profile_url"])
		}
		if Conf["mgrast_oauth_url"] != "" {
			fmt.Printf("type:\tmgrast\noauth_url:\t%s\n\n", Conf["mgrast_oauth_url"])
		}
	}
	fmt.Printf("##### Admin #####\nusers:\t%s\n\n", Conf["admin-users"])
	fmt.Printf("##### Paths #####\nsite:\t%s\ndata:\t%s\nlogs:\t%s\nlocal_paths:\t%s\n\n", Conf["site-path"], Conf["data-path"], Conf["logs-path"], Conf["local-paths"])
	if Bool(Conf["ssl"]) {
		fmt.Printf("##### SSL #####\nenabled:\t%s\nkey:\t%s\ncert:\t%s\n\n", Conf["ssl"], Conf["ssl-key"], Conf["ssl-cert"])
	} else {
		fmt.Printf("##### SSL #####\nenabled:\t%s\n\n", Conf["ssl"])
	}
	fmt.Printf("##### Mongodb #####\nhost(s):\t%s\ndatabase:\t%s\n\n", Conf["mongodb-hosts"], Conf["mongodb-database"])
	fmt.Printf("##### Address #####\nip:\t%s\nport:\t%s\n\n", Conf["api-ip"], Conf["api-port"])
	if Bool(Conf["perf-log"]) {
		fmt.Printf("##### PerfLog enabled #####\n\n")
	}
	if LOG_ROTATE {
		fmt.Printf("##### Log rotation enabled #####\n\n")
	} else {
		fmt.Printf("##### Log rotation disabled #####\n\n")
	}
}
