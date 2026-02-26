package anonymize_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/filter/anonymize"
	"github.com/MG-RAST/Shock/shock-server/test/util"
	"github.com/stretchr/testify/assert"
)

// TestNewReader tests creating a new anonymize reader
func TestNewReader(t *testing.T) {
	// Create test data
	testData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create anonymize reader
	anonymizeReader := anonymize.NewReader(reader)

	// Verify the reader was created
	assert.NotNil(t, anonymizeReader, "Anonymize reader should not be nil")
}

// TestRead tests reading from the anonymize reader
func TestRead(t *testing.T) {
	// Create test FASTQ data
	testData := []byte("@seq1 description\nACGT\n+\nIIII\n@seq2 another description\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create anonymize reader
	anonymizeReader := anonymize.NewReader(reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(anonymizeReader)
	assert.NoError(t, err, "Reading from anonymize reader should not error")

	// The anonymize filter should replace sequence IDs with numbers
	assert.NotContains(t, string(output), "seq1", "Output should not contain original sequence ID")
	assert.NotContains(t, string(output), "seq2", "Output should not contain original sequence ID")
	assert.Contains(t, string(output), "1", "Output should contain numeric ID")
	assert.Contains(t, string(output), "2", "Output should contain numeric ID")
	assert.Contains(t, string(output), "ACGT", "Output should contain sequence data")
	assert.Contains(t, string(output), "TGCA", "Output should contain sequence data")
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

	// Create anonymize reader
	anonymizeReader := anonymize.NewReader(reader)

	// Read with small buffer to force overflow
	smallBuf := make([]byte, 50)
	n, err := anonymizeReader.Read(smallBuf)
	assert.NoError(t, err, "First read should not error")
	assert.Equal(t, 50, n, "Should read full buffer")

	// Read again to get overflow data
	n, err = anonymizeReader.Read(smallBuf)
	assert.NoError(t, err, "Second read should not error")
	assert.Greater(t, n, 0, "Should read some data")

	// Continue reading until EOF
	totalRead := n + 50
	for {
		n, err = anonymizeReader.Read(smallBuf)
		if err == io.EOF {
			break
		}
		assert.NoError(t, err, "Reading should not error")
		totalRead += n
	}

	// Verify we read all the data
	assert.Greater(t, totalRead, 0, "Total bytes read should be greater than 0")
}

// TestClose tests closing the anonymize reader
func TestClose(t *testing.T) {
	// Create test data
	testData := []byte("@seq1\nACGT\n+\nIIII\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create anonymize reader
	anonymizeReader := anonymize.NewReader(reader)

	// Close the reader
	err := anonymizeReader.(io.Closer).Close()
	assert.NoError(t, err, "Closing reader should not error")
}

// TestSequenceCounter tests that sequence IDs are properly incremented
func TestSequenceCounter(t *testing.T) {
	// Create test FASTQ data with multiple sequences
	testData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n@seq3\nGGGG\n+\nJJJJ\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create anonymize reader
	anonymizeReader := anonymize.NewReader(reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(anonymizeReader)
	assert.NoError(t, err, "Reading from anonymize reader should not error")

	// The anonymize filter should replace sequence IDs with sequential numbers
	outputStr := string(output)
	assert.Contains(t, outputStr, "1", "Output should contain first numeric ID")
	assert.Contains(t, outputStr, "2", "Output should contain second numeric ID")
	assert.Contains(t, outputStr, "3", "Output should contain third numeric ID")

	// Verify the order of IDs
	pos1 := bytes.Index(output, []byte("1"))
	pos2 := bytes.Index(output, []byte("2"))
	pos3 := bytes.Index(output, []byte("3"))
	assert.Less(t, pos1, pos2, "First ID should appear before second ID")
	assert.Less(t, pos2, pos3, "Second ID should appear before third ID")
}

// TestErrorHandling tests error handling in the anonymize reader
func TestErrorHandling(t *testing.T) {
	// Create invalid data (not FASTQ or FASTA)
	testData := []byte("This is not a valid sequence file")

	// Create a test reader
	reader := util.NewInMemoryFile("invalid.txt", testData)

	// Create anonymize reader
	anonymizeReader := anonymize.NewReader(reader)

	// Read and verify error
	_, err := ioutil.ReadAll(anonymizeReader)
	assert.Error(t, err, "Reading invalid data should error")
}

// TestEmptyData tests the anonymize reader with empty data
func TestEmptyData(t *testing.T) {
	// Create empty data
	testData := []byte{}

	// Create a test reader
	reader := util.NewInMemoryFile("empty.fastq", testData)

	// Create anonymize reader
	anonymizeReader := anonymize.NewReader(reader)

	// Read and verify output
	output, err := ioutil.ReadAll(anonymizeReader)
	assert.NoError(t, err, "Reading empty data should not error")
	assert.Empty(t, output, "Output should be empty for empty input")
}

// TestMultipleFormats tests the anonymize reader with different sequence formats
func TestMultipleFormats(t *testing.T) {
	// Test with FASTA format
	fastaData := []byte(">seq1 description\nACGT\n>seq2 another description\nTGCA\n")
	fastaReader := util.NewInMemoryFile("test.fasta", fastaData)
	fastaAnonymizeReader := anonymize.NewReader(fastaReader)
	fastaOutput, err := ioutil.ReadAll(fastaAnonymizeReader)
	assert.NoError(t, err, "Reading FASTA data should not error")
	assert.NotContains(t, string(fastaOutput), "seq1", "Output should not contain original sequence ID")
	assert.NotContains(t, string(fastaOutput), "seq2", "Output should not contain original sequence ID")
	assert.Contains(t, string(fastaOutput), "1", "Output should contain numeric ID")
	assert.Contains(t, string(fastaOutput), "2", "Output should contain numeric ID")

	// Test with FASTQ format
	fastqData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n")
	fastqReader := util.NewInMemoryFile("test.fastq", fastqData)
	fastqAnonymizeReader := anonymize.NewReader(fastqReader)
	fastqOutput, err := ioutil.ReadAll(fastqAnonymizeReader)
	assert.NoError(t, err, "Reading FASTQ data should not error")
	assert.NotContains(t, string(fastqOutput), "seq1", "Output should not contain original sequence ID")
	assert.NotContains(t, string(fastqOutput), "seq2", "Output should not contain original sequence ID")
	assert.Contains(t, string(fastqOutput), "1", "Output should contain numeric ID")
	assert.Contains(t, string(fastqOutput), "2", "Output should contain numeric ID")
}
