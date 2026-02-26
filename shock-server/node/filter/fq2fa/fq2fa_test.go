package fq2fa_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/filter/fq2fa"
	"github.com/MG-RAST/Shock/shock-server/test/util"
	"github.com/stretchr/testify/assert"
)

// TestNewReader tests creating a new fq2fa reader
func TestNewReader(t *testing.T) {
	// Create test data
	testData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Verify the reader was created
	assert.NotNil(t, fq2faReader, "Fq2fa reader should not be nil")
}

// TestRead tests reading from the fq2fa reader
func TestRead(t *testing.T) {
	// Create test FASTQ data
	testData := []byte("@seq1 description\nACGT\n+\nIIII\n@seq2 another description\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading from fq2fa reader should not error")

	// The fq2fa filter should convert FASTQ to FASTA format
	assert.Contains(t, string(output), ">seq1", "Output should contain FASTA header")
	assert.Contains(t, string(output), ">seq2", "Output should contain FASTA header")
	assert.Contains(t, string(output), "ACGT", "Output should contain sequence data")
	assert.Contains(t, string(output), "TGCA", "Output should contain sequence data")
	assert.NotContains(t, string(output), "+", "Output should not contain FASTQ separator")
	assert.NotContains(t, string(output), "IIII", "Output should not contain quality scores")
	assert.NotContains(t, string(output), "HHHH", "Output should not contain quality scores")
}

// TestReadWithOverflow tests reading with buffer overflow
func TestReadWithOverflow(t *testing.T) {
	// Create large test FASTQ data
	var buf bytes.Buffer
	for i := 0; i < 100; i++ {
		buf.WriteString("@seq")
		buf.WriteString(string(rune('0' + i%10)))
		buf.WriteString("\nACGTACGTACGTACGT\n+\nIIIIIIIIIIIIIIII\n")
	}
	testData := buf.Bytes()

	// Create a test reader
	reader := util.NewInMemoryFile("large.fastq", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Read with small buffer to force overflow
	smallBuf := make([]byte, 50)
	n, err := fq2faReader.Read(smallBuf)
	assert.NoError(t, err, "First read should not error")
	assert.Equal(t, 50, n, "Should read full buffer")

	// Read again to get overflow data
	n, err = fq2faReader.Read(smallBuf)
	assert.NoError(t, err, "Second read should not error")
	assert.Greater(t, n, 0, "Should read some data")

	// Continue reading until EOF
	totalRead := n + 50
	for {
		n, err = fq2faReader.Read(smallBuf)
		if err == io.EOF {
			break
		}
		assert.NoError(t, err, "Reading should not error")
		totalRead += n
	}

	// Verify we read all the data
	assert.Greater(t, totalRead, 0, "Total bytes read should be greater than 0")
}

// TestClose tests closing the fq2fa reader
func TestClose(t *testing.T) {
	// Create test data
	testData := []byte("@seq1\nACGT\n+\nIIII\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Close the reader
	err := fq2faReader.(io.Closer).Close()
	assert.NoError(t, err, "Closing reader should not error")
}

// TestMultipleSequences tests converting multiple FASTQ sequences to FASTA
func TestMultipleSequences(t *testing.T) {
	// Create test FASTQ data with multiple sequences
	testData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n@seq3\nGGGG\n+\nJJJJ\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading from fq2fa reader should not error")

	// The fq2fa filter should convert all sequences
	outputStr := string(output)
	assert.Contains(t, outputStr, ">seq1", "Output should contain first FASTA header")
	assert.Contains(t, outputStr, ">seq2", "Output should contain second FASTA header")
	assert.Contains(t, outputStr, ">seq3", "Output should contain third FASTA header")
	assert.Contains(t, outputStr, "ACGT", "Output should contain first sequence")
	assert.Contains(t, outputStr, "TGCA", "Output should contain second sequence")
	assert.Contains(t, outputStr, "GGGG", "Output should contain third sequence")

	// Verify the order of sequences
	pos1 := bytes.Index(output, []byte(">seq1"))
	pos2 := bytes.Index(output, []byte(">seq2"))
	pos3 := bytes.Index(output, []byte(">seq3"))
	assert.Less(t, pos1, pos2, "First sequence should appear before second sequence")
	assert.Less(t, pos2, pos3, "Second sequence should appear before third sequence")
}

// TestErrorHandling tests error handling in the fq2fa reader
func TestErrorHandling(t *testing.T) {
	// Create invalid data (not FASTQ)
	testData := []byte("This is not a valid FASTQ file")

	// Create a test reader
	reader := util.NewInMemoryFile("invalid.txt", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Read and verify error
	_, err := ioutil.ReadAll(fq2faReader)
	assert.Error(t, err, "Reading invalid data should error")
}

// TestEmptyData tests the fq2fa reader with empty data
func TestEmptyData(t *testing.T) {
	// Create empty data
	testData := []byte{}

	// Create a test reader
	reader := util.NewInMemoryFile("empty.fastq", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Read and verify output
	output, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading empty data should not error")
	assert.Empty(t, output, "Output should be empty for empty input")
}

// TestSequenceWithDescription tests converting FASTQ with sequence descriptions
func TestSequenceWithDescription(t *testing.T) {
	// Create test FASTQ data with sequence descriptions
	testData := []byte("@seq1 description 1\nACGT\n+\nIIII\n@seq2 description 2\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading from fq2fa reader should not error")

	// The fq2fa filter should preserve sequence descriptions
	assert.Contains(t, string(output), ">seq1 description 1", "Output should contain FASTA header with description")
	assert.Contains(t, string(output), ">seq2 description 2", "Output should contain FASTA header with description")
}

// TestMultilineSequences tests converting FASTQ with multi-line sequences
func TestMultilineSequences(t *testing.T) {
	// Create test FASTQ data with multi-line sequences
	testData := []byte("@seq1\nACGT\nTGCA\n+\nIIII\nHHHH\n@seq2\nGGGG\nAAAA\n+\nJJJJ\nKKKK\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading from fq2fa reader should not error")

	// The fq2fa filter should handle multi-line sequences
	assert.Contains(t, string(output), ">seq1", "Output should contain first FASTA header")
	assert.Contains(t, string(output), ">seq2", "Output should contain second FASTA header")
	assert.Contains(t, string(output), "ACGTTGCA", "Output should contain first complete sequence")
	assert.Contains(t, string(output), "GGGGAAAA", "Output should contain second complete sequence")
}

// TestFastaInput tests the fq2fa reader with FASTA input (should pass through unchanged)
func TestFastaInput(t *testing.T) {
	// Create test FASTA data
	testData := []byte(">seq1\nACGT\n>seq2\nTGCA\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fasta", testData)

	// Create fq2fa reader
	fq2faReader := fq2fa.NewReader(reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading FASTA data should not error")

	// The output should be the same as the input (FASTA passes through)
	assert.Equal(t, string(testData), string(output), "FASTA input should pass through unchanged")
}
