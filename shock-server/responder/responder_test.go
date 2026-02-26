package responder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStandardResponseStruct tests the standardResponse struct field assignments
func TestStandardResponseStruct(t *testing.T) {
	r := standardResponse{
		S: 200,
		D: "test data",
		E: []string{"no error"},
	}

	assert.Equal(t, 200, r.S)
	assert.Equal(t, "test data", r.D)
	assert.Equal(t, []string{"no error"}, r.E)
}

// TestStandardResponseZeroValue tests the zero value of standardResponse
func TestStandardResponseZeroValue(t *testing.T) {
	var r standardResponse
	assert.Equal(t, 0, r.S)
	assert.Nil(t, r.D)
	assert.Nil(t, r.E)
}

// TestPaginatedResponseStruct tests the paginatedResponse struct field assignments
func TestPaginatedResponseStruct(t *testing.T) {
	r := paginatedResponse{
		S:      200,
		D:      []string{"item1", "item2"},
		E:      nil,
		Limit:  10,
		Offset: 0,
		Count:  2,
	}

	assert.Equal(t, 200, r.S)
	assert.Equal(t, []string{"item1", "item2"}, r.D)
	assert.Nil(t, r.E)
	assert.Equal(t, 10, r.Limit)
	assert.Equal(t, 0, r.Offset)
	assert.Equal(t, 2, r.Count)
}

// TestPaginatedResponseZeroValue tests the zero value of paginatedResponse
func TestPaginatedResponseZeroValue(t *testing.T) {
	var r paginatedResponse
	assert.Equal(t, 0, r.S)
	assert.Nil(t, r.D)
	assert.Nil(t, r.E)
	assert.Equal(t, 0, r.Limit)
	assert.Equal(t, 0, r.Offset)
	assert.Equal(t, 0, r.Count)
}

// TestStandardResponseWithNilData tests standardResponse with nil data field
func TestStandardResponseWithNilData(t *testing.T) {
	r := standardResponse{
		S: 404,
		D: nil,
		E: []string{"not found"},
	}

	assert.Equal(t, 404, r.S)
	assert.Nil(t, r.D)
	assert.Equal(t, []string{"not found"}, r.E)
}

// TestStandardResponseWithMultipleErrors tests standardResponse with multiple errors
func TestStandardResponseWithMultipleErrors(t *testing.T) {
	r := standardResponse{
		S: 400,
		D: nil,
		E: []string{"error1", "error2", "error3"},
	}

	assert.Len(t, r.E, 3)
	assert.Contains(t, r.E, "error1")
	assert.Contains(t, r.E, "error2")
	assert.Contains(t, r.E, "error3")
}

// TestGetJsonCodec tests that getJsonCodec returns a non-nil codec service
func TestGetJsonCodec(t *testing.T) {
	codec := getJsonCodec()
	assert.NotNil(t, codec)
}
