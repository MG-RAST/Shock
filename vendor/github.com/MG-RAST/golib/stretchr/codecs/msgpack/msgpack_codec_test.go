package msgpack

import (
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/testify/assert"
	"testing"
)

func TestInterface(t *testing.T) {

	assert.Implements(t, (*codecs.Codec)(nil), new(MsgpackCodec), "MsgpackCodec")

}

func TestMarshal(t *testing.T) {

	codec := new(MsgpackCodec)

	obj := make(map[string]string)
	obj["name"] = "Mat"

	expectedResult := []byte{0x81, 0xa4, 0x6e, 0x61, 0x6d, 0x65, 0xa3, 0x4d, 0x61, 0x74}

	packed, msgpackError := codec.Marshal(obj, nil)

	if msgpackError != nil {
		t.Errorf("Shouldn't return error: %s", msgpackError)
	}

	assert.Equal(t, packed, expectedResult)

}

func TestUnmarshal(t *testing.T) {

	codec := new(MsgpackCodec)
	packed := []byte{0x81, 0xa4, 0x6e, 0x61, 0x6d, 0x65, 0xa3, 0x4d, 0x61, 0x74}
	var object map[string]interface{}

	err := codec.Unmarshal(packed, &object)

	if err != nil {
		t.Errorf("Shouldn't return error: %s", err)
	}

	assert.Equal(t, []byte{0x4d, 0x61, 0x74}, object["name"])

}

func TestResponseContentType(t *testing.T) {

	codec := new(MsgpackCodec)
	assert.Equal(t, codec.ContentType(), constants.ContentTypeMsgpack)

}

func TestFileExtension(t *testing.T) {

	codec := new(MsgpackCodec)
	assert.Equal(t, constants.FileExtensionMsgpack, codec.FileExtension())

}

func TestCanMarshalWithCallback(t *testing.T) {

	codec := new(MsgpackCodec)
	assert.False(t, codec.CanMarshalWithCallback())

}
