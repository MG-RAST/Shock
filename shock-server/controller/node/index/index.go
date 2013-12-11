// Package index implements /node/:id/index resource (UNFINISHED)
package index

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/Shock/shock-server/util"
	"github.com/MG-RAST/golib/goweb"
	"net/http"
	"os"
)

type getRes struct {
	I interface{} `json:"indexes"`
	A interface{} `json:"available_indexers"`
}

type m map[string]string

// GET, POST, PUT, DELETE: /node/{nid}/index/{idxType}
var Controller goweb.ControllerFunc = func(cx *goweb.Context) {
	request.Log(cx.Request)
	u, err := request.Authenticate(cx.Request)
	if err != nil && err.Error() != e.NoAuth {
		request.AuthError(err, cx)
		return
	}

	// Fake public user
	if u == nil {
		u = &user.User{Uuid: ""}
	}

	// Load node and handle user unauthorized
	id := cx.PathParams["nid"]
	n, err := node.Load(id, u.Uuid)
	if err != nil {
		if err.Error() == e.UnAuth {
			cx.RespondWithErrorMessage(err.Error(), http.StatusUnauthorized)
			return
		} else if err.Error() == e.MongoDocNotFound {
			cx.RespondWithNotFound()
			return
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "Err@index:LoadNode: " + err.Error()
			logger.Error(err_msg)
			cx.RespondWithErrorMessage(err_msg, http.StatusInternalServerError)
			return
		}
	}

	idxType, hasType := cx.PathParams["type"]
	query := util.Q(cx.Request.URL.Query())
	switch cx.Request.Method {
	case "GET":
		if hasType {
			if v, has := n.Indexes[idxType]; has {
				cx.RespondWithData(map[string]interface{}{idxType: v})
			} else {
				cx.RespondWithErrorMessage(fmt.Sprintf("Node %s does not have index of type %s.", n.Id, idxType), http.StatusBadRequest)
			}
		} else {
			cx.RespondWithData(getRes{I: n.Indexes, A: filteredIndexes(n.Indexes)})
		}

	case "POST", "PUT":
		if !n.HasFile() {
			cx.RespondWithErrorMessage("Node has no file", http.StatusBadRequest)
			return
		} else if !hasType {
			cx.RespondWithErrorMessage("Index create requires type", http.StatusBadRequest)
			return
		}
		if !contains(filteredIndexes(n.Indexes), idxType) {
			cx.RespondWithErrorMessage(fmt.Sprintf("Index type %s unavailable", idxType), http.StatusBadRequest)
			return
		}

		if conf.Bool(conf.Conf["perf-log"]) {
			logger.Perf("START indexing: " + id)
		}

		if query.Value("index") == "bai" {
			//bam index is created by the command-line tool samtools
			if ext := n.FileExt(); ext == ".bam" {
				if err := index.CreateBamIndex(n.FilePath()); err != nil {
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
		if _, ok := index.Indexers[idxtype]; !ok {
			cx.RespondWithErrorMessage("invalid index type", http.StatusBadRequest)
			return
		}

		newIndexer := index.Indexers[idxtype]
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

	default:
		cx.RespondWithErrorMessage("This request type is not implemented", http.StatusNotImplemented)
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
