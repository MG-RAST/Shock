package csv

import (
	"fmt"
	"github.com/MG-RAST/golib/stretchr/codecs"
	"github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/golib/stretchr/objx"
	"github.com/MG-RAST/golib/stretchr/testify/assert"
	"log"
	"reflect"
	"testing"
)

func TestInterface(t *testing.T) {

	assert.Implements(t, (*codecs.Codec)(nil), new(CsvCodec), "CsvCodec")

}

func TestMapFromFieldsAndRow(t *testing.T) {

	fields := []string{"field1", "field2", "field3"}
	row := []string{"one", "two", "three"}

	m, err := mapFromFieldsAndRow(fields, row)

	if assert.NoError(t, err) && assert.NotNil(t, m) {

		assert.Equal(t, "one", m["field1"])
		assert.Equal(t, "two", m["field2"])
		assert.Equal(t, "three", m["field3"])

	}

}

func TestMarshal_SingleObject(t *testing.T) {

	obj := map[string]interface{}{"field1": "one", "field2": "two", "field3": "three"}

	csvCodec := new(CsvCodec)
	bytes, marshalErr := csvCodec.Marshal(obj, nil)

	if assert.NoError(t, marshalErr) {

		assert.Equal(t, "field1,field2,field3\n\"\"\"one\"\"\",\"\"\"two\"\"\",\"\"\"three\"\"\"\n", string(bytes))

	}

}

func TestMarshal_MultipleObjects(t *testing.T) {

	arr := make([]map[string]interface{}, 3)
	arr[0] = map[string]interface{}{"field1": "oneA", "field2": "twoA", "field3": "threeA"}
	arr[1] = map[string]interface{}{"field1": "oneB", "field2": "twoB", "field3": "threeB"}
	arr[2] = map[string]interface{}{"field1": "oneC", "field2": "twoC", "field3": "threeC"}

	csvCodec := new(CsvCodec)
	bytes, marshalErr := csvCodec.Marshal(arr, nil)

	if assert.NoError(t, marshalErr) {

		assert.Equal(t, "field1,field2,field3\n\"\"\"oneA\"\"\",\"\"\"twoA\"\"\",\"\"\"threeA\"\"\"\n\"\"\"oneB\"\"\",\"\"\"twoB\"\"\",\"\"\"threeB\"\"\"\n\"\"\"oneC\"\"\",\"\"\"twoC\"\"\",\"\"\"threeC\"\"\"\n", string(bytes))

	}

}

func TestMarshal_MultipleObjects_WithDisimilarSchema(t *testing.T) {

	arr := make([]map[string]interface{}, 3)
	arr[0] = map[string]interface{}{"name": "Mat", "age": 30, "language": "en"}
	arr[1] = map[string]interface{}{"first_name": "Tyler", "age": 28, "last_name": "Bunnell"}
	arr[2] = map[string]interface{}{"name": "Ryan", "age": 26, "speaks": "english"}

	csvCodec := new(CsvCodec)
	bytes, marshalErr := csvCodec.Marshal(arr, nil)

	if assert.NoError(t, marshalErr) {

		assert.Equal(t, "name,age,language,first_name,last_name,speaks\n\"\"\"Mat\"\"\",30,\"\"\"en\"\"\",\"\",\"\",\"\"\n\"\",28,\"\",\"\"\"Tyler\"\"\",\"\"\"Bunnell\"\"\",\"\"\n\"\"\"Ryan\"\"\",26,\"\",\"\",\"\",\"\"\"english\"\"\"\n", string(bytes))

	}

}

func TestMarshal_ComplexMap(t *testing.T) {

	obj1 := map[string]interface{}{"name": "Mat", "age": 30, "language": "en"}
	obj2 := map[string]interface{}{"obj": obj1}
	obj3 := map[string]interface{}{"another_obj": obj2}

	csvCodec := new(CsvCodec)
	bytes, marshalErr := csvCodec.Marshal(obj3, nil)

	if assert.NoError(t, marshalErr) {

		assert.Equal(t, "another_obj\n\"{\"\"obj\"\":{\"\"age\"\":30,\"\"language\"\":\"\"en\"\",\"\"name\"\":\"\"Mat\"\"}}\"\n", string(bytes))

	}

}

func TestUnmarshal_SingleObject(t *testing.T) {

	raw := "field_a,field_b,field_c\nrow1a,row1b,row1c\n"

	csvCodec := new(CsvCodec)

	var obj interface{}
	csvCodec.Unmarshal([]byte(raw), &obj)

	if assert.NotNil(t, obj, "Unmarshal should make an object") {
		if object, ok := obj.(map[string]interface{}); ok {

			assert.Equal(t, "row1a", object["field_a"])
			assert.Equal(t, "row1b", object["field_b"])
			assert.Equal(t, "row1c", object["field_c"])

		} else {
			t.Errorf("Expected to be array type, not %s.", reflect.TypeOf(obj).Elem().Name())
		}
	}

}

func TestUnmarshal_SingleObject_WithNoEndLinefeed(t *testing.T) {

	raw := "field_a,field_b,field_c\nrow1a,row1b,row1c"

	csvCodec := new(CsvCodec)

	var obj interface{}
	csvCodec.Unmarshal([]byte(raw), &obj)

	if assert.NotNil(t, obj, "Unmarshal should make an object") {
		if object, ok := obj.(map[string]interface{}); ok {

			assert.Equal(t, "row1a", object["field_a"])
			assert.Equal(t, "row1b", object["field_b"])
			assert.Equal(t, "row1c", object["field_c"])

		} else {
			t.Errorf("Expected to be array type, not %s.", reflect.TypeOf(obj).Elem().Name())
		}
	}

}

