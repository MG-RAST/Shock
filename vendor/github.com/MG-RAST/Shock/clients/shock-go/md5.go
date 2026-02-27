package shock

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
)

// GetMD5FromFile computes the MD5 checksum of a file.
func GetMD5FromFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// GetNodeByMD5 searches for an existing node with the given MD5 checksum.
// Returns the node ID, whether a match was found, and any error.
func (c *Client) GetNodeByMD5(ctx context.Context, md5sum string) (string, bool, error) {
	filters := map[string]string{
		"file.checksum.md5": md5sum,
		"type":              "basic",
	}
	sr, err := c.QueryFull(ctx, filters)
	if err != nil {
		return "", false, err
	}
	if len(sr.Nodes) > 0 {
		c.debugf("GetNodeByMD5: found existing node: %s", sr.Nodes[0].Id)
		return sr.Nodes[0].Id, true, nil
	}
	c.debugf("GetNodeByMD5: file not found in Shock")
	return "", false, nil
}

// PostFileLazy creates a node with a file, but first checks if a node with
// the same MD5 already exists. If so, returns the existing node ID.
// It caches MD5 checksums in .md5 files alongside the source file.
func (c *Client) PostFileLazy(ctx context.Context, filepath, filename string) (string, error) {
	md5Filename := filepath + ".md5"
	var md5sum string

	c.debugf("PostFileLazy: filepath=%s (md5Filename=%s)", filepath, md5Filename)

	if _, err := os.Stat(md5Filename); err != nil {
		// .md5 file not found, calculate
		c.debugf("PostFileLazy: calculating md5sum...")
		var calcErr error
		md5sum, calcErr = GetMD5FromFile(filepath)
		if calcErr != nil {
			return "", fmt.Errorf("GetMD5FromFile: %w", calcErr)
		}
		c.debugf("PostFileLazy: calculated md5sum: %s", md5sum)

		baseName := path.Base(filepath)
		md5sumFileContent := md5sum + " " + baseName
		if writeErr := os.WriteFile(md5Filename, []byte(md5sumFileContent), 0600); writeErr != nil {
			c.debugf("PostFileLazy: could not write md5sum: %s, continuing", writeErr.Error())
		}
	} else {
		// .md5 file exists
		md5sumBytes, readErr := os.ReadFile(md5Filename)
		if readErr != nil {
			return "", fmt.Errorf("could not read md5sum file: %w", readErr)
		}
		if len(md5sumBytes) >= 32 {
			md5sum = string(md5sumBytes[0:32])
		}
		c.debugf("PostFileLazy: got cached md5sum: %s", md5sum)
	}

	if md5sum != "" {
		nodeid, ok, err := c.GetNodeByMD5(ctx, md5sum)
		if err != nil {
			c.debugf("PostFileLazy: GetNodeByMD5 error: %s", err.Error())
		} else if ok {
			return nodeid, nil
		}
	}

	// File not found in Shock, upload it
	var opts []CreateOption
	if filepath != "" {
		opts = append(opts, WithFilePath(filepath))
	}
	if filename != "" {
		opts = append(opts, WithFileName(filename))
	}
	node, err := c.CreateNode(ctx, opts...)
	if err != nil {
		return "", err
	}
	if node != nil {
		return node.Id, nil
	}
	return "", nil
}
