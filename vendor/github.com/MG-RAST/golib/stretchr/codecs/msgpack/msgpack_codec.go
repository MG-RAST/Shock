package msgpack

import (
	"bytes"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/ugorji/go/codec"
)

// MsgpackCodec converts objects to and from Msgpack.
type MsgpackCodec struct{}

var msgpackHandle codec.MsgpackHandle

// Converts an object to Msgpack.
func (c *MsgpackCodec) Marshal(object interface{}, options map[string]interface{}) ([]byte, error) {

	byteBuffer := new(bytes.Buffer)
	enc := codec.NewEncoder(byteBuffer, &msgpackHandle)
	encErr := enc.Encode(object)

	return byteBuffer.Bytes(), encErr
}

// Unmarshal converts Msgpack into an object.
func (c *MsgpackCodec) Unmarshal(data []byte, obj interface{}) error {

	dec := codec.NewDecoder(bytes.NewReader(data), &msgpackHandle)
	return dec.Decode(&obj)
}

// ContentType returns the content type for this codec.
func (c *MsgpackCodec) ContentType() string {
	return constants.ContentTypeMsgpack
}

// FileExtension returns the file extension for this codec.
func (c *MsgpackCodec) FileExtension() string {
	return constants.FileExtensionMsgpack
}

// CanMarshalWithCallback returns whether this codec is capable of marshalling a response containing a callback.
func (c *MsgpackCodec) CanMarshalWithCallback() bool {
	return false
}
