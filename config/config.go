package config

import (
	"github.com/wolferton/quilt/logging"
	"reflect"
	"strings"
)

const JsonPathSeparator string = "."

const (
	JsonUnknown = -1
	JsonInt     = 0
	JsonString  = 1
	JsonArray   = 2
	JsonMap     = 3
	JsonBool    = 4
)

type ConfigValue interface{}

type ConfigAccessor struct {
	JsonData        map[string]interface{}
	FrameworkLogger logging.Logger
}

func (c *ConfigAccessor) PathExists(path string) bool {
	value := c.Value(path)

	return value != nil
}

func (c *ConfigAccessor) Value(path string) ConfigValue {

	splitPath := strings.Split(path, JsonPathSeparator)

	return c.configValue(splitPath, c.JsonData)

}

func (c *ConfigAccessor) ObjectVal(path string) map[string]interface{} {

	value := c.Value(path)

	if value == nil {
		return nil
	} else {
		return value.(map[string]interface{})
	}
}

func (c *ConfigAccessor) StringVal(path string) string {
	return c.Value(path).(string)
}

func (c *ConfigAccessor) StringFieldVal(field string, object map[string]interface{}) string {
	return object[field].(string)
}

func (c *ConfigAccessor) IntValue(path string) int {
	return int(c.Value(path).(float64))
}

func (c *ConfigAccessor) Float64Value(path string) float64 {
	return c.Value(path).(float64)
}

func (c *ConfigAccessor) Array(path string) []interface{} {

	value := c.Value(path)

	if value == nil {
		return nil
	} else {
		return c.Value(path).([]interface{})
	}
}

func (c *ConfigAccessor) BoolValue(path string) bool {
	return c.Value(path).(bool)
}

func JsonType(value interface{}) int {

	switch value.(type) {
	case string:
		return JsonString
	case map[string]interface{}:
		return JsonMap
	case bool:
		return JsonBool
	default:
		return JsonUnknown
	}
}

func (c *ConfigAccessor) configValue(path []string, jsonMap map[string]interface{}) interface{} {

	var result interface{}
	result = jsonMap[path[0]]

	if result == nil {
		return nil
	}

	if len(path) == 1 {
		return result
	} else {
		remainPath := path[1:len(path)]
		return c.configValue(remainPath, result.(map[string]interface{}))
	}
}

func (ca *ConfigAccessor) SetField(fieldName string, path string, target interface{}) {

	targetReflect := reflect.ValueOf(target).Elem()
	targetField := targetReflect.FieldByName(fieldName)

	switch targetField.Type().Kind() {
	case reflect.String:
		targetField.SetString(ca.StringVal(path))
	case reflect.Bool:
		targetField.SetBool(ca.BoolValue(path))
	case reflect.Int:
		targetField.SetInt(int64(ca.IntValue(path)))
	default:
		ca.FrameworkLogger.LogErrorf("Unable to use value at path %s as target field %s is not a suppported type", path, fieldName)
	}

}

func (ca *ConfigAccessor) Populate(path string, target interface{}) {
	exists := ca.PathExists(path)

	if exists {
		targetReflect := reflect.ValueOf(target).Elem()
		targetType := targetReflect.Type()
		numFields := targetType.NumField()

		for i := 0; i < numFields; i++ {

			fieldName := targetType.Field(i).Name

			expectedConfigPath := path + JsonPathSeparator + fieldName

			if ca.PathExists(expectedConfigPath) {
				ca.SetField(fieldName, expectedConfigPath, target)
			}

		}

	} else {
		ca.FrameworkLogger.LogErrorf("Trying to populate an object from a JSON object, but the base path %s does not exist", path)
	}

}
