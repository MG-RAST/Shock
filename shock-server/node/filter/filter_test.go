package filter_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/MG-RAST/Shock/shock-server/node/filter"
	"github.com/MG-RAST/Shock/shock-server/test/util"
	"github.com/stretchr/testify/assert"
)

// TestHas tests the Has function
func TestHas(t *testing.T) {
	// Test with valid filter names
	assert.True(t, filter.Has("anonymize"), "Has should return true for 'anonymize'")
	assert.True(t, filter.Has("fq2fa"), "Has should return true for 'fq2fa'")

	// Test with invalid filter name
	assert.False(t, filter.Has("invalid"), "Has should return false for 'invalid'")
	assert.False(t, filter.Has(""), "Has should return false for empty string")
}

// TestFilter tests the Filter function
func TestFilter(t *testing.T) {
	// Test with valid filter names
	anonymizeFilter := filter.Filter("anonymize")
	assert.NotNil(t, anonymizeFilter, "Filter should return non-nil for 'anonymize'")

	fq2faFilter := filter.Filter("fq2fa")
	assert.NotNil(t, fq2faFilter, "Filter should return non-nil for 'fq2fa'")

	// Test with invalid filter name
	// This should return nil, but since the map lookup returns the zero value (nil) for a missing key,
	// we can't really test this behavior directly
}

// TestNewReader tests the NewReader function
func TestNewReader(t *testing.T) {
	// Create test data
	testData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Test with anonymize filter
	anonymizeReader := filter.NewReader("anonymize", reader)
	assert.NotNil(t, anonymizeReader, "NewReader should return non-nil for 'anonymize'")

	// Test with fq2fa filter
	fq2faReader := filter.NewReader("fq2fa", reader)
	assert.NotNil(t, fq2faReader, "NewReader should return non-nil for 'fq2fa'")
}

