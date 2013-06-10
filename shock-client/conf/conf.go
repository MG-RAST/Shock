package conf

import (
	"flag"
	"fmt"
	"github.com/jaredwilkening/goconfig/config"
	"os"
	"os/user"
)

type server struct {
	Url string
}

type cache struct {
	Dir            string
	MaxConnections int
}

type auth struct {
	Type       string
	TokenUrl   string
	ProfileUrl string
}

var (
	Auth                   = auth{}
	Server                 = server{}
	Cache                  = cache{}
	ConfFile               = ""
	Flags                  = map[string]*string{}
	Examples         *bool = nil
	DOWNLOAD_THREADS       = 4
	// Default Chunksize for size virtual index
	CHUNK_SIZE int64 = 1048576
)

func init() {
	// options
	flag.Bool("h", false, "")
	flag.Bool("help", false, "")
	Examples = flag.Bool("examples", false, "")
	Flags["attributes"] = flag.String("attributes", "", "")
	Flags["full"] = flag.String("full", "", "")
	Flags["parts"] = flag.String("parts", "", "")
	Flags["part"] = flag.String("part", "", "")
	Flags["file"] = flag.String("file", "", "")
	Flags["threads"] = flag.String("threads", "", "")
	Flags["virtual_file"] = flag.String("virtual_file", "", "")
	Flags["remote_path"] = flag.String("remote_path", "", "")
	Flags["index"] = flag.String("index", "", "")
	Flags["index_options"] = flag.String("index_options", "", "")
	flag.StringVar(&ConfFile, "conf", DefaultPath(), "path to config file")
	flag.Parse()
	c, err := config.ReadDefault(ConfFile)
	handle(err)

	// Cache
	Cache.Dir, err = c.String("Cache", "dir")
	Cache.MaxConnections, err = c.Int("Cache", "max_connections")

	// Server
	Server.Url, err = c.String("Server", "url")
	handle(err)

	// Auth
	Auth.Type, _ = c.String("Auth", "type")
	switch Auth.Type {
	case "globus":
		Auth.TokenUrl, _ = c.String("Auth", "token_url")
		Auth.ProfileUrl, _ = c.String("Auth", "profile_url")
	case "basic":
		// nothing yet
	}
}

func handle(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: reading conf file: %v\n", err)
		os.Exit(1)
	}
	return
}

func DefaultPath() (path string) {
	u, _ := user.Current()
	path = u.HomeDir + "/.shock-client.cfg"
	return
}

func Print() {
	fmt.Printf("##### Server #####\nurl:\t%s\n\n", Server.Url)
	fmt.Printf("##### Cache  #####\ndir:\t%s\nmax_connections:\t%d\n\n", Cache.Dir, Cache.MaxConnections)
}
