package request

import (
	"errors"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/node/archive"
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/file/index"
	"github.com/MG-RAST/Shock/shock-server/node/filter"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"io"
	"net/http"
	"os/exec"
)

// MultiStreamer if for taking multiple files and creating one stream through an archive format: zip, tar, etc.

type MultiStreamer struct {
	Files       []*file.FileInfo
	W           http.ResponseWriter
	ContentType string
	Filename    string
	Archive     string
}

type Streamer struct {
	R           []file.SectionReader
	W           http.ResponseWriter
	ContentType string
	Filename    string
	Size        int64
	Filter      filter.FilterFunc
	Compression string
}

func (s *Streamer) Stream(streamRaw bool) (err error) {
	// file download
	if !streamRaw {
		fileName := fmt.Sprintf(" attachment; filename=%s", s.Filename)
		// add extension for compression or archive
		if s.Compression != "" {
			fileName = fmt.Sprintf(" attachment; filename=%s.%s", s.Filename, s.Compression)
		}
		s.W.Header().Set("Content-Disposition", fileName)
	}
	// set headers
	s.W.Header().Set("Content-Type", s.ContentType)
	s.W.Header().Set("Connection", "close")
	s.W.Header().Set("Access-Control-Allow-Headers", "Authorization")
	s.W.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	s.W.Header().Set("Access-Control-Allow-Origin", "*")
	if (s.Size > 0) && (s.Filter == nil) && (s.Compression == "") {
		s.W.Header().Set("Content-Length", fmt.Sprint(s.Size))
	}

	// pipe each SectionReader into one stream
	// run filter pre-pipe
	pReader, pWriter := io.Pipe()
	go func() {
		for _, sr := range s.R {
			var rs io.Reader
			if s.Filter != nil {
				rs = s.Filter(sr)
			} else {
				rs = sr
			}
			io.Copy(pWriter, rs)
		}
		pWriter.Close()
	}()

	// pass pipe to ResponseWriter, go through compression if exists
	cReader := archive.CompressReader(s.Compression, s.Filename, pReader)
	_, err = io.Copy(s.W, cReader)
	cReader.Close()
	pReader.Close()
	return
}

func (m *MultiStreamer) MultiStream() (err error) {
	// set headers
	fileName := fmt.Sprintf(" attachment; filename=%s", m.Filename)
	m.W.Header().Set("Content-Type", m.ContentType)
	m.W.Header().Set("Connection", "close")
	m.W.Header().Set("Access-Control-Allow-Headers", "Authorization")
	m.W.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
	m.W.Header().Set("Access-Control-Allow-Origin", "*")
	m.W.Header().Set("Content-Disposition", fileName)

	// pipe each SectionReader into one stream
	for _, f := range m.Files {
		pReader, pWriter := io.Pipe()
		f.Body = pReader
		go func(lf *file.FileInfo) {
			for _, sr := range lf.R {
				io.Copy(pWriter, sr)
			}
			pWriter.Close()
		}(f)
	}

	// pass pipes through archiver to ResponseWriter
	aReader := archive.ArchiveReader(m.Archive, m.Files)
	_, err = io.Copy(m.W, aReader)
	aReader.Close()
	for _, f := range m.Files {
		f.Body.Close()
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
