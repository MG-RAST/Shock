package line_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/file/format/line"
	"github.com/MG-RAST/Shock/shock-server/test/util"
	"github.com/stretchr/testify/assert"
)

// TestLineReader tests the line reader functionality
func TestLineReader(t *testing.T) {
	// Create test data with multiple lines
	testData := []byte("This is line 1\nThis is line 2\nThis is line 3")

	// Create a test reader
	reader := util.NewInMemoryFile("test.txt", testData)

	// Create a line reader
	lineReader := line.NewReader(reader)

	// Test reading lines
	line1, err := lineReader.ReadLine()
	assert.NoError(t, err, "Reading first line should not error")
	assert.Equal(t, []byte("This is line 1\n"), line1, "First line should match")

	line2, err := lineReader.ReadLine()
	assert.NoError(t, err, "Reading second line should not error")
	assert.Equal(t, []byte("This is line 2\n"), line2, "Second line should match")

	line3, err := lineReader.ReadLine()
	assert.NoError(t, err, "Reading third line should not error")
	assert.Equal(t, []byte("This is line 3"), line3, "Third line should match")

	// Test EOF
	_, err = lineReader.ReadLine()
	assert.Equal(t, io.EOF, err, "Reading past the end should return EOF")
}

// TestGetReadOffset tests the GetReadOffset functionality
func TestGetReadOffset(t *testing.T) {
	// Create test data with multiple lines of different lengths
	testData := []byte("Short line\nLonger line with more text\nMedium line")

	// Create a test reader
	reader := util.NewInMemoryFile("test.txt", testData)

	// Create a line reader
	lineReader := line.NewReader(reader)

	// Test getting offsets
	offset1, err := lineReader.GetReadOffset()
	assert.NoError(t, err, "Getting first offset should not error")
	assert.Equal(t, 11, offset1, "First offset should match line length including newline")

	offset2, err := lineReader.GetReadOffset()
	assert.NoError(t, err, "Getting second offset should not error")
	assert.Equal(t, 27, offset2, "Second offset should match line length including newline")

	offset3, err := lineReader.GetReadOffset()
	assert.NoError(t, err, "Getting third offset should not error")
	assert.Equal(t, 11, offset3, "Third offset should match line length")

	// Test EOF
	_, err = lineReader.GetReadOffset()
	assert.Equal(t, io.EOF, err, "Reading past the end should return EOF")
}

// TestEmptyFile tests reading from an empty file
func TestEmptyFile(t *testing.T) {
	// Create empty test data
	testData := []byte{}

	// Create a test reader
	reader := util.NewInMemoryFile("empty.txt", testData)

	// Create a line reader
	lineReader := line.NewReader(reader)

	// Test reading from empty file
	_, err := lineReader.ReadLine()
	assert.Equal(t, io.EOF, err, "Reading from empty file should return EOF")

	// Test getting offset from empty file
	_, err = lineReader.GetReadOffset()
	assert.Equal(t, io.EOF, err, "Getting offset from empty file should return EOF")
}

// TestSingleLine tests reading a file with a single line
func TestSingleLine(t *testing.T) {
	// Create test data with a single line
	testData := []byte("This is a single line without newline")

	// Create a test reader
	reader := util.NewInMemoryFile("single.txt", testData)

	// Create a line reader
	lineReader := line.NewReader(reader)

	// Test reading the single line
	line1, err := lineReader.ReadLine()
	assert.NoError(t, err, "Reading single line should not error")
	assert.Equal(t, testData, line1, "Line content should match")

	// Test EOF after reading the single line
	_, err = lineReader.ReadLine()
	assert.Equal(t, io.EOF, err, "Reading past the end should return EOF")

	// Reset reader and test GetReadOffset
	reader = util.NewInMemoryFile("single.txt", testData)
	lineReader = line.NewReader(reader)

	offset, err := lineReader.GetReadOffset()
	assert.NoError(t, err, "Getting offset should not error")
	assert.Equal(t, len(testData), offset, "Offset should match data length")
}

// TestReaderReset tests creating a new reader from the same file
func TestReaderReset(t *testing.T) {
	// Create test data
	testData := []byte("Line 1\nLine 2\nLine 3")

	// Create a test reader
	reader := util.NewInMemoryFile("test.txt", testData)

	// Create a line reader and read a line
	lineReader := line.NewReader(reader)
	line1, err := lineReader.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Line 1\n"), line1)

	// Reset the reader by seeking to beginning
	reader.Close()
	reader = util.NewInMemoryFile("test.txt", testData)

	// Create a new line reader and verify we can read from the beginning
	lineReader = line.NewReader(reader)
	line1Again, err := lineReader.ReadLine()
	assert.NoError(t, err)
	assert.Equal(t, []byte("Line 1\n"), line1Again)
}

// TestBufferHandling tests the internal buffer handling
func TestBufferHandling(t *testing.T) {
	// Create test data with very long lines
	var buf bytes.Buffer
	for i := 0; i < 1000; i++ {
		buf.WriteString("This is a very long line that should test the buffer handling in the line reader implementation. ")
	}
	buf.WriteString("\nShort line\n")
	testData := buf.Bytes()

	// Create a test reader
	reader := util.NewInMemoryFile("longlines.txt", testData)

	// Create a line reader
	lineReader := line.NewReader(reader)

	// Read the long line
	longLine, err := lineReader.ReadLine()
	assert.NoError(t, err, "Reading long line should not error")
	assert.Equal(t, testData[:bytes.IndexByte(testData, '\n')+1], longLine, "Long line content should match")

	// Read the short line
	shortLine, err := lineReader.ReadLine()
	assert.NoError(t, err, "Reading short line should not error")
	assert.Equal(t, []byte("Short line\n"), shortLine, "Short line content should match")
}
