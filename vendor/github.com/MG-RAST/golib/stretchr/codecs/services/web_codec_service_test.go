package services

import (
	"fmt"
	"github.com/MG-RAST/golib/stretchr/codecs"
	"github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/golib/stretchr/codecs/json"
	"github.com/MG-RAST/golib/stretchr/codecs/test"
	"github.com/MG-RAST/golib/stretchr/objx"
	"github.com/MG-RAST/golib/stretchr/testify/assert"
	"github.com/MG-RAST/golib/stretchr/testify/mock"
	"strings"
	"testing"
)

/*
	Test code
*/

func TestInterface(t *testing.T) {
	assert.Implements(t, (*CodecService)(nil), NewWebCodecService(), "WebCodecService")
}

func TestNewWebCodecService_DefaultCodecs(t *testing.T) {
	n := NewWebCodecService()
	assert.Equal(t, len(DefaultCodecs), len(n.codecs))
}

func TestAddCodec(t *testing.T) {

	service := NewWebCodecService()

	service.codecs = make([]codecs.Codec, 0)
	jsonCodec := new(json.JsonCodec)

	service.AddCodec(jsonCodec)

	if assert.Equal(t, 1, len(service.Codecs())) {
		assert.Equal(t, jsonCodec, service.codecs[0])
	}

}

func TestGetCodec(t *testing.T) {

	service := NewWebCodecService()
	var codec codecs.Codec

	codec, _ = service.GetCodec(constants.ContentTypeJSON)

	if assert.NotNil(t, codec, "Json should exist") {
		assert.Equal(t, constants.ContentTypeJSON, codec.ContentType(), "ContentTypeJson")
	}

	// case insensitivity
	codec, _ = service.GetCodec(strings.ToUpper(constants.ContentTypeJSON))

	if assert.NotNil(t, codec, "Content case should not matter") {
		assert.Equal(t, constants.ContentTypeJSON, codec.ContentType(), "ContentTypeJson")
	}

	// with noise
	codec, _ = service.GetCodec(fmt.Sprintf("%s; charset=UTF-8", constants.ContentTypeJSON))
	if assert.NotNil(t, codec, "charset in Content-Type should not matter") {
		assert.Equal(t, constants.ContentTypeJSON, codec.ContentType(), "ContentTypeJson")
	}

	// default
	codec, _ = service.GetCodec("")

	if assert.NotNil(t, codec, "Empty contentType string should assume JSON") {
		assert.Equal(t, constants.ContentTypeJSON, codec.ContentType(), "Should assume JSON.")
	}

}

func TestGetCodecForResponding_DefaultCodec(t *testing.T) {

	service := NewWebCodecService()
	var codec codecs.Codec

	codec, _ = service.GetCodecForResponding("", "", false)

	if assert.NotNil(t, codec, "Return of GetCodecForAcceptStringOrExtension should default to JSON") {
		assert.Equal(t, constants.ContentTypeJSON, codec.ContentType(), "Should default to JSON")
	}

}

func TestGetCodecForResponding(t *testing.T) {

	service := NewWebCodecService()
	var codec codecs.Codec

	// JSON - accept header

	codec, _ = service.GetCodecForResponding("something/something,application/json,text/xml", "", false)

	if assert.NotNil(t, codec, "Return of GetCodecForAcceptStringOrExtension") {
		assert.Equal(t, constants.ContentTypeJSON, codec.ContentType(), "ContentTypeJson 1")
	}

	// JSON - accept header (case)

	codec, _ = service.GetCodecForResponding("something/something,application/JSON,text/xml", "", false)

	if assert.NotNil(t, codec, "Case should not matter") {
		assert.Equal(t, constants.ContentTypeJSON, codec.ContentType(), "Case should not matter")
	}

	// JSON - file extension

	codec, _ = service.GetCodecForResponding("", constants.FileExtensionJSON, false)

	if assert.NotNil(t, codec, "Return of GetCodecForAcceptStringOrExtension") {
		assert.Equal(t, constants.ContentTypeJSON, codec.ContentType(), "ContentTypeJson")
	}

	// JSONP - has callback

	codec, _ = service.GetCodecForResponding("", "", true)

	if assert.NotNil(t, codec, "Should return the first codec that can handle a callback") {
		assert.Equal(t, constants.ContentTypeJSONP, codec.ContentType(), "ContentTypeJavaScript")
	}

	// JSONP - file extension

	codec, _ = service.GetCodecForResponding("", constants.FileExtensionJSONP, false)

	if assert.NotNil(t, codec, "Return of GetCodecForAcceptStringOrExtension") {
		assert.Equal(t, constants.ContentTypeJSONP, codec.ContentType(), "ContentTypeJavaScript")
	}

	// JSONP - file extension (case)

	codec, _ = service.GetCodecForResponding("", strings.ToUpper(constants.FileExtensionJSONP), false)

	if assert.NotNil(t, codec, "Return of GetCodecForAcceptStringOrExtension") {
		assert.Equal(t, constants.ContentTypeJSONP, codec.ContentType(), "ContentTypeJavaScript 4")
	}

	// JSONP - Accept header

	codec, _ = service.GetCodecForResponding("something/something,text/javascript,text/xml", "", false)

	if assert.NotNil(t, codec, "Return of GetCodecForAcceptStringOrExtension") {
		assert.Equal(t, constants.ContentTypeJSONP, codec.ContentType(), "ContentTypeJavaScript 5")
	}

	// hasCallback takes precedence over everything else

	codec, _ = service.GetCodecForResponding(constants.ContentTypeJSON, constants.FileExtensionXML, true)

	if assert.NotNil(t, codec, "Return of GetCodecForAcceptStringOrExtension") {
		assert.Equal(t, constants.ContentTypeJSONP, codec.ContentType(), "HasCallback takes precedence over all")
	}

	// File extension takes precedence over accept header

	codec, _ = service.GetCodecForResponding(constants.ContentTypeJSON, constants.FileExtensionXML, false)

	if assert.NotNil(t, codec, "Return of GetCodecForAcceptStringOrExtension") {
		assert.Equal(t, constants.ContentTypeXML, codec.ContentType(), "Extension takes precedence over accept")
	}

}

