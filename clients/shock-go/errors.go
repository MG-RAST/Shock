package shock

import (
	"fmt"
	"strings"
)

// APIError represents an error returned by the Shock API.
type APIError struct {
	StatusCode int
	Errors     []string
}

func (e *APIError) Error() string {
	if len(e.Errors) > 0 {
		return fmt.Sprintf("shock API error (status %d): %s", e.StatusCode, strings.Join(e.Errors, "; "))
	}
	return fmt.Sprintf("shock API error (status %d)", e.StatusCode)
}
