package nested

import (
	"github.com/jinzhu/gorm"
	"reflect"
)

func isZeroValue(v interface{}) bool {
	return reflect.ValueOf(v).Interface() == reflect.Zero(reflect.TypeOf(v)).Interface()
}

func newNodePtrFromValue(value interface{}) Interface {
	v := reflect.ValueOf(value)
	return reflect.New(reflect.TypeOf(reflect.Indirect(v).Interface())).Interface().(Interface)
}

func isTreeNode(v interface{}) bool {
	_, ok := v.(Interface)

	return ok
}

func isUpdateIgnored(scope *gorm.Scope) bool {
	v, ok := scope.Get(settingIgnoreUpdate)
	if !ok {
		return false
	}

	vv, _ := v.(bool)

	return vv
}

func isDeletionIgnored(scope *gorm.Scope) bool {
	v, ok := scope.Get(settingIgnoreDelete)
	if !ok {
		return false
	}

	vv, _ := v.(bool)

	return vv
}

func doubleToSingleIndirect(v interface{}) interface{} {
	iv := reflect.Indirect(reflect.ValueOf(v))
	if iv.Kind() == reflect.Ptr {
		return iv.Interface()
	}

	return v
}

func isNilInterface(i interface{}) bool {
	if i == nil  {
		return true
	}

	return reflect.ValueOf(i).IsNil()
}
