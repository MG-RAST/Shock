package main

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/MG-RAST/Shock/shock-server/auth"
	"github.com/MG-RAST/Shock/shock-server/cache"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/ui"
	ncon "github.com/MG-RAST/Shock/shock-server/controller/node"
	acon "github.com/MG-RAST/Shock/shock-server/controller/node/acl"
	icon "github.com/MG-RAST/Shock/shock-server/controller/node/index"
	pcon "github.com/MG-RAST/Shock/shock-server/controller/preauth"
	"github.com/MG-RAST/Shock/shock-server/db"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/location"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/types"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/Shock/shock-server/versions"
	"github.com/go-chi/chi/v5"
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
	A      []string  `json:"attribute_indexes"`
	C      string    `json:"contact"`
	I      string    `json:"id"`
	O      []string  `json:"auth"`
	P      anonymous `json:"anonymous_permissions"`
	R      []string  `json:"resources"`
	S      string    `json:"server_time"`
	T      string    `json:"type"`
	U      string    `json:"url"`
	Uptime string    `json:"uptime"`
	V      string    `json:"version"`
}

var StartTime = time.Now()

// requestLogger is middleware that logs incoming requests and responses.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, _ := net.SplitHostPort(r.RemoteAddr)
		if host == "::1" {
			host = "localhost"
		}
		suffix := ""
		if _, ok := r.Header["Authorization"]; ok {
			suffix += " AUTH"
		}
		if l, has := r.Header["Content-Length"]; has {
			suffix += " Content-Length: " + l[0]
		}
		logger.Infof("%s REQ RECEIVED \"%s %s%s\"", host, r.Method, r.RequestURI, suffix)
		next.ServeHTTP(w, r)
		logger.Infof("RESPONDED TO %s \"%s %s%s\"", host, r.Method, r.RequestURI, suffix)
	})
}

