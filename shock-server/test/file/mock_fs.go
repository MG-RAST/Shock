// Package file provides mock file system implementations for testing
package file

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MockFileSystem is a mock implementation of the file system for testing
type MockFileSystem struct {
	files    map[string][]byte
	fileLock sync.RWMutex
}

// NewMockFileSystem creates a new mock file system for testing
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
	}
}

// AddFile adds a file to the mock file system
func (fs *MockFileSystem) AddFile(path string, content []byte) {
	fs.fileLock.Lock()
	defer fs.fileLock.Unlock()
	fs.files[path] = content
}

// Open opens a file from the mock file system
func (fs *MockFileSystem) Open(path string) (io.ReadCloser, error) {
	fs.fileLock.RLock()
	defer fs.fileLock.RUnlock()

	content, exists := fs.files[path]
	if !exists {
		return nil, os.ErrNotExist
	}

	return ioutil.NopCloser(bytes.NewReader(content)), nil
}

// Create creates a new file in the mock file system
func (fs *MockFileSystem) Create(path string) (io.WriteCloser, error) {
	return &mockWriteCloser{
		path: path,
		fs:   fs,
		buf:  new(bytes.Buffer),
	}, nil
}

// Remove removes a file from the mock file system
func (fs *MockFileSystem) Remove(path string) error {
	fs.fileLock.Lock()
	defer fs.fileLock.Unlock()

	if _, exists := fs.files[path]; !exists {
		return os.ErrNotExist
	}

	delete(fs.files, path)
	return nil
}

// Stat returns file info for a file in the mock file system
func (fs *MockFileSystem) Stat(path string) (os.FileInfo, error) {
	fs.fileLock.RLock()
	defer fs.fileLock.RUnlock()

	content, exists := fs.files[path]
	if !exists {
		return nil, os.ErrNotExist
	}

	return &mockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(content)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}, nil
}

// MkdirAll creates a directory and all parent directories in the mock file system
func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	// For simplicity, we'll just record that the directory exists
	// by creating an empty file at that path
	fs.fileLock.Lock()
	defer fs.fileLock.Unlock()

	fs.files[path] = []byte{}
	return nil
}

// ReadFile reads a file from the mock file system
func (fs *MockFileSystem) ReadFile(path string) ([]byte, error) {
	fs.fileLock.RLock()
	defer fs.fileLock.RUnlock()

	content, exists := fs.files[path]
	if !exists {
		return nil, os.ErrNotExist
	}

	return content, nil
}

// WriteFile writes a file to the mock file system
func (fs *MockFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	fs.fileLock.Lock()
	defer fs.fileLock.Unlock()

	fs.files[path] = data
	return nil
}

// mockWriteCloser implements io.WriteCloser for the mock file system
type mockWriteCloser struct {
	path string
	fs   *MockFileSystem
	buf  *bytes.Buffer
}

func (mwc *mockWriteCloser) Write(p []byte) (n int, err error) {
	return mwc.buf.Write(p)
}

func (mwc *mockWriteCloser) Close() error {
	mwc.fs.fileLock.Lock()
	defer mwc.fs.fileLock.Unlock()

	mwc.fs.files[mwc.path] = mwc.buf.Bytes()
	return nil
}

// mockFileInfo implements os.FileInfo for the mock file system
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (mfi *mockFileInfo) Name() string       { return mfi.name }
func (mfi *mockFileInfo) Size() int64        { return mfi.size }
func (mfi *mockFileInfo) Mode() os.FileMode  { return mfi.mode }
func (mfi *mockFileInfo) ModTime() time.Time { return mfi.modTime }
func (mfi *mockFileInfo) IsDir() bool        { return mfi.isDir }
func (mfi *mockFileInfo) Sys() interface{}   { return nil }

// Reset clears all files in the mock file system
func (fs *MockFileSystem) Reset() {
	fs.fileLock.Lock()
	defer fs.fileLock.Unlock()

	fs.files = make(map[string][]byte)
}
