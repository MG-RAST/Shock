package router_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/conf"
	ncon "github.com/MG-RAST/Shock/shock-server/controller/node"
	"github.com/MG-RAST/Shock/shock-server/logger"
	"github.com/MG-RAST/Shock/shock-server/responder"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	conf.LOG_OUTPUT = "console"
	conf.PATH_LOGS = "/tmp"
	logger.Initialize()
}

// corsMiddleware matches main.go's CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// setupRouter creates a router matching the production setup for the endpoints we test
func setupRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(corsMiddleware)

	// Root endpoint
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		res := map[string]interface{}{
			"id":        "Shock",
			"type":      "Shock",
			"resources": []string{"node"},
		}
		responder.WriteResponseObject(w, r, http.StatusOK, res)
	})

	// Node CRUD
	nc := &ncon.NodeController{}
	router.Route("/node", func(nr chi.Router) {
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

	// Catch-all
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		responder.RespondWithError(w, r, http.StatusBadRequest, "Parameters do not match a valid Shock request type.")
	})

	return router
}

func TestCORSHeaders(t *testing.T) {
	server := httptest.NewServer(setupRouter())
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST, GET, PUT, DELETE, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Authorization", resp.Header.Get("Access-Control-Allow-Headers"))
	// Note: "Connection: close" is a hop-by-hop header removed by Go's HTTP transport,
	// so we verify it's set by testing with httptest.ResponseRecorder instead.

}

func TestConnectionCloseHeader(t *testing.T) {
	// Use ResponseRecorder to verify the Connection header is set by middleware
	// (hop-by-hop headers are stripped by Go's HTTP client, so we test directly)
	router := setupRouter()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	router.ServeHTTP(rr, req)

	assert.Equal(t, "close", rr.Header().Get("Connection"))
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestOptionsPreflightReturns200(t *testing.T) {
	server := httptest.NewServer(setupRouter())
	defer server.Close()

	tests := []struct {
		path string
	}{
		{"/node"},
		{"/node/abc"},
	}

	client := &http.Client{}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req, err := http.NewRequest("OPTIONS", server.URL+tt.path, nil)
			if !assert.NoError(t, err) {
				return
			}

			resp, err := client.Do(req)
			if !assert.NoError(t, err) {
				return
			}
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestNotFoundHandler(t *testing.T) {
	server := httptest.NewServer(setupRouter())
	defer server.Close()

	resp, err := http.Get(server.URL + "/invalid/path")
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, float64(http.StatusBadRequest), response["status"])
}

func TestNodeOptionsEndpoint(t *testing.T) {
	server := httptest.NewServer(setupRouter())
	defer server.Close()

	client := &http.Client{}
	req, err := http.NewRequest("OPTIONS", server.URL+"/node", nil)
	if !assert.NoError(t, err) {
		return
	}

	resp, err := client.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	// The CORS middleware intercepts OPTIONS and returns 200 with CORS headers
	// before the route handler runs, so there is no JSON body.
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "POST, GET, PUT, DELETE, OPTIONS", resp.Header.Get("Access-Control-Allow-Methods"))
}

func TestUpdateManyNotImplemented(t *testing.T) {
	server := httptest.NewServer(setupRouter())
	defer server.Close()

	client := &http.Client{}
	req, err := http.NewRequest("PUT", server.URL+"/node", nil)
	if !assert.NoError(t, err) {
		return
	}

	resp, err := client.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
}

func TestDeleteManyNotImplemented(t *testing.T) {
	server := httptest.NewServer(setupRouter())
	defer server.Close()

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", server.URL+"/node", nil)
	if !assert.NoError(t, err) {
		return
	}

	resp, err := client.Do(req)
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
}

func TestRootEndpoint(t *testing.T) {
	server := httptest.NewServer(setupRouter())
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if !assert.NoError(t, err) {
		return
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if !assert.NoError(t, err) {
		return
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "Shock", response["id"])
	assert.Equal(t, "Shock", response["type"])

	resources, ok := response["resources"].([]interface{})
	if !assert.True(t, ok, "resources should be an array") {
		return
	}
	assert.Contains(t, resources, "node")
}
