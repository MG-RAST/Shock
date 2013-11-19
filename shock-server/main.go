package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/auth"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/controller"
	"github.com/MG-RAST/Shock/shock-server/db"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/goweb"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

func launchSite(control chan int) {
	goweb.ConfigureDefaultFormatters()
	r := &goweb.RouteManager{}
	r.MapFunc("/raw", RawDir)
	r.MapFunc("/assets", AssetsDir)
	r.MapFunc("*", Site)
	if conf.Bool(conf.Conf["ssl"]) {
		err := goweb.ListenAndServeRoutesTLS(":"+conf.Conf["site-port"], conf.Conf["ssl-cert"], conf.Conf["ssl-key"], r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: site: %v\n", err)
			logger.Error("ERROR: site: " + err.Error())
		}
	} else {
		err := goweb.ListenAndServeRoutes(":"+conf.Conf["site-port"], r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: site: %v\n", err)
			logger.Error("ERROR: site: " + err.Error())
		}
	}
	control <- 1 //we are ending
}

func launchAPI(control chan int) {
	c := controller.New()
	goweb.ConfigureDefaultFormatters()
	r := &goweb.RouteManager{}
	r.MapFunc("/preauth/{id}", c.Preauth, goweb.GetMethod)
	r.Map("/node/{nid}/acl/{type}", c.Acl["typed"])
	r.Map("/node/{nid}/acl", c.Acl["base"])
	r.Map("/node/{nid}/index/{type}", c.Index)
	r.Map("/node/{nid}/index", c.Index)
	r.MapRest("/node", c.Node)
	r.MapFunc("*", ResourceDescription, goweb.GetMethod)
	r.MapFunc("*", RespondOk, goweb.OptionsMethod)
	if conf.Bool(conf.Conf["ssl"]) {
		err := goweb.ListenAndServeRoutesTLS(":"+conf.Conf["api-port"], conf.Conf["ssl-cert"], conf.Conf["ssl-key"], r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: api: %v\n", err)
			logger.Error("ERROR: api: " + err.Error())
		}
	} else {
		err := goweb.ListenAndServeRoutes(":"+conf.Conf["api-port"], r)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: api: %v\n", err)
			logger.Error("ERROR: api: " + err.Error())
		}
	}
	control <- 1 //we are ending
}

func main() {
	// init(s)
	conf.Initialize()
	logger.Initialize()
	if err := db.Initialize(); err != nil {
		logger.Error(err.Error())
	}
	user.Initialize()
	node.Initialize()
	preauth.Initialize()
	auth.Initialize()

	// print conf
	printLogo()
	conf.Print()

	if _, err := os.Stat(conf.Conf["data-path"] + "/temp"); err != nil && os.IsNotExist(err) {
		if err := os.Mkdir(conf.Conf["data-path"]+"/temp", 0777); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			logger.Error("ERROR: " + err.Error())
		}
	}

	// reload
	if conf.RELOAD != "" {
		fmt.Println("####### Reloading #######")
		err := reload(conf.RELOAD)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			logger.Error("ERROR: " + err.Error())
		}
		fmt.Println("Done")
	}

	// setting GOMAXPROCS
	var procs int
	avail := runtime.NumCPU()
	if avail <= 2 {
		procs = 1
	} else if avail == 3 {
		procs = 2
	} else {
		procs = avail - 2
	}

	fmt.Println("##### Procs #####")
	fmt.Printf("Number of available CPUs = %d\n", avail)
	if conf.Conf["GOMAXPROCS"] != "" {
		if setting, err := strconv.Atoi(conf.Conf["GOMAXPROCS"]); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: could not interpret configured GOMAXPROCS value as integer.")
		} else {
			procs = setting
		}
	}

	if procs <= avail {
		fmt.Printf("Running Shock server with GOMAXPROCS = %d\n", procs)
		runtime.GOMAXPROCS(procs)
	} else {
		fmt.Println("GOMAXPROCS config value is greater than available number of CPUs.")
		fmt.Printf("Running Shock server with GOMAXPROCS = %d\n", avail)
		runtime.GOMAXPROCS(avail)
	}

	//launch server
	control := make(chan int)
	go launchSite(control)
	go launchAPI(control)

	//checking to make sure that server has launched
	connect := false
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		_, err := http.Get("http://localhost:" + conf.Conf["api-port"])
		if err == nil {
			connect = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	//exiting if server could not be reached after 1 second
	if connect != true {
		fmt.Fprintln(os.Stderr, "ERROR: server could not be reached at "+"http://localhost:"+conf.Conf["api-port"])
		fmt.Fprintln(os.Stderr, "Exiting!")
		os.Exit(1)
	}

	fmt.Println("\nReady to receive requests...")

	<-control //block till something dies
}
