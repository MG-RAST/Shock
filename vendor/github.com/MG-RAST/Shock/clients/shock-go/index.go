package shock

import (
	"context"
	"net/url"
	"strconv"
)

// PutIndex creates or rebuilds an index on a node.
func (c *Client) PutIndex(ctx context.Context, nodeID, indexName string) error {
	if indexName == "" {
		return nil
	}
	return c.doRequest(ctx, "PUT", "/node/"+nodeID+"/index/"+indexName, nil, nil)
}

// PutIndexQuery creates or rebuilds an index with options.
func (c *Client) PutIndexQuery(ctx context.Context, nodeID, indexName string, force bool, column int) error {
	if indexName == "" {
		return nil
	}
	query := url.Values{}
	if force {
		query.Set("force_rebuild", "1")
	}
	if column > 0 {
		query.Set("number", strconv.Itoa(column))
	}
	return c.doRequest(ctx, "PUT", "/node/"+nodeID+"/index/"+indexName, query, nil)
}
