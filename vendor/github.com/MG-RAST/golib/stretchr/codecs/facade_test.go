package codecs

import (
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/constants"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/codecs/test"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/objx"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/testify/assert"
	"github.com/MG-RAST/Shock/vendor/github.com/MG-RAST/golib/stretchr/testify/mock"
	"reflect"
	"testing"
)

/*
	Tests
*/

func TestPublicData(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o.Mock.On("PublicData", map[string]interface{}{}).Return(objx.New(map[string]interface{}{"theName": "Mat"}), nil)

	public, err := PublicData(o, map[string]interface{}{})

	if assert.Nil(t, err) {
		assert.Equal(t, public.(objx.Map).Get("theName").Str(), "Mat")
	}

	mock.AssertExpectationsForObjects(t, o.Mock)

}

func TestPublicDataMap(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o.Mock.On("PublicData", map[string]interface{}{}).Return(objx.New(map[string]interface{}{"theName": "Mat"}), nil)

	public, err := PublicDataMap(o, map[string]interface{}{})

	if assert.Nil(t, err) {
		assert.Equal(t, public.Get("theName").Str(), "Mat")
	}

	mock.AssertExpectationsForObjects(t, o.Mock)

}

func TestPublicDataMap_WithMSI(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o.Mock.On("PublicData", map[string]interface{}{}).Return(map[string]interface{}{"theName": "Mat"}, nil)

	public, err := PublicDataMap(o, map[string]interface{}{})

	if assert.Nil(t, err) {
		assert.Equal(t, public.Get("theName").Str(), "Mat")
	}

	mock.AssertExpectationsForObjects(t, o.Mock)

}

func TestPublicDataMap_ReturningNil(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o.Mock.On("PublicData", map[string]interface{}{}).Return(nil, nil)

	public, err := PublicDataMap(o, map[string]interface{}{})

	if assert.Nil(t, err) {
		assert.Nil(t, public)
	}

	mock.AssertExpectationsForObjects(t, o.Mock)

}

func TestPublicData_WithArray(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o1 := new(test.TestObjectWithFacade)
	o2 := new(test.TestObjectWithFacade)

	arr := []interface{}{o, o1, o2}

	o.Mock.On("PublicData", map[string]interface{}{}).Return(objx.New(map[string]interface{}{"theName": "1"}), nil)
	o1.Mock.On("PublicData", map[string]interface{}{}).Return(objx.New(map[string]interface{}{"theName": "2"}), nil)
	o2.Mock.On("PublicData", map[string]interface{}{}).Return(objx.New(map[string]interface{}{"theName": "3"}), nil)

	public, err := PublicData(arr, map[string]interface{}{})

	if assert.Nil(t, err) {
		assert.Equal(t, reflect.Slice, reflect.TypeOf(public).Kind(), "Result should be array not %v", reflect.TypeOf(public))
	}

	mock.AssertExpectationsForObjects(t, o.Mock, o1.Mock, o2.Mock)

	publicArray := public.([]interface{})
	if assert.Equal(t, 3, len(publicArray)) {
		assert.Equal(t, publicArray[0].(objx.Map).Get("theName").Str(), "1", "o")
		assert.Equal(t, publicArray[1].(objx.Map).Get("theName").Str(), "2", "o1")
		assert.Equal(t, publicArray[2].(objx.Map).Get("theName").Str(), "3", "o2")
	}

}

func TestPublicData_WithNil(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o.Mock.On("PublicData", map[string]interface{}{}).Return(nil, nil)

	public, err := PublicData(o, map[string]interface{}{})

	if assert.Nil(t, err) {
		assert.Nil(t, public, "Nil is OK")
	}

	mock.AssertExpectationsForObjects(t, o.Mock)

}

func TestPublicData_WithError(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o.Mock.On("PublicData", map[string]interface{}{}).Return(nil, assert.AnError)

	_, err := PublicData(o, map[string]interface{}{})

	assert.Equal(t, assert.AnError, err)
	mock.AssertExpectationsForObjects(t, o.Mock)

}

func TestPublicData_WithRecursion(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o1 := new(test.TestObjectWithFacade)
	o2 := new(test.TestObjectWithFacade)

	o.Mock.On("PublicData", map[string]interface{}{}).Return(o1, nil)
	o1.Mock.On("PublicData", map[string]interface{}{}).Return(o2, nil)
	o2.Mock.On("PublicData", map[string]interface{}{}).Return(objx.New(map[string]interface{}{"theName": "Mat"}), nil)

	public, err := PublicData(o, map[string]interface{}{})

	if assert.Nil(t, err) {
		assert.Equal(t, public.(objx.Map).Get("theName").Str(), "Mat")
	}

	mock.AssertExpectationsForObjects(t, o.Mock, o1.Mock, o2.Mock)

}

func TestPublicData_WithRecursion_WithObjects(t *testing.T) {

	o := new(test.TestObjectWithFacade)
	o1 := new(test.TestObjectWithFacade)
	o2 := new(test.TestObjectWithFacade)

	args := map[string]interface{}{constants.OptionKeyClientCallback: "~d"}

	o.Mock.On("PublicData", args).Return(o1, nil)
	o1.Mock.On("PublicData", args).Return(o2, nil)
	o2.Mock.On("PublicData", args).Return(objx.New(map[string]interface{}{"theName": "Mat"}), nil)

	public, err := PublicData(o, args)

	if assert.Nil(t, err) {
		assert.Equal(t, public.(objx.Map).Get("theName").Str(), "Mat")
	}

	mock.AssertExpectationsForObjects(t, o.Mock, o1.Mock, o2.Mock)

}