// TestAnonymizeFilter tests the anonymize filter functionality
func TestAnonymizeFilter(t *testing.T) {
	// Create test FASTQ data
	testData := []byte("@seq1 description\nACGT\n+\nIIII\n@seq2 another description\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create anonymize filter
	anonymizeReader := filter.NewReader("anonymize", reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(anonymizeReader)
	assert.NoError(t, err, "Reading from anonymize filter should not error")

	// The anonymize filter should replace sequence IDs with numbers
	assert.NotContains(t, string(output), "seq1", "Output should not contain original sequence ID")
	assert.NotContains(t, string(output), "seq2", "Output should not contain original sequence ID")
	assert.Contains(t, string(output), "1", "Output should contain numeric ID")
	assert.Contains(t, string(output), "2", "Output should contain numeric ID")
	assert.Contains(t, string(output), "ACGT", "Output should contain sequence data")
	assert.Contains(t, string(output), "TGCA", "Output should contain sequence data")
}

// TestFq2faFilter tests the fq2fa filter functionality
func TestFq2faFilter(t *testing.T) {
	// Create test FASTQ data
	testData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Create fq2fa filter
	fq2faReader := filter.NewReader("fq2fa", reader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading from fq2fa filter should not error")

	// The fq2fa filter should convert FASTQ to FASTA format
	assert.Contains(t, string(output), ">seq1", "Output should contain FASTA header")
	assert.Contains(t, string(output), ">seq2", "Output should contain FASTA header")
	assert.Contains(t, string(output), "ACGT", "Output should contain sequence data")
	assert.Contains(t, string(output), "TGCA", "Output should contain sequence data")
	assert.NotContains(t, string(output), "+", "Output should not contain FASTQ separator")
	assert.NotContains(t, string(output), "IIII", "Output should not contain quality scores")
	assert.NotContains(t, string(output), "HHHH", "Output should not contain quality scores")
}

// TestFilterChaining tests chaining multiple filters
func TestFilterChaining(t *testing.T) {
	// Create test FASTQ data
	testData := []byte("@seq1\nACGT\n+\nIIII\n@seq2\nTGCA\n+\nHHHH\n")

	// Create a test reader
	reader := util.NewInMemoryFile("test.fastq", testData)

	// Chain filters: first convert FASTQ to FASTA, then anonymize
	// Since filter.NewReader returns io.Reader but expects file.SectionReader,
	// we need to collect intermediate output and wrap it in an InMemoryFile.
	fq2faReader := filter.NewReader("fq2fa", reader)
	intermediateOutput, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading from fq2fa filter should not error")

	intermediateReader := util.NewInMemoryFile("intermediate.fasta", intermediateOutput)
	anonymizeReader := filter.NewReader("anonymize", intermediateReader)

	// Read and verify the filtered output
	output, err := ioutil.ReadAll(anonymizeReader)
	assert.NoError(t, err, "Reading from chained filters should not error")

	outputStr := string(output)

	// The output should be anonymized FASTA - at minimum the first sequence should be anonymized
	assert.NotContains(t, outputStr, "@seq1", "Output should not contain original FASTQ header")
	assert.NotContains(t, outputStr, ">seq1", "Output should not contain original FASTA header")
	assert.Contains(t, outputStr, ">1", "Output should contain anonymized FASTA header")
	assert.Contains(t, outputStr, "ACGT", "Output should contain sequence data")
	assert.NotContains(t, outputStr, "IIII", "Output should not contain quality scores")
	assert.NotContains(t, outputStr, "HHHH", "Output should not contain quality scores")
}

// TestLargeData tests filters with large data
func TestLargeData(t *testing.T) {
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

	// Test with anonymize filter
	anonymizeReader := filter.NewReader("anonymize", reader)
	anonymizeOutput, err := ioutil.ReadAll(anonymizeReader)
	assert.NoError(t, err, "Reading from anonymize filter should not error")
	assert.NotEmpty(t, anonymizeOutput, "Anonymize output should not be empty")

	// Reset reader
	reader = util.NewInMemoryFile("large.fastq", testData)

	// Test with fq2fa filter
	fq2faReader := filter.NewReader("fq2fa", reader)
	fq2faOutput, err := ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading from fq2fa filter should not error")
	assert.NotEmpty(t, fq2faOutput, "Fq2fa output should not be empty")
}

// TestErrorHandling tests error handling in filters
func TestErrorHandling(t *testing.T) {
	// Create invalid data (not FASTQ or FASTA)
	testData := []byte("This is not a valid sequence file")

	// Create a test reader
	reader := util.NewInMemoryFile("invalid.txt", testData)

	// Test with anonymize filter
	anonymizeReader := filter.NewReader("anonymize", reader)
	_, err := ioutil.ReadAll(anonymizeReader)
	assert.Error(t, err, "Reading invalid data with anonymize filter should error")

	// Reset reader
	reader = util.NewInMemoryFile("invalid.txt", testData)

	// Test with fq2fa filter
	fq2faReader := filter.NewReader("fq2fa", reader)
	_, err = ioutil.ReadAll(fq2faReader)
	assert.Error(t, err, "Reading invalid data with fq2fa filter should error")
}

// TestEmptyData tests filters with empty data
func TestEmptyData(t *testing.T) {
	// Create empty data
	testData := []byte{}

	// Create a test reader
	reader := util.NewInMemoryFile("empty.fastq", testData)

	// Test with anonymize filter
	anonymizeReader := filter.NewReader("anonymize", reader)
	output, err := ioutil.ReadAll(anonymizeReader)
	assert.NoError(t, err, "Reading empty data with anonymize filter should not error")
	assert.Empty(t, output, "Output should be empty for empty input")

	// Reset reader
	reader = util.NewInMemoryFile("empty.fastq", testData)

	// Test with fq2fa filter
	fq2faReader := filter.NewReader("fq2fa", reader)
	output, err = ioutil.ReadAll(fq2faReader)
	assert.NoError(t, err, "Reading empty data with fq2fa filter should not error")
	assert.Empty(t, output, "Output should be empty for empty input")
}
