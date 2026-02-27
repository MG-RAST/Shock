package shock

import (
	"context"
	"net/url"
)

// GetNode retrieves a node by ID.
func (c *Client) GetNode(ctx context.Context, id string) (*Node, error) {
	var node Node
	if err := c.doRequest(ctx, "GET", "/node/"+id, nil, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// DeleteNode deletes a node by ID.
func (c *Client) DeleteNode(ctx context.Context, id string) error {
	return c.doRequest(ctx, "DELETE", "/node/"+id, nil, nil)
}

// CreateNode creates a new node with the given options.
func (c *Client) CreateNode(ctx context.Context, opts ...CreateOption) (*Node, error) {
	cfg := buildConfig(opts)
	applyCompression(cfg)

	body, err := c.doMultipart(ctx, "POST", "/node", cfg)
	if err != nil {
		return nil, err
	}

	var node Node
	if err := c.checkErrors(body, 200, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// UpdateNode updates an existing node with the given options.
func (c *Client) UpdateNode(ctx context.Context, id string, opts ...CreateOption) (*Node, error) {
	cfg := buildConfig(opts)
	applyCompression(cfg)

	body, err := c.doMultipart(ctx, "PUT", "/node/"+id, cfg)
	if err != nil {
		return nil, err
	}

	var node Node
	if err := c.checkErrors(body, 200, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// UpdateAttributes updates the attributes of an existing node.
func (c *Client) UpdateAttributes(ctx context.Context, id string, attrFile string, nodeattr map[string]interface{}) error {
	var opts []CreateOption
	if attrFile != "" {
		opts = append(opts, WithAttributesFile(attrFile))
	}
	if len(nodeattr) > 0 {
		opts = append(opts, WithAttributes(nodeattr))
	}
	_, err := c.UpdateNode(ctx, id, opts...)
	return err
}

// PostFileWithAttributes creates a node with a file and attributes.
func (c *Client) PostFileWithAttributes(ctx context.Context, filepath, filename string, nodeattr map[string]interface{}) (*Node, error) {
	opts := []CreateOption{WithFilePath(filepath)}
	if filename != "" {
		opts = append(opts, WithFileName(filename))
	}
	if len(nodeattr) > 0 {
		opts = append(opts, WithAttributes(nodeattr))
	}
	return c.CreateNode(ctx, opts...)
}

// PutOrPostFile provides backward compatibility with the old client.
// If nodeid is empty, creates a new node (POST); otherwise updates (PUT).
func (c *Client) PutOrPostFile(ctx context.Context, filename, nodeid string, rank int, attrfile, ntype string, formopts map[string]string, nodeattr map[string]interface{}) (string, error) {
	var opts []CreateOption

	// Attributes
	if attrfile != "" && rank < 2 {
		opts = append(opts, WithAttributesFile(attrfile))
	}
	if len(nodeattr) > 0 {
		opts = append(opts, WithAttributes(nodeattr))
	}

	// Form options
	if v, ok := formopts["file_name"]; ok {
		opts = append(opts, WithFileName(v))
	}
	if v, ok := formopts["expiration"]; ok {
		opts = append(opts, WithExpiration(v))
	}
	if _, ok := formopts["remove_expiration"]; ok {
		opts = append(opts, WithRemoveExpiration())
	}

	compression := formopts["compression"]

	if rank > 0 && filename == "" {
		// Parts node initialization
		opts = append(opts, WithParts(rank))
		if compression != "" {
			opts = append(opts, WithPartsCompression(compression))
		}
		// Also handle file_name from parts option
		if v, ok := formopts["parts"]; ok {
			n, _ := url.QueryUnescape(v)
			_ = n
		}
	} else if rank > 0 && filename != "" {
		// Part upload
		opts = append(opts, WithPartFile(rank, filename))
	} else {
		// Handle different node types
		switch ntype {
		case "virtual":
			if v, ok := formopts["virtual_file"]; ok {
				opts = append(opts, WithVirtualParts(splitComma(v)))
			}
		case "remote":
			if v, ok := formopts["remote_url"]; ok {
				opts = append(opts, WithRemoteURL(v))
			}
		case "copy":
			if v, ok := formopts["parent_node"]; ok {
				opts = append(opts, WithCopyData(v))
			}
			if _, ok := formopts["copy_indexes"]; ok {
				opts = append(opts, WithCopyIndexes())
			}
			if _, ok := formopts["copy_attributes"]; ok {
				opts = append(opts, WithCopyAttributes())
			}
		case "parts":
			if v, ok := formopts["parts"]; ok {
				opts = append(opts, withPartsString(v))
			}
			if compression != "" {
				opts = append(opts, WithPartsCompression(compression))
			}
		default:
			if filename != "" {
				opts = append(opts, WithFilePath(filename))
				if compression != "" {
					opts = append(opts, WithCompression(compression))
				}
			}
		}
	}

	var node *Node
	var err error
	if nodeid != "" {
		node, err = c.UpdateNode(ctx, nodeid, opts...)
	} else {
		node, err = c.CreateNode(ctx, opts...)
	}
	if err != nil {
		return "", err
	}
	if node != nil {
		return node.Id, nil
	}
	return "", nil
}

func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	parts := make([]string, 0)
	for _, p := range splitStr(s, ',') {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitStr(s string, sep byte) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func withPartsString(s string) CreateOption {
	return func(cfg *createConfig) {
		cfg.params["parts"] = s
	}
}
