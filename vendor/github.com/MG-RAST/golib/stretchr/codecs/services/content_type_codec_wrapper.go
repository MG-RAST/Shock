package services

import (
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs"
)

// contentTypeCodecWrapper is a wrapper for a Codec.  It is used to
// return any given Codec value, but with an overridden ContentType()
// value, usually for the purposes of returning the ContentType that
// was requested in an Accept header.
type contentTypeCodecWrapper struct {
	codec       codecs.Codec
	contentType string
}

// wrapCodecWithContentType takes a codecs.Codec and a mime type
// string, returning a codecs.Codec that acts exactly like the passed
// in codecs.Codec except that the ContentType() method will return
// the passed in mime type instead of the underlying codec's default
// mime type.
func wrapCodecWithContentType(c codecs.Codec, typeString string) codecs.Codec {
	return &contentTypeCodecWrapper{
		codec:       c,
		contentType: typeString,
	}
}

func (c *contentTypeCodecWrapper) Marshal(object interface{}, options map[string]interface{}) ([]byte, error) {
	// Pass the matched content type as a codec option
	if options == nil {
		options = make(map[string]interface{})
	}
	options["matched_type"] = c.contentType
	return c.codec.Marshal(object, options)
}

func (c *contentTypeCodecWrapper) Unmarshal(data []byte, obj interface{}) error {
	return c.codec.Unmarshal(data, obj)
}

func (c *contentTypeCodecWrapper) ContentType() string {
	return c.contentType
}

func (c *contentTypeCodecWrapper) FileExtension() string {
	return c.codec.FileExtension()
}

func (c *contentTypeCodecWrapper) CanMarshalWithCallback() bool {
	return c.codec.CanMarshalWithCallback()
}
