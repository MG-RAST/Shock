package services

import (
	"strings"
)

// ContentType represents a single content type, complete with
// parameters, such as that passed in an HTTP Accept or Content-Type
// header.
type ContentType struct {
	MimeType   string
	Parameters map[string]string
}

func (contentType *ContentType) AddParam(param string) {
	equal := strings.IndexRune(param, '=')
	if equal == -1 {
		return
	}
	name := strings.TrimSpace(param[:equal])
	value := strings.TrimSpace(param[equal+1:])
	contentType.Parameters[name] = value
}

// ParseContentType takes a content-type string and parses it into a
// mimetype and parameters, returning the ContentType representing the
// string.
func ParseContentType(rawType string) (*ContentType, error) {
	rawType = strings.TrimSpace(strings.ToLower(rawType))
	if len(rawType) == 0 {
		return nil, nil
	}
	var (
		end       int
		value     string
		remaining = rawType
	)
	contentType := new(ContentType)
	// Much the same as in accept.go in OrderAcceptHeader,
	// strings.Split was slowing this process down by a decent
	// percentage.  Using IndexRune seems to be much faster.
	for end != -1 {
		end = strings.IndexRune(remaining, ';')
		if end == -1 {
			value = remaining
		} else {
			value = remaining[:end]
		}
		if remaining == rawType {
			contentType.MimeType = value
			if end != -1 {
				contentType.Parameters = make(map[string]string)
			}
		} else {
			contentType.AddParam(value)
		}
		remaining = remaining[end+1:]
	}
	return contentType, nil
}
