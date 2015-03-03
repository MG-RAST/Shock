package filter

import (
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/filter/anonymize"
	"github.com/MG-RAST/Shock/shock-server/node/filter/fq2fa"
	"github.com/MG-RAST/Shock/shock-server/node/filter/gzip"
	"github.com/MG-RAST/Shock/shock-server/node/filter/zip"
	"io"
)

type FilterFunc func(file.SectionReader, string) io.Reader

var (
	filters = map[string]FilterFunc{
		"anonymize": anonymize.NewReader,
		"fq2fa":     fq2fa.NewReader,
		"gzip":      gzip.NewReader,
		"zip":       zip.NewReader,
	}
)

func Has(f string) bool {
	if _, has := filters[f]; has {
		return true
	}
	return false
}

func Filter(f string) FilterFunc {
	return filters[f]
}

func NewReader(f string, fh file.SectionReader, n string) io.Reader {
	return filters[f](fh, n)
}