// corsMiddleware handles CORS headers and OPTIONS preflight globally.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(requestLogger)
	router.Use(corsMiddleware)

	// Specific sub-resource routes first
	router.HandleFunc("/preauth/{id}", pcon.PreAuthRequest)
	router.HandleFunc("/node/{nid}/acl/{type}", acon.AclTypedRequest)
	router.HandleFunc("/node/{nid}/acl/", acon.AclRequest)
	router.HandleFunc("/node/{nid}/index/{idxType}", icon.IndexTypedRequest)
	router.HandleFunc("/node/{nid}/locations/{loc}", node.LocationsRequest)
	router.HandleFunc("/node/{nid}/locations/", node.LocationsRequest)
	router.HandleFunc("/node/{nid}/restore/{val}", node.RestoreRequest)
	router.HandleFunc("/node/{nid}/restore/", node.RestoreRequest)
	router.HandleFunc("/location/{loc}/{function}", location.LocRequest)
	router.HandleFunc("/types/{type}/{function}/", types.TypeRequest)

	// view lock status
	router.Get("/locker", func(w http.ResponseWriter, r *http.Request) {
		ids := locker.NodeLockMgr.GetAll()
		responder.RespondWithData(w, r, ids)
	})
	router.Get("/locked/node", func(w http.ResponseWriter, r *http.Request) {
		ids := locker.NodeLockMgr.GetLocked()
		responder.RespondWithData(w, r, ids)
	})
	router.Get("/locked/file", func(w http.ResponseWriter, r *http.Request) {
		ids := locker.FileLockMgr.GetAll()
		responder.RespondWithData(w, r, ids)
	})
	router.Get("/locked/index", func(w http.ResponseWriter, r *http.Request) {
		ids := locker.IndexLockMgr.GetAll()
		responder.RespondWithData(w, r, ids)
	})

	// admin control of trace file
	router.Get("/trace/start", func(w http.ResponseWriter, r *http.Request) {
		u, err := request.Authenticate(r)
		if err != nil && err.Error() != e.NoAuth {
			request.AuthError(err, w, r)
			return
		}
		if u == nil || !u.Admin {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.NoAuth)
			return
		}
		fname := traceFileName()
		err = startTrace(fname)
		if err != nil {
			responder.RespondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("unable to start trace: %s", err.Error()))
			return
		}
		responder.RespondWithData(w, r, fmt.Sprintf("trace started: %s", fname))
	})
	router.Get("/trace/stop", func(w http.ResponseWriter, r *http.Request) {
		u, err := request.Authenticate(r)
		if err != nil && err.Error() != e.NoAuth {
			request.AuthError(err, w, r)
			return
		}
		if u == nil || !u.Admin {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.NoAuth)
			return
		}
		err = stopTrace()
		if err != nil {
			responder.RespondWithError(w, r, http.StatusInternalServerError, fmt.Sprintf("error stopping trace: %s", err.Error()))
			return
		}
		responder.RespondWithData(w, r, "trace stopped")
	})

	// download latest trace file
	router.Get("/trace/download", func(w http.ResponseWriter, r *http.Request) {
		u, err := request.Authenticate(r)
		if err != nil && err.Error() != e.NoAuth {
			request.AuthError(err, w, r)
			return
		}
		if u == nil || !u.Admin {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.NoAuth)
			return
		}
		path, err := latestTraceFile()
		if err != nil {
			responder.RespondWithError(w, r, http.StatusNotFound, err.Error())
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(path)))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, path)
	})

	// trace summary (footprint)
	router.Get("/trace/summary", func(w http.ResponseWriter, r *http.Request) {
		u, err := request.Authenticate(r)
		if err != nil && err.Error() != e.NoAuth {
			request.AuthError(err, w, r)
			return
		}
		if u == nil || !u.Admin {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.NoAuth)
			return
		}
		path, err := latestTraceFile()
		if err != nil {
			responder.RespondWithError(w, r, http.StatusNotFound, err.Error())
			return
		}
		out, err := runGoToolTrace(path, "footprint")
		if err != nil {
			responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(out)
	})

	// trace parsed events
	router.Get("/trace/events", func(w http.ResponseWriter, r *http.Request) {
		u, err := request.Authenticate(r)
		if err != nil && err.Error() != e.NoAuth {
			request.AuthError(err, w, r)
			return
		}
		if u == nil || !u.Admin {
			responder.RespondWithError(w, r, http.StatusUnauthorized, e.NoAuth)
			return
		}
		path, err := latestTraceFile()
		if err != nil {
			responder.RespondWithError(w, r, http.StatusNotFound, err.Error())
			return
		}
		out, err := runGoToolTrace(path, "parsed")
		if err != nil {
			responder.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write(out)
	})

	// Root endpoint
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		host := util.ApiUrl(r)

		attrs := strings.Split(conf.MONGODB_ATTRIBUTE_INDEXES, ",")
		for k, v := range attrs {
			attrs[k] = strings.TrimSpace(v)
		}

		anonPerms := new(anonymous)
		anonPerms.Read = conf.ANON_READ
		anonPerms.Write = conf.ANON_WRITE
		anonPerms.Delete = conf.ANON_DELETE

		var authList []string
		if conf.AUTH_GLOBUS_TOKEN_URL != "" && conf.AUTH_GLOBUS_PROFILE_URL != "" {
			authList = append(authList, "globus")
		}
		if len(conf.AUTH_OAUTH) > 0 {
			for b := range conf.AUTH_OAUTH {
				authList = append(authList, b)
			}
		}

		res := resource{
			A:      attrs,
			C:      conf.ADMIN_EMAIL,
			I:      "Shock",
			O:      authList,
			P:      *anonPerms,
			R:      []string{"node"},
			S:      time.Now().Format(longDateForm),
			T:      "Shock",
			U:      host + "/",
			Uptime: time.Since(StartTime).String(),
			V:      conf.VERSION,
		}
		responder.WriteResponseObject(w, r, http.StatusOK, res)
	})

	// Embedded web UI
	router.Handle("/ui", http.RedirectHandler("/ui/", http.StatusMovedPermanently))
	router.Handle("/ui/*", http.StripPrefix("/ui", ui.Handler()))

	// Node CRUD (replaces goweb.MapController)
	nc := &ncon.NodeController{}
	router.Route("/node", func(nr chi.Router) {
		nr.Get("/", nc.ReadMany)
		nr.Post("/", nc.Create)
		nr.Put("/", nc.UpdateMany)
		nr.Delete("/", nc.DeleteMany)
		nr.Options("/", nc.Options)
		nr.Get("/{id}", nc.Read)
		nr.Put("/{id}", nc.Replace)
		nr.Delete("/{id}", nc.Delete)
		nr.Options("/{id}", nc.Options)
	})

	// Catch-all handler for everything that we don't understand
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		responder.RespondWithError(w, r, http.StatusBadRequest, "Parameters do not match a valid Shock request type.")
	})

	return router
}

