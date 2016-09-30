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
	"github.com/MG-RAST/Shock/shock-server/versions"
	"github.com/MG-RAST/golib/stretchr/goweb"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	longDateForm = "2006-01-02T15:04:05-07:00"
)

type anonymous struct {
	Read   bool `json:"read"`
	Write  bool `json:"write"`
	Delete bool `json:"delete"`
}

type resource struct {
	A []string  `json:"attribute_indexes"`
	C string    `json:"contact"`
	D string    `json:"documentation"`
	I string    `json:"id"`
	O []string  `json:"auth"`
	P anonymous `json:"anonymous_permissions"`
	R []string  `json:"resources"`
	S string    `json:"server_time"`
	T string    `json:"type"`
	U string    `json:"url"`
	V string    `json:"version"`
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
		logger.Infof("%s REQ RECEIVED \"%s %s%s\"", host, ctx.MethodString(), req.RequestURI, suffix)
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
		logger.Infof("RESPONDED TO %s \"%s %s%s\"", host, ctx.MethodString(), req.RequestURI, suffix)
		return nil
	})

	goweb.Map("/preauth/{id}", func(ctx context.Context) error {
		if ctx.HttpRequest().Method == "OPTIONS" {
			return responder.RespondOK(ctx)
		}
		pcon.PreAuthRequest(ctx)
		return nil
	})

	goweb.Map("/node/{nid}/acl/{type}", func(ctx context.Context) error {
		if ctx.HttpRequest().Method == "OPTIONS" {
			return responder.RespondOK(ctx)
		}
		acon.AclTypedRequest(ctx)
		return nil
	})

	goweb.Map("/node/{nid}/acl/", func(ctx context.Context) error {
		if ctx.HttpRequest().Method == "OPTIONS" {
			return responder.RespondOK(ctx)
		}
		acon.AclRequest(ctx)
		return nil
	})

	goweb.Map("/node/{nid}/index/{idxType}", func(ctx context.Context) error {
		if ctx.HttpRequest().Method == "OPTIONS" {
			return responder.RespondOK(ctx)
		}
		icon.IndexTypedRequest(ctx)
		return nil
	})

	goweb.Map("/openparts", func(ctx context.Context) error {
		ids := node.LockMgr.GetNodes()
		return responder.RespondWithData(ctx, ids)
	})

	goweb.Map("/", func(ctx context.Context) error {
		host := util.ApiUrl(ctx)

		attrs := strings.Split(conf.MONGODB_ATTRIBUTE_INDEXES, ",")
		for k, v := range attrs {
			attrs[k] = strings.TrimSpace(v)
		}

		anonPerms := new(anonymous)
		anonPerms.Read = conf.ANON_READ
		anonPerms.Write = conf.ANON_WRITE
		anonPerms.Delete = conf.ANON_DELETE

		var auth []string
		if conf.AUTH_GLOBUS_TOKEN_URL != "" && conf.AUTH_GLOBUS_PROFILE_URL != "" {
			auth = append(auth, "globus")
		}
		if conf.AUTH_MGRAST_OAUTH_URL != "" {
			auth = append(auth, "mgrast")
		}

		r := resource{
			A: attrs,
			C: conf.ADMIN_EMAIL,
			D: host + "/wiki/",
			I: "Shock",
			O: auth,
			P: *anonPerms,
			R: []string{"node"},
			S: time.Now().Format(longDateForm),
			T: "Shock",
			U: host + "/",
			V: "[% VERSION %]",
		}
		return responder.WriteResponseObject(ctx, http.StatusOK, r)
	})

	nodeController := new(ncon.NodeController)
	goweb.MapController(nodeController)

	goweb.MapStatic("/wiki", conf.PATH_SITE)

	// Map the favicon
	//goweb.MapStaticFile("/favicon.ico", "static-files/favicon.ico")

	// Catch-all handler for everything that we don't understand
	goweb.Map(func(ctx context.Context) error {
		return responder.RespondWithError(ctx, http.StatusBadRequest, "Parameters do not match a valid Shock request type.")
	})
}

