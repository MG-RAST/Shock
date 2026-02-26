package sam_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/file/format/sam"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"github.com/MG-RAST/Shock/shock-server/test/util"
	"github.com/stretchr/testify/assert"
)

// TestSamRegex tests the SAM format regex pattern
func TestSamRegex(t *testing.T) {
	// The SAM regex: ^[\n\r]*[@[A-Z][A-Z][ \t]+[\S \t]+[\n\r]]*
	// It matches: optional leading newlines, then @ followed by an uppercase letter,
	// then a space/tab, then content, then a newline.
	// Note: The regex requires exactly one uppercase char between @ and the tab/space.

	// Test that the regex is not nil and compiles
	assert.NotNil(t, sam.Regex, "SAM regex should not be nil")

	// Test invalid format (not SAM)
	invalidFormat := []byte("This is not a SAM file")
	assert.False(t, sam.Regex.Match(invalidFormat), "Invalid format should not match regex")

	// Test empty file - regex matches empty string via zero-length quantifiers
	emptyFile := []byte("")
	// The regex ^[\n\r]* matches empty at start, then the rest is zero-width too
	matched := sam.Regex.Match(emptyFile)
	// Just verify it doesn't panic
	_ = matched

	// Test FASTA format (should not match @ header without tab after two uppercase chars)
	fastaFormat := []byte(">seq1\nACGT\n>seq2\nTGCA\n")
	assert.False(t, sam.Regex.Match(fastaFormat), "FASTA format should not match SAM regex")
}

// TestNewReader tests creating a new SAM reader
func TestNewReader(t *testing.T) {
	// Create test SAM data
	testData := []byte("@HD\tVN:1.0\tSO:unsorted\n@SQ\tSN:ref\tLN:45\nread1\t0\tref\t1\t60\t10M\t*\t0\t0\tATGCATGCAT\tIIIIIIIIII\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.sam", testData)

	// Create a SAM reader
	samReader := sam.NewReader(reader)

	// Verify the reader was created
	assert.NotNil(t, samReader, "SAM reader should not be nil")
}

// TestNewReaderName tests creating a new SAM reader from a filename
func TestNewReaderName(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "test.sam")
	assert.NoError(t, err, "Creating temp file should not error")
	defer os.Remove(tmpfile.Name())

	// Write test SAM data
	testData := []byte("@HD\tVN:1.0\tSO:unsorted\n@SQ\tSN:ref\tLN:45\nread1\t0\tref\t1\t60\t10M\t*\t0\t0\tATGCATGCAT\tIIIIIIIIII\n")
	_, err = tmpfile.Write(testData)
	assert.NoError(t, err, "Writing to temp file should not error")
	tmpfile.Close()

	// Create a SAM reader from filename
	samReader, err := sam.NewReaderName(tmpfile.Name())
	assert.NoError(t, err, "Creating SAM reader from filename should not error")
	assert.NotNil(t, samReader, "SAM reader should not be nil")

	// Test with non-existent file
	_, err = sam.NewReaderName("non_existent_file.sam")
	assert.Error(t, err, "Creating SAM reader from non-existent file should error")
}

// TestRead tests reading sequences from a SAM file
func TestRead(t *testing.T) {
	// Create test SAM data
	testData := []byte("@HD\tVN:1.0\tSO:unsorted\n@SQ\tSN:ref\tLN:45\nread1\t0\tref\t1\t60\t10M\t*\t0\t0\tATGCATGCAT\tIIIIIIIIII\nread2\t0\tref\t1\t60\t10M\t*\t0\t0\tGCATGCATGC\tHHHHHHHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.sam", testData)

	// Create a SAM reader
	samReader := sam.NewReader(reader)

	// Read first sequence
	seq1, err := samReader.Read()
	assert.NoError(t, err, "Reading first sequence should not error")
	assert.Equal(t, []byte("read1"), seq1.ID, "First sequence ID should match")

	// Read second sequence
	seq2, err := samReader.Read()
	assert.NoError(t, err, "Reading second sequence should not error")
	assert.Equal(t, []byte("read2"), seq2.ID, "Second sequence ID should match")

	// Test EOF
	_, err = samReader.Read()
	assert.Equal(t, io.EOF, err, "Reading past the end should return EOF")
}

