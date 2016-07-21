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

func SetInt64(i interface{}, name string, v int64) {
	r := reflect.ValueOf(i).Elem()
	t := r.FieldByName(name)
	t.SetInt(v)
}

func TargetFieldIsArray(i interface{}, name string) bool {

	return TypeOfField(i, name).Kind() == reflect.Array

}
