package index

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/mgo"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"net/http"
	"os"
	"strconv"
)

type getRes struct {
	I interface{} `json:"indexes"`
	A interface{} `json:"available_indexers"`
}

type m map[string]string

// GET, PUT, DELETE: /node/{nid}/index/{idxType}
func IndexTypedRequest(ctx context.Context) {
	nid := ctx.PathValue("nid")
	idxType := ctx.PathValue("idxType")
	rmeth := ctx.HttpRequest().Method

	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, ctx)
		return
	}

	// public user (no auth) can be used in some cases
	if u == nil {
		if (rmeth == "GET" && conf.ANON_READ) || (rmeth == "PUT" && conf.ANON_WRITE) || (rmeth == "DELETE" && conf.ANON_WRITE) {
			u = &user.User{Uuid: "public"}
		} else {
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.NoAuth)
			return
		}
	}

	// Load node by id
	n, err := node.Load(nid)
	if err != nil {
		if err == mgo.ErrNotFound {
			responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Index:LoadNode: " + nid + ":" + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
		return
	}

	rights := n.Acl.Check(u.Uuid)

	switch rmeth {
	case "DELETE":
		if rights["write"] == false && u.Admin == false && n.Acl.Owner != u.Uuid {
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			return
		}

		if _, has := n.Indexes[idxType]; has {
			if err := n.DeleteIndex(idxType); err != nil {
				err_msg := err.Error()
				logger.Error(err_msg)
				responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
				return
			}
			responder.RespondOK(ctx)
		} else {
			responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Node %s does not have index of type %s.", n.Id, idxType))
		}

	case "GET":
		if rights["read"] == false && u.Admin == false && n.Acl.Owner != u.Uuid {
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			return
		}

		if v, has := n.Indexes[idxType]; has {
			responder.RespondWithData(ctx, map[string]interface{}{idxType: v})
		} else {
			responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Node %s does not have index of type %s.", n.Id, idxType))
		}

	case "PUT":
		if rights["write"] == false && u.Admin == false && n.Acl.Owner != u.Uuid {
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			return
		}

		// Gather query params
		query := ctx.HttpRequest().URL.Query()
		_, forceRebuild := query["force_rebuild"]

		if _, has := n.Indexes[idxType]; has {
			if idxType == "size" {
				responder.RespondOK(ctx)
			} else if !forceRebuild {
				responder.RespondWithError(ctx, http.StatusBadRequest, "This index already exists, please add the parameter 'force_rebuild=1' to force a rebuild of the existing index.")
			}
			return
		}

		if !n.HasFile() {
			responder.RespondWithError(ctx, http.StatusBadRequest, "Node has no file.")
			return
		} else if idxType == "" {
			responder.RespondWithError(ctx, http.StatusBadRequest, "Index create requires type.")
			return
		}
		if _, ok := index.Indexers[idxType]; !ok && idxType != "bai" && idxType != "subset" && idxType != "column" {
			responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Index type %s unavailable.", idxType))
			return
		}
		if idxType == "size" {
			responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Index type size is a virtual index and does not require index building."))
			return
		}

		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("START indexing: " + nid)
		}

		if idxType == "bai" {
			//bam index is created by the command-line tool samtools
			if n.Type == "subset" {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Shock does not support bam index creation on subset nodes.")
				return
			}

			if ext := n.FileExt(); ext == ".bam" {
				if err := index.CreateBamIndex(n.FilePath()); err != nil {
					responder.RespondWithError(ctx, http.StatusInternalServerError, "Error while creating bam index.")
					return
				}
				responder.RespondOK(ctx)
			} else {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Index type bai requires .bam file.")
			}
			return
		}

		subsetSize := int64(0)
		count := int64(0)
		indexFormat := ""
		subsetName := ""
		if idxType == "subset" {
			// Utilizing the multipart form parser since we need to upload a file.
			params, files, err := request.ParseMultipartForm(ctx.HttpRequest())
			if err != nil {
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}

			parentIndex, hasParent := params["parent_index"]
			if !hasParent {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Index type subset requires parent_index param.")
				return
			} else if _, has := n.Indexes[parentIndex]; !has {
				responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Node %s does not have index of type %s.", n.Id, parentIndex))
				return
			}

			newIndex, hasName := params["index_name"]
			if !hasName {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Index type subset requires index_name param.")
				return
			} else if _, reservedName := index.Indexers[newIndex]; reservedName || newIndex == "bai" {
				responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("%s is a reserved index name and cannot be used to create a custom subset index.", newIndex))
				return
			}
			subsetName = newIndex

			subsetIndices, hasFile := files["subset_indices"]
			if !hasFile {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Index type subset requires subset_indices file.")
				return
			}

			f, _ := os.Open(subsetIndices.Path)
			defer f.Close()
			idxer := index.NewSubsetIndexer(f)

			// we default to "array" index format for backwards compatibility
			indexFormat = "array"
			if n.Indexes[parentIndex].Format == "array" || n.Indexes[parentIndex].Format == "matrix" {
				indexFormat = n.Indexes[parentIndex].Format
			}
			count, subsetSize, err = index.CreateSubsetIndex(&idxer, n.IndexPath()+"/"+newIndex+".idx", n.IndexPath()+"/"+parentIndex+".idx", indexFormat, n.Indexes[parentIndex].TotalUnits)
			if err != nil {
				logger.Error("err " + err.Error())
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}

		} else if idxType == "column" {
			// Gather query params
			query := ctx.HttpRequest().URL.Query()

			if n.Type == "subset" {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Shock does not support column index creation on subset nodes.")
				return
			}

			if _, exists := query["number"]; !exists {
				err_msg := "Index type column requires a number parameter in the url."
				logger.Error(err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}

			num_str := query.Get("number")
			idxType = idxType + num_str
			num, err := strconv.Atoi(num_str)
			if err != nil || num < 1 {
				err_msg := "Index type column requires a number parameter in the url of an integer greater than zero."
				logger.Error(err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}

			f, _ := os.Open(n.FilePath())
			defer f.Close()
			idxer := index.NewColumnIndexer(f)
			count, indexFormat, err = index.CreateColumnIndex(&idxer, num, n.IndexPath()+"/"+idxType+".idx")
			if err != nil {
				logger.Error("err " + err.Error())
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			if n.Type == "subset" && (idxType != "chunkrecord" || n.Subset.Parent.IndexName != "record") {
				responder.RespondWithError(ctx, http.StatusBadRequest, "For subset nodes, Shock currently only supports subset and chunkrecord indexes. Also, for a chunkrecord index, the subset node must have been generated from a record index.")
				return
			}

			newIndexer := index.Indexers[idxType]
			f, _ := os.Open(n.FilePath())
			defer f.Close()
			var idxer index.Indexer
			if n.Type == "subset" {
				idxer = newIndexer(f, n.Type, n.Subset.Index.Format, n.IndexPath()+"/"+n.Subset.Parent.IndexName+".idx")
			} else {
				idxer = newIndexer(f, n.Type, "", "")
			}
			count, indexFormat, err = idxer.Create(n.IndexPath() + "/" + idxType + ".idx")
			if err != nil {
				logger.Error("err " + err.Error())
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}
		}

		if count == 0 {
			responder.RespondWithError(ctx, http.StatusBadRequest, "Index empty.")
			return
		}

		idxInfo := node.IdxInfo{
			Type:        idxType,
			TotalUnits:  count,
			AvgUnitSize: n.File.Size / count,
			Format:      indexFormat,
		}

		//if idxType == "chunkrecord" {
		//	idxInfo.AvgUnitSize = conf.CHUNK_SIZE
		//}

		if idxType == "subset" {
			idxType = subsetName
			idxInfo.AvgUnitSize = subsetSize / count
		}

		// reload node by id before updating mongo document (attempting to avoid race conditions)
		n, err := node.Load(nid)
		if err != nil {
			if err == mgo.ErrNotFound {
				responder.RespondWithError(ctx, http.StatusNotFound, "Node deleted during index creation.")
			} else {
				// In theory the db connection could be lost between
				// checking user and load but seems unlikely.
				err_msg := "Err@node_Index:LoadNode: " + nid + ":" + err.Error()
				logger.Error(err_msg)
				responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}
			return
		}

		if err := n.SetIndexInfo(idxType, idxInfo); err != nil {
			logger.Error("err@node.SetIndexInfo: " + err.Error())
		}

		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("END indexing: " + nid)
		}

		responder.RespondOK(ctx)

	default:
		responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented.")
	}
	return
}

func contains(list []string, s string) bool {
	for _, i := range list {
		if i == s {
			return true
		}
	}
	return false
}

func availIndexers() (indexers []string) {
	for name, _ := range index.Indexers {
		indexers = append(indexers, name)
	}
	return
}
