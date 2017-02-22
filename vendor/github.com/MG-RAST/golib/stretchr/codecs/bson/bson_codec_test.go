package bson

import (
	"github.com/MG-RAST/golib/stretchr/codecs"
	"github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/golib/stretchr/testify/assert"
	"testing"
)

func TestInterface(t *testing.T) {

	assert.Implements(t, (*codecs.Codec)(nil), new(BsonCodec))

}

func TestMarshal(t *testing.T) {

	codec := new(BsonCodec)

	obj := make(map[string]string)
	obj["name"] = "Tyler"
	expectedResult := []byte{0x15, 0x0, 0x0, 0x0, 0x2, 0x6e, 0x61, 0x6d, 0x65, 0x0, 0x6, 0x0, 0x0, 0x0, 0x54, 0x79, 0x6c, 0x65, 0x72, 0x0, 0x0}

	bsonData, bsonError := codec.Marshal(obj, nil)

	if bsonError != nil {
		t.Errorf("Shouldn't return error: %s", bsonError)
	}

	assert.Equal(t, bsonData, expectedResult)

}

func TestUnmarshal(t *testing.T) {

	codec := new(BsonCodec)
	bsonData := []byte{0x15, 0x0, 0x0, 0x0, 0x2, 0x6e, 0x61, 0x6d, 0x65, 0x0, 0x6, 0x0, 0x0, 0x0, 0x54, 0x79, 0x6c, 0x65, 0x72, 0x0, 0x0}
	var object map[string]interface{}

	err := codec.Unmarshal(bsonData, &object)

	if assert.Nil(t, err) {
		assert.Equal(t, "Tyler", object["name"])
	}

}

func TestResponseContentType(t *testing.T) {

	codec := new(BsonCodec)
	assert.Equal(t, codec.ContentType(), constants.ContentTypeBSON)
}

func TestFileExtension(t *testing.T) {

	codec := new(BsonCodec)
	assert.Equal(t, constants.FileExtensionBSON, codec.FileExtension())

}

func TestCanMarshalWithCallback(t *testing.T) {

	codec := new(BsonCodec)
	assert.False(t, codec.CanMarshalWithCallback())

}
