package main

import (
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/store"
	"github.com/jaredwilkening/goweb"
	"net/http"
)

func PreAuthRequest(cx *goweb.Context) {
	LogRequest(cx.Request)
	id := cx.PathParams["id"]
	if p, err := store.LoadPreAuth(id); err != nil {
		if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
		} else {
			cx.RespondWithError(http.StatusInternalServerError)
			log.Error("err:@preAuth load: " + err.Error())
		}
		return
	} else {
		if node, err := store.LoadNodeUnauth(p.NodeId); err == nil {
			switch p.Type {
			case "download":
				filename := node.Id
				if fn, has := p.Options["filename"]; has {
					filename = fn
				}
				streamDownload(cx, node, filename)
				store.DeletePreAuth(id)
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

func streamDownload(cx *goweb.Context, node *store.Node, filename string) {
	query := &Query{list: cx.Request.URL.Query()}
	nf, err := node.FileReader()
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
	s := &streamer{rs: []store.SectionReader{nf}, ws: cx.ResponseWriter, contentType: "application/octet-stream", filename: filename, size: node.File.Size, filter: nil}
	err = s.stream()
	if err != nil {
		// causes "multiple response.WriteHeader calls" error but better than no response
		cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
		log.Error("err:@preAuth: s.stream: " + err.Error())
	}
}
