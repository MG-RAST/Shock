package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/conf"
	"github.com/MG-RAST/Shock/logger"
	"github.com/jaredwilkening/goweb"
	"os"
)

var (
	log = logger.New()
)

func launchSite(control chan int, port int) {
	goweb.ConfigureDefaultFormatters()
	r := &goweb.RouteManager{}
	r.MapFunc("/raw", RawDir)
	r.MapFunc("/assets", AssetsDir)
	r.MapFunc("*", Site)
	err := goweb.ListenAndServeRoutes(fmt.Sprintf(":%d", conf.SITEPORT), r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: site: %v\n", err)
	}
	control <- 1 //we are ending
}

func launchAPI(control chan int, port int) {
	goweb.ConfigureDefaultFormatters()
	r := &goweb.RouteManager{}
	r.MapRest("/node", new(NodeController))
	r.MapRest("/user", new(UserController))
	r.MapFunc("*", ResourceDescription, goweb.GetMethod)
	err := goweb.ListenAndServeRoutes(fmt.Sprintf(":%d", port), r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: api: %v\n", err)
	}
	control <- 1 //we are ending
}

func main() {
	fmt.Printf("%s\n######### Conf #########\ndata-root:\t%s\naccess-log:\t%s\nerror-log:\t%s\nmongodb:\t%s\nsecretkey:\t%s\nsite-port:\t%d\napi-port:\t%d\n\n####### Anonymous ######\nread:\t%t\nwrite:\t%t\ncreate-user:\t%t\n\n",
		logo,
		conf.DATAPATH,
		conf.LOGSPATH+"/access.log",
		conf.LOGSPATH+"/error.log",
		conf.MONGODB,
		conf.SECRETKEY,
		conf.SITEPORT,
		conf.APIPORT,
		conf.ANONREAD,
		conf.ANONWRITE,
		conf.ANONCREATEUSER,
	)

	// reload
	if conf.RELOAD != "" {
		fmt.Println("####### Reloading #######")
		err := reload(conf.RELOAD)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		}
		fmt.Println("Done")
	}

	//launch server
	control := make(chan int)
	go log.Handle()
	go launchSite(control, conf.SITEPORT)
	go launchAPI(control, conf.APIPORT)
	<-control //block till something dies
}
