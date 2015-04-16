package node

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"io"
	"math/rand"
	"os"
	"syscall"
	"time"
)

func (node *Node) SetFile(file FormFile) (err error) {
	fileStat, err := os.Stat(file.Path)
	if err != nil {
		return
	}
	os.Rename(file.Path, node.FilePath())
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
	node.Indexes["size"] = IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
		Format:      "dynamic",
		CreatedOn:   time.Now(),
	}
	err = node.Save()
	return
}

func (node *Node) SetFileFromSubset(subsetIndices FormFile) (err error) {
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
	node.Indexes[node.Subset.Parent.IndexName] = IdxInfo{
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
	node.Indexes["size"] = IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
		Format:      "dynamic",
		CreatedOn:   time.Now(),
	}

	err = node.Save()
	return
}

func (node *Node) SetFileFromPath(path string, action string) (err error) {
	fileStat, err := os.Stat(path)
	if err != nil {
		return
	}
	node.File.Name = fileStat.Name()
	node.File.Size = fileStat.Size()
	node.File.CreatedOn = fileStat.ModTime()

	tmpPath := fmt.Sprintf("%s/temp/%d%d", conf.PATH_DATA, rand.Int(), rand.Int())

	if action != "copy_file" && action != "move_file" && action != "keep_file" {
		return errors.New("setting file from path requires action field equal to copy_file, move_file or keep_file")
	}

	if action == "move_file" {
		// Determine if device ID of src and target are the same before proceeding.
		var devID1 uint64
		var f, _ = os.Open(path)
		var fi, _ = f.Stat()
		s := fi.Sys()
		if s, ok := s.(*syscall.Stat_t); ok {
			devID1 = uint64(s.Dev)
		} else {
			return errors.New("Could not determine device ID of data input file, try copy_file action instead.")
		}

		var devID2 uint64
		var f2, _ = os.Open(tmpPath)
		var fi2, _ = f2.Stat()
		s2 := fi2.Sys()
		if s2, ok := s2.(*syscall.Stat_t); ok {
			devID2 = uint64(s2.Dev)
		} else {
			return errors.New("Could not complete move_file action, try copy_file action instead.")
		}

		if devID1 != devID2 {
			return errors.New("Will not be able to rename file because data input file is on a different device than Shock data directory, try copy_file action instead.")
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
	node.Indexes["size"] = IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
		Format:      "dynamic",
		CreatedOn:   time.Now(),
	}

	if action == "copy_file" {
		os.Rename(tmpPath, node.FilePath())
	} else if action == "move_file" {
		os.Rename(path, node.FilePath())
	} else {
		node.File.Path = path
	}
	err = node.Save()
	return
}

func (node *Node) SetFileFromParts(allowEmpty bool) (err error) {
	outf := fmt.Sprintf("%s/%s.data", node.Path(), node.Id)
	outh, oerr := os.Create(outf)
	if oerr != nil {
		return oerr
	}
	defer outh.Close()

	pReader, pWriter := io.Pipe()
	defer pReader.Close()

	cError := make(chan error)
	cKill := make(chan bool)

	// goroutine to add file readers to pipe while gzip reads pipe
	// allows us to only open one file at a time but stream the data to gzip as if it was one reader
	go func() {
		var ferr error
		killed := false
		for i := 1; i <= node.Parts.Count; i++ {
			select {
			case <-cKill:
				killed = true
				break
			default:
			}
			filename := fmt.Sprintf("%s/parts/%d", node.Path(), i)
			// skip this portion unless either
			// 1. file exists, or
			// 2. file does not exist and allowEmpty == false
			if _, errf := os.Stat(filename); errf == nil || (errf != nil && allowEmpty == false) {
				part, err := os.Open(filename)
				if err != nil {
					ferr = err
					break
				}
				_, err = io.Copy(pWriter, part)
				part.Close()
				if err != nil {
					ferr = err
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
	ucReader, ucErr := archive.UncompressReader(node.Parts.Compression, pReader)
	if ucErr != nil {
		close(cKill)
		os.Remove(outf)
		return ucErr
	}
	_, cerr := io.Copy(dst, ucReader)

	// get any errors from channel / finish copy
	if eerr := <-cError; eerr != nil {
		os.Remove(outf)
		return eerr
	}
	if cerr != nil {
		os.Remove(outf)
		return cerr
	}

	// get file info and update node
	fileStat, ferr := os.Stat(outf)
	if ferr != nil {
		return ferr
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
	node.Indexes["size"] = IdxInfo{
		Type:        "size",
		TotalUnits:  totalunits,
		AvgUnitSize: conf.CHUNK_SIZE,
		Format:      "dynamic",
		CreatedOn:   time.Now(),
	}

	err = node.Save()
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
