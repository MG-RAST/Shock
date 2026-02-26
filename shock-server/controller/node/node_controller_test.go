package node_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/controller/node"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// setupTestRouter creates a Chi router with the node controller routes
func setupTestRouter() *chi.Mux {
	nc := &node.NodeController{}
	r := chi.NewRouter()
	r.Route("/node", func(nr chi.Router) {
		nr.Get("/", nc.ReadMany)
		nr.Post("/", nc.Create)
		nr.Put("/", nc.UpdateMany)
		nr.Delete("/", nc.DeleteMany)
		nr.Options("/", nc.Options)
		nr.Get("/{id}", nc.Read)
		nr.Put("/{id}", nc.Replace)
		nr.Delete("/{id}", nc.Delete)
		nr.Options("/{id}", nc.Options)
	})
	return r
}

// TestNodeControllerOptions tests the Options method returns 200
func TestNodeControllerOptions(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	req, err := http.NewRequest("OPTIONS", server.URL+"/node", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := ioutil.ReadAll(resp.Body)
	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["status"])
}

// TestNodeControllerUpdateMany tests the UpdateMany method returns 501
func TestNodeControllerUpdateMany(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	req, err := http.NewRequest("PUT", server.URL+"/node", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
}

// TestNodeControllerDeleteMany tests the DeleteMany method returns 501
func TestNodeControllerDeleteMany(t *testing.T) {
	router := setupTestRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	req, err := http.NewRequest("DELETE", server.URL+"/node", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
}
