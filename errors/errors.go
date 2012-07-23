package errors

import (
	"regexp"
)

var (
	MongoDupKeyRegex = regexp.MustCompile("duplicate\\s+key")
)

const (
	MongoDocNotFound         = "Document not found"
	UnAuth                   = "User Unauthorized"
	NoAuth                   = "No Authorization"
	AttrImut                 = "node attributes immutable"
	FileImut                 = "node file immutable"
	InvalidIndex             = "Invalid Index"
	InvalidFileTypeForFilter = "Invalid file type for filter"
)
