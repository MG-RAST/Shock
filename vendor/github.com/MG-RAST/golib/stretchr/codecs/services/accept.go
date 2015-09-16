package services

import (
	"strconv"
	"strings"
)

// AcceptEntry represents a single entry within an Accept header.  It
// includes both the ContentType and a Quality, parsed from the
// ContentType's parameters.
type AcceptEntry struct {
	ContentType *ContentType

	// Quality is the parsed q value from a ContentType's parameters.
	Quality float32

	// Internal use only, for measuring how "specific" the entry is.
	specificityCount int
}

// NewAcceptEntry returns a new *AcceptEntry with default values.
func NewAcceptEntry() *AcceptEntry {
	entry := &AcceptEntry{
		Quality: 1.0,
	}
	return entry
}

// ParseAcceptEntry parses a single entry within an Accept header into
// a *AcceptEntry value.
func ParseAcceptEntry(accept string) (*AcceptEntry, error) {

	entry := NewAcceptEntry()
	var typeErr error
	entry.ContentType, typeErr = ParseContentType(accept)
	if typeErr != nil {
		return nil, typeErr
	}

	if qualityString, ok := entry.ContentType.Parameters["q"]; ok {
		quality, err := strconv.ParseFloat(qualityString, 32)
		if err != nil {
			return nil, err
		}
		entry.Quality = float32(quality)
	}

	// Parameters are more specific.  Wildcards in the MimeType
	// are less specific.  I can't find anything detailing whether
	// one or the other is more important, so I'm counting them as
	// equals.
	entry.specificityCount += len(entry.ContentType.Parameters)
	entry.specificityCount -= strings.Count(entry.ContentType.MimeType, "*")

	return entry, nil
}

// CompareTo compares two *AcceptEntries and returns an integer
// representing which of the two entries is preferred. Negative return
// values mean that the passed in entry is preferred, positive values
// mean that the target entry is preferred, and zero values mean that
// there is no preference.
func (entry *AcceptEntry) CompareTo(otherEntry *AcceptEntry) int {
	if entry.Quality > otherEntry.Quality {
		return 1
	}
	if entry.Quality < otherEntry.Quality {
		return -1
	}
	return entry.specificityCount - otherEntry.specificityCount
}

// AcceptTree is a binary tree that handles Accept header entries.
// The left-most node will always be the most preferred entry and
// preference will decrease from left to right.
type AcceptTree struct {
	Value *AcceptEntry
	Size  int
	Left  *AcceptTree
	Right *AcceptTree
}

// Add adds an *AcceptEntry to the AcceptTree, putting it in proper
// order of preference.
func (tree *AcceptTree) Add(next *AcceptEntry) {
	if tree.Value == nil {
		tree.Value = next
	} else if next.CompareTo(tree.Value) > 0 {
		if tree.Left == nil {
			tree.Left = &AcceptTree{Value: next, Size: 1}
		} else {
			tree.Left.Add(next)
		}
	} else {
		if tree.Right == nil {
			tree.Right = &AcceptTree{Value: next, Size: 1}
		} else {
			tree.Right.Add(next)
		}
	}
	tree.Size++
}

// Flatten returns the AcceptTree's values in proper order of
// preference as a []*AcceptEntry value.
func (tree *AcceptTree) Flatten() []*AcceptEntry {
	entries := make([]*AcceptEntry, 0, tree.Size)
	tree.flattenToTarget(&entries)
	return entries
}

// flattenToTarget is a helper method for Flatten, which takes a
// pointer to a slice of AcceptEntry pointers and appends all values
// in the tree to it.  This is so that Flatten can allocate all
// required space for its return value, then have flattenToTarget
// populate the pre-allocateed space with values.
func (tree *AcceptTree) flattenToTarget(target *[]*AcceptEntry) {
	if tree.Value != nil {
		if tree.Left != nil {
			tree.Left.flattenToTarget(target)
		}
		*target = append(*target, tree.Value)
		if tree.Right != nil {
			tree.Right.flattenToTarget(target)
		}
	}
}

// OrderAcceptHeader reads an Accept header and pulls out the various
// MIME types in the order of preference, returning the types in that
// order.
//
// The HTTP spec for the Accept header states that multiple MIME types
// can be specified in the Accept header, and that preferred MIME
// types are chosen based on the following criteria:
//
// 1. The q variable for a MIME type in the Accept header defines a
// 'quality', and higher qualities should be chosen over lower
// qualities.  The default quality is 1.0 for any type that doesn't
// state the quality explicitly.
// 2. More specific MIME types should be chosen over less specific
// MIME types, excepting the presence of a q parameter that counters
// this guideline.  For example, assuming equal qualities,
// application/xml should trump application/*.
// 3. Barring the previous two guidelines, MIME types should be chosen
// based on the order that they appear in the Accept header.
//
// For more information, see
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
func OrderAcceptHeader(accept string) ([]*AcceptEntry, error) {
	acceptTree := new(AcceptTree)
	var (
		end       int
		rawEntry  string
		remaining = accept
	)
	// At the time of this writing, strings.Split was using a lot of
	// execution time, relative to the execution time of this
	// function.  This logic is a fair bit faster.
	for end != -1 {
		end = strings.IndexRune(remaining, ',')
		if end == -1 {
			rawEntry = remaining
		} else {
			rawEntry = remaining[:end]
		}
		rawEntry = strings.TrimSpace(rawEntry)
		if rawEntry != "" {
			entry, err := ParseAcceptEntry(rawEntry)
			if err != nil {
				return nil, err
			}
			acceptTree.Add(entry)
		}
		remaining = remaining[end+1:]
	}
	return acceptTree.Flatten(), nil
}
