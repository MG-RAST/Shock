package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/auth"
	"github.com/MG-RAST/Shock/shock-server/conf"
	ncon "github.com/MG-RAST/Shock/shock-server/controller/node"
	acon "github.com/MG-RAST/Shock/shock-server/controller/node/acl"
	icon "github.com/MG-RAST/Shock/shock-server/controller/node/index"
	pcon "github.com/MG-RAST/Shock/shock-server/controller/preauth"
	"github.com/MG-RAST/Shock/shock-server/db"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/stretchr/goweb"
	"github.com/stretchr/goweb/context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
)

type resource struct {
	R []string `json:"resources"`
	U string   `json:"url"`
	D string   `json:"documentation"`
	C string   `json:"contact"`
	I string   `json:"id"`
	T string   `json:"type"`
}

func mapRoutes() {
	goweb.MapBefore(func(ctx context.Context) error {
		req := ctx.HttpRequest()
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		if host == "::1" {
			host = "localhost"
		}
		suffix := ""
		if _, ok := req.Header["Authorization"]; ok {
			suffix += " AUTH"
		}
		if l, has := req.Header["Content-Length"]; has {
			suffix += " Content-Length: " + l[0]
		}
		logger.Info("access", fmt.Sprintf("%s REQ RECEIVED \"%s %s%s\"", host, ctx.MethodString(), req.RequestURI, suffix))
		return nil
	})

	goweb.MapAfter(func(ctx context.Context) error {
		req := ctx.HttpRequest()
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		if host == "::1" {
			host = "localhost"
		}
		suffix := ""
		if _, ok := req.Header["Authorization"]; ok {
			suffix += " AUTH"
		}
		if l, has := req.Header["Content-Length"]; has {
			suffix += " Content-Length: " + l[0]
		}
		logger.Info("access", fmt.Sprintf("RESPONDED TO %s \"%s %s%s\"", host, ctx.MethodString(), req.RequestURI, suffix))
		return nil
	})

	goweb.Map("/preauth/{id}", func(ctx context.Context) error {
		pcon.PreAuthRequest(ctx)
		return nil
	})

	goweb.Map("/node/{nid}/acl/{type}", func(ctx context.Context) error {
		acon.AclTypedRequest(ctx)
		return nil
	})

	goweb.Map("/node/{nid}/acl/", func(ctx context.Context) error {
		acon.AclRequest(ctx)
		return nil
	})

	goweb.Map("/node/{nid}/index/{idxType}", func(ctx context.Context) error {
		icon.IndexTypedRequest(ctx)
		return nil
	})

	goweb.Map("/", func(ctx context.Context) error {
		host := util.ApiUrl(ctx)
		r := resource{
			R: []string{"node"},
			U: host + "/",
			D: host + "/documentation.html",
			C: conf.Conf["admin-email"],
			I: "Shock",
			T: "Shock",
		}
		return responder.WriteResponseObject(ctx, http.StatusOK, r)
	})

	nodeController := new(ncon.NodeController)
	goweb.MapController(nodeController)

	goweb.MapStatic("/assets", conf.Conf["site-path"]+"/assets")
	goweb.MapStaticFile("/documentation.html", conf.Conf["site-path"]+"/pages/main.html")

	// Map the favicon
	//goweb.MapStaticFile("/favicon.ico", "static-files/favicon.ico")

	// Catch-all handler for everything that we don't understand
	goweb.Map(func(ctx context.Context) error {
		return responder.RespondWithError(ctx, http.StatusBadRequest, "Parameters do not match a valid Shock request type.")
	})
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
		fmt.Printf("Running Shock server with GOMAXPROCS = %d\n\n", procs)
		runtime.GOMAXPROCS(procs)
	} else {
		fmt.Println("GOMAXPROCS config value is greater than available number of CPUs.")
		fmt.Printf("Running Shock server with GOMAXPROCS = %d\n\n", avail)
		runtime.GOMAXPROCS(avail)
	}

	f, err := os.Create(conf.Conf["pidfile"])
	if err != nil {
		return
	}
	defer f.Close()

	pid := os.Getpid()
	fmt.Fprintln(f, pid)

	fmt.Println("##### pidfile #####")
	fmt.Printf("pid: %d saved to file: %s\n\n", pid, conf.Conf["pidfile"])

	Address := conf.Conf["api-ip"] + ":" + conf.Conf["api-port"]
	mapRoutes()

	s := &http.Server{
		Addr:           ":" + Address,
		Handler:        goweb.DefaultHttpHandler(),
		ReadTimeout:    0,
		WriteTimeout:   0,
		MaxHeaderBytes: 1 << 20,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	listener, listenErr := net.Listen("tcp", Address)

	if listenErr != nil {
		fmt.Fprintf(os.Stderr, "Could not listen: %s\n", listenErr)
	}

	go func() {
		for _ = range c {
			// sig is a ^C, handle it

			// stop the HTTP server
			fmt.Fprintln(os.Stderr, "Stopping the server...")
			listener.Close()
		}
	}()

	fmt.Fprintf(os.Stderr, "Error in Serve: %s\n", s.Serve(listener))
}
