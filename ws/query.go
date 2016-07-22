package ws

import (
	"errors"
	"fmt"
	"net/url"
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

func (qp *WsQueryParams) FloatNValue(key string, bits int) (float64, error) {

	v := qp.values[key]

	if v == nil {
		return 0.0, qp.noVal(key)
	}

	i, err := strconv.ParseFloat(v[len(v)-1], bits)

	return i, err

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
