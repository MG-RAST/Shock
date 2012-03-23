package filters

import (
	"io"
)

interface Filter {
	io.Reader
}

