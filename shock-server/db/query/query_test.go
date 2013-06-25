package query_test

import (
	//"fmt"
	. "github.com/MG-RAST/Shock/shock-server/query"
	"testing"
)

const testQ = `{{"tags": {"$all": ["metagenome", "soil"]}},{"id": true}}`

func TestParse(t *testing.T) {
}
