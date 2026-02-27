package shock

import (
	"context"
	"fmt"
	"time"
)

var (
	DefaultPollInterval = 30 * time.Second
	DefaultWaitTimeout  = 1 * time.Hour
)

type waitConfig struct {
	interval time.Duration
	timeout  time.Duration
}

// WaitOption configures WaitFile/WaitIndex.
type WaitOption func(*waitConfig)

// WithPollInterval sets the polling interval.
func WithPollInterval(d time.Duration) WaitOption {
	return func(cfg *waitConfig) {
		cfg.interval = d
	}
}

// WithWaitTimeout sets the maximum wait time.
func WithWaitTimeout(d time.Duration) WaitOption {
	return func(cfg *waitConfig) {
		cfg.timeout = d
	}
}

// WaitFile polls until a node's file is no longer locked.
func (c *Client) WaitFile(ctx context.Context, nodeID string, opts ...WaitOption) (*Node, error) {
	cfg := waitConfig{interval: DefaultPollInterval, timeout: DefaultWaitTimeout}
	for _, opt := range opts {
		opt(&cfg)
	}

	startTime := time.Now()
	for {
		node, err := c.GetNode(ctx, nodeID)
		if err != nil {
			return nil, err
		}
		if node.File.Locked == nil {
			return node, nil
		}
		if node.File.Locked.Error != "" {
			return nil, fmt.Errorf("file lock error on node %s: %s", nodeID, node.File.Locked.Error)
		}
		if time.Since(startTime) > cfg.timeout {
			return nil, fmt.Errorf("timeout waiting on file lock: node=%s", nodeID)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(cfg.interval):
			// continue polling
		}
	}
}

// WaitIndex polls until a node's index is no longer locked.
func (c *Client) WaitIndex(ctx context.Context, nodeID, indexName string, opts ...WaitOption) (*IdxInfo, error) {
	if indexName == "" {
		return nil, nil
	}

	cfg := waitConfig{interval: DefaultPollInterval, timeout: DefaultWaitTimeout}
	for _, opt := range opts {
		opt(&cfg)
	}

	startTime := time.Now()
	for {
		node, err := c.GetNode(ctx, nodeID)
		if err != nil {
			return nil, err
		}
		idx, has := node.Indexes[indexName]
		if !has {
			return nil, fmt.Errorf("index does not exist: node=%s, index=%s", nodeID, indexName)
		}
		if idx.Locked != nil && idx.Locked.Error != "" {
			return nil, fmt.Errorf("index error: node=%s, index=%s, error=%s", nodeID, indexName, idx.Locked.Error)
		}
		if idx.Locked == nil {
			return idx, nil
		}
		if time.Since(startTime) > cfg.timeout {
			return nil, fmt.Errorf("timeout waiting on index lock: node=%s, index=%s", nodeID, indexName)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(cfg.interval):
			// continue polling
		}
	}
}
