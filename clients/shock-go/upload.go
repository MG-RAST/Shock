package shock

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

// createConfig holds all options for CreateNode/UpdateNode.
type createConfig struct {
	params map[string]string
	files  []fileEntry
}

type fileEntry struct {
	fieldName string
	filename  string
	path      string    // local file path (mutually exclusive with reader)
	reader    io.Reader // in-memory data (mutually exclusive with path)
}

// CreateOption configures a CreateNode/UpdateNode call.
type CreateOption func(*createConfig)

func buildConfig(opts []CreateOption) *createConfig {
	cfg := &createConfig{params: make(map[string]string)}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// applyCompression renames "upload" field to compression format if set.
func applyCompression(cfg *createConfig) {
	comp, ok := cfg.params["_compression"]
	if !ok {
		return
	}
	delete(cfg.params, "_compression")
	for i := range cfg.files {
		if cfg.files[i].fieldName == "upload" {
			cfg.files[i].fieldName = comp
		}
	}
}

// WithFile sets the upload file from an io.Reader.
func WithFile(r io.Reader, filename string) CreateOption {
	return func(cfg *createConfig) {
		cfg.files = append(cfg.files, fileEntry{
			fieldName: "upload",
			filename:  filename,
			reader:    r,
		})
	}
}

// WithFilePath sets the upload file from a local file path.
func WithFilePath(path string) CreateOption {
	return func(cfg *createConfig) {
		cfg.files = append(cfg.files, fileEntry{
			fieldName: "upload",
			filename:  filepath.Base(path),
			path:      path,
		})
	}
}

// WithFileContent sets the upload file from string content.
func WithFileContent(content string) CreateOption {
	return func(cfg *createConfig) {
		cfg.files = append(cfg.files, fileEntry{
			fieldName: "upload",
			filename:  "upload",
			reader:    bytes.NewBufferString(content),
		})
	}
}

// WithAttributes sets node attributes from a map (serialized as JSON).
func WithAttributes(attrs map[string]interface{}) CreateOption {
	return func(cfg *createConfig) {
		b, _ := json.Marshal(attrs)
		cfg.params["attributes_str"] = string(b)
	}
}

// WithAttributesFile sets node attributes from a JSON file.
// Ignored if WithAttributes was already used.
func WithAttributesFile(path string) CreateOption {
	return func(cfg *createConfig) {
		if _, ok := cfg.params["attributes_str"]; ok {
			return
		}
		if path == "" {
			return
		}
		cfg.files = append(cfg.files, fileEntry{
			fieldName: "attributes",
			filename:  filepath.Base(path),
			path:      path,
		})
	}
}

// WithExpiration sets the expiration for the node.
func WithExpiration(exp string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["expiration"] = exp
	}
}

// WithRemoveExpiration removes the expiration from the node.
func WithRemoveExpiration() CreateOption {
	return func(cfg *createConfig) {
		cfg.params["remove_expiration"] = "1"
	}
}

// WithFileName sets the filename for the node.
func WithFileName(name string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["file_name"] = name
	}
}

// WithCompression sets compression format ("gzip" or "bzip2").
func WithCompression(format string) CreateOption {
	return func(cfg *createConfig) {
		if format != "" {
			cfg.params["_compression"] = format
		}
	}
}

// WithParts initializes a parts node with the given number of parts.
func WithParts(count int) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["parts"] = strconv.Itoa(count)
	}
}

// WithPartsCompression sets the compression format for a parts node.
func WithPartsCompression(format string) CreateOption {
	return func(cfg *createConfig) {
		if format != "" {
			cfg.params["compression"] = format
		}
	}
}

// WithPart uploads a single part from an io.Reader.
func WithPart(partNum int, r io.Reader, filename string) CreateOption {
	return func(cfg *createConfig) {
		cfg.files = append(cfg.files, fileEntry{
			fieldName: strconv.Itoa(partNum),
			filename:  filename,
			reader:    r,
		})
	}
}

// WithPartFile uploads a single part from a file path.
func WithPartFile(partNum int, path string) CreateOption {
	return func(cfg *createConfig) {
		cfg.files = append(cfg.files, fileEntry{
			fieldName: strconv.Itoa(partNum),
			filename:  filepath.Base(path),
			path:      path,
		})
	}
}

// WithRemoteURL sets a remote URL for the node to fetch.
func WithRemoteURL(url string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["upload_url"] = url
	}
}

// WithVirtualParts creates a virtual node from the given node IDs.
func WithVirtualParts(nodeIDs []string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["type"] = "virtual"
		cfg.params["source"] = strings.Join(nodeIDs, ",")
	}
}

// WithCopyData copies data from the specified source node.
func WithCopyData(nodeID string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["copy_data"] = nodeID
	}
}

// WithCopyIndexes copies indexes from the source node.
func WithCopyIndexes() CreateOption {
	return func(cfg *createConfig) {
		cfg.params["copy_indexes"] = "1"
	}
}

// WithCopyAttributes copies attributes from the source node.
func WithCopyAttributes() CreateOption {
	return func(cfg *createConfig) {
		cfg.params["copy_attributes"] = "1"
	}
}

// WithSubset creates a subset node from a parent node's index.
func WithSubset(parentNodeID, parentIndex, subsetFile string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["parent_node"] = parentNodeID
		cfg.params["parent_index"] = parentIndex
		cfg.files = append(cfg.files, fileEntry{
			fieldName: "subset_indices",
			filename:  filepath.Base(subsetFile),
			path:      subsetFile,
		})
	}
}

// WithUnpackNode unpacks an archive node.
func WithUnpackNode(nodeID, archiveFormat string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["unpack_node"] = nodeID
		cfg.params["archive_format"] = archiveFormat
	}
}

// WithChecksumMD5 sets the expected MD5 checksum for verification.
func WithChecksumMD5(md5 string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["checksum-md5"] = md5
	}
}