func main() {
	var err error

	// init config
	err = conf.Initialize()
	if err != nil {
		fmt.Errorf("Err@db.Initialize: %v\n", err)
	}

	// init logging system
	logger.Initialize()
	logger.Info("Starting...")

	if conf.ANON_WRITE {
		warnstr := "Warning: ananoymous write is activated, only use for development !!!!"
		logger.Info(warnstr)
		fmt.Errorf("%s\n", warnstr)
	}

	if conf.ANON_DELETE {
		warnstr := "Warning: ananoymous delete is activated, only use for development !!!!"
		logger.Info(warnstr)
		fmt.Errorf("%s\n", warnstr)
	}

	// init database
	err = db.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@db.Initialize: %v\n", err)
		logger.Error("Err@db.Initialize: " + err.Error())
		os.Exit(1)
	}

	user.Initialize()
	node.Initialize()
	preauth.Initialize()
	auth.Initialize()
	node.InitReaper()
	err = versions.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@versions.Initialize: %v\n", err)
		logger.Error("Err@versions.Initialize: " + err.Error())
		os.Exit(1)
	}
	err = versions.RunVersionUpdates()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@versions.RunVersionUpdates: %v\n", err)
		logger.Error("Err@versions.RunVersionUpdates: " + err.Error())
		os.Exit(1)
	}
	// After version updates have succeeded without error, we can push the configured version numbers into the mongo db
	// Note: configured version numbers are configured in conf.go but are NOT user configurable by design
	err = versions.PushVersionsToDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@versions.PushVersionsToDatabase: %v\n", err)
		logger.Error("Err@versions.PushVersionsToDatabase: " + err.Error())
		os.Exit(1)
	}
	printLogo()
	conf.Print()
	if err := versions.Print(); err != nil {
		fmt.Fprintf(os.Stderr, "Err@versions.Print: %v\n", err)
		logger.Error("Err@versions.Print: " + err.Error())
		os.Exit(1)
	}

	// check if necessary directories exist or created
	for _, path := range []string{conf.PATH_SITE, conf.PATH_DATA, conf.PATH_LOGS, conf.PATH_DATA + "/temp"} {

		err = os.MkdirAll(path, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			logger.Errorf("error createing directory %s: %v", err)
			os.Exit(1)
		}

	}

	// reload
	if conf.RELOAD != "" {
		fmt.Println("####### Reloading #######")
		err := reload(conf.RELOAD)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			logger.Error("ERROR: " + err.Error())
			os.Exit(1)
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
	if conf.GOMAXPROCS != "" {
		if setting, err := strconv.Atoi(conf.GOMAXPROCS); err != nil {
			err_msg := "ERROR: could not interpret configured GOMAXPROCS value as integer.\n"
			fmt.Fprintf(os.Stderr, err_msg)
			logger.Error("ERROR: " + err_msg)
			os.Exit(1)
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

	if conf.PATH_PIDFILE != "" {
		f, err := os.Create(conf.PATH_PIDFILE)
		if err != nil {
			err_msg := "Could not create pid file: " + conf.PATH_PIDFILE + "\n"
			fmt.Fprintf(os.Stderr, err_msg)
			logger.Error("ERROR: " + err_msg)
			os.Exit(1)
		}
		defer f.Close()

		pid := os.Getpid()
		fmt.Fprintln(f, pid)

		fmt.Println("##### pidfile #####")
		fmt.Printf("pid: %d saved to file: %s\n\n", pid, conf.PATH_PIDFILE)
	}

	Address := fmt.Sprintf("%s:%d", conf.API_IP, conf.API_PORT)
	mapRoutes()

	s := &http.Server{
		Addr:           ":" + Address,
		Handler:        goweb.DefaultHttpHandler(),
		ReadTimeout:    48 * time.Hour,
		WriteTimeout:   48 * time.Hour,
		MaxHeaderBytes: 1 << 20,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	listener, listenErr := net.Listen("tcp", Address)

	if listenErr != nil {
		err_msg := "Could not listen - " + listenErr.Error() + "\n"
		fmt.Fprintf(os.Stderr, err_msg)
		logger.Error("ERROR: " + err_msg)
		os.Exit(1)
	}

	go node.Ttl.Handle()
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
