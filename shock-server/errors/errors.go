// Package contains error strings and patterns for common errors
package errors

import (
	"regexp"
)

var (
	MongoDupKeyRegex = regexp.MustCompile("duplicate\\s+key")
)

const (
	// Authentication and authorization
	InvalidAuth    = "Invalid Auth Header"
	UnAuth         = "User Unauthorized"
	NoAuth         = "No Authorization"
	InvalidToken   = "Invalid token"
	NoEligibleTokens      = "No eligible tokens available"
	InvalidAuthHeader     = "Invalid authorization header"
	NoTokenProvidedNoLogin = "No token provided and no login credentials"
	TokenExpired           = "Token has expired"
	TokenFormatIncorrect   = "Token format is incorrect"
	UnableToCreateToken    = "Unable to create token"

	// Node immutability
	AttrImut       = "Node attributes immutable"
	FileImut       = "Node file immutable"
	ProvenanceImut = "Provenance info immutable"

	// Index errors
	InvalidIndex             = "Invalid index type"
	InvalidIndexType         = "Invalid index type specified"
	InvalidIndexRange        = "Invalid index record range"
	InvalidFileTypeForFilter = "Invalid file type for filter"
	InvalidFileTypeForNewIndex = "Invalid file type for new index"
	InvalidFileTypeForNewCopy  = "Invalid file type for new copy"
	InvalidIndexForFilter    = "Invalid index for filter"
	InvalidIndexNoAvail      = "Invalid index not available"
	IndexOutBounds           = "Index record out of bounds"
	IndexNoFile              = "Index file is missing"
	IndexExists              = "Index already exists"
	IndexNotFound            = "Index not found"
	IndexCreationFailed      = "Index creation failed"
	IndexNoAvail             = "Index not available"
	IndexTypeMismatch        = "Index type mismatch"

	// Node state
	NodeReferenced   = "Node referenced by virtual node"
	NodeDoesNotExist = "Node does not exist"
	NodeNotFound     = "Node not found"
	NodeNoFile       = "Node has no file"
	NodeExists       = "Node already exists"
	NodeFileLock     = "Node file is locked"
	NodeFileLocked   = "Node file is locked"
	NodeIndexLock    = "Node index is locked"
	NodeIndexLocked  = "Node index is locked"

	// Node create errors
	NodeCreateLockFailed   = "Node create lock failed"
	NodeCreateUploadFailed = "Node create upload failed"

	// Node update errors
	NodeUpdateLockFailed      = "Node update lock failed"
	NodeUpdateDeleteFailed    = "Node update delete failed"
	NodeUpdateAttributeFailed = "Node update attribute failed"
	NodeUpdateFormFileFailed  = "Node update form file failed"
	NodeUpdateCopyFileFailed  = "Node update copy file failed"
	NodeUpdateParseFailed     = "Node update parse failed"
	NodeUpdatePathFailed      = "Node update path failed"
	NodeUpdateSubsetFailed    = "Node update subset failed"
	NodeUpdateFileFailed      = "Node update file failed"
	NodeUpdateIndexFailed     = "Node update index failed"
	NodeUpdateLinkageFailed   = "Node update linkage failed"
	NodeUpdatePublishFailed   = "Node update publish failed"
	NodeUpdateUnpublishFailed = "Node update unpublish failed"
	NodeUpdateAclFailed       = "Node update ACL failed"

	// Node delete errors
	NodeDeleteLockFailed = "Node delete lock failed"
	NodeDeleteFailed     = "Node delete failed"

	// Node misc errors
	NodeDatabaseError = "Node database error"
	NodeStageFailed   = "Node stage failed"

	// MongoDB errors
	MongoDocNotFound = "MongoDB document not found"

	// Upload errors
	ExpiredUpload     = "Upload has expired"
	InvalidPartNumber = "Invalid part number"
	InvalidPartSize   = "Invalid part size"
	InvalidUpload     = "Invalid upload"
	PartNumberTooHigh = "Part number too high"
	TotalPartsExceeded = "Total parts exceeded"

	// Proxy errors
	ProxyRequest        = "Proxy request"
	ProxyRequestFailed  = "Proxy request failed"
	ProxyMisconfigured  = "Proxy misconfigured"
	ProxyNotFound       = "Proxy not found"
	ProxyForward        = "Proxy forward"

	// Invalid input errors
	InvalidConfig    = "Invalid configuration"
	InvalidFileSize  = "Invalid file size"
	InvalidNodeData  = "Invalid node data"
	InvalidNodeId    = "Invalid node ID"
	InvalidPart      = "Invalid part"
	InvalidUrlPassed = "Invalid URL passed"

	// Missing data errors
	MissingExpiration = "Missing expiration"
	MissingInit       = "Missing initialization"
	NoFileFound       = "No file found"

	// User errors
	NoUserNoLogin        = "User login not found"
	NoUserNoEmail        = "User email not found"
	NoUserExists         = "User does not exist"
	NoUserDeletedUser    = "User has been deleted"
	NoUserNoName         = "User name not found"
	NoUserNoPasswd       = "User password not found"
	NoUserPasswdMismatch = "User password mismatch"
	NoUserUpdateFailure  = "User update failure"
	NoUserCreateFailure  = "User create failure"
	NoUserDeleteFailure  = "User delete failure"
	NoUserNoId           = "User ID not found"
	NoUserUnAuth         = "User not authorized"
	UserResetFailed      = "User reset failed"
	UserUpdateFailed     = "User update failed"
	UserNotFound         = "User not found"
	UserNotAdmin         = "User is not admin"
	UserUnauthorized     = "User unauthorized"

	// Utility errors
	UnableToMakeDirectory  = "Unable to make directory"
	UnableToReadFile       = "Unable to read file"
	UnableToSendMail       = "Unable to send mail"
	UnableToSetExpiration  = "Unable to set expiration"
	UnableToUnmarshal      = "Unable to unmarshal"
	UnableToWriteFile      = "Unable to write file"
	UnsupportedMediaType   = "Unsupported media type"

	// Client errors
	ClientNotFound        = "Client not found"
	ClientNotActive       = "Client not active"
	ClientNoAuth          = "Client not authorized"
	ClientTokenGen        = "Client token generation failed"
	ClientNoEligibleTokens = "Client has no eligible tokens"
)
