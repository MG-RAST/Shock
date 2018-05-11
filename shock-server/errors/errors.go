// Package contains error strings and patterns for common errors
package errors

import (
	"regexp"
)

var (
	MongoDupKeyRegex = regexp.MustCompile("duplicate\\s+key")
)

const (
	InvalidAuth              = "Invalid Auth Header"
	UnAuth                   = "User Unauthorized"
	NoAuth                   = "No Authorization"
	AttrImut                 = "Node attributes immutable"
	FileImut                 = "Node file immutable"
	ProvenanceImut           = "Provenance info immutable"
	InvalidIndex             = "Invalid index type"
	InvalidFileTypeForFilter = "Invalid file type for filter"
	InvalidIndexRange        = "Invalid index record range"
	IndexOutBounds           = "Index record out of bounds"
	IndexNoFile              = "Index file is missing"
	NodeReferenced           = "Node referenced by virtual node"
	NodeDoesNotExist         = "Node does not exist"
	NodeNotFound             = "Node not found"
	NodeNoFile               = "Node has no file"
	NodeDownloadLock         = "Node file is locked from download or indexing"
)
