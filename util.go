package nested

import (
	"reflect"
)

func isZeroValue(v interface{}) bool {
	return reflect.ValueOf(v).Interface() == reflect.Zero(reflect.TypeOf(v)).Interface()
}

func newNodePtrFromValue(value interface{}) Interface {
	v := reflect.ValueOf(value)
	return reflect.New(reflect.TypeOf(reflect.Indirect(v).Interface())).Interface().(Interface)
}

func doubleToSingleIndirect(v interface{}) interface{} {
	iv := reflect.Indirect(reflect.ValueOf(v))
	if iv.Kind() == reflect.Ptr {
		return iv.Interface()
	}

	return v
}

func isNilInterface(i interface{}) bool {
	if i == nil {
		return true
	}

	return reflect.ValueOf(i).IsNil()
}
