// Package util provides testing utilities for Shock server tests
package util

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// We'll define interfaces for the components we need to mock
// This allows us to test without direct dependencies on the actual implementations

// User represents a user for testing purposes
type User struct {
	Uuid         string
	Username     string
	Fullname     string
	Email        string
	Admin        bool
	CustomFields map[string][]string
}

// Node represents a node for testing purposes
type Node struct {
	Id           string
	Version      string
	File         File
	Attributes   interface{}
	Acl          Acl
	CreatedOn    time.Time
	LastModified time.Time
}

// File represents a file for testing purposes
type File struct {
	Name      string
	Size      int64
	Checksum  map[string]string
	Format    string
	Path      string
	CreatedOn time.Time
}

// Acl represents access control for testing purposes
type Acl struct {
	Owner  string
	Read   []string
	Write  []string
	Delete []string
}

// Rights represents a set of access rights
type Rights map[string]bool

// TestUser creates a test user for testing
func TestUser() *User {
	return &User{
		Uuid:         "test_user",
		Username:     "test_user",
		Fullname:     "Test User",
		Email:        "test@example.com",
		Admin:        false,
		CustomFields: map[string][]string{},
	}
}

// TestAdminUser creates a test admin user for testing
func TestAdminUser() *User {
	return &User{
		Uuid:         "test_admin",
		Username:     "test_admin",
		Fullname:     "Test Admin",
		Email:        "admin@example.com",
		Admin:        true,
		CustomFields: map[string][]string{},
	}
}

// CreateTestNode creates a node with test data for testing
func CreateTestNode() *Node {
	return &Node{
		Id:      "test_node_id",
		Version: "1.0",
		File: File{
			Name:      "test_file.txt",
			Size:      12,
			Checksum:  map[string]string{"md5": "test_checksum"},
			Format:    "text",
			CreatedOn: time.Now(),
		},
		Acl: Acl{
			Owner:  "test_user",
			Read:   []string{"test_user"},
			Write:  []string{"test_user"},
			Delete: []string{"test_user"},
		},
		CreatedOn:    time.Now(),
		LastModified: time.Now(),
	}
}

// FormFile represents a form file for testing
type FormFile struct {
	Name     string
	Path     string
	Checksum map[string]string
}

// CreateTestFormFile creates a test form file for testing
func CreateTestFormFile(content string) (FormFile, error) {
	tempDir, err := ioutil.TempDir("", "shock-test")
	if err != nil {
		return FormFile{}, err
	}

	tempFile := filepath.Join(tempDir, "test_file.txt")
	err = ioutil.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		os.RemoveAll(tempDir)
		return FormFile{}, err
	}

	return FormFile{
		Name:     "test_file.txt",
		Path:     tempFile,
		Checksum: map[string]string{},
	}, nil
}

// InMemoryFile represents a file in memory for testing
type InMemoryFile struct {
	name     string
	content  []byte
	position int
	closed   bool
}

// Read implements io.Reader
func (f *InMemoryFile) Read(p []byte) (n int, err error) {
	if f.closed {
		return 0, os.ErrClosed
	}
	if f.position >= len(f.content) {
		return 0, io.EOF
	}
	n = copy(p, f.content[f.position:])
	f.position += n
	return
}

// ReadAt implements io.ReaderAt
func (f *InMemoryFile) ReadAt(p []byte, off int64) (n int, err error) {
	if f.closed {
		return 0, os.ErrClosed
	}
	if off >= int64(len(f.content)) {
		return 0, io.EOF
	}
	n = copy(p, f.content[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return
}

// Close implements io.Closer
func (f *InMemoryFile) Close() error {
	f.closed = true
	return nil
}

// Seek implements io.Seeker
func (f *InMemoryFile) Seek(offset int64, whence int) (int64, error) {
	if f.closed {
		return 0, os.ErrClosed
	}

	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = int64(f.position) + offset
	case io.SeekEnd:
		abs = int64(len(f.content)) + offset
	default:
		return 0, errors.New("invalid whence")
	}

	if abs < 0 {
		return 0, errors.New("negative position")
	}

	if abs > int64(len(f.content)) {
		f.position = len(f.content)
	} else {
		f.position = int(abs)
	}

	return abs, nil
}

// Stat implements the ReaderAt interface
func (f *InMemoryFile) Stat() (os.FileInfo, error) {
	return &InMemoryFileInfo{
		name:    f.name,
		size:    int64(len(f.content)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}, nil
}

// NewInMemoryFile creates a new in-memory file for testing
func NewInMemoryFile(name string, content []byte) *InMemoryFile {
	return &InMemoryFile{
		name:     name,
		content:  content,
		position: 0,
		closed:   false,
	}
}

// InMemoryFileInfo implements os.FileInfo for testing
type InMemoryFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *InMemoryFileInfo) Name() string       { return fi.name }
func (fi *InMemoryFileInfo) Size() int64        { return fi.size }
func (fi *InMemoryFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *InMemoryFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *InMemoryFileInfo) IsDir() bool        { return fi.isDir }
func (fi *InMemoryFileInfo) Sys() interface{}   { return nil }

// CleanupTestDir removes a test directory
func CleanupTestDir(path string) error {
	return os.RemoveAll(path)
}
