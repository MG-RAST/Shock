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
	"github.com/stretchr/goweb/context"
	"net/http"
	"os"
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
			responder.RespondWithError(ctx, http.StatusNotFound, "Node not found")
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
	case "GET":
		if idxType != "" {
			if v, has := n.Indexes[idxType]; has {
				responder.RespondWithData(ctx, map[string]interface{}{idxType: v})
			} else {
				responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Node %s does not have index of type %s.", n.Id, idxType))
			}
		} else {
			responder.RespondWithData(ctx, getRes{I: n.Indexes, A: filteredIndexes(n.Indexes)})
		}

	case "PUT":
		if !n.HasFile() {
			responder.RespondWithError(ctx, http.StatusBadRequest, "Node has no file")
			return
		} else if idxType == "" {
			responder.RespondWithError(ctx, http.StatusBadRequest, "Index create requires type")
			return
		}
		if _, ok := index.Indexers[idxType]; !ok {
			responder.RespondWithError(ctx, http.StatusBadRequest, fmt.Sprintf("Index type %s unavailable", idxType))
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
					responder.RespondWithError(ctx, http.StatusBadRequest, "Error while creating bam index")
					return
				}
				responder.RespondOK(ctx)
				return
			} else {
				responder.RespondWithError(ctx, http.StatusBadRequest, "Index type bai requires .bam file")
				return
			}
		}

		newIndexer := index.Indexers[idxType]
		f, _ := os.Open(n.FilePath())
		defer f.Close()
		idxer := newIndexer(f)
		count, err := idxer.Create(n.IndexPath() + "/" + idxType + ".idx")
		if err != nil {
			logger.Error("err " + err.Error())
			responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
			return
		}

		idxInfo := node.IdxInfo{
			Type:        idxType,
			TotalUnits:  count,
			AvgUnitSize: n.File.Size / count,
		}

		if idxType == "chunkrecord" {
			idxInfo.AvgUnitSize = conf.CHUNK_SIZE
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
		responder.RespondWithError(ctx, http.StatusNotImplemented, "This request type is not implemented")
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

func filteredIndexes(i node.Indexes) (indexes []string) {
	for _, name := range availIndexers() {
		if _, has := i[name]; !has {
			indexes = append(indexes, name)
		}
	}
	return indexes
}

func availIndexers() (indexers []string) {
	for name, _ := range index.Indexers {
		indexers = append(indexers, name)
	}
	return
}
