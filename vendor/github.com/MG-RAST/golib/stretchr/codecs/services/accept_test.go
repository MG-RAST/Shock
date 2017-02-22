package services

import (
	"github.com/MG-RAST/golib/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParseAccept_NoParams(t *testing.T) {
	acceptString := "application/json"
	accept, err := ParseAcceptEntry(acceptString)
	assert.NoError(t, err, acceptString+" should parse with no errors")
	assert.Equal(t, acceptString, accept.ContentType.MimeType)
	assert.Equal(t, accept.Quality, 1.0)
}

func TestParseAccept_QualityParam(t *testing.T) {
	acceptString := "application/json; q=0.4"
	expectedQuality := 0.4

	accept, err := ParseAcceptEntry(acceptString)
	assert.NoError(t, err, acceptString+" should parse with no errors")
	assert.Equal(t, accept.Quality, expectedQuality)
}

func TestAcceptEntry_Equal(t *testing.T) {
	entryA := &AcceptEntry{
		Quality:          1.0,
		specificityCount: 0,
	}
	entryB := &AcceptEntry{
		Quality:          1.0,
		specificityCount: 0,
	}
	assert.Equal(t, entryA.CompareTo(entryB), 0, "Entries with equal quality and specificity should be equal")
}

func TestAcceptEntry_Quality(t *testing.T) {
	greater := &AcceptEntry{
		Quality:          0.8,
		specificityCount: 0,
	}
	lesser := &AcceptEntry{
		Quality:          0.3,
		specificityCount: 10,
	}
	assert.True(t, greater.CompareTo(lesser) > 0, "Higher quality should come out greater")
	assert.True(t, lesser.CompareTo(greater) < 0, "Comparing in opposite direction should provide opposite result")
}

func TestAcceptEntry_Specificity(t *testing.T) {
	greater := &AcceptEntry{
		Quality:          1.0,
		specificityCount: 2,
	}
	lesser := &AcceptEntry{
		Quality:          1.0,
		specificityCount: 1,
	}
	assert.True(t, greater.CompareTo(lesser) > 0, "At equal quality, higher specificity should come out greater")
	assert.True(t, lesser.CompareTo(greater) < 0, "Comparing in opposite direction should provide opposite result")
}

func TestOrderAcceptHeader_EqualQualityAndSpecificity(t *testing.T) {
	expectedOrder := []string{"application/json", "application/xml", "text/xml"}
	header := strings.Join(expectedOrder, ", ")
	orderedAccept, err := OrderAcceptHeader(header)
	assert.NoError(t, err)
	for index, expectedType := range expectedOrder {
		entry := orderedAccept[index]
		assert.Equal(t, entry.ContentType.MimeType, expectedType)
	}
}

func TestOrderAcceptHeader_VariedQualityAndSpecificity(t *testing.T) {
	header := "application/xml; q=0.7, */*; q=0.1, text/*; q=0.1, application/json, text/xml; q=0.7"
	expectedOrder := []string{
		// Default quality should be 1.0, so json should be first.
		"application/json",

		// application/xml shows up in the list before text/xml and
		// they are at the same q value and the same specificity, so
		// application/xml should show up before text/xml.
		"application/xml",
		"text/xml",

		// text/* is more specific than */* and they are at the same q
		// value, so text/* should show up before */*.
		"text/*",
		"*/*",
	}
	orderedAccept, err := OrderAcceptHeader(header)
	assert.NoError(t, err)
	for index, expectedType := range expectedOrder {
		entry := orderedAccept[index]
		assert.Equal(t, entry.ContentType.MimeType, expectedType)
	}
}

// AcceptTree.Flatten should always allocate exactly as much memory as
// it needs.  If capacity and length of the return value are not
// equal, something is wrong.
func TestOrderAcceptHeader_FlattenPerformance(t *testing.T) {
	testHeaders := []string{
		"",
		"application/json",
		"application/xml; q=0.7, */*; q=0.1, text/*; q=0.1, application/json, text/xml; q=0.7",
	}

	for _, testHeader := range testHeaders {
		orderedAccept, err := OrderAcceptHeader(testHeader)
		assert.NoError(t, err)
		assert.Equal(t, len(orderedAccept), cap(orderedAccept),
			"Flatten should allocate exactly as much memory as it needs; failed header: "+testHeader)
	}
}
