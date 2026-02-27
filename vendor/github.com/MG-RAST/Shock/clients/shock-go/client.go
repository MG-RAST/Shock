package shock

import (
	"net/http"
	"time"
)

// Client is a Shock API client.
type Client struct {
	baseURL    string
	token      string
	authType   string
	httpClient *http.Client
	debug      bool
}

// Option configures a Client.
type Option func(*Client)

// New creates a new Shock client for the given base URL.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:  baseURL,
		authType: "OAuth",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithToken sets the authentication token.
func WithToken(token string) Option {
	return func(c *Client) {
		c.token = token
	}
}

// WithAuthType sets the authorization type (e.g. "mgrast", "OAuth").
func WithAuthType(authType string) Option {
	return func(c *Client) {
		c.authType = authType
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithDebug enables debug logging to stderr.
func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.debug = debug
	}
}

// SetToken updates the authentication token.
func (c *Client) SetToken(token string) {
	c.token = token
}
