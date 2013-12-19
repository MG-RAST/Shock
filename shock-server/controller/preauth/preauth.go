// Package preauth implements /preauth resource
package preauth

import (
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/stretchr/goweb/context"
)

func PreAuthRequest(ctx context.Context) {
	id := ctx.PathValue("id")
	if p, err := preauth.Load(id); err != nil {
		err_msg := "err:@preAuth load: " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, 500, err_msg)
		return
	} else {
		if n, err := node.LoadUnauth(p.NodeId); err == nil {
			switch p.Type {
			case "download":
				filename := n.Id
				if fn, has := p.Options["filename"]; has {
					filename = fn
				}
				streamDownload(ctx, n, filename)
				preauth.Delete(id)
				return
			default:
				responder.RespondWithError(ctx, 500, "Preauthorization type not supported: "+p.Type)
			}
		} else {
			err_msg := "err:@preAuth loadnode: " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, 500, err_msg)
		}
	}
	return
}

func streamDownload(ctx context.Context, n *node.Node, filename string) {
	nf, err := n.FileReader()
	defer nf.Close()
	if err != nil {
		// File not found or some sort of file read error.
		// Probably deserves more checking
		err_msg := "err:@preAuth node.FileReader: " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, 500, err_msg)
		return
	}
	s := &request.Streamer{R: []file.SectionReader{nf}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: nil}
	err = s.Stream()
	if err != nil {
		// causes "multiple response.WriteHeader calls" error but better than no response
		err_msg := "err:@preAuth: s.stream: " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, 500, err_msg)
	}
}
