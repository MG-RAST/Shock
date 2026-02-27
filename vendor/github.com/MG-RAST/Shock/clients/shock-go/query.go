package shock

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// QueryResult holds the result of a node query.
type QueryResult struct {
	Nodes      []Node
	Limit      int
	Offset     int
	TotalCount int
}

type queryConfig struct {
	limit     int
	offset    int
	order     string
	direction string
}

// QueryOption configures a Query call.
type QueryOption func(*queryConfig)

// WithLimit sets the maximum number of results.
func WithLimit(n int) QueryOption {
	return func(cfg *queryConfig) {
		cfg.limit = n
	}
}

// WithOffset sets the starting offset.
func WithOffset(n int) QueryOption {
	return func(cfg *queryConfig) {
		cfg.offset = n
	}
}

// WithOrder sets the field to sort by.
func WithOrder(field string) QueryOption {
	return func(cfg *queryConfig) {
		cfg.order = field
	}
}

// WithDirection sets the sort direction ("asc" or "desc").
func WithDirection(dir string) QueryOption {
	return func(cfg *queryConfig) {
		cfg.direction = dir
	}
}

func buildQueryValues(filters map[string]string, qcfg *queryConfig, queryType string) url.Values {
	query := url.Values{}
	query.Set(queryType, "")
	for k, v := range filters {
		query.Set(k, v)
	}
	if qcfg.limit > 0 {
		query.Set("limit", intToStr(qcfg.limit))
	}
	if qcfg.offset > 0 {
		query.Set("offset", intToStr(qcfg.offset))
	}
	if qcfg.order != "" {
		query.Set("order", qcfg.order)
	}
	if qcfg.direction != "" {
		query.Set("direction", qcfg.direction)
	}
	return query
}

func intToStr(n int) string {
	return url.QueryEscape(json.Number(itoa(n)).String())
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func (c *Client) nodeQuery(ctx context.Context, query url.Values) (*QueryResult, error) {
	var resp nodesResponse
	u := c.buildURL("/node", query)
	req, err := c.rawRequest(ctx, "GET", u)
	if err != nil {
		return nil, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}
	if len(resp.Error) > 0 {
		return nil, &APIError{StatusCode: resp.Status, Errors: resp.Error}
	}
	return &QueryResult{
		Nodes:      resp.Data,
		Limit:      resp.Limit,
		Offset:     resp.Offset,
		TotalCount: resp.TotalCount,
	}, nil
}

func (c *Client) rawRequest(ctx context.Context, method, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	return req, nil
}

// Query searches for nodes with basic field matching.
func (c *Client) Query(ctx context.Context, filters map[string]string, opts ...QueryOption) (*QueryResult, error) {
	qcfg := &queryConfig{}
	for _, opt := range opts {
		opt(qcfg)
	}
	query := buildQueryValues(filters, qcfg, "query")
	return c.nodeQuery(ctx, query)
}

// QueryFull searches for nodes with full querynode matching.
func (c *Client) QueryFull(ctx context.Context, filters map[string]string, opts ...QueryOption) (*QueryResult, error) {
	qcfg := &queryConfig{}
	for _, opt := range opts {
		opt(qcfg)
	}
	query := buildQueryValues(filters, qcfg, "querynode")
	return c.nodeQuery(ctx, query)
}

// QueryDistinct returns distinct values for a field.
func (c *Client) QueryDistinct(ctx context.Context, field string, filters map[string]string) (interface{}, error) {
	query := url.Values{}
	query.Set("distinct", field)
	for k, v := range filters {
		query.Set(k, v)
	}
	var result interface{}
	if err := c.doRequest(ctx, "GET", "/node", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// QueryRaw performs a query with pre-built url.Values (backward compatibility).
func (c *Client) QueryRaw(ctx context.Context, query url.Values) (*QueryResult, error) {
	return c.nodeQuery(ctx, query)
}

// QueryDistinctRaw performs a distinct query with pre-built url.Values (backward compatibility).
func (c *Client) QueryDistinctRaw(ctx context.Context, query url.Values) (interface{}, error) {
	var result interface{}
	if err := c.doRequest(ctx, "GET", "/node", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}
