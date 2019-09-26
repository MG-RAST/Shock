package node

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/filter"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	sc "github.com/MG-RAST/go-shock-client"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	mgo "gopkg.in/mgo.v2"
)

type responseWrapper struct {
	Data   interface{} `json:"data"`
	Error  *[]string   `json:"error"`
	Status int         `json:"status"`
}

// GET: /node/{id}
func (cr *NodeController) Read(id string, ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// public user (no auth) can be used in some cases
	if u == nil {
		if conf.ANON_READ {
			u = &user.User{Uuid: "public"}
		} else {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
		}
	}

	// Load node by id
	n, err := node.Load(id)
	if err != nil {
		if err == mgo.ErrNotFound {
			logger.Error("err@node_Read: (node.Load) id=" + id + ": " + e.NodeNotFound)
			return responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "err@node_Read: (node.Load) id=" + id + ": " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	rights := n.Acl.Check(u.Uuid)
	prights := n.Acl.Check("public")
	if rights["read"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["read"] == false {
		logger.Error("err@node_Read: (Authenticate) id=" + id + ": " + e.UnAuth)
		return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
	}

	// Gather query params
	query := ctx.HttpRequest().URL.Query()
	// set defaults
	filename := n.Id
	if n.File.Name != "" {
		filename = n.File.Name
	}
	var fFunc filter.FilterFunc = nil
	var compressionFormat string = ""
	// use query params if exist
	if _, ok := query["file_name"]; ok {
		filename = query.Get("file_name")
	}
	if _, ok := query["filter"]; ok {
		if filter.Has(query.Get("filter")) {
			fFunc = filter.Filter(query.Get("filter"))
		}
	}
	if _, ok := query["compression"]; ok {
		if archive.IsValidCompress(query.Get("compression")) {
			compressionFormat = query.Get("compression")
		}
	}

	// read restore request
	if _, ok := query["restore"]; ok {

		var data bool
		data = n.GetRestore()
		// this shoudl return true/false
		return responder.RespondWithData(ctx, data)
	}

	if _, ok := query["download_idx"]; ok {

		// create a zip file from the idx_directory
		//var indexfiles []os.FileInfo

		indexfiles := n.IndexFiles()
		// list all index files
		//	logger.Infof("(Single-->Download_idx: n.IndexFiles \n) ")

		//spew.Dump(n)
		//logger.Infof("(Single-->Download_idx: indexfiles) ")

		//spew.Dump(indexfiles)

		var files []*file.FileInfo

		//		logger.Infof("(Single-->Download_idx) ")

		// process nodes
		for _, indexfile := range indexfiles { // loop thru index files

			//logger.Infof("(Single-->Download_idx: file %s) ", indexfile)

			// get node

			// get filereader
			nf, err := os.Open(indexfile)
			if err != nil {
				nf.Close()
				continue
			}
			// add to file array
			var fileInfo file.FileInfo

			fileInfo.R = append(fileInfo.R, nf)

			defer nf.Close()

			// add to file info
			fileInfo.Name = path.Base(indexfile)
			//fileInfo.Size = nf.
			//fileInfo.ModTime = n.File.CreatedOn
			// if _, ok := n.File.Checksum["md5"]; ok {
			// 	fileInfo.Checksum = n.File.Checksum["md5"]
			// }
			files = append(files, &fileInfo)
		}

		zipfilename := fmt.Sprintf("%s.idx.zip", n.Id)

		// if there are no index files, just return

		//	create a read for Zip file	and hand into streamer object
		//if (len(files) > 1) && (archiveFormat != "")
		// create multi node / file streamer, must have archive format
		m := &request.MultiStreamer{
			Files:       files,
			W:           ctx.HttpResponseWriter(),
			ContentType: "application/octet-stream",
			Filename:    zipfilename,
			Archive:     "zip", // supported: tar , zip
		}
		if err := m.MultiStream(); err != nil {
			err_msg := "err:@preAuth: " + err.Error()
			logger.Errorf("(single->download_idx) %s ", err_msg)
			return err
		}

		return nil
	}

	// Switch though param flags
	// ?download=1 or ?download_raw=1
	_, download_raw := query["download_raw"]
	if _, ok := query["download"]; ok || download_raw {
		if n.HasFileLock() {
			logger.Error("err@node_Read: id=" + id + ": " + e.NodeFileLock)
			return responder.RespondWithError(ctx, http.StatusLocked, e.NodeFileLock)
		}
		if !n.HasFile() {
			logger.Error("err@node_Read: id=" + id + ": " + e.NodeNoFile)
			return responder.RespondWithError(ctx, http.StatusBadRequest, e.NodeNoFile)
		}

		_, seek_ok := query["seek"]
		if _, length_ok := query["length"]; seek_ok || length_ok {
			if n.Type == "subset" {
				err_msg := "subset nodes do not currently support seek/length offset retrieval"
				logger.Error("err@node_Read: (subset) id=" + id + ": " + err_msg)
				return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			}

			var seek int64
			var length int64
			if !seek_ok {
				seek = 0
				length_str := query.Get("length")
				length, err = strconv.ParseInt(length_str, 10, 0)
				if err != nil {
					err_msg := "length must be an integer value"
					logger.Error("err@node_Read: (seek/length) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
				if length > n.File.Size {
					length = n.File.Size
				}
			} else if !length_ok {
				seek_str := query.Get("seek")
				seek, err = strconv.ParseInt(seek_str, 10, 0)
				if err != nil {
					err_msg := "seek must be an integer value"
					logger.Error("err@node_Read: (seek/length) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
				length = n.File.Size - seek
			} else {
				seek_str := query.Get("seek")
				seek, err = strconv.ParseInt(seek_str, 10, 0)
				if err != nil {
					err_msg := "seek must be an integer value"
					logger.Error("err@node_Read: (seek/length) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
				length_str := query.Get("length")
				length, err = strconv.ParseInt(length_str, 10, 0)
				if err != nil {
					err_msg := "length must be an integer value"
					logger.Error("err@node_Read: (seek/length) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
				if length > n.File.Size-seek {
					length = n.File.Size - seek
				}
			}
			r, err := n.FileReader()
			defer r.Close()
			if err != nil {
				err_msg := "err@node_Read: (node.FileReader) id=" + id + ": " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}
			s := &request.Streamer{R: []file.SectionReader{}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: length, Filter: fFunc, Compression: compressionFormat}
			s.R = append(s.R, io.NewSectionReader(r, seek, length))
			if err = s.Stream(download_raw); err != nil {
				err_msg := "err:@node_Read: (Stream) id=" + id + ": " + err.Error()
				logger.Error(err_msg)
				return errors.New(err_msg)
			}
		} else if _, ok := query["index"]; ok {
			idxName := query.Get("index")
			// check for lock
			if n.HasIndexLock(idxName) {
				logger.Error("err@node_Read: (index) id=" + id + ": " + e.NodeFileLock)
				return responder.RespondWithError(ctx, http.StatusBadRequest, e.NodeFileLock)
			}
			//handling bam file
			if idxName == "bai" {
				if n.Type == "subset" {
					err_msg := "subset nodes do not support bam indices"
					logger.Error("err@node_Read: (index/bai) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}

				s := &request.Streamer{R: []file.SectionReader{}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: fFunc, Compression: compressionFormat}

				var region string
				if _, ok := query["region"]; ok {
					//retrieve alingments overlapped with specified region
					region = query.Get("region")
				}
				queries := ctx.HttpRequest().URL.Query()
				argv, err := request.ParseSamtoolsArgs(queries)
				if err != nil {
					err_msg := "Invaid args in query url"
					logger.Error("err@node_Read: (index/bai) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
				err = s.StreamSamtools(n.FilePath(), region, argv...)
				if err != nil {
					err_msg := "err:@node_Read: (index/bai) id=" + id + ": " + err.Error()
					logger.Error(err_msg)
					return errors.New(err_msg)
				}
				return nil
			}

			// open file
			r, err := n.FileReader()
			defer r.Close()
			if err != nil {
				err_msg := "err@node_Read: (node.FileReader) id=" + id + ": " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}

			// load index obj and info
			idxInfo, ok := n.Indexes[idxName]

			if !ok {
				if idxName == "size" {
					// auto-generate size index if missing
					totalunits := n.File.Size / conf.CHUNK_SIZE
					m := n.File.Size % conf.CHUNK_SIZE
					if m != 0 {
						totalunits += 1
					}
					n.Indexes["size"] = &node.IdxInfo{
						Type:        "size",
						TotalUnits:  totalunits,
						AvgUnitSize: conf.CHUNK_SIZE,
						Format:      "dynamic",
						CreatedOn:   time.Now(),
					}
					// lock during save
					err = locker.NodeLockMgr.LockNode(n.Id)
					if err != nil {
						err_msg := "err@node_Read: (LockNode) id=" + n.Id + ": " + err.Error()
						logger.Error(err_msg)
						return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
					}
					err = n.Save()
					locker.NodeLockMgr.UnlockNode(n.Id)
					if err != nil {
						err_msg := "Size index could not be auto-generated for node that did not have one."
						logger.Error("err@node_Read: (index/size) id=" + id + ": " + err_msg)
						return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
					}
				} else {
					logger.Error("err@node_Read: (index) id=" + id + ": " + e.InvalidIndex)
					return responder.RespondWithError(ctx, http.StatusBadRequest, e.InvalidIndex)
				}
			}

			idx, err := n.DynamicIndex(idxName)
			if err != nil {
				err_msg := "err@node_Read: (DynamicIndex) id=" + id + ": " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			}

			if idx.Type() == "virtual" {
				if n.Type == "subset" {
					err_msg := "subset nodes do not currently support virtual indices"
					logger.Error("err@node_Read: (index/virtual) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}

				csize := conf.CHUNK_SIZE
				if _, ok := query["chunk_size"]; ok {
					csize, err = strconv.ParseInt(query.Get("chunk_size"), 10, 64)
					if err != nil {
						err_msg := "Invalid chunk_size"
						logger.Error("err@node_Read: (index/virtual) id=" + id + ": " + err_msg)
						return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
					}
				}
				idx.Set(map[string]interface{}{"ChunkSize": csize})
			}

			var size int64 = 0
			s := &request.Streamer{R: []file.SectionReader{}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Filter: fFunc, Compression: compressionFormat}

			_, hasPart := query["part"]
			if n.Type == "subset" && idxName == "chunkrecord" {
				recordIdxName := "record"
				recordIdxInfo, ok := n.Indexes[recordIdxName]
				if !ok {
					err_msg := "Invalid request, record index must exist to retrieve chunkrecord index on a subset node."
					logger.Error("err@node_Read: (index/chunkrecord) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
				recordIdx, err := n.DynamicIndex(recordIdxName)
				if err != nil {
					err_msg := "err@node_Read: (DynamicIndex) id=" + id + ": " + err.Error()
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}

				if !hasPart {
					// download full subset file
					fullRange := "1-" + strconv.FormatInt(recordIdxInfo.TotalUnits, 10)
					recSlice, err := recordIdx.Range(fullRange, n.IndexPath()+"/"+recordIdxName+".idx", recordIdxInfo.TotalUnits)
					if err != nil {
						err_msg := "err@node_Read: (recordIdx.Range) id=" + id + ": " + err.Error()
						logger.Error(err_msg)
						return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
					}
					for _, rec := range recSlice {
						size += rec[1]
						s.R = append(s.R, io.NewSectionReader(r, rec[0], rec[1]))
					}
				} else if hasPart {
					// download parts
					for _, p := range query["part"] {
						chunkRecSlice, err := idx.Range(p, n.IndexPath()+"/"+idxName+".idx", idxInfo.TotalUnits)
						if err != nil {
							err_msg := "err@node_Read: (idx.Range) id=" + id + ": " + err.Error()
							logger.Error(err_msg)
							return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
						}
						// This gets us the parts of the chunkrecord index, but we still need to convert these to record indices.
						for _, chunkRec := range chunkRecSlice {
							start := (chunkRec[0] / 16) + 1
							stop := (start - 1) + (chunkRec[1] / 16)
							recSlice, err := recordIdx.Range(strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(stop, 10), n.IndexPath()+"/"+recordIdxName+".idx", recordIdxInfo.TotalUnits)
							if err != nil {
								err_msg := "err@node_Read: (recordIdx.Range) id=" + id + ": " + err.Error()
								logger.Error(err_msg)
								return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
							}
							for _, rec := range recSlice {
								size += rec[1]
								s.R = append(s.R, io.NewSectionReader(r, rec[0], rec[1]))
							}
						}
					}
				} else {
					// bad request
					err_msg := "Index parameter requires part parameter"
					logger.Error("err@node_Read: (index) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
			} else {
				if (!hasPart) && (idxInfo.Type == "subset") {
					// download full subset file
					fullRange := "1-" + strconv.FormatInt(idxInfo.TotalUnits, 10)
					recSlice, err := idx.Range(fullRange, n.IndexPath()+"/"+idxName+".idx", idxInfo.TotalUnits)
					if err != nil {
						err_msg := "err@node_Read: (idx.Range) id=" + id + ": " + err.Error()
						logger.Error(err_msg)
						return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
					}
					for _, rec := range recSlice {
						size += rec[1]
						s.R = append(s.R, io.NewSectionReader(r, rec[0], rec[1]))
					}
				} else if hasPart {
					// download parts
					for _, p := range query["part"] {
						// special case for subset ranges
						if idxInfo.Type == "subset" {
							recSlice, err := idx.Range(p, n.IndexPath()+"/"+idxName+".idx", idxInfo.TotalUnits)
							if err != nil {
								err_msg := "err@node_Read: (idx.Range) id=" + id + ": " + err.Error()
								logger.Error(err_msg)
								return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
							}
							for _, rec := range recSlice {
								size += rec[1]
								s.R = append(s.R, io.NewSectionReader(r, rec[0], rec[1]))
							}
						} else {
							// empty node has no parts
							if n.File.Size == 0 {
								logger.Error("err@node_Read: (File.Size) id=" + id + ": " + e.IndexOutBounds)
								return responder.RespondWithError(ctx, http.StatusBadRequest, e.IndexOutBounds)
							}
							pos, length, err := idx.Part(p, n.IndexPath()+"/"+idxName+".idx", idxInfo.TotalUnits)
							if err != nil {
								err_msg := "err@node_Read: (idx.Part) id=" + id + ": " + err.Error()
								logger.Error(err_msg)
								return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
							}
							size += length
							s.R = append(s.R, io.NewSectionReader(r, pos, length))
						}
					}
				} else {
					// bad request
					err_msg := "Index parameter requires part parameter"
					logger.Error("err@node_Read: (index) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
			}
			s.Size = size
			if err = s.Stream(download_raw); err != nil {
				err_msg := "err:@node_Read: " + err.Error()
				logger.Error(err_msg)
				return errors.New(err_msg)
			}
			// download full file
		} else {
			nf, err := n.FileReader()
			defer nf.Close()
			if err != nil {
				err_msg := "err:@node_Read: (node.FileReader) " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}
			var s *request.Streamer
			if n.Type == "subset" {
				s = &request.Streamer{R: []file.SectionReader{}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: fFunc, Compression: compressionFormat}
				if n.File.Size == 0 {
					// handle empty subset file
					s.R = append(s.R, nf)
				} else {
					idx := index.New()
					fullRange := "1-" + strconv.FormatInt(n.Subset.Index.TotalUnits, 10)
					recSlice, err := idx.Range(fullRange, n.Path()+"/"+n.Id+".subset.idx", n.Subset.Index.TotalUnits)
					if err != nil {
						err_msg := "err@node_Read: (idx.Range) id=" + id + ": " + err.Error()
						logger.Error(err_msg)
						return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
					}
					for _, rec := range recSlice {
						s.R = append(s.R, io.NewSectionReader(nf, rec[0], rec[1]))
					}
				}
			} else {
				s = &request.Streamer{R: []file.SectionReader{nf}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: fFunc, Compression: compressionFormat}
			}
			if err = s.Stream(download_raw); err != nil {
				err_msg := "err:@node_Read: (Stream) " + err.Error()
				logger.Error(err_msg)
				return errors.New(err_msg)
			}
		}
	} else if _, ok := query["download_url"]; ok {
		if n.HasFileLock() {
			logger.Error("err@node_Read: id=" + id + ": " + e.NodeFileLock)
			return responder.RespondWithError(ctx, http.StatusLocked, e.NodeFileLock)
		} else if !n.HasFile() {
			logger.Error("err:@node_Read: (download_url) " + e.NodeNoFile)
			return responder.RespondWithError(ctx, http.StatusBadRequest, e.NodeNoFile)
		} else {
			preauthFilename := filename
			// add options
			options := map[string]string{}
			options["filename"] = filename
			if fFunc != nil {
				options["filter"] = query.Get("filter")
			}
			if compressionFormat != "" {
				options["compression"] = compressionFormat
				preauthFilename = preauthFilename + "." + compressionFormat
			}
			// set preauth
			if p, err := preauth.New(util.RandString(20), "download", []string{n.Id}, options); err != nil {
				err_msg := "err:@node_Read: (download_url) " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			} else {
				data := preauth.PreAuthResponse{
					Url:       util.ApiUrl(ctx) + "/preauth/" + p.Id,
					ValidTill: p.ValidTill.Format(time.ANSIC),
					Format:    options["compression"],
					Filename:  preauthFilename,
					Files:     1,
					Size:      n.File.Size,
				}
				return responder.RespondWithData(ctx, data)
			}
		}
	} else if _, ok := query["download_post"]; ok {
		// This is a request to post the node to another Shock server. The 'post_url' parameter is required.
		// By default the post operation will include the data file and attributes (these options can be set
		// with post_data=0/1 and post_attr=0/1).
		if n.Type == "subset" {
			err_msg := "subset nodes do not currently support download_post operation"
			logger.Error("err@node_Read: (download_post) id=" + id + ": " + err_msg)
			return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}

		post_url := ""
		if _, ok := query["post_url"]; ok {
			post_url = query.Get("post_url")
		} else {
			err_msg := "Request type requires post_url parameter of where to post new Shock node"
			logger.Error("err@node_Read: (download_post) id=" + id + ": " + err_msg)
			return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}

		post_opts := map[string]int{
			"post_data": 1,
			"post_attr": 1,
		}

		for k, _ := range post_opts {
			if _, ok := query[k]; ok {
				if query.Get(k) == "0" {
					post_opts[k] = 0
				} else if query.Get(k) == "1" {
					post_opts[k] = 1
				} else {
					err_msg := "Parameter " + k + " must be either 0 or 1"
					logger.Error("err@node_Read: (download_post) id=" + id + ": " + err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
			}
		}

		var authToken string
		if _, hasAuth := ctx.HttpRequest().Header["Authorization"]; hasAuth {
			authToken = ctx.HttpRequest().Header.Get("Authorization")
		}
		client := sc.NewShockClient(post_url, authToken, false)

		var uploadPath string
		if (post_opts["post_data"] == 1) && n.HasFile() && !n.HasFileLock() {
			uploadPath = n.FilePath()
		}

		var nodeAttr map[string]interface{}
		if (post_opts["post_attr"]) == 1 && (n.Attributes != nil) {
			var ok bool
			nodeAttr, ok = n.Attributes.(map[string]interface{})
			if !ok {
				err_msg := "Node corrupted, got type " + reflect.TypeOf(n.Attributes).String() + " for attributes"
				logger.Error("err@node_Read: (download_post) id=" + id + ": " + err_msg)
				return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			}
		}

		node, err := client.PostFileWithAttributes(uploadPath, n.File.Name, nodeAttr)
		if err != nil {
			err_msg := "err:@node_Read POST: " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
		return responder.RespondWithData(ctx, node)

	} else {
		// Base case respond with node in json
		return responder.RespondWithData(ctx, n)
	}

	return nil
}
