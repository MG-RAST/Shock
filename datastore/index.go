package datastore

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	//"math/rand"
	//"fmt"
	//"errors"
	//bson "launchpad.net/mgo/bson"
)

func init() {
	/*
		var pos, length uint64 = 0, 0
		bFile, err := os.Create("/tmp/idx.bson"); if err != nil { fmt.Println(err.Error()) }
		jFile, err := os.Create("/tmp/idx.json"); if err != nil { fmt.Println(err.Error()) }

		//checksum := "716484d000da82756fe593bcec43213f" 
		idx := Index{Filename : "test", CType: "None", Idx: RecordIndex{}}
		for i := 0; i < 1024000; i++ {
			length = uint64(rand.Int63n(10240))
			idx.Idx = append(idx.Idx, Record{pos,length})
			pos = pos+length
		}

		b, err := bson.Marshal(idx); if err != nil { fmt.Println(err.Error()) }
		bFile.Write(b)
		j, err := json.Marshal(idx); if err != nil { fmt.Println(err.Error()) }
		jFile.Write(j)
		fmt.Println("done create")
		bFile.Close()
		jFile.Close()

		bf, err := ioutil.ReadFile("/tmp/idx.json"); if err != nil { fmt.Println(err.Error()) }
		idx = Index{}
		err = bson.Unmarshal(bf, &idx); if err != nil { fmt.Println(err.Error()) }
		fmt.Println("done")
		panic("this is a panic folks")
	*/
}

/*	
Shock Index format:
<position> - unsigned 64bit int
<length>   - unsigned 64bit int
<checksum> - optional (none,md5,sha1,sha256)

#filename=<filename>:checksum=<type>\n
<position><length><checksum><position><length><checksum>...

Json representation:
{
	index_type : <type>,
	filename : <filename>,
	checksum_type : <type>,
	version : <version>,
	index : [
		[<position>,<length>,<optional_checksum>]...
	]
}
*/

type Index struct {
	Type     string      `bson:"index_type" json:"index_type"`
	Filename string      `bson:"filename" json:"filename"`
	CType    string      `bson:"checksum_type" json:"checksum_type"`
	Idx      RecordIndex `bson:"index" json:"index"`
	Version  int         `bson:"version" json:"version"`
}

type BinaryIndex struct {
	Idx    [][]int64
	Length int
}

func NewBinaryIndex() *BinaryIndex {
	return &BinaryIndex{
		Idx:    [][]int64{},
		Length: 0,
	}
}

func (i *BinaryIndex) Append(rec []int64) {
	i.Idx = append(i.Idx, rec)
	i.Length += 1
}

func (i *BinaryIndex) Part(part string) (pos int64, length int64, err error) {
	if strings.Contains(part, "-") {
		startend := strings.Split(part, "-")
		start, startEr := strconv.ParseInt(startend[0], 10, 64)
		end, endEr := strconv.ParseInt(startend[1], 10, 64)
		if startEr != nil || endEr != nil || start <= 0 || start > int64(i.Length) || end <= 0 || end > int64(i.Length) {
			err = errors.New("")
			return
		}
		pos = i.Idx[(start - 1)][0]
		length = (i.Idx[(end - 1)][0] - i.Idx[(start - 1)][0]) + i.Idx[(end - 1)][1]
	} else {
		p, er := strconv.ParseInt(part, 10, 64)
		if er != nil || p <= 0 || p > int64(i.Length) {
			err = errors.New("")
			return
		}
		pos = i.Idx[(p - 1)][0]
		length = i.Idx[(p - 1)][1]
	}
	return
}

func (i *BinaryIndex) Dump(file string) (err error) {
	f, err := os.Create(file)
	defer f.Close()
	if err != nil {
		return
	}
	for _, rec := range i.Idx {
		binary.Write(f, binary.LittleEndian, rec[0])
		binary.Write(f, binary.LittleEndian, rec[1])
	}
	return
}

func (i *BinaryIndex) Load(file string) (err error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return
	}
	for {
		rec := make([]int64, 2)
		er := binary.Read(f, binary.LittleEndian, &rec[0])
		if er != nil {
			if er != io.EOF {
				err = er
			}
			return
		}
		er = binary.Read(f, binary.LittleEndian, &rec[1])
		if er != nil {
			if er != io.EOF {
				err = er
			}
			return
		}
		i.Append(rec)
	}
	return
}

type RecordIndex []Record

type Record []interface{}

func NewIndex() *Index {
	return &Index{Filename: "", CType: "", Idx: RecordIndex{}}
}

func (idx *Index) Save(filename string) (err error) {
	jFile, err := os.Create(filename)
	if err != nil {
		return
	}
	defer jFile.Close()
	j, err := json.Marshal(idx)
	if err != nil {
		return
	}
	jFile.Write(j)
	return
}

func (idx *Index) Load(filename string) (err error) {
	bf, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	err = json.Unmarshal(bf, idx)
	return
}