func TestMarshalWithCodec(t *testing.T) {

	testCodec := new(test.TestCodec)
	service := NewWebCodecService()

	// make some test stuff
	var bytesToReturn []byte = []byte("Hello World")
	var object objx.Map = objx.MSI("Name", "Mat")
	var option1 string = "Option One"
	var option2 string = "Option Two"

	args := map[string]interface{}{option1: option1, option2: option2}

	// setup expectations
	testCodec.On("Marshal", object, args).Return(bytesToReturn, nil)

	bytes, err := service.MarshalWithCodec(testCodec, object, args)

	if assert.Nil(t, err) {
		assert.Equal(t, string(bytesToReturn), string(bytes))
	}

	// assert that our expectations were met
	mock.AssertExpectationsForObjects(t, testCodec.Mock)

}

func TestMarshalWithCodec_WithFacade(t *testing.T) {

	// func (s *WebCodecService) MarshalWithCodec(codec codecs.Codec, object interface{}, options ...interface{}) ([]byte, error) {

	testCodec := new(test.TestCodec)
	service := NewWebCodecService()

	// make some test stuff
	var bytesToReturn []byte = []byte("Hello World")
	testObjectWithFacade := new(test.TestObjectWithFacade)
	object := objx.MSI("Name", "Mat")
	var option1 string = "Option One"
	var option2 string = "Option Two"

	args := map[string]interface{}{option1: option1, option2: option2}

	// setup expectations
	testObjectWithFacade.On("PublicData", args).Return(object, nil)
	testCodec.On("Marshal", object, args).Return(bytesToReturn, nil)

	bytes, err := service.MarshalWithCodec(testCodec, testObjectWithFacade, args)

	if assert.Nil(t, err) {
		assert.Equal(t, string(bytesToReturn), string(bytes))
	}

	// assert that our expectations were met
	mock.AssertExpectationsForObjects(t, testCodec.Mock, testObjectWithFacade.Mock)

}

func TestMarshalWithCodec_WithFacade_AndError(t *testing.T) {

	// func (s *WebCodecService) MarshalWithCodec(codec codecs.Codec, object interface{}, options ...interface{}) ([]byte, error) {

	testCodec := new(test.TestCodec)
	service := NewWebCodecService()

	// make some test stuff
	testObjectWithFacade := new(test.TestObjectWithFacade)
	var option1 string = "Option One"
	var option2 string = "Option Two"

	args := map[string]interface{}{option1: option1, option2: option2}

	// setup expectations
	testObjectWithFacade.On("PublicData", args).Return(nil, assert.AnError)

	_, err := service.MarshalWithCodec(testCodec, testObjectWithFacade, args)

	assert.Equal(t, assert.AnError, err)

}

func TestMarshalWithCodec_WithError(t *testing.T) {

	// func (s *WebCodecService) MarshalWithCodec(codec codecs.Codec, object interface{}, options ...interface{}) ([]byte, error) {

	testCodec := new(test.TestCodec)
	service := NewWebCodecService()

	// make some test stuff
	object := objx.MSI("Name", "Mat")
	var option1 string = "Option One"
	var option2 string = "Option Two"

	args := map[string]interface{}{option1: option1, option2: option2}

	// setup expectations
	testCodec.On("Marshal", object, args).Return(nil, assert.AnError)

	_, err := service.MarshalWithCodec(testCodec, object, args)

	assert.Equal(t, assert.AnError, err, "The error should get returned")

	// assert that our expectations were met
	mock.AssertExpectationsForObjects(t, testCodec.Mock)

}

func TestUnmarshalWithCodec(t *testing.T) {

	// func (s *WebCodecService) UnmarshalWithCodec(codec codecs.Codec, data []byte, object interface{}) error {

	testCodec := new(test.TestCodec)
	service := NewWebCodecService()

	// some test objects
	object := struct{}{}
	data := []byte("Some bytes")

	// setup expectations
	testCodec.On("Unmarshal", data, object).Return(nil)

	// call the target method
	err := service.UnmarshalWithCodec(testCodec, data, object)

	assert.Nil(t, err)
	mock.AssertExpectationsForObjects(t, testCodec.Mock)

}

func TestUnmarshalWithCodec_WithError(t *testing.T) {

	// func (s *WebCodecService) UnmarshalWithCodec(codec codecs.Codec, data []byte, object interface{}) error {

	testCodec := new(test.TestCodec)
	service := NewWebCodecService()

	// some test objects
	object := struct{}{}
	data := []byte("Some bytes")

	// setup expectations
	testCodec.On("Unmarshal", data, object).Return(assert.AnError)

	// call the target method
	err := service.UnmarshalWithCodec(testCodec, data, object)

	assert.Equal(t, assert.AnError, err)
	mock.AssertExpectationsForObjects(t, testCodec.Mock)

}
