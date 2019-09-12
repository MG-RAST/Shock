package node

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"syscall"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/locker"
)

func (node *Node) SetFile(file file.FormFile) (err error) {
	fileStat, err := os.Stat(file.Path)
	if err != nil {
		return
	}

	err = os.Rename(file.Path, node.FilePath())
	if err != nil {
		return
	}
	node.File.Name = file.Name
	node.File.Size = fileStat.Size()
	node.File.Checksum = file.Checksum
	node.File.CreatedOn = fileStat.ModTime()

	//fill size index info
	totalunits := node.File.Size / conf.CHUNK_SIZE
	m := node.File.Size % conf.CHUNK_SIZE
	if m != 0 {
		totalunits += 1
	}
	node.Indexes["size"] = &IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
		Format:      "dynamic",
		CreatedOn:   time.Now(),
	}

	return
}

func (node *Node) SetFileFromSubset(subsetIndices file.FormFile) (err error) {
	// load parent node
	var n *Node
	n, err = Load(node.Subset.Parent.Id)
	if err != nil {
		return err
	}

	if _, indexExists := n.Indexes[node.Subset.Parent.IndexName]; !indexExists {
		return errors.New("Index '" + node.Subset.Parent.IndexName + "' does not exist for parent node.")
	}

	parentIndexFile := n.IndexPath() + "/" + node.Subset.Parent.IndexName + ".idx"
	if _, statErr := os.Stat(parentIndexFile); statErr != nil {
		return errors.New("Could not stat index file for parent node where parent node = '" + node.Subset.Parent.Id + "' and index = '" + node.Subset.Parent.IndexName + "'.")
	}

	// we default to "array" index format for backwards compatibility
	indexFormat := "array"
	if n.Indexes[node.Subset.Parent.IndexName].Format == "array" || n.Indexes[node.Subset.Parent.IndexName].Format == "matrix" {
		indexFormat = n.Indexes[node.Subset.Parent.IndexName].Format
	}

	if fi, statErr := os.Stat(subsetIndices.Path); statErr != nil {
		return errors.New("Could not stat uploaded subset_indices file.")
	} else {
		if fi.Size() == 0 {
			return errors.New("Uploaded subset_indices file is size zero.  This is prohibited.")
		}
	}

	f, _ := os.Open(subsetIndices.Path)
	defer f.Close()
	idxer := index.NewSubsetIndexer(f)
	coIndexPath := node.Path() + "/" + node.Id + ".subset.idx"
	oIndexPath := node.Path() + "/idx/" + node.Subset.Parent.IndexName + ".idx"

	coCount, oCount, oSize, err := index.CreateSubsetNodeIndexes(&idxer, coIndexPath, oIndexPath, parentIndexFile, indexFormat, n.Indexes[node.Subset.Parent.IndexName].TotalUnits)
	if err != nil {
		return
	}

	// this info refers to the compressed index for the subset node's data file
	node.Subset.Index.Path = coIndexPath
	node.Subset.Index.TotalUnits = coCount
	node.Subset.Index.AvgUnitSize = oSize / coCount
	node.Subset.Index.Format = "array"
	node.File.Size = oSize
	node.File.CreatedOn = time.Now()

	// this info is for the subset index that's been created in the index folder
	node.Indexes[node.Subset.Parent.IndexName] = &IdxInfo{
		Type:        "subset",
		TotalUnits:  oCount,
		AvgUnitSize: oSize / oCount,
		Format:      indexFormat,
		CreatedOn:   time.Now(),
	}

	// fill size index info
	totalunits := node.File.Size / conf.CHUNK_SIZE
	m := node.File.Size % conf.CHUNK_SIZE
	if m != 0 {
		totalunits += 1
	}
	node.Indexes["size"] = &IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
		Format:      "dynamic",
		CreatedOn:   time.Now(),
	}

	return
}

