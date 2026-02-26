package seq_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"github.com/stretchr/testify/assert"
)

// TestNew tests creating a new sequence
func TestNew(t *testing.T) {
	// Test creating a sequence with all fields
	id := []byte("seq1")
	sequence := []byte("ACGT")
	quality := []byte("IIII")

	s := seq.New(id, sequence, quality)

	assert.Equal(t, id, s.ID, "Sequence ID should match")
	assert.Equal(t, sequence, s.Seq, "Sequence data should match")
	assert.Equal(t, quality, s.Qual, "Sequence quality should match")

	// Test creating a sequence without quality
	s = seq.New(id, sequence, nil)

	assert.Equal(t, id, s.ID, "Sequence ID should match")
	assert.Equal(t, sequence, s.Seq, "Sequence data should match")
	assert.Nil(t, s.Qual, "Sequence quality should be nil")
}

// TestReadFormater tests the ReadFormater interface
func TestReadFormater(t *testing.T) {
	// Create a mock implementation of ReadFormater
	mockReader := &mockReadFormater{
		readFunc: func() (*seq.Seq, error) {
			return seq.New([]byte("seq1"), []byte("ACGT"), []byte("IIII")), nil
		},
		formatFunc: func(s *seq.Seq, w io.Writer) (int, error) {
			return w.Write([]byte(">seq1\nACGT\n"))
		},
	}

	// Test Read method
	s, err := mockReader.Read()
	assert.NoError(t, err, "Reading sequence should not error")
	assert.Equal(t, []byte("seq1"), s.ID, "Sequence ID should match")
	assert.Equal(t, []byte("ACGT"), s.Seq, "Sequence data should match")
	assert.Equal(t, []byte("IIII"), s.Qual, "Sequence quality should match")

	// Test Format method
	var buf bytes.Buffer
	n, err := mockReader.Format(s, &buf)
	assert.NoError(t, err, "Formatting sequence should not error")
	assert.Equal(t, 11, n, "Number of bytes written should match")
	assert.Equal(t, ">seq1\nACGT\n", buf.String(), "Formatted output should match")
}

// TestReader tests the Reader interface
func TestReader(t *testing.T) {
	// Create a mock implementation of Reader
	mockReader := &mockReader{
		readFunc: func() (*seq.Seq, error) {
			return seq.New([]byte("seq1"), []byte("ACGT"), []byte("IIII")), nil
		},
		getReadOffsetFunc: func() (int, error) {
			return 10, nil
		},
		seekChunkFunc: func(offset int64, lastIndex bool) (int64, error) {
			return 20, nil
		},
	}

	// Test Read method
	s, err := mockReader.Read()
	assert.NoError(t, err, "Reading sequence should not error")
	assert.Equal(t, []byte("seq1"), s.ID, "Sequence ID should match")
	assert.Equal(t, []byte("ACGT"), s.Seq, "Sequence data should match")
	assert.Equal(t, []byte("IIII"), s.Qual, "Sequence quality should match")

	// Test GetReadOffset method
	offset, err := mockReader.GetReadOffset()
	assert.NoError(t, err, "Getting read offset should not error")
	assert.Equal(t, 10, offset, "Read offset should match")

	// Test SeekChunk method
	pos, err := mockReader.SeekChunk(5, false)
	assert.NoError(t, err, "Seeking chunk should not error")
	assert.Equal(t, int64(20), pos, "Seek position should match")
}

