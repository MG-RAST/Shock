// Package preauth implements /preauth resource
package preauth

import (
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/jaredwilkening/goweb"
	"net/http"
)

func PreAuthRequest(cx *goweb.Context) {
	request.Log(cx.Request)
	id := cx.PathParams["id"]
	if p, err := preauth.Load(id); err != nil {
		if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
		} else {
			cx.RespondWithError(http.StatusInternalServerError)
			logger.Error("err:@preAuth load: " + err.Error())
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
			logger.Error("err:@preAuth loadnode: " + err.Error())
		}
	}
	return
}

func streamDownload(cx *goweb.Context, n *node.Node, filename string) {
	// Set connection to close when done
	cx.ResponseWriter.Header().Set("Connection", "close")
	
	query := util.Q(cx.Request.URL.Query())
	nf, err := n.FileReader()
	defer nf.Close()
	if err != nil {
		// File not found or some sort of file read error.
		// Probably deserves more checking
		logger.Error("err:@preAuth node.FileReader: " + err.Error())
		cx.RespondWithError(http.StatusBadRequest)
		return
	}
	if query.Has("filename") {
		filename = query.Value("filename")
	}
	s := &request.Streamer{R: []file.SectionReader{nf}, W: cx.ResponseWriter, ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: nil}
	err = s.Stream()
	if err != nil {
		// causes "multiple response.WriteHeader calls" error but better than no response
		cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
		logger.Error("err:@preAuth: s.stream: " + err.Error())
	}
}