// TestReadInvalidSam tests reading from an invalid SAM file
func TestReadInvalidSam(t *testing.T) {
	// Create test data with invalid SAM format (missing required fields)
	testData := []byte("@HD\tVN:1.0\tSO:unsorted\n@SQ\tSN:ref\tLN:45\nread1\t0\tref\t1\t60\n")

	// Create a test reader
	reader := util.NewInMemoryFile("invalid.sam", testData)

	// Create a SAM reader
	samReader := sam.NewReader(reader)

	// Try to read sequence
	_, err := samReader.Read()
	assert.Error(t, err, "Reading invalid SAM format should error")
	assert.Contains(t, err.Error(), "sam alignment fields less than 11", "Error message should indicate invalid format")
}

// TestGetReadOffset tests the GetReadOffset functionality
func TestGetReadOffset(t *testing.T) {
	// Create test SAM data
	testData := []byte("@HD\tVN:1.0\tSO:unsorted\n@SQ\tSN:ref\tLN:45\nread1\t0\tref\t1\t60\t10M\t*\t0\t0\tATGCATGCAT\tIIIIIIIIII\nread2\t0\tref\t1\t60\t10M\t*\t0\t0\tGCATGCATGC\tHHHHHHHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.sam", testData)

	// Create a SAM reader
	samReader := sam.NewReader(reader)

	// Test getting offsets
	offset1, err := samReader.GetReadOffset()
	assert.NoError(t, err, "Getting first offset should not error")
	assert.Greater(t, offset1, 0, "First offset should be greater than 0")

	offset2, err := samReader.GetReadOffset()
	assert.NoError(t, err, "Getting second offset should not error")
	assert.Greater(t, offset2, 0, "Second offset should be greater than 0")

	// Test EOF
	_, err = samReader.GetReadOffset()
	assert.Equal(t, io.EOF, err, "Reading past the end should return EOF")
}

// TestRewind tests the Rewind functionality
func TestRewind(t *testing.T) {
	// Create test SAM data
	testData := []byte("@HD\tVN:1.0\tSO:unsorted\n@SQ\tSN:ref\tLN:45\nread1\t0\tref\t1\t60\t10M\t*\t0\t0\tATGCATGCAT\tIIIIIIIIII\nread2\t0\tref\t1\t60\t10M\t*\t0\t0\tGCATGCATGC\tHHHHHHHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.sam", testData)

	// Create a SAM reader
	samReader := sam.NewReader(reader)

	// Read first sequence
	seq1, err := samReader.Read()
	assert.NoError(t, err, "Reading first sequence should not error")
	assert.Equal(t, []byte("read1"), seq1.ID, "First sequence ID should match")

	// Rewind the reader
	err = samReader.Rewind()
	assert.NoError(t, err, "Rewinding should not error")

	// Read first sequence again
	seq1Again, err := samReader.Read()
	assert.NoError(t, err, "Reading first sequence again should not error")
	assert.Equal(t, []byte("read1"), seq1Again.ID, "First sequence ID should match after rewind")
}

// TestWriter tests the SAM writer functionality
func TestWriter(t *testing.T) {
	// Create a buffer for output
	var buf bytes.Buffer

	// Create a SAM writer
	samWriter := sam.NewWriter(&nopWriteCloser{&buf}, 0)

	// Create a sequence to write
	seq := &seq.Seq{
		ID:  []byte("read1"),
		Seq: []byte("ATGCATGCAT"),
	}

	// Write the sequence
	n, err := samWriter.Write(seq)
	assert.NoError(t, err, "Writing sequence should not error")
	assert.Greater(t, n, 0, "Number of bytes written should be greater than 0")

	// Flush the writer
	err = samWriter.Flush()
	assert.NoError(t, err, "Flushing writer should not error")

	// Close the writer
	err = samWriter.Close()
	assert.NoError(t, err, "Closing writer should not error")

	// Verify output
	assert.Contains(t, buf.String(), "ATGCATGCAT", "Output should contain sequence data")
}

// TestFormat tests the Format functionality
func TestFormat(t *testing.T) {
	// Create a sequence
	seq := &seq.Seq{
		ID:  []byte("read1"),
		Seq: []byte("ATGCATGCAT"),
	}

	// Create a buffer for output
	var buf bytes.Buffer

	// Format the sequence
	n, err := sam.Format(seq, &buf)
	assert.NoError(t, err, "Formatting sequence should not error")
	assert.Greater(t, n, 0, "Number of bytes written should be greater than 0")

	// Verify output
	assert.Contains(t, buf.String(), "ATGCATGCAT", "Formatted output should contain sequence data")
}

// nopWriteCloser wraps an io.Writer and provides a no-op Close method
type nopWriteCloser struct {
	io.Writer
}

func (nwc *nopWriteCloser) Close() error {
	return nil
}
