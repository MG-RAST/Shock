package xml

import (
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/objx"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/testify/assert"
	"testing"
)

var xmlCodec SimpleXmlCodec

func TestInterface(t *testing.T) {

	assert.Implements(t, (*codecs.Codec)(nil), new(SimpleXmlCodec), "XmlCodec")

}

func TestCanMarshalWithCallback(t *testing.T) {
	assert.False(t, xmlCodec.CanMarshalWithCallback(), "SimpleXmlCodec cannot marshal with callback")
}

func TestContentType(t *testing.T) {
	assert.Equal(t, constants.ContentTypeXML, xmlCodec.ContentType())
}

func TestExtension(t *testing.T) {
	assert.Equal(t, constants.FileExtensionXML, xmlCodec.FileExtension())
}

func TestMarshalAndUnmarshal(t *testing.T) {

	// make a big object
	obj := map[string]interface{}{}
	obj["name"] = "Mat"
	obj["age"] = 30
	obj["address"] = map[string]interface{}{
		"street":  "Pearl Street",
		"city":    "Boulder",
		"state":   "CO",
		"country": "USA",
	}
	obj["animals"] = map[string]interface{}{
		"favourite": []string{"Dog", "Cat"},
	}

	bytes, marshalErr := xmlCodec.Marshal(obj, nil)

	if assert.NoError(t, marshalErr) {
		assert.Contains(t, string(bytes), "<?xml version=\"1.0\"?>", "Output")
	}

	// unmarshal it
	var newObj interface{}
	if assert.NoError(t, xmlCodec.Unmarshal(bytes, &newObj)) {

		assert.NotNil(t, newObj)

	}

}

func TestMarshal_map(t *testing.T) {

	data := map[string]interface{}{"name": "Mat", "age": 30, "yesOrNo": true}
	bytes, marshalErr := marshal(data, false, 0, nil)

	if assert.NoError(t, marshalErr) {
		assert.Equal(t, "<object><name>Mat</name><age>30</age><yesOrNo>true</yesOrNo></object>", string(bytes), "Output")
	}

}

func TestMarshal_mapWithTypes(t *testing.T) {

	data := map[string]interface{}{"name": "Mat", "age": 30, "yesOrNo": true}
	options := objx.MSI(OptionIncludeTypeAttributes, true)
	bytes, marshalErr := marshal(data, false, 0, options)

	if assert.NoError(t, marshalErr) {
		assert.Equal(t, "<object><name type=\"string\">Mat</name><age type=\"int\">30</age><yesOrNo type=\"bool\">true</yesOrNo></object>", string(bytes), "Output")
	}

}

func TestMarshal_arrayOfMaps(t *testing.T) {

	data1 := map[string]interface{}{"name": "Mat"}
	data2 := map[string]interface{}{"name": "Tyler"}
	data3 := map[string]interface{}{"name": "Ryan"}
	array := []map[string]interface{}{data1, data2, data3}
	bytes, marshalErr := marshal(array, false, 0, nil)

	if assert.NoError(t, marshalErr) {
		assert.Equal(t, "<objects><object><name>Mat</name></object><object><name>Tyler</name></object><object><name>Ryan</name></object></objects>", string(bytes), "Output")
	}

}

func TestUnmarshal_map(t *testing.T) {

	xml := `<object><name>Mat</name><age type='int'>30</age><yesOrNo type='bool'>true</yesOrNo><address><city>Boulder</city><state>CO</state></address></object>`

	obj, err := unmarshal(xml, nil)

	if assert.NoError(t, err) {
		if assert.NotNil(t, obj) {

			o := obj.(map[string]interface{})

			assert.Equal(t, "Mat", o["name"])
			assert.Equal(t, 30, o["age"])
			assert.Equal(t, true, o["yesOrNo"])

		}
	}

}

func TestResolveValue(t *testing.T) {

	assert.Equal(t, "Hello", resolveValue("Hello"))
	assert.Equal(t, 30, resolveValue(map[string]interface{}{"-type": "int", "#text": "30"}))
	assert.Equal(t, 30.5, resolveValue(map[string]interface{}{"-type": "float", "#text": "30.5"}))
	assert.Equal(t, "30", resolveValue(map[string]interface{}{"-type": "string", "#text": "30"}))
	assert.Equal(t, true, resolveValue(map[string]interface{}{"-type": "bool", "#text": "true"}))
	assert.Equal(t, "true", resolveValue(map[string]interface{}{"-type": "true", "#text": "true"}))

}

func TestResolveValues_SingleObject(t *testing.T) {

	m1 := map[string]interface{}{"name": "Mat", "age": map[string]interface{}{"-type": "int", "#text": "30"}}

	m := resolveValues(m1)

	assert.Equal(t, m.(map[string]interface{})["name"], "Mat")
	assert.Equal(t, m.(map[string]interface{})["age"], 30)

}

func TestResolveValues_MultipleObjects(t *testing.T) {

	m1 := map[string]interface{}{"name": "Mat", "age": map[string]interface{}{"-type": "int", "#text": "30"}}
	m2 := map[string]interface{}{"name": "Tyler", "english": map[string]interface{}{"-type": "bool", "#text": "false"}}
	m3 := map[string]interface{}{"name": "Ryan", "weight": map[string]interface{}{"-type": "float", "#text": "180.22"}}

	m := resolveValues([]interface{}{m1, m2, m3}).([]interface{})

	assert.Equal(t, m[0].(map[string]interface{})["name"], "Mat")
	assert.Equal(t, m[0].(map[string]interface{})["age"], 30)
	assert.Equal(t, m[1].(map[string]interface{})["name"], "Tyler")
	assert.Equal(t, m[1].(map[string]interface{})["english"], false)
	assert.Equal(t, m[2].(map[string]interface{})["name"], "Ryan")
	assert.Equal(t, m[2].(map[string]interface{})["weight"], 180.22)

}

func TestUnmarshal_arrayOfMaps(t *testing.T) {

	xml := `<objects><object><name>Mat</name><age type="int">30</age><yesOrNo type="bool">true</yesOrNo><address><city>Boulder</city><state>CO</state></address></object><object><name>Tyler</name><age type="int">28</age><yesOrNo type="bool">false</yesOrNo><address><city>Salt Lake City</city><state>UT</state></address></object></objects>`

	obj, err := unmarshal(xml, nil)

	if assert.NoError(t, err) {
		if assert.NotNil(t, obj) {

			os := obj.([]interface{})

			o1 := os[0].(map[string]interface{})
			assert.Equal(t, "Mat", o1["name"])
			assert.Equal(t, 30, o1["age"])
			assert.Equal(t, true, o1["yesOrNo"])

			o2 := os[1].(map[string]interface{})
			assert.Equal(t, "Tyler", o2["name"])
			assert.Equal(t, 28, o2["age"])
			assert.Equal(t, false, o2["yesOrNo"])

		}
	}

}

func TestGetTypeString(t *testing.T) {

	assert.Equal(t, "string", getTypeString("Hello"))
	assert.Equal(t, "int", getTypeString(10))
	assert.Equal(t, "int", getTypeString(int8(10)))
	assert.Equal(t, "int", getTypeString(int16(10)))
	assert.Equal(t, "uint", getTypeString(uint64(10)))
	assert.Equal(t, "uint", getTypeString(uint16(10)))
	assert.Equal(t, "uint", getTypeString(uint8(10)))
	assert.Equal(t, "bool", getTypeString(true))
	assert.Equal(t, "bool", getTypeString(false))
	assert.Equal(t, "float", getTypeString(10.2))
	assert.Equal(t, "float", getTypeString(10.23294))

}
