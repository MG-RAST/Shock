// Package preauth implements /preauth resource
package preauth

import (
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
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
		switch p.Type {
		case "download":
			streamDownload(ctx, id, p.Nodes, p.Options)
			preauth.Delete(id)
		default:
			responder.RespondWithError(ctx, http.StatusNotFound, "Preauthorization type not supported: "+p.Type)
		}
	}
	return
}

// handle download and its options
func streamDownload(ctx context.Context, pid string, nodes []string, options map[string]string) {
	// get defaults
	filename := pid
	var filterFunc filter.FilterFunc = nil
	var compressionFormat string = ""
	var archiveFormat string = ""

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
		compressionFormat = cp
	}
	if ar, has := options["archive"]; has {
		archiveFormat = ar
	}
	var files []*file.FileInfo

	// process nodes
	for _, nid := range nodes {
		// get node
		n, err := node.Load(nid)
		if (err != nil) || !n.HasFile() || n.File.LockDownload {
			continue
		}
		// get filereader
		nf, err := n.FileReader()
		if err != nil {
			nf.Close()
			continue
		}
		// add to file array
		var fileInfo file.FileInfo
		if n.Type == "subset" {
			if n.File.Size == 0 {
				// handle empty subset file
				fileInfo.R = append(fileInfo.R, nf)
			} else {
				idx := index.New()
				fullRange := "1-" + strconv.FormatInt(n.Subset.Index.TotalUnits, 10)
				recSlice, err := idx.Range(fullRange, n.Path()+"/"+n.Id+".subset.idx", n.Subset.Index.TotalUnits)
				if err != nil {
					nf.Close()
					continue
				}
				for _, rec := range recSlice {
					fileInfo.R = append(fileInfo.R, io.NewSectionReader(nf, rec[0], rec[1]))
				}
			}
		} else {
			fileInfo.R = append(fileInfo.R, nf)
		}
		defer nf.Close()
		// add to file info
		fileInfo.Name = n.File.Name
		fileInfo.Size = n.File.Size
		fileInfo.ModTime = n.File.CreatedOn
		if _, ok := n.File.Checksum["md5"]; ok {
			fileInfo.Checksum = n.File.Checksum["md5"]
		}
		files = append(files, &fileInfo)
	}

	if (len(nodes) == 1) && (len(files) == 1) {
		// create single node / file streamer
		s := &request.Streamer{
			R:           files[0].R,
			W:           ctx.HttpResponseWriter(),
			ContentType: "application/octet-stream",
			Filename:    filename,
			Size:        files[0].Size,
			Filter:      filterFunc,
			Compression: compressionFormat,
		}
		if err := s.Stream(false); err != nil {
			err_msg := "err:@preAuth: " + err.Error()
			logger.Error(err_msg)
			return
		}
	} else if (len(files) > 1) && (archiveFormat != "") {
		// create multi node / file streamer, must have archive format
		m := &request.MultiStreamer{
			Files:       files,
			W:           ctx.HttpResponseWriter(),
			ContentType: "application/octet-stream",
			Filename:    filename,
			Archive:     archiveFormat,
		}
		if err := m.MultiStream(); err != nil {
			err_msg := "err:@preAuth: " + err.Error()
			logger.Error(err_msg)
			return
		}
	} else {
		// something broke
		err_msg := "err:@preAuth: no files available to download for given combination of options"
		logger.Error(err_msg)
		responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
	}
	return
}
