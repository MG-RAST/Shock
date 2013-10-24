package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/jaredwilkening/goweb"
	"net"
	"net/http"
	"strings"
)

const logo = `
 +-------------+  +----+    +----+  +--------------+  +--------------+  +----+      +----+
 |             |  |    |    |    |  |              |  |              |  |    |      |    |
 |    +--------+  |    |    |    |  |    +----+    |  |    +---------+  |    |      |    |
 |    |           |    +----+    |  |    |    |    |  |    |            |    |     |    |
 |    +--------+  |              |  |    |    |    |  |    |            |    |    |    |
 |             |  |    +----+    |  |    |    |    |  |    |            |    |   |    |
 +--------+    |  |    |    |    |  |    |    |    |  |    |            |    +---+    +-+
          |    |  |    |    |    |  |    |    |    |  |    |            |               |
 +--------+    |  |    |    |    |  |    +----+    |  |    +---------+  |    +-----+    |
 |             |  |    |    |    |  |              |  |              |  |    |     |    |
 +-------------+  +----+    +----+  +--------------+  +--------------+  +----+     +----+`

func printLogo() {
	fmt.Println(logo)
	return
}

type urlResponse struct {
	Url       string `json:"url"`
	ValidTill string `json:"validtill"`
}

type resource struct {
	R []string `json:"resources"`
	U string   `json:"url"`
	D string   `json:"documentation"`
	C string   `json:"contact"`
	I string   `json:"id"`
	T string   `json:"type"`
}

func RespondOk(cx *goweb.Context) {
	LogRequest(cx.Request)
	cx.RespondWithOK()
	return
}

func ResourceDescription(cx *goweb.Context) {
	LogRequest(cx.Request)
	r := resource{
		R: []string{"node", "user"},
		U: apiUrl(cx) + "/",
		D: siteUrl(cx) + "/",
		C: conf.Conf["admin-email"],
		I: "Shock",
		T: "Shock",
	}
	cx.WriteResponse(r, 200)
}

func apiUrl(cx *goweb.Context) string {
	if conf.Conf["api-url"] != "" {
		return conf.Conf["api-url"]
	}
	return "http://" + cx.Request.Host
}

func siteUrl(cx *goweb.Context) string {
	if conf.Conf["site-url"] != "" {
		return conf.Conf["site-url"]
	} else if strings.Contains(cx.Request.Host, ":") {
		return fmt.Sprintf("http://%s:%s", strings.Split(cx.Request.Host, ":")[0], conf.Conf["site-port"])
	}
	return "http://" + cx.Request.Host
}

func Site(cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Connection", "close")
	LogRequest(cx.Request)
	http.ServeFile(cx.ResponseWriter, cx.Request, conf.Conf["site-path"]+"/pages/main.html")
}

func RawDir(cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Connection", "close")
	LogRequest(cx.Request)
	http.ServeFile(cx.ResponseWriter, cx.Request, fmt.Sprintf("%s%s", conf.Conf["data-path"], cx.Request.URL.Path))
}

func AssetsDir(cx *goweb.Context) {
	cx.ResponseWriter.Header().Set("Connection", "close")
	LogRequest(cx.Request)
	http.ServeFile(cx.ResponseWriter, cx.Request, conf.Conf["site-path"]+cx.Request.URL.Path)
}

func LogRequest(req *http.Request) {
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	// failed attempt to get the host in ipv4
	//addrs, _ := net.LookupIP(host)
	//fmt.Println(addrs)
	suffix := ""
	if _, auth := req.Header["Authorization"]; auth {
		suffix += " AUTH"
	}

	if l, has := req.Header["Content-Length"]; has {
		suffix += " Content-Length: " + l[0]
	}
	url := ""
	if req.URL.RawQuery != "" {
		url = fmt.Sprintf("%s %s?%s", req.Method, req.URL.Path, req.URL.RawQuery)
	} else {
		url = fmt.Sprintf("%s %s", req.Method, req.URL.Path)
	}
	logger.Info("access", host+" \""+url+suffix+"\"")
}
