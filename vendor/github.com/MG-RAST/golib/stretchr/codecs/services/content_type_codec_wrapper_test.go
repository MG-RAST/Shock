package services

import (
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/json"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/testify/assert"
	"testing"
)

// testCodec is a json codec that marshals the passed in options
// instead of the data, for testing the wrapped codec.
type testCodec struct {
	json.JsonCodec
}

// Marshal passes the options as both the data and options to the
// JsonCodec's Marshal, for testing purposes.
func (c *testCodec) Marshal(data interface{}, options map[string]interface{}) ([]byte, error) {
	return c.JsonCodec.Marshal(options, options)
}

func TestWrapCodec_ContentType(t *testing.T) {
	codec := new(json.JsonCodec)
	testContentType := "application/vnd.stretchr.test+json"
	var target interface{} = wrapCodecWithContentType(codec, testContentType)

	wrappedCodec, ok := target.(codecs.Codec)
	assert.True(t, ok, "A wrapped codec should still be a Codec")

	if ok {
		assert.Equal(t, testContentType, wrappedCodec.ContentType())
	}
}

func TestWrapCodec_Marshal(t *testing.T) {
	codec := new(testCodec)
	testContentType := "application/vnd.stretchr.test+json"
	wrappedCodec := wrapCodecWithContentType(codec, testContentType)

	response, err := wrappedCodec.Marshal(nil, nil)
	assert.NoError(t, err)
	expectedResponse := `{"matched_type":"` + testContentType + `"}`
	assert.Equal(t, response, []byte(expectedResponse),
		"The wrapped codec should add the matched content type to options on unmarshal")
}
