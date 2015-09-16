package csv

import (
	"reflect"
)

// An InvalidUnmarshalError describes an invalid argument passed to Unmarshal.
// (The argument to Unmarshal must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "codecs: csv: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "codecs: csv: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "codecs: csv: Unmarshal(nil " + e.Type.String() + ")"
}
