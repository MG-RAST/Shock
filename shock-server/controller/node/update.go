package node

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/indexer"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/jaredwilkening/goweb"
	"net/http"
	"os"
)

// PUT: /node/{id} -> multipart-form
func (cr *Controller) Update(id string, cx *goweb.Context) {
	// Log Request and check for Auth
	request.Log(cx.Request)
	u, err := request.Authenticate(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, cx)
		return
	}

	// Gather query params
	query := request.Q(cx.Request.URL.Query())

	// Fake public user
	if u == nil {
		u = &user.User{Uuid: ""}
	}

	n, err := node.Load(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			cx.RespondWithError(http.StatusUnauthorized)
			return
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			logger.Error("Err@node_Update:LoadNode: " + err.Error())
			cx.RespondWithError(http.StatusInternalServerError)
			return
		}
	}

	if query.Has("index") {
		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("START indexing: " + id)
		}

		if !n.HasFile() {
			cx.RespondWithErrorMessage("node file empty", http.StatusBadRequest)
			return
		}

		if query.Value("index") == "bai" {
			//bam index is created by the command-line tool samtools
			if ext := n.FileExt(); ext == ".bam" {
				if err := request.CreateBamIndex(n.FilePath()); err != nil {
					cx.RespondWithErrorMessage("Error while creating bam index", http.StatusBadRequest)
					return
				}
				return
			} else {
				cx.RespondWithErrorMessage("Index type bai requires .bam file", http.StatusBadRequest)
				return
			}
		}

		idxtype := query.Value("index")
		if _, ok := indexer.Indexers[idxtype]; !ok {
			cx.RespondWithErrorMessage("invalid index type", http.StatusBadRequest)
			return
		}

		newIndexer := indexer.Indexer(idxtype)
		f, _ := os.Open(n.FilePath())
		defer f.Close()
		idxer := newIndexer(f)
		count, err := idxer.Create()
		if err != nil {
			logger.Error("err " + err.Error())
			cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
			return
		}

		if err := idxer.Dump(n.IndexPath() + "/" + query.Value("index") + ".idx"); err != nil {
			logger.Error("err " + err.Error())
			cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
			return
		}

		idxInfo := node.IdxInfo{
			Type:        query.Value("index"),
			TotalUnits:  count,
			AvgUnitSize: n.File.Size / count,
		}

		if idxtype == "chunkrecord" {
			idxInfo.AvgUnitSize = conf.CHUNK_SIZE
		}

		if err := n.SetIndexInfo(query.Value("index"), idxInfo); err != nil {
			logger.Error("err@node.SetIndexInfo: " + err.Error())
		}

		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("END indexing: " + id)
		}

		cx.RespondWithOK()
		return

	} else {
		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("START PUT data: " + id)
		}
		params, files, err := request.ParseMultipartForm(cx.Request)
		if err != nil {
			logger.Error("err@node_ParseMultipartForm: " + err.Error())
			cx.RespondWithError(http.StatusBadRequest)
			return
		}

		err = n.Update(params, files)
		if err != nil {
			errors := []string{e.FileImut, e.AttrImut, "parts cannot be less than 1"}
			for e := range errors {
				if err.Error() == errors[e] {
					cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
					return
				}
			}
			logger.Error("err@node_Update: " + id + ":" + err.Error())
			cx.RespondWithErrorMessage(err.Error(), http.StatusBadRequest)
			return
		}
		cx.RespondWithData(n)
		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("END PUT data: " + id)
		}
	}
	return
}
