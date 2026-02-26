package responder

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

// TestRespondOK tests the RespondOK function produces valid JSON
func TestRespondOK(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := RespondOK(w, req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp standardResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.S)
	assert.Nil(t, resp.D)
	assert.Nil(t, resp.E)
}

// TestRespondWithData tests the RespondWithData function produces valid JSON
func TestRespondWithData(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := RespondWithData(w, req, "hello")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp standardResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.S)
	assert.Equal(t, "hello", resp.D)
	assert.Nil(t, resp.E)
}

// TestRespondWithError tests the RespondWithError function produces valid JSON
func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	err := RespondWithError(w, req, http.StatusBadRequest, "something went wrong")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp standardResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.S)
	assert.Nil(t, resp.D)
	assert.Equal(t, []string{"something went wrong"}, resp.E)
}
