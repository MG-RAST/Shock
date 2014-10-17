package util

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/golib/stretchr/goweb/context"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

// Arrays to check for valid param and file form names for node creation and updating, and also acl modification.
// Note: indexing and querying do not use functions that use these arrays and thus we don't have to include those field names.
var validParams = []string{"action", "all", "attributes_str", "copy_data", "copy_indexes", "delete", "file_name", "format", "ids", "index_name", "linkage", "operation", "owner", "parent_index", "parent_node", "parts", "path", "read", "source", "tags", "type", "users", "write"}
var validFiles = []string{"attributes", "subset_indices", "upload", "gzip", "bzip2"}
var validUpload = []string{"upload", "gzip", "bzip2"}

type UrlResponse struct {
	Url       string `json:"url"`
	ValidTill string `json:"validtill"`
}

type Query struct {
	list map[string][]string
}

func Q(l map[string][]string) *Query {
	return &Query{list: l}
}

func (q *Query) Has(key string) bool {
	if _, has := q.list[key]; has {
		return true
	}
	return false
}

func (q *Query) Value(key string) string {
	return q.list[key][0]
}

func (q *Query) List(key string) []string {
	return q.list[key]
}

func (q *Query) All() map[string][]string {
	return q.list
}

func RandString(l int) (s string) {
	rand.Seed(time.Now().UTC().UnixNano())
	c := make([]byte, l)
	for i := 0; i < l; i++ {
		c[i] = chars[rand.Intn(len(chars))]
	}
	return string(c)
}

func ToInt(s string) (i int) {
	i, _ = strconv.Atoi(s)
	return
}

func ApiUrl(ctx context.Context) string {
	if conf.Conf["api-url"] != "" {
		return conf.Conf["api-url"]
	}
	return "http://" + ctx.HttpRequest().Host
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StripSuffix(file string) string {
	if i := strings.LastIndex(file, "."); i > -1 {
		return file[:i]
	}
	return file
}

func IsValidParamName(a string) bool {
	for _, b := range validParams {
		if b == a {
			return true
		}
	}
	return false
}

func IsValidFileName(a string) bool {
	for _, b := range validFiles {
		if b == a {
			return true
		}
	}
	return false
}

func IsValidUploadFile(a string) bool {
	for _, b := range validUpload {
		if b == a {
			return true
		}
	}
	return false
}

func CopyFile(src string, dst string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}