func main() {
	var err error

	// init config
	err = conf.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@conf.Initialize: %s\n", err.Error())
		os.Exit(1)
	}

	// init profile
	go func() {
		fmt.Println(http.ListenAndServe(conf.API_IP+":6060", nil))
	}()

	// init logging system
	logger.Initialize()
	logger.Info("Starting...")

	if conf.ANON_WRITE {
		warnstr := "Warning: anonymous write is activated, only use for development !!!!"
		logger.Info(warnstr)
	}
	if conf.ANON_DELETE {
		warnstr := "Warning: anonymous delete is activated, only use for development !!!!"
		logger.Info(warnstr)
	}

	// init database
	err = db.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@db.Initialize: %s\n", err.Error())
		logger.Error("Err@db.Initialize: " + err.Error())
		os.Exit(1)
	}

	err = user.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@user.Initialize: %s\n", err.Error())
		logger.Error("Err@user.Initialize: " + err.Error())
		os.Exit(1)
	}

	node.Initialize()
	preauth.Initialize()
	auth.Initialize()

	node.InitReaper()
	node.InitUploader()
	cache.InitCacheReaper()

	err = versions.Initialize()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@versions.Initialize: %s\n", err.Error())
		logger.Error("Err@versions.Initialize: " + err.Error())
		os.Exit(1)
	}
	err = versions.RunVersionUpdates()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@versions.RunVersionUpdates: %s\n", err.Error())
		logger.Error("Err@versions.RunVersionUpdates: " + err.Error())
		os.Exit(1)
	}
	// After version updates have succeeded without error, we can push the configured version numbers into the mongo db
	// Note: configured version numbers are configured in conf.go but are NOT user configurable by design
	err = versions.PushVersionsToDatabase()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err@versions.PushVersionsToDatabase: %s\n", err.Error())
		logger.Error("Err@versions.PushVersionsToDatabase: " + err.Error())
		os.Exit(1)
	}
	printLogo()
	conf.Print()

	if err := versions.Print(); err != nil {
		fmt.Fprintf(os.Stderr, "Err@versions.Print: %s\n", err.Error())
		logger.Error("Err@versions.Print: " + err.Error())
		os.Exit(1)
	}

	// check if necessary directories exist or created
	for _, path := range []string{conf.PATH_SITE, conf.PATH_DATA, conf.PATH_LOGS, conf.PATH_DATA + "/temp"} {
		err = os.MkdirAll(path, 0777)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
			logger.Errorf("error createing directory %s: %s", path, err.Error())
			os.Exit(1)
		}

	}

	// reload
	if conf.RELOAD != "" {
		fmt.Println("####### Reloading #######")
		err := reload(conf.RELOAD)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
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

	s := &http.Server{
		Addr:           ":" + Address,
		Handler:        newRouter(),
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
	if cache.CacheReaper != nil {
		go cache.CacheReaper.Handle()
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
