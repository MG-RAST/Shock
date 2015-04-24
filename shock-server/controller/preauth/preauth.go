// Package preauth implements /preauth resource
package preauth

import (
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/filter"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
)

func PreAuthRequest(ctx context.Context) {
	id := ctx.PathValue("id")
	if p, err := preauth.Load(id); err != nil {
		err_msg := "err:@preAuth load: " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, 500, err_msg)
	} else {
		if n, err := node.Load(p.NodeId); err == nil {
			switch p.Type {
			case "download":
				streamDownload(ctx, n, p.Options)
				preauth.Delete(id)
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

// handle download and its options
func streamDownload(ctx context.Context, n *node.Node, options map[string]string) {
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
	// set defaults
	filename := n.Id
	var filterFunc filter.FilterFunc = nil
	var compressionFormat string = ""
	// use options if exist
	if fn, has := options["filename"]; has {
		filename = fn
	}
	if fl, has := options["filter"]; has {
		if filter.Has(fl) {
			filterFunc = filter.Filter(fl)
		}
	}
	if cp, has := options["compression"]; has {
		if archive.IsValidCompress(cp) {
			compressionFormat = cp
		}
	}
	// stream it
	s := &request.Streamer{R: []file.SectionReader{nf}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: filterFunc, Compression: compressionFormat}
	err = s.Stream(false)
	if err != nil {
		// causes "multiple response.WriteHeader calls" error but better than no response
		err_msg := "err:@preAuth: s.stream: " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, 500, err_msg)
	}
	return
}
