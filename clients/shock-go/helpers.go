package shock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

// doRequest performs a JSON request with no body and parses the envelope response.
func (c *Client) doRequest(ctx context.Context, method, path string, query url.Values, result interface{}) error {
	u := c.buildURL(path, query)
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return err
	}
	c.setAuth(req)
	c.debugf("%s %s", method, u)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return c.checkErrors(body, res.StatusCode, result)
}

// doJSON sends a JSON body request and parses the envelope response.
func (c *Client) doJSON(ctx context.Context, method, path string, payload interface{}, result interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	u := c.buildURL(path, nil)
	req, err := http.NewRequestWithContext(ctx, method, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)
	c.debugf("%s %s (json body)", method, u)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return c.checkErrors(body, res.StatusCode, result)
}

// doMultipart builds a multipart form and sends it. Returns raw response body.
func (c *Client) doMultipart(ctx context.Context, method, urlPath string, cfg *createConfig) ([]byte, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Write form in a goroutine to stream data without buffering
	errCh := make(chan error, 1)
	go func() {
		var writeErr error
		defer func() {
			// Close writer first to finalize the multipart boundary
			writer.Close()
			if writeErr != nil {
				pw.CloseWithError(writeErr)
			} else {
				pw.Close()
			}
			errCh <- writeErr
		}()

		// Write string params
		for k, v := range cfg.params {
			if err := writer.WriteField(k, v); err != nil {
				writeErr = err
				return
			}
		}

		// Write file fields
		for _, f := range cfg.files {
			part, err := writer.CreateFormFile(f.fieldName, f.filename)
			if err != nil {
				writeErr = err
				return
			}
			if f.reader != nil {
				if _, err := io.Copy(part, f.reader); err != nil {
					writeErr = err
					return
				}
			} else if f.path != "" {
				fh, err := os.Open(f.path)
				if err != nil {
					writeErr = err
					return
				}
				_, err = io.Copy(part, fh)
				fh.Close()
				if err != nil {
					writeErr = err
					return
				}
			}
		}
	}()

	u := c.buildURL(urlPath, nil)
	req, err := http.NewRequestWithContext(ctx, method, u, pr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.setAuth(req)
	c.debugf("%s %s (multipart)", method, u)

	res, err := c.httpClient.Do(req)
	if err != nil {
		// Also check goroutine error
		if wErr := <-errCh; wErr != nil {
			return nil, fmt.Errorf("multipart write error: %w; http error: %w", wErr, err)
		}
		return nil, err
	}
	defer res.Body.Close()

	// Wait for goroutine to finish
	<-errCh

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// checkErrors parses the envelope response and extracts .Data into result.
func (c *Client) checkErrors(body []byte, statusCode int, result interface{}) error {
	var envelope struct {
		Status int             `json:"status"`
		Data   json.RawMessage `json:"data"`
		Error  []string        `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return &APIError{StatusCode: statusCode, Errors: []string{string(body)}}
	}
	if len(envelope.Error) > 0 {
		return &APIError{StatusCode: envelope.Status, Errors: envelope.Error}
	}
	if result != nil && len(envelope.Data) > 0 {
		return json.Unmarshal(envelope.Data, result)
	}
	return nil
}

// setAuth adds the Authorization header if a token is set.
func (c *Client) setAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", c.authType+" "+c.token)
	}
}

// buildURL constructs a full URL from base + path + query params.
func (c *Client) buildURL(path string, query url.Values) string {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		// Fallback: concatenate directly
		s := c.baseURL + path
		if query != nil && len(query) > 0 {
			s += "?" + query.Encode()
		}
		return s
	}
	u.Path = path
	if query != nil && len(query) > 0 {
		u.RawQuery = query.Encode()
	}
	return u.String()
}

// debugf prints debug output to stderr when debug mode is enabled.
func (c *Client) debugf(format string, args ...interface{}) {
	if c.debug {
		fmt.Fprintf(os.Stderr, "[shock-client] "+format+"\n", args...)
	}
}
