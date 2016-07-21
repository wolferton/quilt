package reflecttools

import (
	"reflect"
)

func HasFieldOfName(i interface{}, fieldName string) bool {
	r := reflect.ValueOf(i).Elem()
	f := r.FieldByName(fieldName)

	return f.IsValid()
}

func TypeOfField(i interface{}, name string) reflect.Type {
	r := reflect.ValueOf(i).Elem()
	return r.FieldByName(name).Type()
}
