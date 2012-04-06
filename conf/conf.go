package conf

import (
	"flag"
	"os"
)

// Command line options
var (
	PORT      = flag.Int("port", 8000, "port to listen on")
	DATAROOT  = flag.String("data", "", "data directory to store on disk files")
	MONGODB   = flag.String("mongodb", "localhost", "hostname(s) of mongodb")
	SECRETKEY = flag.String("secretkey", "supersupersecret", "secret key")
	SITEPATH  = os.ExpandEnv("$GOPATH") + "/src/github.com/MG-RAST/Shock/site"
)

func init() {
	flag.Parse()
}
