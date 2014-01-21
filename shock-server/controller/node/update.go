package node

import (
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

// PUT: /node/{id} -> multipart-form
func (cr *NodeController) Replace(id string, ctx context.Context) error {
	u, err := request.Authenticate(ctx.HttpRequest())
	if err != nil && err.Error() != e.NoAuth {
		return request.AuthError(err, ctx)
	}

	// Gather query params
	query := ctx.HttpRequest().URL.Query()

	// Fake public user
	if u == nil {
		u = &user.User{Uuid: ""}
	}

	n, err := node.Load(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			return responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
		} else if err.Error() == e.MongoDocNotFound {
			return responder.RespondWithError(ctx, http.StatusNotFound, "Node not found")
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@node_Update:LoadNode: " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
	}

	if _, ok := query["index"]; ok {
		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("START indexing: " + id)
		}

		if !n.HasFile() {
			return responder.RespondWithError(ctx, http.StatusBadRequest, "node file empty")
		}

		if query.Get("index") == "bai" {
			//bam index is created by the command-line tool samtools
			if ext := n.FileExt(); ext == ".bam" {
				if err := index.CreateBamIndex(n.FilePath()); err != nil {
					return responder.RespondWithError(ctx, http.StatusBadRequest, "Error while creating bam index")
				}
				return nil
			} else {
				return responder.RespondWithError(ctx, http.StatusBadRequest, "Index type bai requires .bam file")
			}
		}

		idxtype := ctx.QueryValue("index")
		if _, ok := index.Indexers[idxtype]; !ok {
			return responder.RespondWithError(ctx, http.StatusBadRequest, "invalid index type")
		}

		newIndexer := index.Indexers[idxtype]
		f, _ := os.Open(n.FilePath())
		defer f.Close()
		idxer := newIndexer(f)

		count, err := idxer.Create(n.IndexPath() + "/" + ctx.QueryValue("index") + ".idx")
		if err != nil {
			logger.Error("err " + err.Error())
			return responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
		}

		idxInfo := node.IdxInfo{
			Type:        ctx.QueryValue("index"),
			TotalUnits:  count,
			AvgUnitSize: n.File.Size / count,
		}

		if idxtype == "chunkrecord" {
			idxInfo.AvgUnitSize = conf.CHUNK_SIZE
		}

		if err := n.SetIndexInfo(ctx.QueryValue("index"), idxInfo); err != nil {
			logger.Error("err@node.SetIndexInfo: " + err.Error())
		}

		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("END indexing: " + id)
		}

		return responder.RespondOK(ctx)

	} else {
		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("START PUT data: " + id)
		}
		params, files, err := request.ParseMultipartForm(ctx.HttpRequest())
		if err != nil {
			err_msg := "err@node_ParseMultipartForm: " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}

		err = n.Update(params, files)
		if err != nil {
			err_msg := "err@node_Update: " + id + ": " + err.Error()
			logger.Error(err_msg)
			return responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}
		responder.RespondWithData(ctx, n)
		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("END PUT data: " + id)
		}
	}
	return nil
}
