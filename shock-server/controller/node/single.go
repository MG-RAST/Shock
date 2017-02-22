package node

import (
	"encoding/json"
	client "github.com/MG-RAST/Shock/shock-client/lib/httpclient"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/filter"
	"github.com/MG-RAST/Shock/shock-server/preauth"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	mgo "gopkg.in/mgo.v2"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
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
			return responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Read:LoadNode: " + id + ":" + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	rights := n.Acl.Check(u.Uuid)
	prights := n.Acl.Check("public")
	if rights["read"] == false && u.Admin == false && n.Acl.Owner != u.Uuid && prights["read"] == false {
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
	if _, ok := query["filename"]; ok {
		filename = query.Get("filename")
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

	// Switch though param flags
	// ?download=1 or ?download_raw=1
	_, download_raw := query["download_raw"]
	if _, ok := query["download"]; ok || download_raw {
		if !n.HasFile() {
			return responder.RespondWithError(ctx, http.StatusBadRequest, e.NodeNoFile)
		}

		_, seek_ok := query["seek"]
		if _, length_ok := query["length"]; seek_ok || length_ok {
			if n.Type == "subset" {
				return responder.RespondWithError(ctx, http.StatusBadRequest, "subset nodes do not currently support seek/length offset retrieval")
			}

			var seek int64
			var length int64
			if !seek_ok {
				seek = 0
				length_str := query.Get("length")
				length, err = strconv.ParseInt(length_str, 10, 0)
				if err != nil {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "length must be an integer value")
				}
				if length > n.File.Size {
					length = n.File.Size
				}
			} else if !length_ok {
				seek_str := query.Get("seek")
				seek, err = strconv.ParseInt(seek_str, 10, 0)
				if err != nil {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "seek must be an integer value")
				}
				length = n.File.Size - seek
			} else {
				seek_str := query.Get("seek")
				seek, err = strconv.ParseInt(seek_str, 10, 0)
				if err != nil {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "seek must be an integer value")
				}
				length_str := query.Get("length")
				length, err = strconv.ParseInt(length_str, 10, 0)
				if err != nil {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "length must be an integer value")
				}
				if length > n.File.Size-seek {
					length = n.File.Size - seek
				}
			}
			r, err := n.FileReader()
			defer r.Close()
			if err != nil {
				err_msg := "Err@node_Read:Open: " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}
			s := &request.Streamer{R: []file.SectionReader{}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: length, Filter: fFunc, Compression: compressionFormat}
			s.R = append(s.R, io.NewSectionReader(r, seek, length))
			if err = s.Stream(download_raw); err != nil {
				// causes "multiple response.WriteHeader calls" error but better than no response
				err_msg := "err:@node_Read s.Stream: " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			}
		} else if _, ok := query["index"]; ok {
			//handling bam file
			if query.Get("index") == "bai" {
				if n.Type == "subset" {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "subset nodes do not support bam indices")
				}

				s := &request.Streamer{R: []file.SectionReader{}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: fFunc, Compression: compressionFormat}

				var region string
				if _, ok := query["region"]; ok {
					//retrieve alingments overlapped with specified region
					region = query.Get("region")
				}
				argv, err := request.ParseSamtoolsArgs(ctx)
				if err != nil {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "Invaid args in query url")
				}
				err = s.StreamSamtools(n.FilePath(), region, argv...)
				if err != nil {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "error while invoking samtools")
				}
				return nil
			}

			// open file
			r, err := n.FileReader()
			defer r.Close()
			if err != nil {
				err_msg := "Err@node_Read:Open: " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}

			// load index obj and info
			idxName := query.Get("index")
			idxInfo, ok := n.Indexes[idxName]

			if !ok {
				if idxName == "size" {
					totalunits := n.File.Size / conf.CHUNK_SIZE
					m := n.File.Size % conf.CHUNK_SIZE
					if m != 0 {
						totalunits += 1
					}
					n.Indexes["size"] = node.IdxInfo{
						Type:        "size",
						TotalUnits:  totalunits,
						AvgUnitSize: conf.CHUNK_SIZE,
						Format:      "dynamic",
						CreatedOn:   time.Now(),
					}
					err = n.Save()
					if err != nil {
						err_msg := "Size index could not be auto-generated for node that did not have one."
						logger.Error(err_msg)
						return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
					}
				} else {
					return responder.RespondWithError(ctx, http.StatusBadRequest, e.InvalidIndex)
				}
			}

			idx, err := n.DynamicIndex(idxName)
			if err != nil {
				return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
			}

			if idx.Type() == "virtual" {
				if n.Type == "subset" {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "subset nodes do not currently support virtual indices")
				}

				csize := conf.CHUNK_SIZE
				if _, ok := query["chunk_size"]; ok {
					csize, err = strconv.ParseInt(query.Get("chunk_size"), 10, 64)
					if err != nil {
						return responder.RespondWithError(ctx, http.StatusBadRequest, "Invalid chunk_size")
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
					return responder.RespondWithError(ctx, http.StatusBadRequest, "Invalid request, record index must exist to retrieve chunkrecord index on a subset node.")
				}
				recordIdx, err := n.DynamicIndex(recordIdxName)
				if err != nil {
					return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				}

				if !hasPart {
					// download full subset file
					fullRange := "1-" + strconv.FormatInt(recordIdxInfo.TotalUnits, 10)
					recSlice, err := recordIdx.Range(fullRange, n.IndexPath()+"/"+recordIdxName+".idx", recordIdxInfo.TotalUnits)
					if err != nil {
						return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
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
							return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
						}
						// This gets us the parts of the chunkrecord index, but we still need to convert these to record indices.
						for _, chunkRec := range chunkRecSlice {
							start := (chunkRec[0] / 16) + 1
							stop := (start - 1) + (chunkRec[1] / 16)
							recSlice, err := recordIdx.Range(strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(stop, 10), n.IndexPath()+"/"+recordIdxName+".idx", recordIdxInfo.TotalUnits)
							if err != nil {
								return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
							}
							for _, rec := range recSlice {
								size += rec[1]
								s.R = append(s.R, io.NewSectionReader(r, rec[0], rec[1]))
							}
						}
					}
				} else {
					// bad request
					return responder.RespondWithError(ctx, http.StatusBadRequest, "Index parameter requires part parameter")
				}
			} else {
				if (!hasPart) && (idxInfo.Type == "subset") {
					// download full subset file
					fullRange := "1-" + strconv.FormatInt(idxInfo.TotalUnits, 10)
					recSlice, err := idx.Range(fullRange, n.IndexPath()+"/"+idxName+".idx", idxInfo.TotalUnits)
					if err != nil {
						return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
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
								return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
							}
							for _, rec := range recSlice {
								size += rec[1]
								s.R = append(s.R, io.NewSectionReader(r, rec[0], rec[1]))
							}
						} else {
							// empty node has no parts
							if n.File.Size == 0 {
								return responder.RespondWithError(ctx, http.StatusBadRequest, e.IndexOutBounds)
							}
							pos, length, err := idx.Part(p, n.IndexPath()+"/"+idxName+".idx", idxInfo.TotalUnits)
							if err != nil {
								return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
							}
							size += length
							s.R = append(s.R, io.NewSectionReader(r, pos, length))
						}
					}
				} else {
					// bad request
					return responder.RespondWithError(ctx, http.StatusBadRequest, "Index parameter requires part parameter")
				}
			}
			s.Size = size
			if err = s.Stream(download_raw); err != nil {
				// causes "multiple response.WriteHeader calls" error but better than no response
				err_msg := "err:@node_Read s.Stream: " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			}
			// download full file
		} else {
			if n.Type == "subset" {
				// open file
				r, err := n.FileReader()
				defer r.Close()
				if err != nil {
					err_msg := "Err@node_Read:Open: " + err.Error()
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
				}

				s := &request.Streamer{R: []file.SectionReader{}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: fFunc, Compression: compressionFormat}

				if n.File.Size == 0 {
					// handle empty subset file
					s.R = append(s.R, r)
				} else {
					idx := index.New()
					fullRange := "1-" + strconv.FormatInt(n.Subset.Index.TotalUnits, 10)
					recSlice, err := idx.Range(fullRange, n.Path()+"/"+n.Id+".subset.idx", n.Subset.Index.TotalUnits)
					if err != nil {
						return responder.RespondWithError(ctx, http.StatusInternalServerError, err.Error())
					}
					for _, rec := range recSlice {
						s.R = append(s.R, io.NewSectionReader(r, rec[0], rec[1]))
					}
				}
				if err = s.Stream(download_raw); err != nil {
					// causes "multiple response.WriteHeader calls" error but better than no response
					err_msg := "err:@node_Read s.Stream: " + err.Error()
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
			} else {
				nf, err := n.FileReader()
				defer nf.Close()
				if err != nil {
					// File not found or some sort of file read error.
					// Probably deserves more checking
					err_msg := "err:@node_Read node.FileReader: " + err.Error()
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
				s := &request.Streamer{R: []file.SectionReader{nf}, W: ctx.HttpResponseWriter(), ContentType: "application/octet-stream", Filename: filename, Size: n.File.Size, Filter: fFunc, Compression: compressionFormat}
				if err = s.Stream(download_raw); err != nil {
					// causes "multiple response.WriteHeader calls" error but better than no response
					err_msg := "err:@node_Read s.Stream: " + err.Error()
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				}
			}
		}
	} else if _, ok := query["download_url"]; ok {
		if n.Type == "subset" {
			return responder.RespondWithError(ctx, http.StatusBadRequest, "subset nodes do not currently support download_url operation")
		}

		if !n.HasFile() {
			return responder.RespondWithError(ctx, http.StatusBadRequest, e.NodeNoFile)
		} else {
			// add options
			options := map[string]string{}
			options["filename"] = filename
			if fFunc != nil {
				options["filter"] = query.Get("filter")
			}
			if compressionFormat != "" {
				options["compression"] = compressionFormat
			}
			// set preauth
			if p, err := preauth.New(util.RandString(20), "download", n.Id, options); err != nil {
				err_msg := "err:@node_Read download_url: " + err.Error()
				logger.Error(err_msg)
				return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			} else {
				return responder.RespondWithData(ctx, util.UrlResponse{Url: util.ApiUrl(ctx) + "/preauth/" + p.Id, ValidTill: p.ValidTill.Format(time.ANSIC)})
			}
		}
	} else if _, ok := query["download_post"]; ok {
		// This is a request to post the node to another Shock server. The 'post_url' parameter is required.
		// By default the post operation will include the data file and attributes (these options can be set
		// with post_data=0/1 and post_attr=0/1).
		if n.Type == "subset" {
			return responder.RespondWithError(ctx, http.StatusBadRequest, "subset nodes do not currently support download_post operation")
		}

		post_url := ""
		if _, ok := query["post_url"]; ok {
			post_url = query.Get("post_url")
		} else {
			return responder.RespondWithError(ctx, http.StatusBadRequest, "Request type requires post_url parameter of where to post new Shock node")
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
					return responder.RespondWithError(ctx, http.StatusBadRequest, "Parameter "+k+" must be either 0 or 1")
				}
			}
		}

		form := client.NewForm()
		form.AddParam("file_name", n.File.Name)

		if post_opts["post_data"] == 1 {
			form.AddFile("upload", n.FilePath())
		}

		if post_opts["post_attr"] == 1 && n.Attributes != nil {
			attr, _ := json.Marshal(n.Attributes)
			form.AddParam("attributes_str", string(attr[:]))
		}

		err = form.Create()
		if err != nil {
			err_msg := "could not create multipart form for posting to Shock server: " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}

		headers := client.Header{
			"Content-Type":   form.ContentType,
			"Content-Length": strconv.FormatInt(form.Length, 10),
		}

		if _, hasAuth := ctx.HttpRequest().Header["Authorization"]; hasAuth {
			headers["Authorization"] = ctx.HttpRequest().Header.Get("Authorization")
		}

		if res, err := client.Do("POST", post_url, headers, form.Reader); err == nil {
			if res.StatusCode == 200 {
				r := responseWrapper{}
				body, _ := ioutil.ReadAll(res.Body)
				if err = json.Unmarshal(body, &r); err != nil {
					err_msg := "err:@node_Read POST: " + err.Error()
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
				} else {
					return responder.WriteResponseObject(ctx, http.StatusOK, r)
				}
			} else {
				r := responseWrapper{}
				body, _ := ioutil.ReadAll(res.Body)
				if err = json.Unmarshal(body, &r); err == nil {
					err_msg := "err:@node_Read POST " + post_url + " = " + res.Status + ":" + (*r.Error)[0]
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
				} else {
					err_msg := "err:@node_Read POST " + post_url + " = " + res.Status
					logger.Error(err_msg)
					return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
				}
			}
		} else {
			return err
		}
	} else {
		// Base case respond with node in json
		return responder.RespondWithData(ctx, n)
	}

	return nil
}
