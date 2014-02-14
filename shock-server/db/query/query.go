// Package query to parse complex mongodb queries (UNFINISHED)
package query

/*
import (
	"github.com/MG-RAST/golib/mgo/bson"
	"io"
	"ioutil"
)

type mQuery struct {
	Query      *bson.M
	Projection *map[string]bool
}

func Parse(r *io.Reader) (q *mQuery, err error) {
    body := []byte{}
    if body, err = ioutil.ReadAll(r); err != nil {
        return nil, err
    }
    return parseBytes(body)
}

func parseBytes(b []bytes) (q *mQuery, err error) {
    i := interface{}
    err = json.Unmarshal(b, &i)
    if err != nil {
        return nil, err
    }

}
*/
