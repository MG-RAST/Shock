package shock

import (
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type downloadConfig struct {
	uncompress string
	computeMD5 bool
}

// DownloadOption configures a DownloadToFile call.
type DownloadOption func(*downloadConfig)

// WithUncompress sets the decompression method for downloads (e.g. "gzip").
func WithUncompress(method string) DownloadOption {
	return func(cfg *downloadConfig) {
		cfg.uncompress = method
	}
}

// WithComputeMD5 enables MD5 checksum computation during download.
func WithComputeMD5() DownloadOption {
	return func(cfg *downloadConfig) {
		cfg.computeMD5 = true
	}
}

// Download streams a node's file data. The caller must close the returned ReadCloser.
func (c *Client) Download(ctx context.Context, nodeID string) (io.ReadCloser, error) {
	query := url.Values{"download": {""}}
	u := c.buildURL("/node/"+nodeID, query)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)
	c.debugf("GET %s (download)", u)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusOK {
		return res.Body, nil
	}

	// Error response — parse JSON envelope
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var envelope struct {
		Error []string `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Error) > 0 {
		return nil, &APIError{StatusCode: res.StatusCode, Errors: envelope.Error}
	}
	return nil, &APIError{StatusCode: res.StatusCode, Errors: []string{string(body)}}
}

// DownloadRange streams a byte range from a node's file.
func (c *Client) DownloadRange(ctx context.Context, nodeID string, seek, length int64) (io.ReadCloser, error) {
	query := url.Values{
		"download": {""},
		"seek":     {strconv.FormatInt(seek, 10)},
		"length":   {strconv.FormatInt(length, 10)},
	}
	u := c.buildURL("/node/"+nodeID, query)
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusOK {
		return res.Body, nil
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	var envelope struct {
		Error []string `json:"error"`
	}
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Error) > 0 {
		return nil, &APIError{StatusCode: res.StatusCode, Errors: envelope.Error}
	}
	return nil, &APIError{StatusCode: res.StatusCode, Errors: []string{string(body)}}
}

// DownloadToFile downloads a node's file to a local path, optionally decompressing and computing MD5.
func (c *Client) DownloadToFile(ctx context.Context, nodeID, filepath string, opts ...DownloadOption) (int64, string, error) {
	cfg := &downloadConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	body, err := c.Download(ctx, nodeID)
	if err != nil {
		return 0, "", err
	}
	defer body.Close()

	localFile, err := os.Create(filepath)
	if err != nil {
		return 0, "", err
	}
	defer localFile.Close()

	md5h := md5.New()
	var size int64

	if cfg.uncompress == "gzip" {
		var src io.Reader = body
		if cfg.computeMD5 {
			// MD5 on compressed stream, then decompress
			pReader, pWriter := io.Pipe()
			defer pReader.Close()
			dst := io.MultiWriter(pWriter, md5h)
			go func() {
				io.Copy(dst, body)
				pWriter.Close()
			}()
			src = pReader
		}
		gr, gerr := gzip.NewReader(src)
		if gerr != nil {
			return 0, "", gerr
		}
		defer gr.Close()
		size, err = io.Copy(localFile, gr)
		if err != nil {
			return 0, "", err
		}
	} else {
		var dst io.Writer
		if cfg.computeMD5 {
			dst = io.MultiWriter(localFile, md5h)
		} else {
			dst = localFile
		}
		size, err = io.Copy(dst, body)
		if err != nil {
			return 0, "", err
		}
	}

	var md5sum string
	if cfg.computeMD5 {
		md5sum = fmt.Sprintf("%x", md5h.Sum(nil))
	}
	return size, md5sum, nil
}

// GetDownloadURL gets a pre-authenticated download URL for a node.
func (c *Client) GetDownloadURL(ctx context.Context, nodeID string) (*PreAuthResponse, error) {
	query := url.Values{"download_url": {""}}
	var resp PreAuthResponse
	if err := c.doRequest(ctx, "GET", "/node/"+nodeID, query, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DownloadURL returns the direct download URL for a node (no pre-auth).
func (c *Client) DownloadURL(nodeID string) string {
	query := url.Values{"download": {""}}
	return c.buildURL("/node/"+nodeID, query)
}
