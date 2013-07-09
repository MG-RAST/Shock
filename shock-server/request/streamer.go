package request

import (
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/filter"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
)

type Streamer struct {
	R           []file.SectionReader
	W           http.ResponseWriter
	ContentType string
	Filename    string
	Size        int64
	Filter      filter.FilterFunc
}

func (s *Streamer) Stream() (err error) {
	s.W.Header().Set("Content-Type", s.ContentType)
	s.W.Header().Set("Content-Disposition", fmt.Sprintf(" attachment; filename=%s", s.Filename))
	if s.Size > 0 && s.Filter == nil {
		s.W.Header().Set("Content-Length", fmt.Sprint(s.Size))
	}
	for _, sr := range s.R {
		var rs io.Reader
		if s.Filter != nil {
			rs = s.Filter(sr)
		} else {
			rs = sr
		}
		_, err := io.Copy(s.W, rs)
		if err != nil {
			return err
		}
	}
	return
}

func (s *Streamer) StreamSamtools(filePath string, region string, args ...string) (err error) {
	//involking samtools in command line:
	//samtools view [-c] [-H] [-f INT] ... filname.bam [region]

	argv := []string{}
	argv = append(argv, "view")
	argv = append(argv, args...)
	argv = append(argv, filePath)

	if region != "" {
		argv = append(argv, region)
	}

	LoadBamIndex(filePath)

	cmd := exec.Command("samtools", argv...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	go io.Copy(s.W, stdout)

	err = cmd.Wait()
	if err != nil {
		return err
	}

	UnLoadBamIndex(filePath)

	return
}

//helper function to translate args in URL query to samtools args
//manual: http://samtools.sourceforge.net/samtools.shtml
func ParseSamtoolsArgs(query *Query) (argv []string, err error) {

	var (
		filter_options = map[string]string{
			"head":     "-h",
			"headonly": "-H",
			"count":    "-c",
		}
		valued_options = map[string]string{
			"flag":      "-f",
			"lib":       "-l",
			"mapq":      "-q",
			"readgroup": "-r",
		}
	)

	for src, des := range filter_options {
		if query.Has(src) {
			argv = append(argv, des)
		}
	}

	for src, des := range valued_options {
		if query.Has(src) {
			if val := query.Value(src); val != "" {
				argv = append(argv, des)
				argv = append(argv, val)
			} else {
				return nil, errors.New(fmt.Sprintf("required value not found for query arg: %s ", src))
			}
		}
	}
	return argv, nil
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
