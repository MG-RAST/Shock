package json

import (
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/testify/assert"
	"testing"
)

var codec JsonCodec

func TestInterface(t *testing.T) {

	assert.Implements(t, (*codecs.Codec)(nil), new(JsonCodec), "JsonCodec")

}

func TestMarshal(t *testing.T) {

	obj := make(map[string]string)
	obj["name"] = "Mat"

	jsonString, jsonError := codec.Marshal(obj, nil)

	if jsonError != nil {
		t.Errorf("Shouldn't return error: %s", jsonError)
	}

	assert.Equal(t, string(jsonString), `{"name":"Mat"}`)

}

func TestUnmarshal(t *testing.T) {

	jsonString := `{"name":"Mat"}`
	var object map[string]interface{}

	err := codec.Unmarshal([]byte(jsonString), &object)

	if err != nil {
		t.Errorf("Shouldn't return error: %s", err)
	}

	assert.Equal(t, "Mat", object["name"])

}

func TestResponseContentType(t *testing.T) {

	assert.Equal(t, codec.ContentType(), constants.ContentTypeJSON)

}

func TestFileExtension(t *testing.T) {

	assert.Equal(t, constants.FileExtensionJSON, codec.FileExtension())

}

func TestCanMarshalWithCallback(t *testing.T) {

	assert.False(t, codec.CanMarshalWithCallback())

}
