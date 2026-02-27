package shock

import (
	"context"
	"path/filepath"
)

// UnpackArchiveNode unpacks an archive node into multiple nodes.
func (c *Client) UnpackArchiveNode(ctx context.Context, nodeID, format, attrFile string) ([]Node, error) {
	cfg := &createConfig{params: make(map[string]string)}
	cfg.params["unpack_node"] = nodeID
	cfg.params["archive_format"] = format
	if attrFile != "" {
		cfg.files = append(cfg.files, fileEntry{
			fieldName: "attributes",
			filename:  filepath.Base(attrFile),
			path:      attrFile,
		})
	}

	body, err := c.doMultipart(ctx, "POST", "/node", cfg)
	if err != nil {
		return nil, err
	}

	var nodes []Node
	if err := c.checkErrors(body, 200, &nodes); err != nil {
		return nil, err
	}
	return nodes, nil
}