// TestReadRewinder tests the ReadRewinder interface
func TestReadRewinder(t *testing.T) {
	// Create a mock implementation of ReadRewinder
	mockRewinder := &mockReadRewinder{
		readFunc: func() (*seq.Seq, error) {
			return seq.New([]byte("seq1"), []byte("ACGT"), []byte("IIII")), nil
		},
		getReadOffsetFunc: func() (int, error) {
			return 10, nil
		},
		seekChunkFunc: func(offset int64, lastIndex bool) (int64, error) {
			return 20, nil
		},
		rewindFunc: func() error {
			return nil
		},
	}

	// Test Read method
	s, err := mockRewinder.Read()
	assert.NoError(t, err, "Reading sequence should not error")
	assert.Equal(t, []byte("seq1"), s.ID, "Sequence ID should match")
	assert.Equal(t, []byte("ACGT"), s.Seq, "Sequence data should match")
	assert.Equal(t, []byte("IIII"), s.Qual, "Sequence quality should match")

	// Test GetReadOffset method
	offset, err := mockRewinder.GetReadOffset()
	assert.NoError(t, err, "Getting read offset should not error")
	assert.Equal(t, 10, offset, "Read offset should match")

	// Test SeekChunk method
	pos, err := mockRewinder.SeekChunk(5, false)
	assert.NoError(t, err, "Seeking chunk should not error")
	assert.Equal(t, int64(20), pos, "Seek position should match")

	// Test Rewind method
	err = mockRewinder.Rewind()
	assert.NoError(t, err, "Rewinding should not error")
}

// TestErrorHandling tests error handling in the interfaces
func TestErrorHandling(t *testing.T) {
	// Create a mock implementation that returns errors
	mockWithErrors := &mockReadRewinder{
		readFunc: func() (*seq.Seq, error) {
			return nil, io.EOF
		},
		getReadOffsetFunc: func() (int, error) {
			return 0, io.ErrUnexpectedEOF
		},
		seekChunkFunc: func(offset int64, lastIndex bool) (int64, error) {
			return 0, io.ErrShortBuffer
		},
		rewindFunc: func() error {
			return io.ErrClosedPipe
		},
	}

	// Test Read method with error
	_, err := mockWithErrors.Read()
	assert.Equal(t, io.EOF, err, "Read should return EOF")

	// Test GetReadOffset method with error
	_, err = mockWithErrors.GetReadOffset()
	assert.Equal(t, io.ErrUnexpectedEOF, err, "GetReadOffset should return ErrUnexpectedEOF")

	// Test SeekChunk method with error
	_, err = mockWithErrors.SeekChunk(5, false)
	assert.Equal(t, io.ErrShortBuffer, err, "SeekChunk should return ErrShortBuffer")

	// Test Rewind method with error
	err = mockWithErrors.Rewind()
	assert.Equal(t, io.ErrClosedPipe, err, "Rewind should return ErrClosedPipe")
}

// Mock implementations for testing

// mockReadFormater implements the ReadFormater interface
type mockReadFormater struct {
	readFunc   func() (*seq.Seq, error)
	formatFunc func(*seq.Seq, io.Writer) (int, error)
}

func (m *mockReadFormater) Read() (*seq.Seq, error) {
	return m.readFunc()
}

func (m *mockReadFormater) Format(s *seq.Seq, w io.Writer) (int, error) {
	return m.formatFunc(s, w)
}

// mockReader implements the Reader interface
type mockReader struct {
	readFunc          func() (*seq.Seq, error)
	getReadOffsetFunc func() (int, error)
	seekChunkFunc     func(int64, bool) (int64, error)
}

func (m *mockReader) Read() (*seq.Seq, error) {
	return m.readFunc()
}

func (m *mockReader) GetReadOffset() (int, error) {
	return m.getReadOffsetFunc()
}

func (m *mockReader) SeekChunk(offset int64, lastIndex bool) (int64, error) {
	return m.seekChunkFunc(offset, lastIndex)
}

// mockReadRewinder implements the ReadRewinder interface
type mockReadRewinder struct {
	readFunc          func() (*seq.Seq, error)
	getReadOffsetFunc func() (int, error)
	seekChunkFunc     func(int64, bool) (int64, error)
	rewindFunc        func() error
}

func (m *mockReadRewinder) Read() (*seq.Seq, error) {
	return m.readFunc()
}

func (m *mockReadRewinder) GetReadOffset() (int, error) {
	return m.getReadOffsetFunc()
}

func (m *mockReadRewinder) SeekChunk(offset int64, lastIndex bool) (int64, error) {
	return m.seekChunkFunc(offset, lastIndex)
}

func (m *mockReadRewinder) Rewind() error {
	return m.rewindFunc()
}
