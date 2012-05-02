package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/conf"
	"github.com/MG-RAST/Shock/goweb"
)

func main() {
	fmt.Printf("%s\n######### Conf #########\ndata-root:\t%s\nmongodb:\t%s\nsecretkey:\t%s\nsite-port:\t%d\napi-port:\t%d\n\n####### Starting #######\n",
		logo,
		conf.DATAPATH,
		conf.MONGODB,
		conf.SECRETKEY,
		conf.SITEPORT,
		conf.APIPORT,
	)

	c := make(chan int)
	goweb.ConfigureDefaultFormatters()
	// start site
	go func() {
		r := &goweb.RouteManager{}
		r.MapFunc("/raw", RawDir)
		r.MapFunc("/assets", AssetsDir)
		r.MapFunc("*", Site)
		c <- 1
		goweb.ListenAndServeRoutes(fmt.Sprintf(":%d", conf.SITEPORT), r)
		c <- 1
	}()
	<-c
	fmt.Printf("site :%d... running\n", conf.SITEPORT)

	// start api
	go func() {
		r := &goweb.RouteManager{}
		r.MapRest("/node", new(NodeController))
		r.MapRest("/user", new(UserController))
		r.MapFunc("*", ResourceDescription, goweb.GetMethod)
		c <- 1
		goweb.ListenAndServeRoutes(fmt.Sprintf(":%d", conf.APIPORT), r)
		c <- 1
	}()
	<-c
	fmt.Printf("api  :%d... running\n", conf.APIPORT)
	fmt.Printf("\n######### Log  #########\n")
	<-c
}
