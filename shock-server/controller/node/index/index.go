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

	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, ctx)
		return
	}

	// Fake public user
	if u == nil {
		u = &user.User{Uuid: ""}
	}

	// Load node and handle user unauthorized
	n, err := node.Load(nid, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			return
		} else if err.Error() == e.MongoDocNotFound {
			responder.RespondWithError(ctx, http.StatusNotFound, "Node not found.")
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@index:LoadNode: " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			return
		}
	}

	switch ctx.HttpRequest().Method {
	case "DELETE":
		if _, has := n.Indexes[idxType]; has {
			if err := n.DeleteIndex(idxType); err != nil {
				err_msg := err.Error()
				logger.Error(err_msg)
				responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
			}
			responder.RespondOK(ctx)
		} else {
			responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Node %s does not have index of type %s.", n.Id, idxType))
		}

	case "GET":
		if v, has := n.Indexes[idxType]; has {
			responder.RespondWithData(ctx, map[string]interface{}{idxType: v})
		} else {
			responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Node %s does not have index of type %s.", n.Id, idxType))
		}

	case "PUT":
		if _, has := n.Indexes[idxType]; has {
			responder.RespondOK(ctx)
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
			if ext := n.FileExt(); ext == ".bam" {
				if err := index.CreateBamIndex(n.FilePath()); err != nil {
					responder.RespondWithError(ctx, http.StatusBadRequest, "Error while creating bam index.")
					return
				}
				responder.RespondOK(ctx)
				return
			} else {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Index type bai requires .bam file.")
				return
			}
		}

		count := int64(0)
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
			count, _, err = index.CreateSubsetIndex(&idxer, n.IndexPath()+"/"+newIndex+".idx", n.IndexPath()+"/"+parentIndex+".idx")
			if err != nil {
				logger.Error("err " + err.Error())
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}

		} else if idxType == "column" {
			// Gather query params
			query := ctx.HttpRequest().URL.Query()

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
			count, err = index.CreateColumnIndex(&idxer, num, n.IndexPath()+"/"+idxType+".idx")
			if err != nil {
				logger.Error("err " + err.Error())
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}
		} else {
			newIndexer := index.Indexers[idxType]
			f, _ := os.Open(n.FilePath())
			defer f.Close()
			idxer := newIndexer(f)
			count, err = idxer.Create(n.IndexPath() + "/" + idxType + ".idx")
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
		}

		//if idxType == "chunkrecord" {
		//	idxInfo.AvgUnitSize = conf.CHUNK_SIZE
		//}

		if idxType == "subset" {
			idxInfo.AvgUnitSize = -1
			idxType = subsetName
		}

		if err := n.SetIndexInfo(idxType, idxInfo); err != nil {
			logger.Error("err@node.SetIndexInfo: " + err.Error())
		}

		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("END indexing: " + nid)
		}

		responder.RespondOK(ctx)
		return

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
