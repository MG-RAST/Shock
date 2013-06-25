package main

import (
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/jaredwilkening/goweb"
	"net/http"
)

func PreAuthRequest(cx *goweb.Context) {
	LogRequest(cx.Request)
	id := cx.PathParams["id"]
	if p, err := preauth.Load(id); err != nil {
		if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
		} else {
			cx.RespondWithError(http.StatusInternalServerError)
			log.Error("err:@preAuth load: " + err.Error())
		}
		return
	} else {
		if n, err := node.LoadUnauth(p.NodeId); err == nil {
			switch p.Type {
			case "download":
				filename := n.Id
				if fn, has := p.Options["filename"]; has {
					filename = fn
				}
				streamDownload(cx, n, filename)
				preauth.Delete(id)
				return
			default:
				cx.RespondWithError(http.StatusInternalServerError)
			}
		} else {
			cx.RespondWithError(http.StatusInternalServerError)
			log.Error("err:@preAuth loadnode: " + err.Error())
		}
	}
	return
}

func streamDownload(cx *goweb.Context, n *node.Node, filename string) {
	query := &Query{list: cx.Request.URL.Query()}
	nf, err := n.FileReader()
	if err != nil {
		// File not found or some sort of file read error.
		// Probably deserves more checking
		log.Error("err:@preAuth node.FileReader: " + err.Error())
		cx.RespondWithError(http.StatusBadRequest)
		return
	}
	if query.Has("filename") {
		filename = query.Value("filename")
	}
	s := &streamer{rs: []file.SectionReader{nf}, ws: cx.ResponseWriter, contentType: "application/octet-stream", filename: filename, size: n.File.Size, filter: nil}
	err = s.stream()
	if err != nil {
		// causes "multiple response.WriteHeader calls" error but better than no response
		cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
		log.Error("err:@preAuth: s.stream: " + err.Error())
	}
}
