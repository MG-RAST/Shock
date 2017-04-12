// Package preauth implements /preauth resource
package preauth

import (
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/filter"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"io"
	"net/http"
	"strconv"
)

func PreAuthRequest(ctx context.Context) {
	id := ctx.PathValue("id")
	if p, err := preauth.Load(id); err != nil {
		err_msg := "err:@preAuth load: " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
	} else {
		if n, err := node.Load(p.NodeId); err == nil {
			switch p.Type {
			case "download":
				streamDownload(ctx, n, p.Options)
				preauth.Delete(id)
			default:
				responder.RespondWithError(ctx, http.StatusNotFound, "Preauthorization type not supported: "+p.Type)
			}
		} else {
			err_msg := "err:@preAuth loadnode: " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}
	return
}

// handle download and its options
func streamDownload(ctx context.Context, n *node.Node, options map[string]string) {
	nf, err := n.FileReader()
	defer nf.Close()
	if err != nil {
		err_msg := "err:@preAuth node.FileReader: " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
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
	var s *request.Streamer
	if n.Type == "subset" {
		s = &request.Streamer{R: []file.SectionReader{}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: filterFunc, Compression: compressionFormat}
		if n.File.Size == 0 {
			// handle empty subset file
			s.R = append(s.R, nf)
		} else {
			idx := index.New()
			fullRange := "1-" + strconv.FormatInt(n.Subset.Index.TotalUnits, 10)
			recSlice, err := idx.Range(fullRange, n.Path()+"/"+n.Id+".subset.idx", n.Subset.Index.TotalUnits)
			if err != nil {
				responder.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
				return
			}
			for _, rec := range recSlice {
				s.R = append(s.R, io.NewSectionReader(nf, rec[0], rec[1]))
			}
		}
	} else {
		s = &request.Streamer{R: []file.SectionReader{nf}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: filterFunc, Compression: compressionFormat}
	}
	if err = s.Stream(false); err != nil {
		// causes "multiple response.WriteHeader calls" error but better than no response
		err_msg := "err:@preAuth: s.stream: " + err.Error()
		logger.Error(err_msg)
		responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	return
}
