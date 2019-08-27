package node

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
	"github.com/MG-RAST/Shock/shock-server/request"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
)

func AsyncIndexer(idxType string, nid string, colNum int, ctx context.Context) {
	// use function closure to get current value of err at return time
	var err error
	defer func() {
		locker.IndexLockMgr.Error(nid, idxType, err)
	}()

	// get node - this is read only
	var n *Node
	n, err = Load(nid)
	if err != nil {
		return
	}

	avgSize := int64(0)
	subsetSize := int64(0)
	count := int64(0)
	indexFormat := ""
	subsetName := ""

	var f *os.File
	if idxType == "bai" {
		err = index.CreateBamIndex(n.FilePath())
		if err != nil {
			return
		}
		indexFormat = "bai"
	} else if idxType == "column" {
		idxType = fmt.Sprintf("%s%d", idxType, colNum)
		f, err = os.Open(n.FilePath())
		if err != nil {
			return
		}
		defer f.Close()
		idxer := index.NewColumnIndexer(f)
		count, indexFormat, err = index.CreateColumnIndex(&idxer, colNum, n.IndexPath()+"/"+idxType+".idx")
		if err != nil {
			return
		}
	} else if idxType == "subset" {
		// Utilizing the multipart form parser since we need to upload a file.
		var params map[string]string
		var files file.FormFiles
		params, files, err = request.ParseMultipartForm(ctx.HttpRequest())
		// clean up temp dir !!
		defer file.RemoveAllFormFiles(files)
		if err != nil {
			return
		}

		parentIndex, hasParent := params["parent_index"]
		if !hasParent {
			err = errors.New("Index type subset requires parent_index param.")
			return
		} else if _, has := n.Indexes[parentIndex]; !has {
			err = fmt.Errorf("Node %s does not have index of type %s.", n.Id, parentIndex)
			return
		}

		newIndex, hasName := params["index_name"]
		if !hasName {
			err = errors.New("Index type subset requires index_name param.")
			return
		} else if _, reservedName := index.Indexers[newIndex]; reservedName || newIndex == "bai" {
			err = fmt.Errorf("%s is a reserved index name and cannot be used to create a custom subset index.", newIndex)
			return
		}
		subsetName = newIndex

		subsetIndices, hasFile := files["subset_indices"]
		if !hasFile {
			err = errors.New("Index type subset requires subset_indices file.")
			return
		}

		f, err = os.Open(subsetIndices.Path)
		if err != nil {
			return
		}
		defer f.Close()
		idxer := index.NewSubsetIndexer(f)

		// we default to "array" index format for backwards compatibility
		indexFormat = "array"
		if n.Indexes[parentIndex].Format == "array" || n.Indexes[parentIndex].Format == "matrix" {
			indexFormat = n.Indexes[parentIndex].Format
		}
		count, subsetSize, err = index.CreateSubsetIndex(&idxer, n.IndexPath()+"/"+newIndex+".idx", n.IndexPath()+"/"+parentIndex+".idx", indexFormat, n.Indexes[parentIndex].TotalUnits)
		if err != nil {
			return
		}
	} else if n.File.Size > 0 {
		newIndexer := index.Indexers[idxType] // newIndexer is only a constructor function
		f, err = os.Open(n.FilePath())
		if err != nil {
			return
		}
		defer f.Close()
		var idxer index.Indexer
		if n.Type == "subset" {
			idxer = newIndexer(f, n.Type, n.Subset.Index.Format, n.IndexPath()+"/"+n.Subset.Parent.IndexName+".idx")
		} else {
			idxer = newIndexer(f, n.Type, "", "")
		}
		count, indexFormat, err = idxer.Create(n.IndexPath() + "/" + idxType + ".idx")
		if err != nil {
			return
		}
	}

	// lock node for writes
	err = locker.NodeLockMgr.LockNode(nid)
	if err != nil {
		return
	}
	defer locker.NodeLockMgr.UnlockNode(nid)

	// refresh node
	n, err = Load(nid)
	if err != nil {
		return
	}

	if idxType == "bai" {
		avgSize = 0
	} else if (count == 0) && (n.File.Size > 0) {
		err = errors.New("Index is empty.")
		return
	} else if count > 0 {
		if idxType == "subset" {
			avgSize = subsetSize / count
		} else {
			avgSize = n.File.Size / count
		}
	}

	idxInfo := &IdxInfo{
		Type:        idxType,
		TotalUnits:  count,
		AvgUnitSize: avgSize,
		Format:      indexFormat,
		CreatedOn:   time.Now(),
		Locked:      nil,
	}
	if idxType == "subset" {
		idxType = subsetName
	}

	n.SetIndexInfo(idxType, idxInfo)
	err = n.Save()
	if err != nil {
		return
	}

	locker.IndexLockMgr.Remove(nid, idxType)
	return
}
