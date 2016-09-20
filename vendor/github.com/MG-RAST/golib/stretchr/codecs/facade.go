package codecs

import (
	"errors"
	"fmt"
	"github.com/MG-RAST/golib/stretchr/objx"
	"reflect"
)

const (
	// facadeMaxRecursionLevel is the maximum number of recursions it will make before
	// giving up and assuming circular recusion.
	facadeMaxRecursionLevel int = 100
)

var (
	// PublicDataDidNotFindMap is returned when the PublicData func fails to discover an appropriate
	// public data object, which must end up being a map[string]interface{}.
	PublicDataDidNotFindMap = errors.New("codecs: Object doesn't implement Facade interface and is not a Data object. PublicData(object) failed.")

	// PublicDataTooMuchRecursion is returned when there is too much recursion when
	// calling the Facade interfaces PublicData function.  The PublicData must return either another
	// object that implements Facade, or a map[string]interface{} that will be used for
	// public data.
	PublicDataTooMuchRecursion = errors.New("codecs: Facade object's PublicData() method caused too much recursion.  Does one of your PublicData funcs return itself?")
)

// Facade is the interface objects should implement if they
// want to be responsible for providing an alternative data object
// to the codecs.  Without this interface, the codecs will attempt to
// work on the object itself, whereas if an object implements this interface,
// the PublicData() method will be called instead, and the resulting object
// will instead be marshalled.
type Facade interface {

	// PublicData should return an object containing the
	// data to be marshalled.  If the method returns an error, the codecs
	// will send this error back to the calling code.
	//
	// The method may return either a final map[string]interface{} object,
	// or else another object that implements the Facade interface.
	//
	// The PublicData method should return a new object, and not the original
	// object, as it is possible that the response from PublicData will be modified
	// before being used, and it is bad practice for these methods to alter the
	// original data object.
	PublicData(options map[string]interface{}) (publicData interface{}, err error)
}

// PublicData gets the data that is considered public for the specified object.
// If the object implements the Facade interface, its PublicData method is called
// until the returning object no longer implements the Facade interface at which point
// it is considered to have returned the public data.
//
// If the object passed in is an array or slice, PublicData is called on each object
// to build up an array of public versions of the objects, and an array will be
// returned.
//
// If the resulting object is not of the appropriate type, the PublicDataDidNotFindMap error will
// be returned.
//
// If one of the PublicData methods returns itself (or another object already in the path)
// thus resulting in too much recursion, the PublicDataTooMuchRecursion error is returned.
//
// If any of the objects' PublicData() method returns an error, that is directly returned.
func PublicData(object interface{}, options map[string]interface{}) (interface{}, error) {
	return publicData(object, 0, options)
}

// PublicDataMap calls PublicData and returns the result after type asserting to objx.Map
func PublicDataMap(object interface{}, options map[string]interface{}) (objx.Map, error) {

	data, err := publicData(object, 0, options)

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	switch data.(type) {
	case map[string]interface{}:
		return objx.New(data.(map[string]interface{})), nil
	case objx.Map:
		return data.(objx.Map), nil
	default:
		if dataMap, ok := data.(objx.Map); ok {
			return dataMap, nil
		} else {
			panic(fmt.Sprintf("codecs: PublicDataMap must refer to a map[string]interface{} or objx.Map, not %s.  Did you mean to implement the Facade interface?", reflect.TypeOf(data)))
		}
	}
}

// publicData performs the work of PublicData keeping track of the level in order
// to ensure the code doesn't recurse too much.
func publicData(object interface{}, level int, options map[string]interface{}) (interface{}, error) {

	// make sure we don't end up with too much recusrion
	if level > facadeMaxRecursionLevel {
		return nil, PublicDataTooMuchRecursion
	}

	// if object is nil, that's OK - we'll just return nil
	if object == nil {
		return nil, nil
	}

	// handle arrays and slices - https://github.com/stretchr/goweb/issues/27
	objectValue := reflect.ValueOf(object)
	objectKind := objectValue.Kind()
	if objectKind == reflect.Array || objectKind == reflect.Slice {

		// make an array to hold the items
		length := objectValue.Len()
		arr := make([]interface{}, length)

		// get the public data for each item
		for subObjIndex := 0; subObjIndex < length; subObjIndex++ {

			// get this object
			subObj := objectValue.Index(subObjIndex).Interface()

			// ask for the object's public data
			subPublic, subPublicErr := publicData(subObj, level+1, options)

			// throw an error if there is one
			if subPublicErr != nil {
				return nil, subPublicErr
			}

			// add the item to the array
			arr[subObjIndex] = subPublic

		}

		// return the object
		return arr, nil

	}

	// cast the object
	if objectWithFacade, ok := object.(Facade); ok {

		publicObject, err := objectWithFacade.PublicData(options)

		// return the public data error if there was one
		if err != nil {
			return nil, err
		}

		// recursivly call publicData until the object no longer
		// implements the Facade interface.
		return publicData(publicObject, level+1, options)
	}

	// we can't do anything - so just return the object back
	return object, nil
}
