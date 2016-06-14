package config
import (
    "strings"
)

const JsonPathSeparator string = "."

const (
    JsonUnknown = -1
    JsonInt = 0
    JsonString = 1
    JsonArray = 2
    JsonMap = 3
)


type ConfigValue interface{}

type ConfigAccessor struct{
    JsonData map[string]interface{}
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

func (c *ConfigAccessor) StringFieldVal(field string, object map[string]interface{} ) string {
    return object[field].(string)
}

func (c *ConfigAccessor) IntValue(path string) int {
    return int(c.Value(path).(float64))
}

func (c *ConfigAccessor) Float64Value(path string) float64 {
    return c.Value(path).(float64)
}

func (c *ConfigAccessor) Array(path string) []interface{} {
    return c.Value(path).([]interface{})
}

func  (c *ConfigAccessor) BoolValue(path string) bool {
	return c.Value(path).(bool)
}

func JsonType(value interface{}) int{

    switch value.(type) {
        case string:
            return JsonString
        case map[string]interface{}:
            return JsonMap
        default:
            return JsonUnknown
    }
}

func (c *ConfigAccessor) configValue(path []string, jsonMap map[string]interface{}) interface{} {

    var result interface{}
    result = jsonMap[path[0]]

    if len(path) == 1 {
        return result
    } else {
        remainPath := path[1 : len(path)]
        return c.configValue(remainPath, result.(map[string]interface{}))
    }
}




func check(e error) {
    if e != nil {
        panic(e)
    }
}


