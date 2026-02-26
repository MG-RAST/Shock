package responder

import (
	"encoding/json"
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

func RespondOK(w http.ResponseWriter, r *http.Request) error {
	response := new(standardResponse)
	response.S = http.StatusOK
	response.D = nil
	response.E = nil
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func WriteResponseObject(w http.ResponseWriter, r *http.Request, status int, responseObject interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(responseObject)
}

func RespondWithData(w http.ResponseWriter, r *http.Request, data interface{}) error {
	response := new(standardResponse)
	response.S = http.StatusOK
	response.D = data
	response.E = nil
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func RespondWithError(w http.ResponseWriter, r *http.Request, status int, err string) error {
	response := new(standardResponse)
	response.S = status
	response.D = nil
	response.E = append(response.E, err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(response)
}

func RespondWithPaginatedData(w http.ResponseWriter, r *http.Request, data interface{}, limit, offset, count int) error {
	response := new(paginatedResponse)
	response.S = http.StatusOK
	response.D = data
	response.E = nil
	response.Limit = limit
	response.Offset = offset
	response.Count = count
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}
