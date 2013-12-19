package responder

import (
	"github.com/stretchr/goweb"
	"github.com/stretchr/goweb/context"
	"net/http"
)

// The standard API response object
type standardResponse struct {
	S int         `json:"status"`
	D interface{} `json:"data"`
	E []string    `json:"error"`
}

// The standard API response object
type paginatedResponse struct {
	S      int         `json:"status"`
	D      interface{} `json:"data"`
	E      []string    `json:"error"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Count  int         `json:"total_count"`
}

func RespondOK(ctx context.Context) error {
	response := new(standardResponse)
	response.S = http.StatusOK
	response.D = nil
	response.E = nil
	return goweb.API.WriteResponseObject(ctx, http.StatusOK, response)
}

func WriteResponseObject(ctx context.Context, status int, responseObject interface{}) error {
	return goweb.API.WriteResponseObject(ctx, status, responseObject)
}

func RespondWithData(ctx context.Context, data interface{}) error {
	response := new(standardResponse)
	response.S = http.StatusOK
	response.D = data
	response.E = nil
	return goweb.API.WriteResponseObject(ctx, http.StatusOK, response)
}

func RespondWithError(ctx context.Context, status int, err string) error {
	response := new(standardResponse)
	response.S = status
	response.D = nil
	response.E = append(response.E, err)
	return goweb.API.WriteResponseObject(ctx, http.StatusOK, response)
}

func RespondWithPaginatedData(ctx context.Context, data interface{}, limit, offset, count int) error {
	// make the standard response object
	response := new(paginatedResponse)
	response.S = http.StatusOK
	response.D = data
	response.E = nil
	response.Limit = limit
	response.Offset = offset
	response.Count = count

	return goweb.API.WriteResponseObject(ctx, http.StatusOK, response)
}
