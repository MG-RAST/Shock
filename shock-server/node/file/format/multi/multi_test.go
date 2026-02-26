package multi_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/file/format/multi"
	"github.com/MG-RAST/Shock/shock-server/node/file/format/seq"
	"github.com/MG-RAST/Shock/shock-server/test/util"
	"github.com/stretchr/testify/assert"
)

// TestDetermineFormatFasta tests format detection for FASTA files
func TestDetermineFormatFasta(t *testing.T) {
	// Create test FASTA data
	testData := []byte(">seq1\nACGT\n>seq2\nTGCA\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fasta", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Test format determination
	err := multiReader.DetermineFormat()
	assert.NoError(t, err, "Format determination should not error")

	// Read sequences
	seq1, err := multiReader.Read()
	assert.NoError(t, err, "Reading first sequence should not error")
	assert.Equal(t, []byte("seq1"), seq1.ID, "First sequence ID should match")
	assert.Equal(t, []byte("ACGT"), seq1.Seq, "First sequence data should match")

	seq2, err := multiReader.Read()
	assert.NoError(t, err, "Reading second sequence should not error")
	assert.Equal(t, []byte("seq2"), seq2.ID, "Second sequence ID should match")
	assert.Equal(t, []byte("TGCA"), seq2.Seq, "Second sequence data should match")

	// Test EOF
	_, err = multiReader.Read()
	assert.Equal(t, io.EOF, err, "Reading past the end should return EOF")
}

// TestDetermineFormatFastq tests format detection for FASTQ files
func TestDetermineFormatFastq(t *testing.T) {
	// Create test FASTQ data
	testData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Test format determination
	err := multiReader.DetermineFormat()
	assert.NoError(t, err, "Format determination should not error")

	// Read sequences
	seq1, err := multiReader.Read()
	assert.NoError(t, err, "Reading first sequence should not error")
	assert.Equal(t, []byte("seq1"), seq1.ID, "First sequence ID should match")
	assert.Equal(t, []byte("ACGT"), seq1.Seq, "First sequence data should match")
	assert.Equal(t, []byte("IIII"), seq1.Qual, "First sequence quality should match")

	seq2, err := multiReader.Read()
	assert.NoError(t, err, "Reading second sequence should not error")
	assert.Equal(t, []byte("seq2"), seq2.ID, "Second sequence ID should match")
	assert.Equal(t, []byte("TGCA"), seq2.Seq, "Second sequence data should match")
	assert.Equal(t, []byte("HHHH"), seq2.Qual, "Second sequence quality should match")

	// Test EOF
	_, err = multiReader.Read()
	assert.Equal(t, io.EOF, err, "Reading past the end should return EOF")
}

// TestDetermineFormatSam tests format detection for SAM files
func TestDetermineFormatSam(t *testing.T) {
	// Create test SAM data
	testData := []byte("@HD\tVN:1.0\tSO:unsorted\n@SQ\tSN:ref\tLN:45\nread1\t0\tref\t1\t60\t10M\t*\t0\t0\tATGCATGCAT\tIIIIIIIIII\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.sam", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Test format determination
	err := multiReader.DetermineFormat()
	assert.NoError(t, err, "Format determination should not error")

	// Read sequence
	seq1, err := multiReader.Read()
	assert.NoError(t, err, "Reading sequence should not error")
	assert.Equal(t, []byte("read1"), seq1.ID, "Sequence ID should match")
}

// TestInvalidFormat tests handling of invalid format
func TestInvalidFormat(t *testing.T) {
	// Create test data with invalid format
	testData := []byte("This is not a valid sequence format file")

	// Create a test reader
	reader := util.NewInMemoryFile("test.txt", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Test format determination
	err := multiReader.DetermineFormat()
	assert.Error(t, err, "Format determination should error for invalid format")
}

// TestEmptyFile tests handling of empty files
func TestEmptyFile(t *testing.T) {
	// Create empty test data
	testData := []byte{}

	// Create a test reader
	reader := util.NewInMemoryFile("empty.txt", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Test format determination
	err := multiReader.DetermineFormat()
	assert.Error(t, err, "Format determination should error for empty file")
}

// TestGetReadOffset tests the GetReadOffset functionality
func TestGetReadOffset(t *testing.T) {
	// Create test FASTA data
	testData := []byte(">seq1\nACGT\n>seq2\nTGCA\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fasta", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Determine format
	err := multiReader.DetermineFormat()
	assert.NoError(t, err, "Format determination should not error")

	// Test getting offsets
	offset1, err := multiReader.GetReadOffset()
	assert.NoError(t, err, "Getting first offset should not error")
	assert.Greater(t, offset1, 0, "First offset should be greater than 0")

	offset2, err := multiReader.GetReadOffset()
	assert.NoError(t, err, "Getting second offset should not error")
	assert.Greater(t, offset2, 0, "Second offset should be greater than 0")

	// Test EOF
	_, err = multiReader.GetReadOffset()
	assert.Equal(t, io.EOF, err, "Reading past the end should return EOF")
}

// TestSeekChunk tests the SeekChunk functionality
func TestSeekChunk(t *testing.T) {
	// Create test FASTA data
	testData := []byte(">seq1\nACGT\n>seq2\nTGCA\n>seq3\nGGGG\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fasta", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Determine format
	err := multiReader.DetermineFormat()
	assert.NoError(t, err, "Format determination should not error")

	// Test seeking to a chunk
	n, err := multiReader.SeekChunk(0, false)
	assert.NoError(t, err, "SeekChunk should not error")
	assert.GreaterOrEqual(t, n, int64(0), "SeekChunk should return a valid position")
}

// TestFormat tests the Format functionality
func TestFormat(t *testing.T) {
	// Create test FASTA data
	testData := []byte(">seq1\nACGT\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fasta", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Determine format
	err := multiReader.DetermineFormat()
	assert.NoError(t, err, "Format determination should not error")

	// Read sequence
	sequence, err := multiReader.Read()
	assert.NoError(t, err, "Reading sequence should not error")

	// Test formatting
	var buf bytes.Buffer
	n, err := multiReader.Format(sequence, &buf)
	assert.NoError(t, err, "Formatting should not error")
	assert.Greater(t, n, 0, "Formatted output should have length > 0")
	assert.Contains(t, buf.String(), "seq1", "Formatted output should contain sequence ID")
	assert.Contains(t, buf.String(), "ACGT", "Formatted output should contain sequence data")
}

// TestReadWithoutDetermineFormat tests reading without calling DetermineFormat first
func TestReadWithoutDetermineFormat(t *testing.T) {
	// Create test FASTA data
	testData := []byte(">seq1\nACGT\n>seq2\nTGCA\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fasta", testData)

	// Create a multi reader
	multiReader := multi.NewReader(reader)

	// Read without determining format first (should auto-determine)
	seq1, err := multiReader.Read()
	assert.NoError(t, err, "Reading first sequence should not error")
	assert.Equal(t, []byte("seq1"), seq1.ID, "First sequence ID should match")
	assert.Equal(t, []byte("ACGT"), seq1.Seq, "First sequence data should match")
}

// TestUnknownFormat tests handling of unknown format in Format method
func TestUnknownFormat(t *testing.T) {
	// Create a sequence
	sequence := seq.New([]byte("test"), []byte("ACGT"), []byte("IIII"))

	// Create a buffer for output
	var buf bytes.Buffer

	// Create a multi reader with a reader that won't match any format
	reader := util.NewInMemoryFile("test.txt", []byte("Invalid format"))
	multiReader := multi.NewReader(reader)

	// Try to format without determining format
	_, err := multiReader.Format(sequence, &buf)
	assert.Error(t, err, "Formatting with unknown format should error")
	assert.Contains(t, err.Error(), "unknown sequence format", "Error message should indicate unknown format")
}
