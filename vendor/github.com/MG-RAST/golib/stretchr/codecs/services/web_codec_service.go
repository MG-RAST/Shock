package services

import (
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/bson"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/csv"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/json"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/jsonp"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/msgpack"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/xml"
	"strings"
)

type ContentTypeNotSupportedError struct {
	ContentType string
}

func (e *ContentTypeNotSupportedError) Error() string {
	return "Content type " + e.ContentType + " is not supported."
}

// DefaultCodecs represents the list of Codecs that get added automatically by
// a call to NewWebCodecService.
var DefaultCodecs = []codecs.Codec{new(json.JsonCodec), new(jsonp.JsonPCodec), new(msgpack.MsgpackCodec), new(bson.BsonCodec), new(csv.CsvCodec), new(xml.SimpleXmlCodec)}

// WebCodecService represents the default implementation for providing access to the
// currently installed web codecs.
type WebCodecService struct {
	// codecs holds the installed codecs for this service.
	codecs []codecs.Codec
}

// NewWebCodecService makes a new WebCodecService with the default codecs
// added.
func NewWebCodecService() *WebCodecService {
	s := new(WebCodecService)
	s.codecs = DefaultCodecs
	return s
}

// Codecs gets all currently installed codecs.
func (s *WebCodecService) Codecs() []codecs.Codec {
	return s.codecs
}

// AddCodec adds the specified codec to the installed codecs list.
func (s *WebCodecService) AddCodec(codec codecs.Codec) {
	s.codecs = append(s.codecs, codec)
}

func (s *WebCodecService) assertCodecs() {
	if len(s.codecs) == 0 {
		panic("codecs: No codecs are installed - use AddCodec to add some or use NewWebCodecService for default codecs.")
	}
}

// GetCodecForResponding gets the codec to use to respond based on the
// given accept string, the extension provided and whether it has a callback
// or not.
//
// As of now, if hasCallback is true, the JSONP codec will be returned.
// This may be changed if additional callback capable codecs are added.
func (s *WebCodecService) GetCodecForResponding(accept, extension string, hasCallback bool) (codecs.Codec, error) {

	// make sure we have at least one codec
	s.assertCodecs()

	if hasCallback {
		for _, codec := range s.codecs {
			if codec.CanMarshalWithCallback() {
				return codec, nil
			}
		}
	}

	if extension != "" {
		for _, codec := range s.codecs {
			if strings.ToLower(codec.FileExtension()) == strings.ToLower(extension) {
				return codec, nil
			}
		}
	}

	if accept != "" {
		// Try the simple case first
		if !(strings.ContainsRune(accept, ',') || strings.ContainsRune(accept, ';')) {
			accept = strings.TrimSpace(accept)
			codec, err := s.getCodecByMimeString(accept)
			if codec != nil {
				return codec, err
			}
		}

		// If this doesn't match the simple case or simple matching
		// failed to find a matching codec, do a full header parse
		orderedAccept, err := OrderAcceptHeader(accept)
		if err != nil {
			return nil, err
		}
		for _, entry := range orderedAccept {
			if codec, err := s.getCodecByContentType(entry.ContentType); err == nil {
				return codec, nil
			}
		}
	}

	// return the first installed codec by default
	return s.codecs[0], nil
}

// GetCodec gets the codec to use to interpret the request based on the
// content type.
func (s *WebCodecService) GetCodec(contentType string) (codecs.Codec, error) {

	// make sure we have at least one codec
	s.assertCodecs()

	parsedContentType, err := ParseContentType(contentType)
	if err != nil {
		return nil, err
	}

	return s.getCodecByContentType(parsedContentType)
}

// getCodecByMimeString is a helper method to retrieve a codec that
// can handle the passed in mime type string.
func (s *WebCodecService) getCodecByMimeString(mime string) (codecs.Codec, error) {

	for _, codec := range s.codecs {

		// default codec
		if mime == "" && codec.ContentType() == constants.ContentTypeJSON {
			return codec, nil
		}

		// match the content type
		if matcher, ok := codec.(codecs.ContentTypeMatcherCodec); ok {
			if matcher.ContentTypeSupported(mime) {
				// For codecs.ContentTypeMatcherCodec values, the
				// matched content type could be different than the
				// codec's ContentType return value.  The
				// wrapCodecWithContentType function will override the
				// default return value of ContentType() with the
				// matched content type.
				return wrapCodecWithContentType(codec, mime), nil
			}
		} else if mime == strings.ToLower(codec.ContentType()) {
			return codec, nil
		}

	}

	return nil, &ContentTypeNotSupportedError{mime}

}

// getCodecByContentType is a helper method to retrieve a codec that
// can handle the passed in *ContentType value.
func (s *WebCodecService) getCodecByContentType(contentType *ContentType) (codecs.Codec, error) {
	if contentType == nil {
		return s.getCodecByMimeString("")
	}
	return s.getCodecByMimeString(contentType.MimeType)
}

// MarshalWithCodec marshals the specified object with the specified codec and options.
// If the object implements the Facade interface, the PublicData object should be
// marshalled instead.
func (s *WebCodecService) MarshalWithCodec(codec codecs.Codec, object interface{}, options map[string]interface{}) ([]byte, error) {

	// make sure we have at least one codec
	s.assertCodecs()

	// get the public data
	publicData, err := codecs.PublicData(object, options)

	// if there was an error - return it
	if err != nil {
		return nil, err
	}

	// let the codec do its work
	return codec.Marshal(publicData, options)
}

// UnmarshalWithCodec unmarshals the specified data into the object with the specified codec.
func (s *WebCodecService) UnmarshalWithCodec(codec codecs.Codec, data []byte, object interface{}) error {

	// make sure we have at least one codec
	s.assertCodecs()

	return codec.Unmarshal(data, object)
}
