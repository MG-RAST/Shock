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
	AttrImut                 = "Node attributes immutable"
	FileImut                 = "Node file immutable"
	ProvenanceImut           = "Provenance info immutable"
	InvalidIndex             = "Invalid Index"
	InvalidFileTypeForFilter = "Invalid file type for filter"
	NodeReferenced           = "Node referenced by virtual node"
)