// SetFileFromPath _
func (node *Node) SetFileFromPath(path string, action string) (err error) {
	if action != "copy_file" && action != "move_file" && action != "keep_file" {
		err = errors.New("setting file from path requires action field equal to copy_file, move_file or keep_file")
		return
	}

	fileStat, err := os.Stat(path)
	if err != nil {
		return
	}
	node.File.Name = fileStat.Name()
	node.File.Size = fileStat.Size()
	node.File.CreatedOn = fileStat.ModTime()

	// only for copy_file and move_file
	tmpPath := fmt.Sprintf("%s/temp/%d%d", conf.PATH_DATA, rand.Int(), rand.Int())

	if action == "move_file" {
		// Determine if device ID of src and target are the same before proceeding.
		var devID1 uint64
		var f, _ = os.Open(path)
		var fi, _ = f.Stat()
		sIf := fi.Sys()
		s, ok := sIf.(*syscall.Stat_t)
		if !ok {
			err = errors.New("Could not determine device ID of data input file, try copy_file action instead.")
			return
		}

		devID1 = uint64(s.Dev)

		var devID2 uint64
		var f2, _ = os.Open(tmpPath)
		var fi2, _ = f2.Stat()
		s2If := fi2.Sys()
		s2, ok := s2If.(*syscall.Stat_t)
		if !ok {
			return errors.New("Could not complete move_file action, try copy_file action instead.")
		}
		devID2 = uint64(s2.Dev)

		if devID1 != devID2 {
			err = errors.New("Will not be able to rename file because data input file is on a different device than Shock data directory, try copy_file action instead.")
			return
		}
	}

	// Open file for reading.
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Set writer
	var dst io.Writer
	md5h := md5.New()

	if action == "copy_file" {
		tmpFile, err := os.Create(tmpPath)
		if err != nil {
			return err
		}
		defer tmpFile.Close()
		dst = io.MultiWriter(tmpFile, md5h)
	} else {
		dst = md5h
	}

	if _, err = io.Copy(dst, f); err != nil {
		return err
	}

	node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))

	//fill size index info
	totalunits := node.File.Size / conf.CHUNK_SIZE
	m := node.File.Size % conf.CHUNK_SIZE
	if m != 0 {
		totalunits += 1
	}
	node.Indexes["size"] = &IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
		Format:      "dynamic",
		CreatedOn:   time.Now(),
	}

	if action == "copy_file" {
		err = os.Rename(tmpPath, node.FilePath())
		if err != nil {
			return
		}
	} else if action == "move_file" {
		err = os.Rename(path, node.FilePath())
		if err != nil {
			return
		}
	} else {
		node.File.Path = path
	}

	return
}

// this runs asynchronously and uses FileLockMgr
func SetFileFromParts(id string, compress string, numParts int, allowEmpty bool) {
	// use function closure to get current value of err at return time
	var err error
	defer func() {
		locker.FileLockMgr.Error(id, err)
	}()

	outf := fmt.Sprintf("%s/%s.data", getPath(id), id)

	var outh *os.File
	outh, err = os.Create(outf)
	if err != nil {
		return
	}
	defer outh.Close()

	pReader, pWriter := io.Pipe()
	defer pReader.Close()

	cError := make(chan error)
	cKill := make(chan bool)

	// goroutine to add file readers to pipe while reading pipe
	// allows us to only open one file at a time but stream the data as if it was one reader
	go func() {
		var ferr error
		killed := false
		for i := 1; i <= numParts; i++ {
			select {
			case <-cKill:
				killed = true
				break
			default:
			}
			filename := fmt.Sprintf("%s/parts/%d", getPath(id), i)
			// skip this portion unless either
			// 1. file exists, or
			// 2. file does not exist and allowEmpty == false
			if _, errf := os.Stat(filename); errf == nil || (errf != nil && allowEmpty == false) {
				var part *os.File
				part, ferr = os.Open(filename)
				if ferr != nil {
					break
				}
				_, ferr = io.Copy(pWriter, part)
				part.Close()
				if ferr != nil {
					break
				}
			}
		}
		pWriter.Close()
		select {
		case <-cKill:
			killed = true
		default:
		}
		if killed {
			return
		}
		cError <- ferr
	}()

	md5h := md5.New()
	dst := io.MultiWriter(outh, md5h)

	// write from pipe to outfile / md5
	// handle optional compression
	var ucReader io.Reader
	ucReader, err = archive.UncompressReader(compress, pReader)
	if err != nil {
		close(cKill)
		os.Remove(outf)
		return
	}
	_, cerr := io.Copy(dst, ucReader)

	// get any errors from channel / finish copy
	err = <-cError
	if err != nil {
		os.Remove(outf)
		return
	}
	if cerr != nil {
		os.Remove(outf)
		err = cerr
		return
	}

	// get file info and update node
	var fileStat os.FileInfo
	fileStat, err = outh.Stat()
	if err != nil {
		return
	}

	// lock node while updating
	err = locker.NodeLockMgr.LockNode(id)
	if err != nil {
		return
	}
	defer locker.NodeLockMgr.UnlockNode(id)

	var node *Node
	node, err = Load(id)
	if err != nil {
		return
	}

	if node.File.Name == "" {
		node.File.Name = node.Id
	}
	node.File.Size = fileStat.Size()
	node.File.Checksum["md5"] = fmt.Sprintf("%x", md5h.Sum(nil))
	node.File.CreatedOn = fileStat.ModTime()

	//fill size index info
	totalunits := node.File.Size / conf.CHUNK_SIZE
	m := node.File.Size % conf.CHUNK_SIZE
	if m != 0 {
		totalunits += 1
	}
	node.Indexes["size"] = &IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
		Format:      "dynamic",
		CreatedOn:   time.Now(),
	}

	err = os.RemoveAll(node.Path() + "/parts/")
	if err != nil {
		return
	}
	err = node.Save()
	if err != nil {
		return
	}
	locker.FileLockMgr.Remove(id)

	return
}

func (node *Node) Rmdir() (err error) {
	return os.RemoveAll(node.Path())
}

func (node *Node) Mkdir() (err error) {
	err = os.MkdirAll(node.Path(), 0777)
	if err != nil {
		return
	}
	err = os.MkdirAll(node.IndexPath(), 0777)
	if err != nil {
		return
	}
	return
}
