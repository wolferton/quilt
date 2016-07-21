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

func (qp *WsQueryParams) IntValue(key string) (int, error) {

	v := qp.values[key]

	if v == nil {
		return 0, qp.noVal(key)
	}

	i, err := strconv.Atoi(v[len(v)-1])

	return i, err

}

func (qp *WsQueryParams) noVal(key string) error {
	message := fmt.Sprintf("No value available for key %s", key)
	return errors.New(message)
}

type QueryBinder struct {
	FrameworkLogger logging.Logger
}

func (qb *QueryBinder) AutoBind(wsReq *WsRequest) {

	t := wsReq.RequestBody
	qp := wsReq.QueryParams

	for _, field := range qp.ParamNames() {

		if rt.HasFieldOfName(t, field) {
			fErr := qb.bindValueToField(field, field, qp, t)

			if fErr != nil {
				wsReq.AddFrameworkError(fErr)
			}

		}

	}
}

func (qb *QueryBinder) bindValueToField(paramName string, fieldName string, qp *WsQueryParams, t interface{}) *WsFrameworkError {

	if !rt.TargetFieldIsArray(t, fieldName) && qp.MultipleValues(paramName) {
		message := fmt.Sprintf("Multiple values for query parameter %s, but the target field can only accept a single value.", fieldName)
		return NewQueryBindFrameworkError(message, paramName, fieldName)
	}

	switch rt.TypeOfField(t, fieldName).Kind() {
	case reflect.Int:

		i, err := qp.IntValue(paramName)

		if err != nil {
			return NewQueryBindFrameworkError(err.Error(), paramName, fieldName)
		}

		rt.SetInt(t, fieldName, i)

	}

	return nil

}
