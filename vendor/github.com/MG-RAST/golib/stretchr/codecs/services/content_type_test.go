package services

import (
	"github.com/MG-RAST/golib/stretchr/testify/assert"
	"testing"
)

func TestParseContentType_NoParams(t *testing.T) {
	contentTypeString := "application/json"
	contentType, err := ParseContentType(contentTypeString)
	assert.NoError(t, err)
	assert.Equal(t, contentTypeString, contentType.MimeType)
}

func TestParseContentType_WithParams(t *testing.T) {
	contentTypeString := "application/xml; q=0.7; test=hello"
	expectedMimeType := "application/xml"
	expectedParams := map[string]string{
		"q":    "0.7",
		"test": "hello",
	}
	contentType, err := ParseContentType(contentTypeString)
	assert.NoError(t, err)
	assert.Equal(t, expectedMimeType, contentType.MimeType)
	for index, value := range expectedParams {
		parsedValue, ok := contentType.Parameters[index]
		assert.True(t, ok, "All expected values should have been parsed as params")
		assert.Equal(t, parsedValue, value)
	}
}
