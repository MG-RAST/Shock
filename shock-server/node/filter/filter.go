package filter

import (
	"github.com/MG-RAST/Shock/shock-server/node/file"
	"github.com/MG-RAST/Shock/shock-server/node/filter/anonymize"
	"github.com/MG-RAST/Shock/shock-server/node/filter/fq2fa"
	"io"
)

type FilterFunc func(file.SectionReader) io.Reader

var (
	filters = map[string]FilterFunc{
		"anonymize": anonymize.NewReader,
		"fq2fa":     fq2fa.NewReader,
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

func NewReader(f string, fh file.SectionReader) io.Reader {
	return filters[f](fh)
}
