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
	CHUNK_SIZE int64 = 10485760
)

func Initialize(args []string) []string {
	// options
	fs := flag.FlagSet{}
	Flags["attributes"] = fs.String("attributes", "", "")
	Flags["full"] = fs.String("full", "", "")
	Flags["parts"] = fs.String("parts", "", "")
	Flags["part"] = fs.String("part", "", "")
	Flags["file"] = fs.String("file", "", "")
	Flags["threads"] = fs.String("threads", "", "")
	Flags["virtual_file"] = fs.String("virtual_file", "", "")
	Flags["remote_path"] = fs.String("remote_path", "", "")
	Flags["index"] = fs.String("index", "", "")
	Flags["index_options"] = fs.String("index_options", "", "")
	fs.StringVar(&ConfFile, "conf", DefaultPath(), "path to config file")
	fs.Parse(args)
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
	return fs.Args()
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
