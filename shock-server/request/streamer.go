package request

import (
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/filter"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"io"
	"net/http"
	"os/exec"
)

type Streamer struct {
	R           []file.SectionReader
	W           http.ResponseWriter
	ContentType string
	Filename    string
	Size        int64
	Filter      filter.FilterFunc
	FilterType  string
}

func (s *Streamer) Stream() (err error) {
	fileName := fmt.Sprintf(" attachment; filename=%s", s.Filename)
	if (s.FilterType == "gzip") || (s.FilterType == "zip") {
		fileName = fmt.Sprintf(" attachment; filename=%s.%s", s.Filename, s.FilterType)
	}

	s.W.Header().Set("Content-Type", s.ContentType)
	s.W.Header().Set("Content-Disposition", fileName)
	s.W.Header().Set("Connection", "close")
	s.W.Header().Set("Access-Control-Allow-Headers", "Authorization")
	s.W.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	s.W.Header().Set("Access-Control-Allow-Origin", "*")

	if s.Size > 0 && s.Filter == nil {
		s.W.Header().Set("Content-Length", fmt.Sprint(s.Size))
	}

	for i, sr := range s.R {
		n := s.Filename
		if len(s.R) > 1 {
			n = fmt.Sprintf("%s.%d", s.Filename, i+1)
		}
		var rs io.Reader
		if s.Filter != nil {
			rs = s.Filter(sr, n)
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

func (s *Streamer) StreamRaw() (err error) {
	s.W.Header().Set("Content-Type", s.ContentType)
	s.W.Header().Set("Connection", "close")
	s.W.Header().Set("Access-Control-Allow-Headers", "Authorization")
	s.W.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	s.W.Header().Set("Access-Control-Allow-Origin", "*")
	if s.Size > 0 && s.Filter == nil {
		s.W.Header().Set("Content-Length", fmt.Sprint(s.Size))
	}
	for i, sr := range s.R {
		n := s.Filename
		if len(s.R) > 1 {
			n = fmt.Sprintf("%s.%d", s.Filename, i+1)
		}
		var rs io.Reader
		if s.Filter != nil {
			rs = s.Filter(sr, n)
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

	index.LoadBamIndex(filePath)

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

	index.UnLoadBamIndex(filePath)

	return
}

//helper function to translate args in URL query to samtools args
//manual: http://samtools.sourceforge.net/samtools.shtml
func ParseSamtoolsArgs(ctx context.Context) (argv []string, err error) {

	query := ctx.HttpRequest().URL.Query()
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
		if _, ok := query[src]; ok {
			argv = append(argv, des)
		}
	}

	for src, des := range valued_options {
		if _, ok := query[src]; ok {
			if val := query.Get(src); val != "" {
				argv = append(argv, des)
				argv = append(argv, val)
			} else {
				return nil, errors.New(fmt.Sprintf("required value not found for query arg: %s ", src))
			}
		}
	}
	return argv, nil
}
