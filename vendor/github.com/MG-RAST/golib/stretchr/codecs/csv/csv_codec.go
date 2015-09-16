package csv

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/objx"
	"reflect"
	"strings"
)

var validCsvContentTypes = []string{
	"application/csv",
	"text/csv",
}

// CsvCodec converts objects to and from CSV format.
type CsvCodec struct{}

// Converts an object to CSV data.
func (c *CsvCodec) Marshal(object interface{}, options map[string]interface{}) ([]byte, error) {

	// collect the data rows in a consistent type

	dataRows := make([]map[string]interface{}, 0)

	switch object.(type) {
	case objx.Map:
		dataRows = append(dataRows, object.(objx.Map).Value().ObjxMap())
	case map[string]interface{}:
		dataRows = append(dataRows, object.(map[string]interface{}))
	case []map[string]interface{}:
		dataRows = object.([]map[string]interface{})
	case []objx.Map:
		for _, item := range object.([]objx.Map) {
			dataRows = append(dataRows, item.Value().ObjxMap())
		}
	case []interface{}:
		for _, item := range object.([]interface{}) {
			dataRows = append(dataRows, item.(map[string]interface{}))
		}
	}

	// collect the fields
	var fields []string
	for _, m := range dataRows {

		// for each field
		for k, _ := range m {

			shouldAdd := true
			for _, field := range fields {
				if strings.ToLower(field) == strings.ToLower(k) {
					shouldAdd = false
					break
				}
			}

			if shouldAdd {
				// add this new field
				fields = append(fields, k)
			}

		}

	}

	// make a new CSV writer
	byteBuffer := new(bytes.Buffer)
	writer := csv.NewWriter(byteBuffer)

	// write the fields
	writer.Write(fields)

	// now write the data
	for _, row := range dataRows {

		rowData := make([]string, len(fields))

		// do it each field at a time
		for k, v := range row {

			// find the field index
			var fieldIndex int
			for index, f := range fields {
				if strings.ToLower(f) == strings.ToLower(k) {
					fieldIndex = index
					break
				}
			}

			// set the field
			str, strErr := marshalValue(v)

			if strErr != nil {
				return nil, strErr
			}

			rowData[fieldIndex] = string(str)

		}

		// write the row
		writer.Write(rowData)

	}

	// finish writing
	writer.Flush()
	if writer.Error() != nil {
		return nil, writer.Error()
	}

	return byteBuffer.Bytes(), nil
}

// Unmarshal converts CSV data into an object.
func (c *CsvCodec) Unmarshal(data []byte, obj interface{}) error {

	// check the value
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(obj)}
	}

	reader := csv.NewReader(bytes.NewReader(data))
	records, readErr := reader.ReadAll()

	if readErr != nil {
		return readErr
	}

	lenRecords := len(records)

	if lenRecords == 0 {

		// no records
		return nil

	} else if lenRecords == 1 {

		// no records (first line should be header)
		return nil

	} else if lenRecords == 2 {

		// one record

		// get the object
		object, err := mapFromFieldsAndRow(records[0], records[1])

		if err != nil {
			return err
		}

		// set the obj value
		rv.Elem().Set(reflect.ValueOf(object))

	} else {

		// multiple records

		// make a new array to hold the data
		rows := make([]interface{}, lenRecords-1)

		// collect the fields
		fields := records[0]

		// add each row
		var err error
		for i := 1; i < lenRecords; i++ {

			rows[i-1], err = mapFromFieldsAndRow(fields, records[i])

			if err != nil {
				return err
			}

		}

		// set the obj value
		rv.Elem().Set(reflect.ValueOf(rows))

	}

	return nil
}

// ContentType returns the content type for this codec.
func (c *CsvCodec) ContentType() string {
	return constants.ContentTypeCSV
}

// FileExtension returns the file extension for this codec.
func (c *CsvCodec) FileExtension() string {
	return constants.FileExtensionCSV
}

// CanMarshalWithCallback returns whether this codec is capable of marshalling a response containing a callback.
func (c *CsvCodec) CanMarshalWithCallback() bool {
	return false
}

func (c *CsvCodec) ContentTypeSupported(contentType string) bool {
	for _, supportedType := range validCsvContentTypes {
		if supportedType == contentType {
			return true
		}
	}
	return contentType == c.ContentType()
}

// mapFromFieldsAndRow makes a map[string]interface{} from the given fields and
// row data.
func mapFromFieldsAndRow(fields, row []string) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	for index, item := range row {

		value, unmarshalErr := unmarshalValue(item)

		if unmarshalErr != nil {
			return nil, unmarshalErr
		}

		m[fields[index]] = value
	}

	return m, nil
}

// marshalValue generates a string of the specified object by
// using the JSON encoding capabilities.  If it fails, the object
// is just returned as a raw string.
func marshalValue(obj interface{}) (string, error) {
	s, e := json.Marshal(obj)
	if e != nil {
		return fmt.Sprintf("%v", s), nil
	}
	return string(s), nil
}

// unmarshalValue creates an object from the specified string by
// using the JSON encoding capabilities.  If it fails, the raw value is
// returned as a string.
func unmarshalValue(value string) (interface{}, error) {

	var obj interface{}
	err := json.Unmarshal([]byte(value), &obj)

	if err != nil {
		return value, nil
	}

	return obj, nil

}
