package xml

import (
	"fmt"
	xml "github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/clbanning/x2j"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/objx"
	"reflect"
	"strconv"
	"strings"
)

const (
	OptionIncludeTypeAttributes string = "types"
)

var (
	Indentation                               string = "  "
	XMLDeclaration                            string = "<?xml version=\"1.0\"?>"
	XMLTagFormat                              string = "<%s>"
	XMLClosingTagFormat                       string = "</%s>"
	XMLElementFormat                          string = "<%s>%s</%s>"
	XMLElementFormatIndented                  string = "<%s>\n%s%s\n</%s>"
	XMLElementWithTypeAttributeFormat         string = "<%s type=\"%s\">%s</%s>"
	XMLElementWithTypeAttributeFormatIndented string = "<%s type=\"%s\">\n%s%s\n</%s>"
	XMLObjectElementName                      string = "object"
	XMLObjectsElementName                     string = "objects"
)

var validXmlContentTypes = []string{
	"text/xml",
	"application/xml",
}

// SimpleXmlCodec converts objects to and from simple XML.
type SimpleXmlCodec struct{}

// Marshal converts an object to a []byte representation.
// You can optionally pass additional arguments to further customize this call.
func (c *SimpleXmlCodec) Marshal(object interface{}, options map[string]interface{}) ([]byte, error) {

	var output []string

	// add the declaration
	output = append(output, XMLDeclaration)

	// add the rest of the XML
	bytes, err := marshal(object, true, 0, objx.New(options))

	if err != nil {
		return nil, err
	}

	output = append(output, string(bytes))

	// return the output
	return []byte(strings.Join(output, "")), nil
}

// Unmarshal converts a []byte representation into an object.
func (c *SimpleXmlCodec) Unmarshal(data []byte, obj interface{}) error {

	// check the value
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(obj)}
	}

	obj, err := unmarshal(string(data), nil)

	if err != nil {
		return err
	}

	// set the obj value
	rv.Elem().Set(reflect.ValueOf(obj))

	// no errors
	return nil

}

// ContentType gets the content type that this codec handles.
func (c *SimpleXmlCodec) ContentType() string {
	return constants.ContentTypeXML
}

// FileExtension returns the file extension by which this codec is represented.
func (c *SimpleXmlCodec) FileExtension() string {
	return constants.FileExtensionXML
}

// CanMarshalWithCallback indicates whether this codec is capable of marshalling a response with
// a callback parameter.
func (c *SimpleXmlCodec) CanMarshalWithCallback() bool {
	return false
}

func (c *SimpleXmlCodec) ContentTypeSupported(contentType string) bool {
	for _, supportedType := range validXmlContentTypes {
		if supportedType == contentType {
			return true
		}
	}
	return contentType == c.ContentType()
}

// unmarshal generates an object from the specified XML bytes.
func unmarshal(data string, options objx.Map) (interface{}, error) {

	m, err := xml.DocToMap(data)

	if err != nil {
		return nil, err
	}

	if object, ok := m[XMLObjectElementName]; ok {
		return resolveValues(object), nil
	} else if objects, ok := m[XMLObjectsElementName]; ok {
		return resolveValues(objects.(map[string]interface{})[XMLObjectElementName]), nil
	}

	return nil, nil

}

func resolveValues(object interface{}) interface{} {

	switch object.(type) {
	case map[string]interface{}:

		obj := object.(map[string]interface{})

		// one object
		for k, v := range obj {
			obj[k] = resolveValue(v)
		}

		return obj

	case []interface{}:

		objArr := object.([]interface{})

		for i, obj := range objArr {
			objArr[i] = resolveValues(obj)
		}

		return objArr

	}

	// we failed - just return it
	return object
}

func resolveValue(value interface{}) interface{} {

	switch value.(type) {
	case map[string]interface{}:

		valueMap := value.(map[string]interface{})

		if explicitType, ok := valueMap["-type"]; ok {

			switch explicitType {
			case "int":

				val, err := strconv.ParseInt(valueMap["#text"].(string), 10, 64)

				if err == nil {
					return val
				}

			case "bool":

				val, err := strconv.ParseBool(valueMap["#text"].(string))

				if err == nil {
					return val
				}

			case "float":

				val, err := strconv.ParseFloat(valueMap["#text"].(string), 64)

				if err == nil {
					return val
				}

			case "uint":

				val, err := strconv.ParseUint(valueMap["#text"].(string), 10, 64)

				if err == nil {
					return val
				}

			}

			return valueMap["#text"].(string)

		} else {

			// normal map - do each value too
			for k, v := range valueMap {
				valueMap[k] = resolveValue(v)
			}

		}

	}

	return value
}

/*
  Custom XML marshalling
*/

// marshal generates XML bytes from the specified object.
func marshal(object interface{}, doIndent bool, indentLevel int, options objx.Map) ([]byte, error) {

	var nextIndent int = indentLevel + 1
	var output []string

	switch object.(type) {
	case map[string]interface{}:

		var objects []string
		for k, v := range object.(map[string]interface{}) {

			valueBytes, valueMarshalErr := marshal(v, doIndent, nextIndent, options)

			// handle errors
			if valueMarshalErr != nil {
				return nil, valueMarshalErr
			}

			// add the key and value
			el := element(k, v, string(valueBytes), doIndent, nextIndent, options)
			objects = append(objects, el)

		}

		output = append(output, element(XMLObjectElementName, nil, strings.Join(objects, ""), doIndent, nextIndent, nil))

	case []map[string]interface{}:

		var objects []string
		for _, v := range object.([]map[string]interface{}) {

			valueBytes, err := marshal(v, doIndent, nextIndent, options)

			if err != nil {
				return nil, err
			}

			objects = append(objects, string(valueBytes))

		}

		el := strings.Join(objects, "")
		output = append(output, element(XMLObjectsElementName, nil, el, doIndent, nextIndent, nil))

	default:
		// return the value
		output = append(output, fmt.Sprintf("%v", object))
	}

	return []byte(strings.Join(output, "")), nil

}

func element(k string, v interface{}, vString string, doIndent bool, indentLevel int, options objx.Map) string {

	var typeString string
	if v != nil && options.Has(OptionIncludeTypeAttributes) {
		typeString = getTypeString(v)
	}

	if doIndent {
		indent := strings.Repeat(Indentation, indentLevel)

		if options.Has(OptionIncludeTypeAttributes) {
			return fmt.Sprintf(XMLElementWithTypeAttributeFormatIndented, k, typeString, indent, vString, k)
		} else {
			return fmt.Sprintf(XMLElementFormatIndented, k, indent, vString, k)
		}

	}

	if options.Has(OptionIncludeTypeAttributes) {
		return fmt.Sprintf(XMLElementWithTypeAttributeFormat, k, typeString, vString, k)
	} else {
		return fmt.Sprintf(XMLElementFormat, k, vString, k)
	}

}

// getTypeString gets a simple string describing the type of the object
// passed in.
//
// For simplicity sake, the type size is omitted.
//
// For example, all int types (int8, int16, int32, int64) will be represented
// as "int".
func getTypeString(obj interface{}) string {

	typeString := reflect.TypeOf(obj).Name()

	// trim off the numbers - no need to worry users about that level of
	// detail
	typeString = strings.TrimRight(typeString, "0123456789")

	return typeString

}
