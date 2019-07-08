package index

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	e "github.com/MG-RAST/Shock/shock-server/errors"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/node"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/MG-RAST/Shock/shock-server/user"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	mgo "gopkg.in/mgo.v2"
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

	if idxType == "" {
		responder.RespondWithError(ctx, http.StatusInternalServerError, "idxType empty")
		return
	}

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
			logger.Error("err@node_Index: (node.Load) id=" + nid + ": " + e.NodeNotFound)
			responder.RespondWithError(ctx, http.StatusNotFound, e.NodeNotFound)
		} else {
			// In theory the db connection could be lost between
			// checking user and load but seems unlikely.
			err_msg := "err@node_Index: (node.Load) id=" + nid + ":" + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
		}
		return
	}

	rights := n.Acl.Check(u.Uuid)

	switch rmeth {
	case "DELETE":
		if rights["write"] == false && u.Admin == false && n.Acl.Owner != u.Uuid {
			logger.Error("err@node_Index: (Authenticate) id=" + nid + ": " + e.UnAuth)
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			return
		}

		if _, has := n.Indexes[idxType]; has {
			if err = n.DeleteIndex(idxType); err != nil {
				err_msg := "err@node_Index: (node.DeleteIndex) id=" + nid + ":" + err.Error()
				logger.Error(err_msg)
				responder.RespondWithError(ctx, http.StatusInternalServerError, err_msg)
				return
			}
			responder.RespondOK(ctx)
		} else {
			err_msg := fmt.Sprintf("Node %s does not have index of type %s.", n.Id, idxType)
			logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}

	case "GET":
		if rights["read"] == false && u.Admin == false && n.Acl.Owner != u.Uuid {
			logger.Error("err@node_Index: (Authenticate) id=" + nid + ": " + e.UnAuth)
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			return
		}

		if v, has := n.Indexes[idxType]; has {
			responder.RespondWithData(ctx, map[string]interface{}{idxType: v})
		} else {
			err_msg := fmt.Sprintf("Node %s does not have index of type %s.", n.Id, idxType)
			logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
		}

	case "PUT": // PUT should be idempotent
		if rights["write"] == false && u.Admin == false && n.Acl.Owner != u.Uuid {
			logger.Error("err@node_Index: (Authenticate) id=" + nid + ": " + e.UnAuth)
			responder.RespondWithError(ctx, http.StatusUnauthorized, e.UnAuth)
			return
		}

		// lock node
		err = locker.NodeLockMgr.LockNode(nid)
		if err != nil {
			err_msg := "err@node_Index: (LockNode) id=" + nid + ": " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		}
		defer locker.NodeLockMgr.UnlockNode(nid)

		// check for locks and file
		if n.HasFileLock() {
			logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + e.NodeFileLock)
			responder.RespondWithError(ctx, http.StatusLocked, e.NodeFileLock)
			return
		} else if n.HasIndexLock(idxType) {
			logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + e.NodeIndexLock)
			responder.RespondWithError(ctx, http.StatusLocked, e.NodeIndexLock)
			return
		} else if !n.HasFile() {
			logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + e.NodeNoFile)
			responder.RespondWithError(ctx, http.StatusBadRequest, e.NodeNoFile)
			return
		} else if idxType == "" {
			logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + e.InvalidIndex)
			responder.RespondWithError(ctx, http.StatusBadRequest, e.InvalidIndex+" , idxType is empty")
			return
		}

		// Gather query params
		query := ctx.HttpRequest().URL.Query()
		forceRebuildStr, forceRebuild := query["force_rebuild"]
		if forceRebuild {
			if forceRebuildStr[0] == "0" {
				forceRebuild = false
			}
		}

		// does it already exist
		if _, has := n.Indexes[idxType]; has {
			if idxType == "size" {
				responder.RespondOK(ctx)
				return
			} else if !forceRebuild {
				err_msg := "This index already exists, please add the parameter 'force_rebuild=1' to force a rebuild of the existing index."
				logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
		}

		_, isUpload := query["upload"]

		if !isUpload {
			// check for invalid combinations
			if _, ok := index.Indexers[idxType]; !ok && idxType != "bai" && idxType != "subset" && idxType != "column" {
				logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + e.InvalidIndex)
				responder.RespondWithError(ctx, http.StatusBadRequest, e.InvalidIndex+" , invalid combination")
				return
			}
		}
		if idxType == "size" {
			err_msg := fmt.Sprintf("Index type size is a virtual index and does not require index building.")
			logger.Error("err@node_Index: (node.Indexes) id=" + nid + ": " + err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		}
		if idxType == "bai" {
			if n.Type == "subset" {
				err_msg := "subset nodes do not support bam indices"
				logger.Error("err@node_Index: (index/bai) id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
			if ext := n.FileExt(); ext != ".bam" {
				err_msg := "Index type bai requires .bam file."
				logger.Error("err@node_Index: (index/bai) id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
		}
		if n.Type == "subset" && (idxType != "chunkrecord" || n.Subset.Parent.IndexName != "record") {
			err_msg := "For subset nodes, Shock currently only supports subset and chunkrecord indexes. Also, for a chunkrecord index, the subset node must have been generated from a record index."
			logger.Error("err@node_Index: (index/subset) id=" + nid + ": " + err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		}
		var colNum int
		if idxType == "column" {
			if n.Type == "subset" {
				err_msg := "Shock does not support column index creation on subset nodes."
				logger.Error("err@node_Index: (index/column) id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}

			if _, exists := query["number"]; !exists {
				err_msg := "Index type column requires a number parameter in the url."
				logger.Error("err@node_Index: (index/column) id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
			num_str := query.Get("number")
			idxType = idxType + num_str
			colNum, err = strconv.Atoi(num_str)
			if err != nil || colNum < 1 {
				err_msg := "Index type column requires a number parameter in the url of an integer greater than zero."
				logger.Error("err@node_Index: (index/column) id=" + nid + ": " + err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
		}

		if isUpload {

			indexFormat := ""
			indexFormatValue, hasIndexFormat := query["indexFormat"]
			if hasIndexFormat {
				indexFormat = indexFormatValue[0]
			}

			var avgUnitSize int64
			avgUnitSize = 0
			avgUnitSizeValue, hasAvgUnitSize := query["avgUnitSize"]
			if hasAvgUnitSize {
				avgUnitSize, err = strconv.ParseInt(avgUnitSizeValue[0], 10, 64)
				if err != nil {
					err = fmt.Errorf("Could not parse avgUnitSize: %s", err.Error())
					responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
					return
				}
			}

			var totalUnits int64
			totalUnits = 0
			totalUnitsValue, hasTotalUnits := query["totalUnits"]
			if hasTotalUnits {
				totalUnits, err = strconv.ParseInt(totalUnitsValue[0], 10, 64)
				if err != nil {
					err = fmt.Errorf("Could not parse totalUnits: %s", err.Error())
					responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
					return
				}
			}

			var files file.FormFiles
			_, files, err = request.ParseMultipartForm(ctx.HttpRequest())
			if err != nil {
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}
			indexUpload, hasIndexUpload := files["upload"]

			if !hasIndexUpload {
				err = fmt.Errorf("index file missing, use upload field")
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}
			// move index file
			newIndexFilePath := n.IndexPath() + "/" + idxType + ".idx"
			err = os.Rename(indexUpload.Path, newIndexFilePath)
			if err != nil {
				responder.RespondWithError(ctx, http.StatusBadRequest, err.Error())
				return
			}

			idxInfo := &node.IdxInfo{
				Type:        idxType,
				TotalUnits:  totalUnits,
				AvgUnitSize: avgUnitSize,
				Format:      indexFormat,
				CreatedOn:   time.Now(),
				Locked:      nil,
			}

			n.SetIndexInfo(idxType, idxInfo)
			if err = n.Save(); err != nil {
				err_msg := "err@node_Index (node.Save): id=" + nid + ": " + err.Error()
				logger.Error(err_msg)
				responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
				return
			}
			//responder.RespondOK(ctx)
			responder.RespondWithData(ctx, idxInfo)
			return
		}
		// lock this index, trigger async indexing, save state
		idxInfo := &node.IdxInfo{
			Type:   idxType,
			Locked: locker.IndexLockMgr.Add(nid, idxType),
		}
		n.SetIndexInfo(idxType, idxInfo)
		if err = n.Save(); err != nil {
			err_msg := "err@node_Index (node.Save): id=" + nid + ": " + err.Error()
			logger.Error(err_msg)
			responder.RespondWithError(ctx, http.StatusBadRequest, err_msg)
			return
		}

		go node.AsyncIndexer(idxType, nid, colNum, ctx)
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
