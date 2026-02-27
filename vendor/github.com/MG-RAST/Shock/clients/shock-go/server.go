package shock

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ServerInfo retrieves server information from the root endpoint.
// The root endpoint returns a bare struct, not envelope-wrapped.
func (c *Client) ServerInfo(ctx context.Context) (*ServerInfo, error) {
	u := c.buildURL("/", nil)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	c.debugf("GET %s (server info)", u)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var info ServerInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode server info: %w", err)
	}
	return &info, nil
}
