package errors

import (
	"regexp"
)

var (
	MongoDupKeyRegex = regexp.MustCompile("duplicate\\s+key")
)

const (
	MongoDocNotFound         = "not found"
	InvalidAuth              = "Invalid Auth Header"
	UnAuth                   = "User Unauthorized"
	NoAuth                   = "No Authorization"
	AttrImut                 = "node attributes immutable"
	FileImut                 = "node file immutable"
	ProvenanceImut           = "provenance info immutable"
	InvalidIndex             = "Invalid Index"
	InvalidFileTypeForFilter = "Invalid file type for filter"
)
