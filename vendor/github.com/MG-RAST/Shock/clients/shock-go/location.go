package shock

import "context"

// GetLocations retrieves storage locations for a node.
func (c *Client) GetLocations(ctx context.Context, nodeID string) ([]Location, error) {
	var locs []Location
	if err := c.doRequest(ctx, "GET", "/node/"+nodeID+"/locations/", nil, &locs); err != nil {
		return nil, err
	}
	return locs, nil
}

// AddLocation adds a storage location to a node.
func (c *Client) AddLocation(ctx context.Context, nodeID string, loc Location) ([]Location, error) {
	var locs []Location
	if err := c.doJSON(ctx, "POST", "/node/"+nodeID+"/locations/", loc, &locs); err != nil {
		return nil, err
	}
	return locs, nil
}

// DeleteLocation removes a storage location from a node.
func (c *Client) DeleteLocation(ctx context.Context, nodeID, locationID string) ([]Location, error) {
	var locs []Location
	if err := c.doRequest(ctx, "DELETE", "/node/"+nodeID+"/locations/"+locationID, nil, &locs); err != nil {
		return nil, err
	}
	return locs, nil
}

// DeleteAllLocations removes all storage locations from a node.
func (c *Client) DeleteAllLocations(ctx context.Context, nodeID string) error {
	return c.doRequest(ctx, "DELETE", "/node/"+nodeID+"/locations/", nil, nil)
}
