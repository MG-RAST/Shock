package preauth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPreAuthResponseStruct tests the PreAuthResponse struct field assignments
func TestPreAuthResponseStruct(t *testing.T) {
	response := PreAuthResponse{
		Url:       "http://example.com/preauth/123",
		ValidTill: "2023-12-31",
		Format:    "fasta",
		Filename:  "test.fasta",
		Files:     2,
		Size:      1024,
	}

	assert.Equal(t, "http://example.com/preauth/123", response.Url)
	assert.Equal(t, "2023-12-31", response.ValidTill)
	assert.Equal(t, "fasta", response.Format)
	assert.Equal(t, "test.fasta", response.Filename)
	assert.Equal(t, 2, response.Files)
	assert.Equal(t, int64(1024), response.Size)
}

// TestPreAuthStruct tests the PreAuth struct field assignments
func TestPreAuthStruct(t *testing.T) {
	now := time.Now()
	p := PreAuth{
		Id:        "test_id",
		Type:      "download",
		Nodes:     []string{"node1", "node2"},
		Options:   map[string]string{"format": "fasta"},
		ValidTill: now,
	}

	assert.Equal(t, "test_id", p.Id)
	assert.Equal(t, "download", p.Type)
	assert.Equal(t, []string{"node1", "node2"}, p.Nodes)
	assert.Equal(t, map[string]string{"format": "fasta"}, p.Options)
	assert.Equal(t, now, p.ValidTill)
}

// TestPreAuthZeroValue tests the zero value of PreAuth
func TestPreAuthZeroValue(t *testing.T) {
	var p PreAuth
	assert.Empty(t, p.Id)
	assert.Empty(t, p.Type)
	assert.Nil(t, p.Nodes)
	assert.Nil(t, p.Options)
	assert.True(t, p.ValidTill.IsZero())
}

// TestPreAuthResponseZeroValue tests the zero value of PreAuthResponse
func TestPreAuthResponseZeroValue(t *testing.T) {
	var r PreAuthResponse
	assert.Empty(t, r.Url)
	assert.Empty(t, r.ValidTill)
	assert.Empty(t, r.Format)
	assert.Empty(t, r.Filename)
	assert.Equal(t, 0, r.Files)
	assert.Equal(t, int64(0), r.Size)
}
