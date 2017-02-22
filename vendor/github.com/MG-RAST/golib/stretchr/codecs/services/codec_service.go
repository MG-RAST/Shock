package services

import (
	"github.com/MG-RAST/golib/stretchr/codecs"
)

// CodecService is the interface for a service responsible for providing Codecs.
type CodecService interface {
	// GetCodecForResponding gets the codec to use to respond based on the
	// given accept string, the extension provided and whether it has a callback
	// or not.
	GetCodecForResponding(accept, extension string, hasCallback bool) (codecs.Codec, error)

	// GetCodec gets the codec to use to interpret the request based on the
	// content type.
	GetCodec(contentType string) (codecs.Codec, error)

	// MarshalWithCodec marshals the specified object with the specified codec and options.
	// If the object implements the Facade interface, the PublicData object should be
	// marshalled instead.
	MarshalWithCodec(codec codecs.Codec, object interface{}, options map[string]interface{}) ([]byte, error)

	// UnmarshalWithCodec unmarshals the specified data into the object with the specified codec.
	UnmarshalWithCodec(codec codecs.Codec, data []byte, object interface{}) error

	// Codecs gets all currently installed codecs.
	Codecs() []codecs.Codec

	// AddCodec adds the specified codec to the installed codecs list.
	AddCodec(codecs.Codec)
}
