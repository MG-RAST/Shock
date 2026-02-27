package shock

import (
	"context"
	"net/url"
)

// GetAcl retrieves the ACL for a node.
func (c *Client) GetAcl(ctx context.Context, nodeID string) (*DisplayAcl, error) {
	var acl DisplayAcl
	if err := c.doRequest(ctx, "GET", "/node/"+nodeID+"/acl", nil, &acl); err != nil {
		return nil, err
	}
	return &acl, nil
}

// PutAcl adds users to an ACL type on a node.
func (c *Client) PutAcl(ctx context.Context, nodeID, aclType, users string) (*DisplayAcl, error) {
	query := url.Values{}
	if users != "" {
		query.Set("users", users)
	}
	var acl DisplayAcl
	if err := c.doRequest(ctx, "PUT", "/node/"+nodeID+"/acl/"+aclType, query, &acl); err != nil {
		return nil, err
	}
	return &acl, nil
}

// DeleteAcl removes users from an ACL type on a node.
func (c *Client) DeleteAcl(ctx context.Context, nodeID, aclType, users string) (*DisplayAcl, error) {
	query := url.Values{}
	if users != "" {
		query.Set("users", users)
	}
	var acl DisplayAcl
	if err := c.doRequest(ctx, "DELETE", "/node/"+nodeID+"/acl/"+aclType, query, &acl); err != nil {
		return nil, err
	}
	return &acl, nil
}

// MakePublic sets a node's read ACL to public.
func (c *Client) MakePublic(ctx context.Context, nodeID string) (*DisplayAcl, error) {
	var acl DisplayAcl
	if err := c.doRequest(ctx, "PUT", "/node/"+nodeID+"/acl/public_read", nil, &acl); err != nil {
		return nil, err
	}
	return &acl, nil
}

// ChownNode changes the owner of a node.
func (c *Client) ChownNode(ctx context.Context, nodeID, user string) (*DisplayAcl, error) {
	query := url.Values{}
	query.Set("users", user)
	var acl DisplayAcl
	if err := c.doRequest(ctx, "PUT", "/node/"+nodeID+"/acl/owner", query, &acl); err != nil {
		return nil, err
	}
	return &acl, nil
}