func TestUnmarshal_MultipleObjects(t *testing.T) {

	raw := "field_a,field_b,field_c\nrow1a,row1b,row1c\nrow2a,row2b,row2c\nrow3a,row3b,row3c"

	csvCodec := new(CsvCodec)

	var obj interface{}
	csvCodec.Unmarshal([]byte(raw), &obj)

	if assert.NotNil(t, obj, "Unmarshal should make an object") {
		if array, ok := obj.([]interface{}); ok {

			if assert.Equal(t, 3, len(array), "Should be 3 items") {

				assert.Equal(t, "row1a", array[0].(map[string]interface{})["field_a"])
				assert.Equal(t, "row1b", array[0].(map[string]interface{})["field_b"])
				assert.Equal(t, "row1c", array[0].(map[string]interface{})["field_c"])

				assert.Equal(t, "row2a", array[1].(map[string]interface{})["field_a"])
				assert.Equal(t, "row2b", array[1].(map[string]interface{})["field_b"])
				assert.Equal(t, "row2c", array[1].(map[string]interface{})["field_c"])

				assert.Equal(t, "row3a", array[2].(map[string]interface{})["field_a"])
				assert.Equal(t, "row3b", array[2].(map[string]interface{})["field_b"])
				assert.Equal(t, "row3c", array[2].(map[string]interface{})["field_c"])

			}

		} else {
			t.Errorf("Expected to be array type, not %s.", reflect.TypeOf(obj).Elem().Name())
		}
	}

}

func TestUnMarshal_ObjxMap(t *testing.T) {

	obj1 := objx.MSI("name", "Mat", "age", 30, "language", "en")
	obj2 := objx.MSI("obj", obj1)
	obj3 := objx.MSI("another_obj", obj2)

	csvCodec := new(CsvCodec)
	bytes, _ := csvCodec.Marshal(obj3, nil)

	log.Printf("bytes = %s", string(bytes))

	// unmarshal it back
	var obj interface{}
	csvCodec.Unmarshal(bytes, &obj)

	if objmap, ok := obj.(map[string]interface{}); ok {
		if objmap2, ok := objmap["another_obj"].(map[string]interface{}); ok {
			if objmap3, ok := objmap2["obj"].(map[string]interface{}); ok {

				assert.Equal(t, "Mat", objmap3["name"])
				assert.Equal(t, 30, objmap3["age"])
				assert.Equal(t, "en", objmap3["language"])

			} else {
				assert.True(t, false, "another_obj.obj should be msi")
			}
		} else {
			assert.True(t, false, "another_obj should be msi")
		}
	} else {
		assert.True(t, false, "obj should be msi")
	}

}

func TestUnMarshal_ComplexMap(t *testing.T) {

	obj1 := map[string]interface{}{"name": "Mat", "age": 30, "language": "en"}
	obj2 := map[string]interface{}{"obj": obj1}
	obj3 := map[string]interface{}{"another_obj": obj2}

	csvCodec := new(CsvCodec)
	bytes, _ := csvCodec.Marshal(obj3, nil)

	// unmarshal it back
	var obj interface{}
	csvCodec.Unmarshal(bytes, &obj)

	if objmap, ok := obj.(map[string]interface{}); ok {
		if objmap2, ok := objmap["another_obj"].(map[string]interface{}); ok {
			if objmap3, ok := objmap2["obj"].(map[string]interface{}); ok {

				assert.Equal(t, "Mat", objmap3["name"])
				assert.Equal(t, 30, objmap3["age"])
				assert.Equal(t, "en", objmap3["language"])

			} else {
				assert.True(t, false, "another_obj.obj should be msi")
			}
		} else {
			assert.True(t, false, "another_obj should be msi")
		}
	} else {
		assert.True(t, false, "obj should be msi")
	}

}

func getMarshalValue(v interface{}) string {
	s, e := marshalValue(v)
	if e != nil {
		panic(fmt.Sprintf("Failed to marshal: %v (%s)", v, e))
	}
	return string(s)
}

func getUnmarshalValue(s string) interface{} {
	obj, e := unmarshalValue(s)
	if e != nil {
		panic(fmt.Sprintf("Failed to unmarshal: %v (%s)", s, e))
	}
	return obj
}

func TestMarshalValue(t *testing.T) {

	assert.Equal(t, "\"str\"", getMarshalValue("str"))
	assert.Equal(t, "18", getMarshalValue(18))
	assert.Equal(t, "true", getMarshalValue(true))

}

func TestUnmarshalValue(t *testing.T) {

	assert.Equal(t, "str", getUnmarshalValue("\"str\""))
	assert.Equal(t, "str", getUnmarshalValue("str"))
	assert.Equal(t, 18, getUnmarshalValue("18"))
	assert.Equal(t, true, getUnmarshalValue("true"))

}

func TestResponseContentType(t *testing.T) {

	codec := new(CsvCodec)
	assert.Equal(t, codec.ContentType(), constants.ContentTypeCSV)

}

func TestFileExtension(t *testing.T) {

	codec := new(CsvCodec)
	assert.Equal(t, constants.FileExtensionCSV, codec.FileExtension())

}

func TestCanMarshalWithCallback(t *testing.T) {

	codec := new(CsvCodec)
	assert.False(t, codec.CanMarshalWithCallback())

}
