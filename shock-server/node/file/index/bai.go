package index

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type bai struct{}

func NewBaiIndexer(f *os.File) Indexer {
	return &bai{}
}

func (i *bai) Create(string) (count int64, err error) {
	return
}

func (i *bai) Close() (err error) {
	return
}

func CreateBamIndex(bamFile string) (err error) {
	err = exec.Command("samtools", "index", bamFile).Run()
	if err != nil {
		return err
	}

	baiFile := fmt.Sprintf("%s.bai", bamFile)
	idxPath := fmt.Sprintf("%s/idx/", filepath.Dir(bamFile))

	err = exec.Command("mv", baiFile, idxPath).Run()
	if err != nil {
		return err
	}

	return
}

func LoadBamIndex(bamFile string) (err error) {
	bamFileDir := filepath.Dir(bamFile)
	bamFileName := filepath.Base(bamFile)
	targetBai := fmt.Sprintf("%s/%s.bai", bamFileDir, bamFileName)
	srcBai := fmt.Sprintf("%s/idx/%s.bai", bamFileDir, bamFileName)
	err = exec.Command("ln", "-s", srcBai, targetBai).Run()
	if err != nil {
		return err
	}
	return
}

func UnLoadBamIndex(bamFile string) (err error) {
	bamFileDir := filepath.Dir(bamFile)
	bamFileName := filepath.Base(bamFile)
	targetBai := fmt.Sprintf("%s/%s.bai", bamFileDir, bamFileName)
	err = exec.Command("rm", targetBai).Run()
	if err != nil {
		return err
	}
	return
}
