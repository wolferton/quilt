package ws

import (
	"errors"
	"fmt"
	"github.com/wolferton/quilt/logging"
	rt "github.com/wolferton/quilt/reflecttools"
	"net/url"
	"reflect"
	"strconv"
)

func NewWsQueryParams(values url.Values) *WsQueryParams {

	qp := new(WsQueryParams)
	qp.values = values

	var names []string

	for k, _ := range values {
		names = append(names, k)
	}

	qp.paramNames = names

	return qp

}

type WsQueryParams struct {
	values     url.Values
	paramNames []string
}

func (qp *WsQueryParams) ParamNames() []string {
	return qp.paramNames
}

func (qp *WsQueryParams) Exists(key string) bool {
	return qp.values[key] != nil
}

func (qp *WsQueryParams) MultipleValues(key string) bool {

	value := qp.values[key]

	return value != nil && len(value) > 1

}

func (qp *WsQueryParams) StringValue(key string) (string, error) {

	s := qp.values[key]

	if s == nil {
		return "", qp.noVal(key)
	}

	return s[len(s)-1], nil

}

func (qp *WsQueryParams) BoolValue(key string) (bool, error) {

	v := qp.values[key]

	if v == nil {
		return false, qp.noVal(key)
	}

	b, err := strconv.ParseBool(v[len(v)-1])

	return b, err

}

func (qp *WsQueryParams) IntNValue(key string, bits int) (int64, error) {

	v := qp.values[key]

	if v == nil {
		return 0, qp.noVal(key)
	}

	i, err := strconv.ParseInt(v[len(v)-1], 10, bits)

	return i, err

}

func (qp *WsQueryParams) UIntNValue(key string, bits int) (uint64, error) {

	v := qp.values[key]

	if v == nil {
		return 0, qp.noVal(key)
	}

	i, err := strconv.ParseUint(v[len(v)-1], 10, bits)

	return i, err

}

func (qp *WsQueryParams) noVal(key string) error {
	message := fmt.Sprintf("No value available for key %s", key)
	return errors.New(message)
}

type QueryBinder struct {
	FrameworkLogger logging.Logger
}

func (qb *QueryBinder) AutoBindQueryParameters(wsReq *WsRequest) {

	t := wsReq.RequestBody
	p := wsReq.QueryParams

	for _, paramName := range p.ParamNames() {

		if rt.HasFieldOfName(t, paramName) {

			if !rt.TargetFieldIsArray(t, paramName) && p.MultipleValues(paramName) {
				message := fmt.Sprintf("Multiple values for query parameter %s. The target field can only accept a single value.", paramName)
				wsReq.AddFrameworkError(NewQueryBindFrameworkError(message, paramName, paramName))
				continue
			}

			fErr := qb.bindValueToField(paramName, paramName, p, t)

			if fErr != nil {
				wsReq.AddFrameworkError(fErr)
			}

		}

	}
}

func (qb *QueryBinder) queryParamError(paramName string, fieldName string, typeName string) *WsFrameworkError {

	message := fmt.Sprintf("Unable to convert the value of query parameter %s to %s", paramName, typeName)
	return NewQueryBindFrameworkError(message, paramName, fieldName)

}

func (qb *QueryBinder) bindValueToField(paramName string, fieldName string, p *WsQueryParams, t interface{}) *WsFrameworkError {

	switch rt.TypeOfField(t, fieldName).Kind() {
	case reflect.Int:
		return qb.setIntNField(paramName, fieldName, p, t, 0)
	case reflect.Int8:
		return qb.setIntNField(paramName, fieldName, p, t, 8)
	case reflect.Int16:
		return qb.setIntNField(paramName, fieldName, p, t, 16)
	case reflect.Int32:
		return qb.setIntNField(paramName, fieldName, p, t, 32)
	case reflect.Int64:
		return qb.setIntNField(paramName, fieldName, p, t, 64)
	case reflect.Bool:
		return qb.setBoolField(paramName, fieldName, p, t)
	case reflect.String:
		return qb.setStringField(paramName, fieldName, p, t)
	case reflect.Uint8:
		return qb.setUintNField(paramName, fieldName, p, t, 8)
	case reflect.Uint16:
		return qb.setUintNField(paramName, fieldName, p, t, 16)
	case reflect.Uint32:
		return qb.setUintNField(paramName, fieldName, p, t, 32)
	case reflect.Uint64:
		return qb.setUintNField(paramName, fieldName, p, t, 64)
	}

	return nil

}

func (qb *QueryBinder) setStringField(paramName string, fieldName string, qp *WsQueryParams, t interface{}) *WsFrameworkError {
	s, err := qp.StringValue(paramName)

	if err != nil {
		return NewQueryBindFrameworkError(err.Error(), paramName, fieldName)
	}

	rt.SetString(t, fieldName, s)

	return nil
}

func (qb *QueryBinder) setBoolField(paramName string, fieldName string, qp *WsQueryParams, t interface{}) *WsFrameworkError {
	b, err := qp.BoolValue(paramName)

	if err != nil {
		return NewQueryBindFrameworkError(err.Error(), paramName, fieldName)
	}

	rt.SetBool(t, fieldName, b)
	return nil
}

func (qb *QueryBinder) setIntNField(paramName string, fieldName string, qp *WsQueryParams, t interface{}, bits int) *WsFrameworkError {
	i, err := qp.IntNValue(paramName, bits)

	if err != nil {
		return NewQueryBindFrameworkError(err.Error(), paramName, fieldName)
	}

	rt.SetInt64(t, fieldName, i)
	return nil
}

func (qb *QueryBinder) setUintNField(paramName string, fieldName string, qp *WsQueryParams, t interface{}, bits int) *WsFrameworkError {
	i, err := qp.UIntNValue(paramName, bits)

	if err != nil {
		return NewQueryBindFrameworkError(err.Error(), paramName, fieldName)
	}

	rt.SetUint64(t, fieldName, i)
	return nil
}
